package utils

import (
	"fmt"
	"sync"
	"time"
)

// SentPacket tracks information about a sent packet
type SentPacket struct {
	Packet       *Packet
	SentTime     time.Time
	LastSentTime time.Time
	RetryCount   int
	Acked        bool
	AckTime      time.Time
}

// GetLatency returns the latency in milliseconds for this packet
func (sp *SentPacket) GetLatency() time.Duration {
	if sp.Acked {
		return sp.AckTime.Sub(sp.SentTime)
	}
	return 0
}

// PacketMetrics represents metrics about packet transmission
type PacketMetrics struct {
	TotalSent       int
	TotalReceived   int
	TotalLost       int
	TotalRetransmit int
	TotalACKed      int
	AverageLatency  time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	LossRate        float64
	RetransmitRate  float64
}

// String returns a formatted string representation of metrics
func (pm *PacketMetrics) String() string {
	return fmt.Sprintf(
		"Metrics{Sent:%d, Received:%d, Acked:%d, Lost:%d, Retrans:%d, Loss:%.2f%%, Avg:%dms}",
		pm.TotalSent,
		pm.TotalReceived,
		pm.TotalACKed,
		pm.TotalLost,
		pm.TotalRetransmit,
		pm.LossRate*100,
		pm.AverageLatency.Milliseconds(),
	)
}

// ReliabilityManager manages reliable delivery of UDP packets
type ReliabilityManager struct {
	sentPackets     map[uint32]*SentPacket
	ackedPackets    map[uint32]bool
	lostPackets     map[uint32]bool
	mutex           sync.RWMutex
	ackTimeout      time.Duration
	maxRetries      int
	simulateLoss    bool
	lossRate        float64
	startTime       time.Time
	totalRetransmit int
	latencies       []time.Duration
}

// NewReliabilityManager creates a new reliability manager
func NewReliabilityManager(ackTimeout time.Duration, maxRetries int) *ReliabilityManager {
	return &ReliabilityManager{
		sentPackets:  make(map[uint32]*SentPacket),
		ackedPackets: make(map[uint32]bool),
		lostPackets:  make(map[uint32]bool),
		ackTimeout:   ackTimeout,
		maxRetries:   maxRetries,
		simulateLoss: false,
		lossRate:     0.0,
		startTime:    time.Now(),
		latencies:    make([]time.Duration, 0),
	}
}

// TrackSent records that a packet has been sent
func (rm *ReliabilityManager) TrackSent(packet *Packet) error {
	if packet == nil {
		return fmt.Errorf("packet is nil")
	}

	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	now := time.Now()

	// Check if packet already exists
	if sp, exists := rm.sentPackets[packet.ID]; exists {
		// This is a retransmission
		sp.RetryCount++
		sp.LastSentTime = now
		rm.totalRetransmit++
	} else {
		// New packet
		rm.sentPackets[packet.ID] = &SentPacket{
			Packet:       packet,
			SentTime:     now,
			LastSentTime: now,
			RetryCount:   0,
			Acked:        false,
		}
	}

	return nil
}

// MarkACK marks a packet as acknowledged
func (rm *ReliabilityManager) MarkACK(packetID uint32) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	sp, exists := rm.sentPackets[packetID]
	if !exists {
		return fmt.Errorf("packet ID %d not found in tracking", packetID)
	}

	now := time.Now()
	sp.Acked = true
	sp.AckTime = now

	// Record latency
	latency := now.Sub(sp.SentTime)
	rm.latencies = append(rm.latencies, latency)

	rm.ackedPackets[packetID] = true

	return nil
}

// MarkLost marks a packet as lost
func (rm *ReliabilityManager) MarkLost(packetID uint32) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	_, exists := rm.sentPackets[packetID]
	if !exists {
		return fmt.Errorf("packet ID %d not found in tracking", packetID)
	}

	rm.lostPackets[packetID] = true

	return nil
}

// GetRetransmitCandidates returns packets that should be retransmitted
// Returns packets that have not been ACKed and whose timeout has expired
func (rm *ReliabilityManager) GetRetransmitCandidates() []*Packet {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	var candidates []*Packet
	now := time.Now()

	for _, sp := range rm.sentPackets {
		if sp.Acked {
			continue
		}

		// Check if timeout has expired
		timeSinceLastSent := now.Sub(sp.LastSentTime)
		expectedTimeout := rm.ackTimeout * time.Duration(sp.RetryCount+1)

		if timeSinceLastSent > expectedTimeout && sp.RetryCount < rm.maxRetries {
			candidates = append(candidates, sp.Packet)
		}
	}

	return candidates
}

// IsPacketAcked checks if a packet has been acknowledged
func (rm *ReliabilityManager) IsPacketAcked(packetID uint32) bool {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	return rm.ackedPackets[packetID]
}

