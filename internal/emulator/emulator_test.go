package emulator

import (
	"context"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stderr)

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid config - zero base addr",
			config: &Config{
				BaseAddr:  0,
				StackAddr: 0x20000000,
				StackSize: 1024 * 1024,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emu, err := New(tt.config, logger)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, emu)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, emu)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "same base and stack addr",
			config: &Config{
				BaseAddr:  0x10000000,
				StackAddr: 0x10000000,
				StackSize: 1024,
			},
			wantErr: true,
		},
		{
			name: "unsupported architecture",
			config: &Config{
				BaseAddr:     0x10000000,
				StackAddr:    0x20000000,
				StackSize:    1024,
				Architecture: "x86",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmulatorError(t *testing.T) {
	originalErr := assert.AnError
	emulatorErr := NewEmulatorError("test operation", originalErr)

	assert.Contains(t, emulatorErr.Error(), "test operation")
	assert.Contains(t, emulatorErr.Error(), originalErr.Error())
	assert.Equal(t, originalErr, emulatorErr.Unwrap())
}

// Integration test - requires test files
func TestEmulator_Integration(t *testing.T) {
	// Skip if test files don't exist
	if _, err := os.Stat("../../libil2cpp.so"); os.IsNotExist(err) {
		t.Skip("Skipping integration test - libil2cpp.so not found")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during test

	config := DefaultConfig()

	emu, err := New(config, logger)
	require.NoError(t, err)
	defer emu.Close()

	err = emu.Initialize()
	require.NoError(t, err)

	err = emu.Load("../../libil2cpp.so")
	require.NoError(t, err)

	// Test function invocation
	req := &InvokeRequest{
		Offset:  0x6268754, // Example offset
		Args:    []uint64{0, 1, 1014},
		Context: context.Background(),
	}

	result, err := emu.Invoke(req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
}
