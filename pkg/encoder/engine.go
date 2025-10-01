package encoder

import (
	"fmt"

	"github.com/arisu-archive/protocol-encoder-go/internal/emulator"
)

type EmulatorRunner interface {
	Initialize() error
	Load(binaryPath string) error
	Close() error
	Invoke(req *emulator.InvokeRequest) (*emulator.InvokeResponse, error)
}

type Encoder struct {
	cfg    *Config
	runner EmulatorRunner
}

func NewEncoder(cfg *Config, opts ...Option) (*Encoder, error) {
	// Apply options to config
	for _, opt := range opts {
		opt(cfg)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	runner, err := newEmulatorRunner(cfg, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create emulator runner: %w", err)
	}
	if err := runner.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize emulator runner: %w", err)
	}
	if err := runner.Load(cfg.Binary); err != nil {
		return nil, fmt.Errorf("failed to load emulator runner: %w", err)
	}
	return &Encoder{runner: runner, cfg: cfg}, nil
}

func (e *Encoder) Encode(protocol, crc32 uint64) (*emulator.InvokeResponse, error) {
	return e.runner.Invoke(&emulator.InvokeRequest{
		Offset: e.cfg.Offset,
		Args:   []uint64{protocol, crc32},
	})
}

func (e *Encoder) Close() error {
	if err := e.runner.Close(); err != nil {
		return fmt.Errorf("failed to close emulator runner: %w", err)
	}
	return nil
}
