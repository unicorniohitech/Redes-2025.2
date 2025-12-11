package client

import (
	"strconv"
	"sync"
	"udp/utils"

	"go.uber.org/zap"
)

type Config struct {
	Address        string
	Port           int
	partialPackets map[string][]utils.Packet
	mux            sync.Mutex
}

func NewConfig() *Config {
	return DefaultConfig()
}

func DefaultConfig() *Config {
	return &Config{
		Address:        "localhost",
		Port:           8080,
		partialPackets: make(map[string][]utils.Packet),
	}
}

func (c *Config) SetAddress(address string) {
	c.Address = address
}

func (c *Config) SetPort(port int) {
	c.Port = port
}

func (c *Config) AddressString() string {
	return c.Address + ":" + strconv.Itoa(c.Port)
}

func (c *Config) AddPartialPacket(key string, packet utils.Packet) {
	c.mux.Lock()
	defer func() {
		logger := utils.GetLogger()
		logger.Info(
			"Added partial packet",
			zap.String("key", key),
			zap.Uint16("current_count", packet.Control),
			zap.Uint16("expected_length", packet.Length),
		)
		c.mux.Unlock()
	}()
	if _, exists := c.partialPackets[key]; !exists {
		c.partialPackets[key] = make([]utils.Packet, 0)
	}
	c.partialPackets[key] = append(c.partialPackets[key], packet)
}

func (c *Config) IsPacketComplete(key string) bool {
	c.mux.Lock()
	defer c.mux.Unlock()
	if packets, exists := c.partialPackets[key]; exists {
		return uint16(len(packets)) == packets[0].Length+1
	}
	return false
}
