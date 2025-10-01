package cli

import (
	"context"

	"github.com/sirupsen/logrus"
)

type loggerContextKey struct{}

func WithLogger(ctx context.Context, logger *logrus.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

func GetLogger(ctx context.Context) *logrus.Logger {
	return ctx.Value(loggerContextKey{}).(*logrus.Logger)
}
