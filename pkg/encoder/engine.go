package encoder

import (
	"fmt"

	"github.com/arisu-archive/protocol-encoder-go/internal/emulator"
	"github.com/sirupsen/logrus"
)

type EmulatorRunner interface {
	Initialize() error
	Load(binaryPath string) error
	Close() error
	Invoke(req *emulator.InvokeRequest) (*emulator.InvokeResponse, error)
}

type Encoder struct {
	cfg    EncoderConfig
	runner EmulatorRunner
}

func NewEncoder(commonCfg EmulatorConfig, cfg EncoderConfig, logger *logrus.Logger) (*Encoder, error) {
	runner, err := newEmulatorRunner(commonCfg, cfg, logger)
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
	// __thiscall. But the funciton doesn't use 'this', so just pass 0.
	return e.runner.Invoke(&emulator.InvokeRequest{
		Offset: e.cfg.Offset,
		Args:   []uint64{0, crc32, protocol},
	})
}

func (e *Encoder) Close() error {
	if err := e.runner.Close(); err != nil {
		return fmt.Errorf("failed to close emulator runner: %w", err)
	}
	return nil
}
