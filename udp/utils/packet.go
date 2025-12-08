package utils

import (
	"fmt"

	"go.uber.org/zap"
)

type Packet struct {
	Control uint16
	Length  uint16
	Payload []byte
	CRC     uint16
}

func (p Packet) Bytes() []byte {
	data := make([]byte, 2+2+len(p.Payload)+2)
	// Control
	data[0] = byte(p.Control >> 8)
	data[1] = byte(p.Control & 0xFF)
	// Length
	data[2] = byte(p.Length >> 8)
	data[3] = byte(p.Length & 0xFF)
	// Payload
	copy(data[4:], p.Payload)
	// CRC
	data[len(data)-2] = byte(p.CRC >> 8)
	data[len(data)-1] = byte(p.CRC & 0xFF)
	return data
}

func ParsePacket(data []byte) (*Packet, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid packet format")
	}
	control := uint16(data[0])<<8 | uint16(data[1])
	length := uint16(data[2])<<8 | uint16(data[3])
	crcStart := len(data) - 2
	payload := data[4:crcStart]
	crc := uint16(data[crcStart])<<8 | uint16(data[crcStart+1])
	return &Packet{
		Control: control,
		Length:  length,
		Payload: payload,
		CRC:     crc,
	}, nil
}

func CalculateCRC(packet *Packet) uint16 {
	var crc uint16 = 0xFFFF
	crcCalculator := NewCRC()
	packet.CRC = 0
	crc = crcCalculator.Compute(crc, packet.Bytes())
	return uint16(crc)
}

func NewPacket(payload []byte) []*Packet {
	logger := GetLogger()
	logger.Info("Creating packets", zap.Int("payload_length", len(payload)))
	partsQt := len(payload) / 1024
	packets := make([]*Packet, partsQt+1)
	for i := 0; i <= partsQt; i++ {
		start := i * 1024
		end := start + 1024
		if end > len(payload) {
			end = len(payload)
		}
		partPayload := payload[start:end]
		p := &Packet{
			Control: uint16(i),
			Length:  uint16(partsQt),
			Payload: partPayload,
		}
		p.CRC = CalculateCRC(p)
		packets[i] = p
	}
	logger.Info("Packets created", zap.Int("packets_count", len(packets)))
	return packets
}
