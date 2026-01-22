package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/middleware"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestWithLogging_Table(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
	}{
		{"GET root", "GET", "/", http.StatusOK},
		{"POST root", "POST", "/", http.StatusCreated},
		{"POST shorten", "POST", "/shorten", http.StatusCreated},
		{"not found", "GET", "/unknown", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService(config.DefaultBaseURL, repo)
			_ = handler.New(*svc)

			core, obs := observer.New(zapcore.InfoLevel)
			testLogger := zap.New(core).Sugar()

			oldLogger := logger.Get()
			logger.Set(testLogger)
			defer logger.Set(oldLogger)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// act
			h := middleware.WithLogging(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("ok"))
			})

			h.ServeHTTP(w, req)

			// assert
			resp := w.Result()
			defer resp.Body.Close()

			body := w.Body.String()

			if resp.StatusCode != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, resp.StatusCode)
			}
			if body != "ok" {
				t.Errorf("expected body %q, got %q", "ok", body)
			}

			entries := obs.All()
			if len(entries) == 0 {
				t.Errorf("expected at least one log entry, got none")
				return
			}

			e := entries[len(entries)-1]
			ctxMap := e.ContextMap()

			if ctxMap["method"] != tt.method {
				t.Errorf("expected method %s, got %v", tt.method, ctxMap["method"])
			}
			if ctxMap["uri"] != tt.path {
				t.Errorf("expected uri %s, got %v", tt.path, ctxMap["uri"])
			}
			if ctxMap["status"] != int64(tt.statusCode) {
				t.Errorf("expected status %d, got %v", tt.statusCode, ctxMap["status"])
			}
			if ctxMap["size"] != int64(len("ok")) {
				t.Errorf("expected size %d, got %v", len("ok"), ctxMap["size"])
			}
			if _, ok := ctxMap["duration"]; !ok {
				t.Errorf("expected duration field in log")
			}
		})
	}
}
