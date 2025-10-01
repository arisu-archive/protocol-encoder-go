package emulator

import (
	"fmt"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// EmulatorPool manages a pool of emulator instances for high RPS
type EmulatorPool struct {
	pool   chan *Emulator
	config *Config
	logger *logrus.Logger
}

// NewPool creates a new emulator pool for high-performance concurrent execution
func NewPool(config *Config, logger *logrus.Logger, poolSize int) (*EmulatorPool, error) {
	if poolSize <= 0 {
		poolSize = runtime.NumCPU() * 2 // Default to 2x CPU cores
	}

	pool := &EmulatorPool{
		pool:   make(chan *Emulator, poolSize),
		config: config,
		logger: logger,
	}

	logger.WithField("pool_size", poolSize).Info("Emulator pool created successfully")
	return pool, nil
}

// Initialize initializes all emulators in the pool
func (p *EmulatorPool) Initialize() error {
	for i := 0; i < cap(p.pool); i++ {
		emu, err := New(p.config, p.logger)
		if err != nil {
			return err
		}

		if err := emu.Initialize(); err != nil {
			return err
		}

		p.pool <- emu
	}
	return nil
}

// Load loads a binary into all emulators in the pool
func (p *EmulatorPool) Load(binaryPath string) error {
	poolSize := cap(p.pool)

	// Collect all emulators
	emulators := make([]*Emulator, poolSize)
	for i := 0; i < poolSize; i++ {
		select {
		case emu := <-p.pool:
			emulators[i] = emu
		default:
			return NewEmulatorError(ErrInitialization, fmt.Errorf("failed to get emulator %d from pool", i))
		}
	}

	// Load binary into each emulator
	for i, emu := range emulators {
		if err := emu.Load(binaryPath); err != nil {
			// Return all emulators to pool before returning error
			for j := 0; j <= i; j++ {
				p.pool <- emulators[j]
			}
			return fmt.Errorf("failed to load binary into emulator %d: %w", i, err)
		}
	}

	// Return all emulators to pool
	for _, emu := range emulators {
		p.pool <- emu
	}

	return nil
}

// Invoke executes a function using the emulator pool
func (p *EmulatorPool) Invoke(req *InvokeRequest) (*InvokeResponse, error) {
	// Get an emulator from the pool
	var emu *Emulator
	select {
	case emu = <-p.pool:
		// Got an emulator
	case <-time.After(time.Second * 10): // Timeout after 10 seconds
		return &InvokeResponse{
			Success: false,
			Error:   "timeout waiting for available emulator",
		}, NewEmulatorError(ErrEmulationStart, ErrPoolTimeout)
	}

	// Ensure we return the emulator to the pool
	defer func() {
		p.pool <- emu
	}()

	// Execute the function
	response, err := emu.Invoke(req)

	return response, err
}

// Close closes all emulators in the pool
func (p *EmulatorPool) Close() error {
	close(p.pool)

	var lastError error
	for emu := range p.pool {
		if err := emu.Close(); err != nil {
			lastError = err
			p.logger.WithError(err).Error("Failed to close emulator")
		}
	}

	return lastError
}
