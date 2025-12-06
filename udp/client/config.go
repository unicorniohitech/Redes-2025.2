package client

import (
	"strconv"
	"time"
)

type Config struct {
	Address       string
	Port          int
	MaxPacketSize int
	AckTimeout    time.Duration
	MaxRetries    int
	SimulateLoss  bool
	LossRate      float64
}

func NewConfig() *Config {
	return DefaultConfig()
}

func DefaultConfig() *Config {
	return &Config{
		Address:       "localhost",
		Port:          8000,
		MaxPacketSize: 1024,
		AckTimeout:    2 * time.Second,
		MaxRetries:    3,
		SimulateLoss:  false,
		LossRate:      0.0,
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

func (c *Config) SetMaxPacketSize(size int) {
	c.MaxPacketSize = size
}

func (c *Config) SetAckTimeout(timeout time.Duration) {
	c.AckTimeout = timeout
}

func (c *Config) SetMaxRetries(retries int) {
	c.MaxRetries = retries
}

func (c *Config) SetSimulateLoss(simulate bool) {
	c.SimulateLoss = simulate
}

func (c *Config) SetLossRate(rate float64) {
	c.LossRate = rate
}
