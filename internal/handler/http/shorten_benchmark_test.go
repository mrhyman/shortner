// handler_benchmark_test.go
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrhyman/shortner/api"
	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler/http"
	"github.com/mrhyman/shortner/internal/middleware/http"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
)

func BenchmarkShortenHandler(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)
	handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.ShortenHandler)

	requestBody := api.ShortenRequest{
		URL: "https://example.com/very/long/url/that/needs/to/be/shortened",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.ReportAllocs()

	// act
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rec, req)

		// assert
		if rec.Code != http.StatusCreated {
			b.Fatalf("unexpected status code: got %d, want %d", rec.Code, http.StatusCreated)
		}
	}
}

func BenchmarkShortenHandlerParallel(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)
	handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.ShortenHandler)

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		requestBody := api.ShortenRequest{
			URL: "https://example.com/test/url",
		}
		bodyBytes, _ := json.Marshal(requestBody)

		// act
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			handlerFunc.ServeHTTP(rec, req)
		}
	})
}

func BenchmarkShortenHandlerURLSizes(b *testing.B) {
	testCases := []struct {
		name string
		url  string
	}{
		{"Short", "https://example.com"},
		{"Medium", "https://example.com/path/to/resource?param1=value1&param2=value2"},
		{"Long", "https://example.com/very/long/path/to/some/resource/with/many/segments?param1=value1&param2=value2&param3=value3&param4=value4"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// arrange
			baseURL := config.DefaultBaseURL
			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService(baseURL, repo)
			h := handler.New(*svc)
			handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.ShortenHandler)

			requestBody := api.ShortenRequest{URL: tc.url}
			bodyBytes, _ := json.Marshal(requestBody)

			b.ResetTimer()
			b.ReportAllocs()

			// act
			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				handlerFunc.ServeHTTP(rec, req)
			}
		})
	}
}

func BenchmarkShortenHandlerWithMiddleware(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)
	handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.ShortenHandler)

	requestBody := api.ShortenRequest{
		URL: "https://example.com/test",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.ReportAllocs()

	// act
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rec, req)
	}
}

func BenchmarkShortenHandlerWithoutMiddleware(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)

	requestBody := api.ShortenRequest{
		URL: "https://example.com/test",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.ReportAllocs()

	// act
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		h.ShortenHandler(rec, req)
	}
}
