package encoder

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Emulator EmulatorConfig            `mapstructure:"emulator" validate:"required"`
	Encoders map[string]*EncoderConfig `mapstructure:"encoders" validate:"required,dive,keys,required,endkeys,required"`
	Logger   *logrus.Logger            `mapstructure:"-"` // not from config file
}

type EmulatorConfig struct {
	PoolSize int           `mapstructure:"pool_size" validate:"required,min=1"`
	Timeout  time.Duration `mapstructure:"timeout" validate:"required,min=1"`
}

type EncoderConfig struct {
	BinaryPath string        `mapstructure:"binary" validate:"required"`
	OffsetPath string        `mapstructure:"offset" validate:"required"`
	Aliases    []string      `mapstructure:"aliases"`
	PoolSize   int           `mapstructure:"pool_size" validate:"omitempty,min=1"`
	Timeout    time.Duration `mapstructure:"timeout" validate:"omitempty,min=1"`

	parsedOffset uint64
	offsetOnce   sync.Once
}

func (c *Config) Validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	return v.Struct(c)
}

func parseOffsetFile(path string) (uint64, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read offset file: %w", err)
	}
	var offset uint64
	_, err = fmt.Sscanf(string(buf), "0x%X", &offset)
	if err != nil {
		return 0, fmt.Errorf("failed to parse offset: %w", err)
	}
	return offset, nil
}

// GetOffset returns the parsed offset. If OffsetPath is empty, returns 0.
// If OffsetPath is a file path, reads the file and parses the offset.
// Caches the result after the first call.
func (e *EncoderConfig) GetOffset() uint64 {
	e.offsetOnce.Do(func() {
		offset, err := parseOffsetFile(e.OffsetPath)
		if err != nil {
			panic(fmt.Sprintf("failed to parse offset file %s: %v", e.OffsetPath, err))
		}
		e.parsedOffset = offset
	})
	return e.parsedOffset
}
