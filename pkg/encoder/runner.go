package encoder

import (
	"github.com/arisu-archive/protocol-encoder-go/internal/emulator"
	"github.com/sirupsen/logrus"
)

func newEmulatorRunner(cfg *Config, logger *logrus.Logger) (EmulatorRunner, error) {
	// If pool size is less than 2, we use a single emulator
	if cfg.PoolSize > 1 {
		return emulator.NewPool(emulator.NewConfig(), logger, cfg.PoolSize)
	}
	return emulator.New(emulator.NewConfig(), logger)
}
