package utils

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

// PacketType represents the type of UDP message
type PacketType uint8

const (
	PacketTypeRequest   PacketType = 0
	PacketTypeResponse  PacketType = 1
	PacketTypeACK       PacketType = 2
	PacketTypeHeartbeat PacketType = 3
)

// String returns a string representation of PacketType
func (pt PacketType) String() string {
	switch pt {
	case PacketTypeRequest:
		return "REQUEST"
	case PacketTypeResponse:
		return "RESPONSE"
	case PacketTypeACK:
		return "ACK"
	case PacketTypeHeartbeat:
		return "HEARTBEAT"
	default:
		return "UNKNOWN"
	}
}

// Packet represents a UDP packet with custom header
// Header structure (13 bytes):
// Bytes 0-3:    Packet ID (uint32)
// Byte 4:       Message Type (uint8)
// Bytes 5-6:    Data Size (uint16)
// Bytes 7-8:    Total Packets (uint16)
// Bytes 9-10:   Packet Number (uint16)
// Bytes 11-12:  Checksum (uint16)
// Bytes 13+:    Payload (variável)
type Packet struct {
	ID           uint32
	MessageType  PacketType
	DataSize     uint16
	TotalPackets uint16
	PacketNumber uint16
	Checksum     uint16
	Payload      []byte
}

// NewPacket creates a new packet with the given parameters
func NewPacket(id uint32, msgType PacketType, payload []byte) *Packet {
	return &Packet{
		ID:           id,
		MessageType:  msgType,
		DataSize:     uint16(len(payload)),
		TotalPackets: 1,
		PacketNumber: 1,
		Payload:      payload,
	}
}

// Bytes serializes the packet to a byte slice
func (p *Packet) Bytes() []byte {
	// Header size: 13 bytes + payload
	headerSize := 13
	totalSize := headerSize + len(p.Payload)
	buf := make([]byte, totalSize)

	// Write header
	binary.BigEndian.PutUint32(buf[0:4], p.ID)            // ID
	buf[4] = byte(p.MessageType)                          // MessageType
	binary.BigEndian.PutUint16(buf[5:7], p.DataSize)      // DataSize
	binary.BigEndian.PutUint16(buf[7:9], p.TotalPackets)  // TotalPackets
	binary.BigEndian.PutUint16(buf[9:11], p.PacketNumber) // PacketNumber
	binary.BigEndian.PutUint16(buf[11:13], p.Checksum)    // Checksum (placeholder)

	// Copy payload
	copy(buf[13:], p.Payload)

	// Calculate and update checksum
	checksum := calculateChecksum(buf)
	binary.BigEndian.PutUint16(buf[11:13], checksum)

	return buf
}

// FromBytes deserializes a byte slice into a Packet
func FromBytes(data []byte) (*Packet, error) {
	if len(data) < 13 {
		return nil, fmt.Errorf("packet too small: expected at least 13 bytes, got %d", len(data))
	}

	packet := &Packet{
		ID:           binary.BigEndian.Uint32(data[0:4]),
		MessageType:  PacketType(data[4]),
		DataSize:     binary.BigEndian.Uint16(data[5:7]),
		TotalPackets: binary.BigEndian.Uint16(data[7:9]),
		PacketNumber: binary.BigEndian.Uint16(data[9:11]),
		Checksum:     binary.BigEndian.Uint16(data[11:13]),
	}

	// Validate checksum before extracting payload
	if !packet.ValidateChecksum(data) {
		return nil, fmt.Errorf("checksum validation failed for packet ID %d", packet.ID)
	}

	// Extract payload
	if len(data) > 13 {
		packet.Payload = make([]byte, len(data)-13)
		copy(packet.Payload, data[13:])
	}

	return packet, nil
}

// ValidateChecksum checks if the packet checksum is valid
func (p *Packet) ValidateChecksum(data []byte) bool {
	if len(data) < 13 {
		return false
	}

	// Create a copy of the data with checksum set to 0 for validation
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	binary.BigEndian.PutUint16(dataCopy[11:13], 0)

	calculatedChecksum := calculateChecksum(dataCopy)
	return calculatedChecksum == p.Checksum
}

// calculateChecksum computes CRC32 checksum for a packet
func calculateChecksum(data []byte) uint16 {
	// Use CRC32 and take lower 16 bits for checksum
	crc := crc32.ChecksumIEEE(data)
	return uint16(crc & 0xFFFF)
}

// IsComplete returns true if all packets in a fragmented message are complete
func (p *Packet) IsComplete() bool {
	return p.PacketNumber == p.TotalPackets
}

// IsFragmented returns true if the packet is part of a fragmented message
func (p *Packet) IsFragmented() bool {
	return p.TotalPackets > 1
}

// String returns a string representation of the packet
func (p *Packet) String() string {
	return fmt.Sprintf(
		"Packet{ID:%d, Type:%s, Size:%d, Fragmented:%d/%d, Checksum:%d}",
		p.ID,
		p.MessageType.String(),
		p.DataSize,
		p.PacketNumber,
		p.TotalPackets,
		p.Checksum,
	)
}
