package client

import "strconv"

type Config struct {
	Address string
	Port    int
}

func NewConfig() *Config {
	return DefaultConfig()
}

func DefaultConfig() *Config {
	return &Config{
		Address: "localhost",
		Port:    8000,
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