// IsPacketLost checks if a packet has been marked as lost
func (rm *ReliabilityManager) IsPacketLost(packetID uint32) bool {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	return rm.lostPackets[packetID]
}

// CanRetransmit checks if a packet can still be retransmitted
func (rm *ReliabilityManager) CanRetransmit(packetID uint32) bool {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	sp, exists := rm.sentPackets[packetID]
	if !exists {
		return false
	}

	return !sp.Acked && sp.RetryCount < rm.maxRetries
}

// SetSimulateLoss configures packet loss simulation for testing
func (rm *ReliabilityManager) SetSimulateLoss(simulate bool, lossRate float64) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.simulateLoss = simulate
	if lossRate < 0 {
		rm.lossRate = 0
	} else if lossRate > 1 {
		rm.lossRate = 1
	} else {
		rm.lossRate = lossRate
	}
}

// GetMetrics returns current metrics
func (rm *ReliabilityManager) GetMetrics() *PacketMetrics {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	metrics := &PacketMetrics{
		TotalSent:       len(rm.sentPackets),
		TotalACKed:      len(rm.ackedPackets),
		TotalLost:       len(rm.lostPackets),
		TotalRetransmit: rm.totalRetransmit,
	}

	// Calculate received (acked + lost)
	metrics.TotalReceived = metrics.TotalACKed + metrics.TotalLost

	// Calculate loss rate
	if metrics.TotalSent > 0 {
		metrics.LossRate = float64(metrics.TotalLost) / float64(metrics.TotalSent)
	}

	// Calculate retransmit rate
	if metrics.TotalSent > 0 {
		metrics.RetransmitRate = float64(metrics.TotalRetransmit) / float64(metrics.TotalSent)
	}

	// Calculate latencies
	if len(rm.latencies) > 0 {
		var sum time.Duration
		min := rm.latencies[0]
		max := rm.latencies[0]

		for _, lat := range rm.latencies {
			sum += lat
			if lat < min {
				min = lat
			}
			if lat > max {
				max = lat
			}
		}

		metrics.AverageLatency = sum / time.Duration(len(rm.latencies))
		metrics.MinLatency = min
		metrics.MaxLatency = max
	}

	return metrics
}

// GetUptime returns the time since manager was created
func (rm *ReliabilityManager) GetUptime() time.Duration {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	return time.Since(rm.startTime)
}

// Clear resets all tracking data (use with caution)
func (rm *ReliabilityManager) Clear() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.sentPackets = make(map[uint32]*SentPacket)
	rm.ackedPackets = make(map[uint32]bool)
	rm.lostPackets = make(map[uint32]bool)
	rm.totalRetransmit = 0
	rm.latencies = make([]time.Duration, 0)
	rm.startTime = time.Now()
}

// GetAckWaitTime returns how long to wait for an ACK considering retry count
func (rm *ReliabilityManager) GetAckWaitTime(packetID uint32) time.Duration {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	sp, exists := rm.sentPackets[packetID]
	if !exists {
		return rm.ackTimeout
	}

	// Exponential backoff: timeout * (retryCount + 1)
	return rm.ackTimeout * time.Duration(sp.RetryCount+1)
}

// GetPendingAcks returns the number of packets waiting for ACK
func (rm *ReliabilityManager) GetPendingAcks() int {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	pending := 0
	for _, sp := range rm.sentPackets {
		if !sp.Acked {
			pending++
		}
	}
	return pending
}

// CleanupOldEntries removes old tracking entries (older than 10x timeout)
func (rm *ReliabilityManager) CleanupOldEntries() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	now := time.Now()
	cutoff := 10 * rm.ackTimeout

	var toDelete []uint32
	for id, sp := range rm.sentPackets {
		if now.Sub(sp.LastSentTime) > cutoff && sp.Acked {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		delete(rm.sentPackets, id)
	}
}

// IsHealthy returns true if system is operating normally
// Considers loss rate and retry rate as health indicators
func (rm *ReliabilityManager) IsHealthy() bool {
	metrics := rm.GetMetrics()

	// Unhealthy if loss rate > 50% or retry rate > 100%
	return metrics.LossRate < 0.5 && metrics.RetransmitRate < 1.0
}

// SessionStats contains comprehensive session information
type SessionStats struct {
	Uptime      time.Duration
	Metrics     *PacketMetrics
	IsHealthy   bool
	PendingAcks int
	AckWaitTime time.Duration
}

// GetSessionStats returns comprehensive session statistics
func (rm *ReliabilityManager) GetSessionStats() *SessionStats {
	metrics := rm.GetMetrics()
	pendingAcks := rm.GetPendingAcks()

	return &SessionStats{
		Uptime:      rm.GetUptime(),
		Metrics:     metrics,
		IsHealthy:   rm.IsHealthy(),
		PendingAcks: pendingAcks,
		AckWaitTime: rm.ackTimeout,
	}
}
