package server

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"udp/utils"

	"go.uber.org/zap"
)

var (
	dict          = NewDictionary()
	dictMutex     sync.RWMutex
	packetCounter uint32
)

// ClientSession represents a UDP client session
type ClientSession struct {
	RemoteAddr       *net.UDPAddr
	LastActivity     time.Time
	ReliabilityMgr   *utils.ReliabilityManager
	PacketBuffer     *utils.PacketBuffer
	TotalReceived    int
	TotalSent        int
	SessionStartTime time.Time
}

// SessionManager manages multiple client sessions
type SessionManager struct {
	sessions map[string]*ClientSession
	mutex    sync.RWMutex
	timeout  time.Duration
	logger   *zap.Logger
}

// NewSessionManager creates a new session manager
func NewSessionManager(timeout time.Duration, logger *zap.Logger) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*ClientSession),
		timeout:  timeout,
		logger:   logger,
	}
}

// GetOrCreateSession gets or creates a session for a client
func (sm *SessionManager) GetOrCreateSession(remoteAddr *net.UDPAddr) *ClientSession {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	key := remoteAddr.String()

	if session, exists := sm.sessions[key]; exists {
		session.LastActivity = time.Now()
		return session
	}

	// Create new session
	session := &ClientSession{
		RemoteAddr:       remoteAddr,
		LastActivity:     time.Now(),
		ReliabilityMgr:   utils.NewReliabilityManager(2*time.Second, 3),
		PacketBuffer:     utils.NewPacketBuffer(5 * time.Second),
		SessionStartTime: time.Now(),
	}

	sm.sessions[key] = session
	sm.logger.Info("New client session created", zap.String("remote", key))

	return session
}

// CleanupExpiredSessions removes sessions that have timed out
func (sm *SessionManager) CleanupExpiredSessions() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	var expiredKeys []string

	for key, session := range sm.sessions {
		if now.Sub(session.LastActivity) > sm.timeout {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(sm.sessions, key)
		sm.logger.Info("Client session expired", zap.String("remote", key))
	}
}

// GetSessionStats returns statistics for all sessions
func (sm *SessionManager) GetSessionStats() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_sessions"] = len(sm.sessions)
	stats["active_sessions"] = 0
	stats["total_packets_received"] = 0
	stats["total_packets_sent"] = 0

	now := time.Now()
	activeSessions := 0
	totalReceived := 0
	totalSent := 0

	for _, session := range sm.sessions {
		if now.Sub(session.LastActivity) <= sm.timeout {
			activeSessions++
		}
		totalReceived += session.TotalReceived
		totalSent += session.TotalSent
	}

	stats["active_sessions"] = activeSessions
	stats["total_packets_received"] = totalReceived
	stats["total_packets_sent"] = totalSent

	return stats
}

// StartServer starts the UDP server
func StartServer(config *Config) error {
	logger := utils.GetLogger()
	defer logger.Sync()

	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", config.AddressString())
	if err != nil {
		logger.Error("Failed to resolve UDP address", zap.Error(err))
		return err
	}

	// Listen on UDP
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logger.Error("Failed to listen on UDP", zap.Error(err))
		return err
	}
	defer conn.Close()

	logger.Info("UDP server started",
		zap.String("address", config.AddressString()),
		zap.Int("max_packet_size", config.MaxPacketSize),
		zap.Duration("ack_timeout", config.AckTimeout),
		zap.Int("max_retries", config.MaxRetries),
	)

	// Initialize session manager
	sessionMgr := NewSessionManager(30*time.Second, logger)

	// Cleanup goroutine
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			sessionMgr.CleanupExpiredSessions()
		}
	}()

	// Retransmission goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			handleRetransmissions(conn, sessionMgr, logger)
		}
	}()

	// Stats goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			stats := sessionMgr.GetSessionStats()
			logger.Info("Server statistics",
				zap.Int("active_sessions", stats["active_sessions"].(int)),
				zap.Int("total_packets_received", stats["total_packets_received"].(int)),
				zap.Int("total_packets_sent", stats["total_packets_sent"].(int)),
			)
		}
	}()

	// Main receive loop
	buffer := make([]byte, 2048)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			logger.Warn("Error reading from UDP", zap.Error(err))
			continue
		}

		// Get or create session
		session := sessionMgr.GetOrCreateSession(remoteAddr)
		session.TotalReceived++

		// Handle packet in goroutine
		go handleClientPacket(conn, remoteAddr, buffer[:n], session, logger, config)
	}
}

