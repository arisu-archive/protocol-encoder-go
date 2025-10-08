package cli

import (
	"context"

	"github.com/arisu-archive/protocol-encoder-go/pkg/encoder"
	"github.com/sirupsen/logrus"
)

type loggerContextKey struct{}

func WithLogger(ctx context.Context, logger *logrus.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

func GetLogger(ctx context.Context) *logrus.Logger {
	return ctx.Value(loggerContextKey{}).(*logrus.Logger)
}

type configContextKey struct{}

func WithConfig(ctx context.Context, cfg *encoder.Config) context.Context {
	return context.WithValue(ctx, configContextKey{}, cfg)
}

func GetConfig(ctx context.Context) *encoder.Config {
	return ctx.Value(configContextKey{}).(*encoder.Config)
}
