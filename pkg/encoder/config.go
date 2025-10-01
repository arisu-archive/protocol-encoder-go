package encoder

import (
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Binary   string         `json:"binary" validate:"required,file"`
	Offset   uint64         `json:"offset" validate:"required"`
	PoolSize int            `json:"pool_size" validate:"required,min=1"`
	Logger   *logrus.Logger `json:"-" validate:"required"`
}

type Option func(*Config)

func WithPoolSize(poolSize int) Option {
	return func(cfg *Config) {
		cfg.PoolSize = poolSize
	}
}

func (c *Config) Validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	return v.Struct(c)
}
