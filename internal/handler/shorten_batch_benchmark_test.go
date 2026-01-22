package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrhyman/shortner/api"
	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/middleware"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
)

func BenchmarkShortenBatchHandler(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)
	handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.ShortenBatchHandler)

	requestBody := []api.ShortenBatchRequest{
		{CorrelationID: "1", OriginalURL: "https://example.com/url1"},
		{CorrelationID: "2", OriginalURL: "https://example.com/url2"},
		{CorrelationID: "3", OriginalURL: "https://example.com/url3"},
	}
	bodyBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.ReportAllocs()

	// act
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rec, req)

		// assert
		if rec.Code != http.StatusCreated {
			b.Fatalf("unexpected status code: got %d, want %d", rec.Code, http.StatusCreated)
		}
	}
}

func BenchmarkShortenBatchHandlerParallel(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)
	handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.ShortenBatchHandler)

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		requestBody := []api.ShortenBatchRequest{
			{CorrelationID: "1", OriginalURL: "https://example.com/url1"},
			{CorrelationID: "2", OriginalURL: "https://example.com/url2"},
			{CorrelationID: "3", OriginalURL: "https://example.com/url3"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		// act
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			handlerFunc.ServeHTTP(rec, req)
		}
	})
}

func BenchmarkShortenBatchHandlerBatchSizes(b *testing.B) {
	testCases := []struct {
		name  string
		count int
	}{
		{"Small_5", 5},
		{"Medium_20", 20},
		{"Large_50", 50},
		{"XLarge_100", 100},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// arrange
			baseURL := config.DefaultBaseURL
			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService(baseURL, repo)
			h := handler.New(*svc)
			handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.ShortenBatchHandler)

			// Генерируем batch нужного размера
			requestBody := make([]api.ShortenBatchRequest, tc.count)
			for i := 0; i < tc.count; i++ {
				requestBody[i] = api.ShortenBatchRequest{
					CorrelationID: string(rune('A' + i)),
					OriginalURL:   "https://example.com/url" + string(rune('0'+i)),
				}
			}
			bodyBytes, _ := json.Marshal(requestBody)

			b.ResetTimer()
			b.ReportAllocs()

			// act
			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				handlerFunc.ServeHTTP(rec, req)
			}
		})
	}
}

func BenchmarkShortenBatchHandlerEmpty(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)
	handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.ShortenBatchHandler)

	requestBody := []api.ShortenBatchRequest{}
	bodyBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.ReportAllocs()

	// act
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rec, req)

		// assert
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status code: got %d, want %d", rec.Code, http.StatusOK)
		}
	}
}

func BenchmarkShortenBatchHandlerSingleVsBatch(b *testing.B) {
	baseURL := config.DefaultBaseURL
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)

	b.Run("Single_3_Requests", func(b *testing.B) {
		handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.ShortenHandler)

		requests := []api.ShortenRequest{
			{URL: "https://example.com/url1"},
			{URL: "https://example.com/url2"},
			{URL: "https://example.com/url3"},
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			for _, reqBody := range requests {
				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")

				rec := httptest.NewRecorder()
				handlerFunc.ServeHTTP(rec, req)
			}
		}
	})

	b.Run("Batch_3_URLs", func(b *testing.B) {
		handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.ShortenBatchHandler)

		requestBody := []api.ShortenBatchRequest{
			{CorrelationID: "1", OriginalURL: "https://example.com/url1"},
			{CorrelationID: "2", OriginalURL: "https://example.com/url2"},
			{CorrelationID: "3", OriginalURL: "https://example.com/url3"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			handlerFunc.ServeHTTP(rec, req)
		}
	})
}

func BenchmarkShortenBatchHandlerWithMiddleware(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)
	handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.ShortenBatchHandler)

	requestBody := []api.ShortenBatchRequest{
		{CorrelationID: "1", OriginalURL: "https://example.com/url1"},
		{CorrelationID: "2", OriginalURL: "https://example.com/url2"},
		{CorrelationID: "3", OriginalURL: "https://example.com/url3"},
	}
	bodyBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.ReportAllocs()

	// act
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rec, req)
	}
}

func BenchmarkShortenBatchHandlerWithoutMiddleware(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)

	requestBody := []api.ShortenBatchRequest{
		{CorrelationID: "1", OriginalURL: "https://example.com/url1"},
		{CorrelationID: "2", OriginalURL: "https://example.com/url2"},
		{CorrelationID: "3", OriginalURL: "https://example.com/url3"},
	}
	bodyBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.ReportAllocs()

	// act
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		h.ShortenBatchHandler(rec, req)
	}
}
