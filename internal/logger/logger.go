package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type ctxKey struct{}

var (
	globalLogger     *zap.SugaredLogger
	globalLoggerOnce sync.Once
	loggerMutex      sync.RWMutex
)

func Init() error {
	var err error
	globalLoggerOnce.Do(func() {
		var log *zap.Logger
		log, err = zap.NewProduction()
		if err != nil {
			return
		}
		loggerMutex.Lock()
		globalLogger = log.Sugar()
		loggerMutex.Unlock()
	})
	return err
}

func Get() *zap.SugaredLogger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	if globalLogger == nil {
		loggerMutex.RUnlock()
		_ = Init()
		loggerMutex.RLock()
	}
	return globalLogger
}

func Set(log *zap.SugaredLogger) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	globalLogger = log
}

func WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, Get())
}

func Sync() error {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}
