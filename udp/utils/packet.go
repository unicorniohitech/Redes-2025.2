package utils

import (
	"fmt"
	"sync"
	"time"
)

// FragmentedMessage represents a fragmented message being reassembled
type FragmentedMessage struct {
	ID           uint32
	Packets      map[uint16]*Packet
	TotalPackets uint16
	ReceivedAt   time.Time
	LastUpdate   time.Time
}

// IsComplete returns true if all packets have been received
func (fm *FragmentedMessage) IsComplete() bool {
	if uint16(len(fm.Packets)) != fm.TotalPackets {
		return false
	}
	// Verify all packet numbers are present
	for i := uint16(1); i <= fm.TotalPackets; i++ {
		if _, exists := fm.Packets[i]; !exists {
			return false
		}
	}
	return true
}

// GetMissingPackets returns a list of missing packet numbers
func (fm *FragmentedMessage) GetMissingPackets() []uint16 {
	var missing []uint16
	for i := uint16(1); i <= fm.TotalPackets; i++ {
		if _, exists := fm.Packets[i]; !exists {
			missing = append(missing, i)
		}
	}
	return missing
}

// Assemble combines all packets into a single payload
func (fm *FragmentedMessage) Assemble() []byte {
	if !fm.IsComplete() {
		return nil
	}

	var result []byte
	for i := uint16(1); i <= fm.TotalPackets; i++ {
		if packet, exists := fm.Packets[i]; exists {
			result = append(result, packet.Payload...)
		}
	}
	return result
}

// PacketBuffer manages fragmented messages and reassembly
type PacketBuffer struct {
	messages  map[uint32]*FragmentedMessage
	mutex     sync.RWMutex
	timeout   time.Duration
	maxSize   int
	lastClean time.Time
}

// NewPacketBuffer creates a new packet buffer with specified timeout
func NewPacketBuffer(timeout time.Duration) *PacketBuffer {
	return &PacketBuffer{
		messages:  make(map[uint32]*FragmentedMessage),
		timeout:   timeout,
		maxSize:   10 * 1024 * 1024, // 10MB max buffer
		lastClean: time.Now(),
	}
}

// Add adds a packet to the buffer
// Returns error if packet is invalid or corrupted
// Returns nil if packet was added successfully (or message is not complete yet)
// Returns the assembled payload if message is now complete
func (pb *PacketBuffer) Add(packet *Packet) ([]byte, error) {
	if packet == nil {
		return nil, fmt.Errorf("packet is nil")
	}

	pb.mutex.Lock()
	defer pb.mutex.Unlock()

	// Cleanup expired messages periodically (every 100 operations)
	now := time.Now()
	if now.Sub(pb.lastClean) > pb.timeout {
		pb.cleanupExpired()
		pb.lastClean = now
	}

	// Handle single-packet messages
	if packet.TotalPackets == 1 {
		return packet.Payload, nil
	}

	// Handle fragmented messages
	msg, exists := pb.messages[packet.ID]
	if !exists {
		// Create new fragmented message
		msg = &FragmentedMessage{
			ID:           packet.ID,
			Packets:      make(map[uint16]*Packet),
			TotalPackets: packet.TotalPackets,
			ReceivedAt:   now,
			LastUpdate:   now,
		}
		pb.messages[packet.ID] = msg
	}

	// Add packet to message
	msg.Packets[packet.PacketNumber] = packet
	msg.LastUpdate = now

	// Check if message is complete
	if msg.IsComplete() {
		assembled := msg.Assemble()
		delete(pb.messages, packet.ID)
		return assembled, nil
	}

	// Message not complete yet
	return nil, nil
}

// IsComplete checks if all packets for a message have been received
func (pb *PacketBuffer) IsComplete(id uint32) bool {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()

	msg, exists := pb.messages[id]
	if !exists {
		return false
	}
	return msg.IsComplete()
}

// Retrieve gets the assembled payload for a complete message
// Returns error if message is not complete or does not exist
func (pb *PacketBuffer) Retrieve(id uint32) ([]byte, error) {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()

	msg, exists := pb.messages[id]
	if !exists {
		return nil, fmt.Errorf("message ID %d not found in buffer", id)
	}

	if !msg.IsComplete() {
		return nil, fmt.Errorf("message ID %d is not complete", id)
	}

	assembled := msg.Assemble()
	delete(pb.messages, id)
	return assembled, nil
}

// GetLostPackets returns a list of lost packet numbers for a message
func (pb *PacketBuffer) GetLostPackets(id uint32) []uint16 {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()

	msg, exists := pb.messages[id]
	if !exists {
		return nil
	}

	return msg.GetMissingPackets()
}

// GetProgress returns the progress of a fragmented message
// Returns (received, total, percent)
func (pb *PacketBuffer) GetProgress(id uint32) (int, int, float64) {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()

	msg, exists := pb.messages[id]
	if !exists {
		return 0, 0, 0
	}

	received := len(msg.Packets)
	total := int(msg.TotalPackets)
	percent := (float64(received) / float64(total)) * 100.0

	return received, total, percent
}

// Cleanup removes all expired messages from the buffer
func (pb *PacketBuffer) Cleanup() {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()

	pb.cleanupExpired()
}

// cleanupExpired removes expired messages (must be called with lock held)
func (pb *PacketBuffer) cleanupExpired() {
	now := time.Now()
	var expiredIDs []uint32

	for id, msg := range pb.messages {
		if now.Sub(msg.LastUpdate) > pb.timeout {
			expiredIDs = append(expiredIDs, id)
		}
	}

	for _, id := range expiredIDs {
		delete(pb.messages, id)
	}
}

// GetPendingMessages returns the number of incomplete messages in the buffer
func (pb *PacketBuffer) GetPendingMessages() int {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()

	return len(pb.messages)
}

// GetBufferStats returns statistics about the current buffer state
type BufferStats struct {
	PendingMessages int
	TotalPackets    int
	AverageProgress float64
}

// GetStats returns current buffer statistics
func (pb *PacketBuffer) GetStats() BufferStats {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()

	stats := BufferStats{
		PendingMessages: len(pb.messages),
	}

	if len(pb.messages) == 0 {
		return stats
	}

	totalPackets := 0
	totalProgress := 0.0

	for _, msg := range pb.messages {
		totalPackets += len(msg.Packets)
		totalProgress += float64(len(msg.Packets)) / float64(msg.TotalPackets)
	}

	stats.TotalPackets = totalPackets
	stats.AverageProgress = totalProgress / float64(len(pb.messages))

	return stats
}

// Fragment divides a payload into multiple packets if it exceeds maxSize
// Returns a slice of packets with proper fragmentation headers
func FragmentPayload(id uint32, payload []byte, maxPayloadSize int) []*Packet {
	if maxPayloadSize <= 0 {
		maxPayloadSize = 1024
	}

	// If payload fits in one packet
	if len(payload) <= maxPayloadSize {
		return []*Packet{NewPacket(id, PacketTypeRequest, payload)}
	}

	// Split payload into fragments
	var packets []*Packet
	totalPackets := (len(payload) + maxPayloadSize - 1) / maxPayloadSize

	for i := 0; i < len(payload); i += maxPayloadSize {
		end := i + maxPayloadSize
		if end > len(payload) {
			end = len(payload)
		}

		fragment := payload[i:end]
		packet := NewPacket(id, PacketTypeRequest, fragment)
		packet.TotalPackets = uint16(totalPackets)
		packet.PacketNumber = uint16(len(packets) + 1)

		packets = append(packets, packet)
	}

	return packets
}