// handleClientPacket processes a single packet from a client
func handleClientPacket(conn *net.UDPConn, remoteAddr *net.UDPAddr, data []byte, session *ClientSession, logger *zap.Logger, config *Config) {
	// Deserialize packet
	packet, err := utils.FromBytes(data)
	if err != nil {
		logger.Warn("Failed to deserialize packet",
			zap.String("remote", remoteAddr.String()),
			zap.Error(err),
		)
		return
	}

	// Log packet receipt
	logger.Debug("Packet received",
		zap.String("remote", remoteAddr.String()),
		zap.Uint32("packet_id", packet.ID),
		zap.String("type", packet.MessageType.String()),
		zap.Uint16("data_size", packet.DataSize),
	)

	// Handle different packet types
	switch packet.MessageType {
	case utils.PacketTypeRequest:
		handleRequest(conn, remoteAddr, packet, session, logger, config)

	case utils.PacketTypeACK:
		logger.Debug("ACK received",
			zap.String("remote", remoteAddr.String()),
			zap.Uint32("packet_id", packet.ID),
		)
		session.ReliabilityMgr.MarkACK(packet.ID)

	case utils.PacketTypeHeartbeat:
		logger.Debug("Heartbeat received",
			zap.String("remote", remoteAddr.String()),
			zap.Uint32("packet_id", packet.ID),
		)
		// Echo back heartbeat
		ackPacket := utils.NewPacket(packet.ID, utils.PacketTypeACK, nil)
		sendPacket(conn, remoteAddr, ackPacket, session, logger)

	default:
		logger.Warn("Unknown packet type",
			zap.String("remote", remoteAddr.String()),
			zap.Uint8("type", uint8(packet.MessageType)),
		)
	}
}

// handleRequest processes a REQUEST packet
func handleRequest(conn *net.UDPConn, remoteAddr *net.UDPAddr, reqPacket *utils.Packet, session *ClientSession, logger *zap.Logger, config *Config) {
	// Send ACK immediately
	ackPacket := utils.NewPacket(reqPacket.ID, utils.PacketTypeACK, nil)
	sendPacket(conn, remoteAddr, ackPacket, session, logger)

	// Add packet to buffer (in case of fragmentation)
	payload, err := session.PacketBuffer.Add(reqPacket)
	if err != nil {
		logger.Warn("Failed to add packet to buffer", zap.Error(err))
		return
	}

	// If payload is nil, message is incomplete
	if payload == nil {
		logger.Debug("Fragmented message incomplete",
			zap.String("remote", remoteAddr.String()),
			zap.Uint32("packet_id", reqPacket.ID),
		)
		return
	}

	// Process complete command
	command := string(payload)
	logger.Debug("Processing command",
		zap.String("remote", remoteAddr.String()),
		zap.String("command", command),
	)

	response := ProcessDictCommand(command, dict, &dictMutex)
	if response == nil {
		response = &Response{
			StatusCode: 500,
			Message:    "Internal server error",
		}
	}

	// Prepare response packet
	responsePayload := []byte(fmt.Sprintf("%d %s", response.StatusCode, response.Message))

	// Fragment response if needed
	packets := utils.FragmentPayload(reqPacket.ID, responsePayload, config.MaxPacketSize)

	// Send response packets
	for _, respPacket := range packets {
		respPacket.MessageType = utils.PacketTypeResponse
		sendPacket(conn, remoteAddr, respPacket, session, logger)
	}

	logger.Debug("Response sent",
		zap.String("remote", remoteAddr.String()),
		zap.Uint32("packet_id", reqPacket.ID),
		zap.Int("packets", len(packets)),
	)
}

// sendPacket sends a packet to a client and tracks it
func sendPacket(conn *net.UDPConn, remoteAddr *net.UDPAddr, packet *utils.Packet, session *ClientSession, logger *zap.Logger) error {
	// Serialize packet
	data := packet.Bytes()

	// Send packet
	_, err := conn.WriteToUDP(data, remoteAddr)
	if err != nil {
		logger.Warn("Failed to send packet",
			zap.String("remote", remoteAddr.String()),
			zap.Error(err),
		)
		return err
	}

	// Track sent packet
	session.ReliabilityMgr.TrackSent(packet)
	session.TotalSent++

	logger.Debug("Packet sent",
		zap.String("remote", remoteAddr.String()),
		zap.Uint32("packet_id", packet.ID),
		zap.String("type", packet.MessageType.String()),
	)

	return nil
}

// handleRetransmissions checks for packets that need retransmission
func handleRetransmissions(conn *net.UDPConn, sessionMgr *SessionManager, logger *zap.Logger) {
	sessionMgr.mutex.RLock()
	defer sessionMgr.mutex.RUnlock()

	for _, session := range sessionMgr.sessions {
		candidates := session.ReliabilityMgr.GetRetransmitCandidates()
		for _, packet := range candidates {
			_, err := conn.WriteToUDP(packet.Bytes(), session.RemoteAddr)
			if err != nil {
				logger.Warn("Failed to retransmit packet",
					zap.String("remote", session.RemoteAddr.String()),
					zap.Uint32("packet_id", packet.ID),
					zap.Error(err),
				)
				continue
			}

			session.ReliabilityMgr.TrackSent(packet)
			logger.Debug("Packet retransmitted",
				zap.String("remote", session.RemoteAddr.String()),
				zap.Uint32("packet_id", packet.ID),
			)
		}
	}
}

// GetNextPacketID generates the next packet ID
func GetNextPacketID() uint32 {
	return atomic.AddUint32(&packetCounter, 1)
}
