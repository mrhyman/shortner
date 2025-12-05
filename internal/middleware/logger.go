package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

type logWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (lw *logWriter) WriteHeader(statusCode int) {
	if lw.status != 0 {
		return
	}
	lw.status = statusCode
	lw.ResponseWriter.WriteHeader(statusCode)
}

func (lw *logWriter) Write(b []byte) (int, error) {
	if lw.status == 0 {
		lw.status = http.StatusOK
	}
	n, err := lw.ResponseWriter.Write(b)
	lw.size += n
	return n, err
}

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

		logger.FromContext(req.Context()).With(
			"uri", routePattern,
			"method", method,
			"status", rw.status,
			"size", rw.size,
			"duration", duration.String(),
			"userID", userID,
		).Info()
	}
}
