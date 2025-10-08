package encoder

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Emulator EmulatorConfig           `mapstructure:"emulator" validate:"required"`
	Encoders map[string]EncoderConfig `mapstructure:"encoders" validate:"required,dive,keys,required,endkeys,required"`
	Logger   *logrus.Logger           `mapstructure:"-"` // not from config file
}

type EmulatorConfig struct {
	PoolSize int           `mapstructure:"pool_size" validate:"required,min=1"`
	Timeout  time.Duration `mapstructure:"timeout" validate:"required,min=1"`
}

type EncoderConfig struct {
	Binary   string        `mapstructure:"binary" validate:"required,file"`
	Offset   uint64        `mapstructure:"offset"`
	Aliases  []string      `mapstructure:"aliases"`
	PoolSize int           `mapstructure:"pool_size" validate:"omitempty,min=1"`
	Timeout  time.Duration `mapstructure:"timeout" validate:"omitempty,min=1"`
}

func (c *Config) Validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	return v.Struct(c)
}
