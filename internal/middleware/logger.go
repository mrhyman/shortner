// Package middleware реализует промежуточное ПО для обработки HTTP запросов.
package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

// logWriter оборачивает http.ResponseWriter для сбора информации о HTTP ответе.
type logWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// WriteHeader записывает код состояния HTTP в ответ и сохраняет его для логирования.
func (lw *logWriter) WriteHeader(statusCode int) {
	if lw.status != 0 {
		return
	}
	lw.status = statusCode
	lw.ResponseWriter.WriteHeader(statusCode)
}

// Write записывает данные в ответ и сохраняет размер для логирования.
func (lw *logWriter) Write(b []byte) (int, error) {
	if lw.status == 0 {
		lw.status = http.StatusOK
	}
	n, err := lw.ResponseWriter.Write(b)
	lw.size += n
	return n, err
}

// WithLogging это middleware для логирования HTTP запросов и ответов.
// Он записывает информацию о каждом запросе, включая метод, URI, статус ответа, размер и время выполнения.
func WithLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()

		routePattern := chi.RouteContext(req.Context()).RoutePattern()
		if routePattern == "" {
			routePattern = req.URL.Path
		}
		method := req.Method

		rw := &logWriter{ResponseWriter: res}

		next.ServeHTTP(rw, req)

		duration := time.Since(start)

		userID := req.Context().Value(model.UserIDKey)

		logger.Get().With(
			"uri", routePattern,
			"method", method,
			"status", rw.status,
			"size", rw.size,
			"duration", duration.String(),
			"userID", userID,
		).Info()
	}
}
