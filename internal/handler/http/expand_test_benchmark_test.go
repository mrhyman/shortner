// expand_benchmark_test.go
package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler/http"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
)

func BenchmarkExpandHandler(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	ctx := context.Background()
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)

	// Предварительно создаём короткую ссылку
	store.Store(ctx, model.Link{
		UUID:        uuid.New(),
		ShortURL:    baseURL + "/abc123",
		OriginalURL: "https://example.com/original",
	})

	b.ResetTimer()
	b.ReportAllocs()

	// act
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		rec := httptest.NewRecorder()

		h.ExpandHandler(rec, req)

		// assert
		if rec.Code != http.StatusTemporaryRedirect {
			b.Fatalf("unexpected status code: got %d, want %d", rec.Code, http.StatusTemporaryRedirect)
		}
	}
}

func BenchmarkExpandHandlerParallel(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	ctx := context.Background()
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)

	// Предварительно создаём короткую ссылку
	store.Store(ctx, model.Link{
		UUID:        uuid.New(),
		ShortURL:    baseURL + "/abc123",
		OriginalURL: "https://example.com/original",
	})

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		// act
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
			rec := httptest.NewRecorder()

			h.ExpandHandler(rec, req)
		}
	})
}

func BenchmarkExpandHandlerCacheHit(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	ctx := context.Background()
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)

	store.Store(ctx, model.Link{
		UUID:        uuid.New(),
		ShortURL:    baseURL + "/popular",
		OriginalURL: "https://example.com/popular-page",
	})

	b.ResetTimer()
	b.ReportAllocs()

	// act - многократное обращение к одной ссылке (тест кэширования)
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/popular", nil)
		rec := httptest.NewRecorder()

		h.ExpandHandler(rec, req)
	}
}

func BenchmarkExpandHandlerShortCodeLengths(b *testing.B) {
	testCases := []struct {
		name      string
		shortCode string
	}{
		{"Short_3chars", "abc"},
		{"Medium_6chars", "abc123"},
		{"Long_10chars", "abc1234567"},
		{"VeryLong_20chars", "abc12345678901234567"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// arrange
			baseURL := config.DefaultBaseURL
			ctx := context.Background()
			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService(baseURL, repo)
			h := handler.New(*svc)

			store.Store(ctx, model.Link{
				UUID:        uuid.New(),
				ShortURL:    baseURL + "/" + tc.shortCode,
				OriginalURL: "https://example.com/test",
			})

			b.ResetTimer()
			b.ReportAllocs()

			// act
			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodGet, "/"+tc.shortCode, nil)
				rec := httptest.NewRecorder()

				h.ExpandHandler(rec, req)
			}
		})
	}
}
