package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mrhyman/shortner/internal/logger"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if rw.status != 0 {
		return
	}
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

func WithLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		if routePattern == "" {
			routePattern = r.URL.Path
		}
		method := r.Method

		rw := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)


		logger.FromContext(r.Context()).With(
			"uri", routePattern,
			"method", method,
			"status", rw.status,
			"size", rw.size,
			"duration", duration.String(),
		).Info()

	}
}
