package handler_test

import (
	"os"
	"testing"

	"github.com/mrhyman/shortner/internal/logger"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Устанавливаем no-op logger для всех тестов в этом пакете
	logger.Set(zap.NewNop().Sugar())

	// Запускаем тесты
	code := m.Run()

	// Выходим с кодом результата
	os.Exit(code)
}
