package encoder

import (
	"github.com/arisu-archive/protocol-encoder-go/internal/emulator"
	"github.com/sirupsen/logrus"
)

func newEmulatorRunner(commonCfg EmulatorConfig, encCfg *EncoderConfig, logger *logrus.Logger) (EmulatorRunner, error) {
	timeout := encCfg.Timeout
	if timeout <= 0 {
		timeout = commonCfg.Timeout
	}
	poolSize := encCfg.PoolSize
	if poolSize <= 0 {
		poolSize = commonCfg.PoolSize
	}
	emulatorConfig := emulator.NewConfig(emulator.WithTimeout(timeout))
	// If pool size is less than 2, we use a single emulator
	if poolSize > 1 {
		return emulator.NewPool(emulatorConfig, logger, poolSize)
	}
	return emulator.New(emulatorConfig, logger)
}
