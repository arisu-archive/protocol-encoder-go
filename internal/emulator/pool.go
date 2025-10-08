package emulator

import (
	"fmt"
	"runtime"
	"sync"
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

// Initialize initializes all emulators in the pool concurrently
func (p *EmulatorPool) Initialize() error {
	poolSize := cap(p.pool)
	errChan := make(chan error, poolSize)
	var wg sync.WaitGroup

	// Initialize emulators concurrently
	for i := 0; i < poolSize; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			emu, err := New(p.config, p.logger)
			if err != nil {
				errChan <- fmt.Errorf("failed to create emulator %d: %w", index, err)
				return
			}

			if err := emu.Initialize(); err != nil {
				errChan <- fmt.Errorf("failed to initialize emulator %d: %w", index, err)
				return
			}

			p.pool <- emu
		}(i)
	}

	// Wait for all initializations to complete
	wg.Wait()
	close(errChan)

	// Check for any errors
	if err := <-errChan; err != nil {
		return err
	}

	return nil
}

// Load loads a binary into all emulators in the pool concurrently
func (p *EmulatorPool) Load(binaryPath string) error {
	poolSize := cap(p.pool)

	// Collect all emulators
	emulators := make([]*Emulator, poolSize)
	for i := range poolSize {
		select {
		case emu := <-p.pool:
			emulators[i] = emu
		default:
			return NewEmulatorError(ErrInitialization, fmt.Errorf("failed to get emulator %d from pool", i))
		}
	}

	// Load binary into each emulator concurrently
	errChan := make(chan error, poolSize)
	var wg sync.WaitGroup

	for i, emu := range emulators {
		wg.Add(1)
		go func(index int, e *Emulator) {
			defer wg.Done()
			if err := e.Load(binaryPath); err != nil {
				errChan <- fmt.Errorf("failed to load binary into emulator %d: %w", index, err)
			}
		}(i, emu)
	}

	// Wait for all loads to complete
	wg.Wait()
	close(errChan)

	// Return all emulators to pool
	for _, emu := range emulators {
		p.pool <- emu
	}

	// Check for any errors
	if err := <-errChan; err != nil {
		return err
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
	case <-time.After(time.Second * p.config.Timeout):
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
