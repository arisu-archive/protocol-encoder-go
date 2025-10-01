package emulator

import (
	"errors"
	"fmt"
)

// Config holds the configuration for the ARM emulator
type Config struct {
	BaseAddr     uint64 `mapstructure:"base_addr" json:"base_addr"`
	StackAddr    uint64 `mapstructure:"stack_addr" json:"stack_addr"`
	StackSize    uint64 `mapstructure:"stack_size" json:"stack_size"`
	ReturnAddr   uint64 `mapstructure:"return_addr" json:"return_addr"`
	Architecture string `mapstructure:"architecture" json:"architecture"`
	Mode         string `mapstructure:"mode" json:"mode"`
}

// ConfigOption represents a configuration option
type ConfigOption func(*Config)

// WithBaseAddr sets the base address for binary mapping
func WithBaseAddr(addr uint64) ConfigOption {
	return func(c *Config) {
		c.BaseAddr = addr
	}
}

// WithStackAddr sets the stack base address
func WithStackAddr(addr uint64) ConfigOption {
	return func(c *Config) {
		c.StackAddr = addr
	}
}

// WithStackSize sets the stack size
func WithStackSize(size uint64) ConfigOption {
	return func(c *Config) {
		c.StackSize = size
	}
}

// WithReturnAddr sets the return address sentinel
func WithReturnAddr(addr uint64) ConfigOption {
	return func(c *Config) {
		c.ReturnAddr = addr
	}
}

// WithArchitecture sets the CPU architecture
func WithArchitecture(arch string) ConfigOption {
	return func(c *Config) {
		c.Architecture = arch
	}
}

// WithMode sets the CPU mode
func WithMode(mode string) ConfigOption {
	return func(c *Config) {
		c.Mode = mode
	}
}

// NewConfig creates a new configuration with options
func NewConfig(options ...ConfigOption) *Config {
	config := &Config{
		BaseAddr:     0x10000000,
		StackAddr:    0x20000000,
		StackSize:    1024 * 1024, // 1MB
		ReturnAddr:   0xDEADBEEF,
		Architecture: "arm64",
		Mode:         "arm",
	}

	for _, option := range options {
		option(config)
	}

	return config
}

// DefaultConfig returns a default configuration (deprecated, use NewConfig)
func DefaultConfig() *Config {
	return NewConfig()
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.BaseAddr == 0 {
		return errors.New("base_addr must be non-zero")
	}
	if c.StackAddr == 0 {
		return errors.New("stack_addr must be non-zero")
	}
	if c.StackSize == 0 {
		return errors.New("stack_size must be non-zero")
	}
	if c.BaseAddr == c.StackAddr {
		return errors.New("base_addr and stack_addr cannot be the same")
	}
	if c.Architecture != "arm64" {
		return fmt.Errorf("unsupported architecture: %s", c.Architecture)
	}
	if c.Mode != "arm" {
		return fmt.Errorf("unsupported mode: %s", c.Mode)
	}
	return nil
}
