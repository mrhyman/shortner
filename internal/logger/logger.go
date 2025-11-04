package logger

import (
	"context"

	"log/slog"

	"github.com/mrhyman/shortner/internal/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey struct{}

func WithinContext(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	if l, ok := ctx.Value(ctxKey{}).(*zap.SugaredLogger); ok && l != nil {
		return l
	}
	return New()
}

func New() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		slog.ErrorContext(context.Background(), model.ErrLoggerSetup.Error(), slog.String("err", err.Error()))
	}

	return logger.Sugar()
}

func NewWithCore(core zapcore.Core) *zap.SugaredLogger {
	return zap.New(core).Sugar()
}
