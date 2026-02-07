package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/middleware"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
)

func BenchmarkGetUserLinksHandler(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	ctx := context.Background()
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)
	handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.GetUserLinksHandler)

	userID := uuid.New().String()

	for i := 0; i < 5; i++ {
		store.Store(ctx, model.Link{
			UUID:        uuid.New(),
			ShortURL:    baseURL + "/link" + string(rune('0'+i)),
			OriginalURL: "https://example.com/url" + string(rune('0'+i)),
			UserID:      userID,
		})
	}

	b.ResetTimer()
	b.ReportAllocs()

	// act
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		req = req.WithContext(context.WithValue(req.Context(), model.UserIDKey, userID))

		rec := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rec, req)

		// assert
		if rec.Code != http.StatusNoContent {
			b.Fatalf("unexpected status code: got %d, want %d", rec.Code, http.StatusNoContent)
		}
	}
}

func BenchmarkGetUserLinksHandlerParallel(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	ctx := context.Background()
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)
	handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.GetUserLinksHandler)

	userID := uuid.New().String()

	for i := 0; i < 5; i++ {
		store.Store(ctx, model.Link{
			UUID:        uuid.New(),
			ShortURL:    baseURL + "/link" + string(rune('0'+i)),
			OriginalURL: "https://example.com/url" + string(rune('0'+i)),
			UserID:      userID,
		})
	}

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		// act
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			req = req.WithContext(context.WithValue(req.Context(), model.UserIDKey, userID))

			rec := httptest.NewRecorder()
			handlerFunc.ServeHTTP(rec, req)
		}
	})
}

func BenchmarkGetUserLinksHandlerUserLinkCounts(b *testing.B) {
	testCases := []struct {
		name  string
		count int
	}{
		{"Empty", 0},
		{"Few_5", 5},
		{"Medium_50", 50},
		{"Many_200", 200},
		{"Large_1000", 1000},
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
			handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.GetUserLinksHandler)

			userID := uuid.New().String()

			for i := 0; i < tc.count; i++ {
				shortID, _ := service.GenerateShortID()
				store.Store(ctx, model.Link{
					UUID:        uuid.New(),
					ShortURL:    baseURL + "/" + shortID,
					OriginalURL: "https://example.com/url" + shortID,
					UserID:      userID,
				})
			}

			b.ResetTimer()
			b.ReportAllocs()

			// act
			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
				req = req.WithContext(context.WithValue(req.Context(), model.UserIDKey, userID))

				rec := httptest.NewRecorder()
				handlerFunc.ServeHTTP(rec, req)

				// assert
				expectedStatus := http.StatusNoContent
				if tc.count == 0 {
					expectedStatus = http.StatusNoContent
				}
				if rec.Code != expectedStatus {
					b.Fatalf("unexpected status code: got %d, want %d", rec.Code, expectedStatus)
				}
			}
		})
	}
}

func BenchmarkGetUserLinksHandlerMultipleUsers(b *testing.B) {
	testCases := []struct {
		name         string
		userCount    int
		linksPerUser int
	}{
		{"Users_10_Links_5", 10, 5},
		{"Users_100_Links_10", 100, 10},
		{"Users_1000_Links_5", 1000, 5},
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
			handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.GetUserLinksHandler)

			userIDs := make([]string, tc.userCount)
			for u := 0; u < tc.userCount; u++ {
				userID := uuid.New().String()
				userIDs[u] = userID

				for i := 0; i < tc.linksPerUser; i++ {
					shortID, _ := service.GenerateShortID()
					store.Store(ctx, model.Link{
						UUID:        uuid.New(),
						ShortURL:    baseURL + "/" + shortID,
						OriginalURL: "https://example.com/url" + shortID,
						UserID:      userID,
					})
				}
			}

			b.ResetTimer()
			b.ReportAllocs()

			// act - запросы от разных пользователей
			for i := 0; i < b.N; i++ {
				userID := userIDs[i%tc.userCount]
				req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
				req = req.WithContext(context.WithValue(req.Context(), model.UserIDKey, userID))

				rec := httptest.NewRecorder()
				handlerFunc.ServeHTTP(rec, req)
			}
		})
	}
}

func BenchmarkGetUserLinksHandlerNoContent(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)
	handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.GetUserLinksHandler)

	userID := uuid.New().String()

	b.ResetTimer()
	b.ReportAllocs()

	// act
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		req = req.WithContext(context.WithValue(req.Context(), model.UserIDKey, userID))

		rec := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rec, req)

		// assert
		if rec.Code != http.StatusNoContent {
			b.Fatalf("unexpected status code: got %d, want %d", rec.Code, http.StatusNoContent)
		}
	}
}

func BenchmarkGetUserLinksHandlerWithMiddleware(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	ctx := context.Background()
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)
	handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.GetUserLinksHandler)

	userID := uuid.New().String()

	for i := 0; i < 10; i++ {
		store.Store(ctx, model.Link{
			UUID:        uuid.New(),
			ShortURL:    baseURL + "/link" + string(rune('0'+i)),
			OriginalURL: "https://example.com/url" + string(rune('0'+i)),
			UserID:      userID,
		})
	}

	b.ResetTimer()
	b.ReportAllocs()

	// act
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		req = req.WithContext(context.WithValue(req.Context(), model.UserIDKey, userID))

		rec := httptest.NewRecorder()
		handlerFunc.ServeHTTP(rec, req)
	}
}

func BenchmarkGetUserLinksHandlerWithoutMiddleware(b *testing.B) {
	// arrange
	baseURL := config.DefaultBaseURL
	ctx := context.Background()
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(baseURL, repo)
	h := handler.New(*svc)

	userID := uuid.New().String()

	for i := 0; i < 10; i++ {
		store.Store(ctx, model.Link{
			UUID:        uuid.New(),
			ShortURL:    baseURL + "/link" + string(rune('0'+i)),
			OriginalURL: "https://example.com/url" + string(rune('0'+i)),
			UserID:      userID,
		})
	}

	b.ResetTimer()
	b.ReportAllocs()

	// act
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		req = req.WithContext(context.WithValue(req.Context(), model.UserIDKey, userID))

		rec := httptest.NewRecorder()
		h.GetUserLinksHandler(rec, req)
	}
}

func BenchmarkGetUserLinksHandlerURLLengths(b *testing.B) {
	testCases := []struct {
		name      string
		urlLength int
	}{
		{"Short_URLs", 30},
		{"Medium_URLs", 100},
		{"Long_URLs", 500},
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
			handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.GetUserLinksHandler)

			userID := uuid.New().String()

			for i := 0; i < 10; i++ {
				longURL := "https://example.com/" + string(make([]byte, tc.urlLength))
				store.Store(ctx, model.Link{
					UUID:        uuid.New(),
					ShortURL:    baseURL + "/link" + string(rune('0'+i)),
					OriginalURL: longURL,
					UserID:      userID,
				})
			}

			b.ResetTimer()
			b.ReportAllocs()

			// act
			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
				req = req.WithContext(context.WithValue(req.Context(), model.UserIDKey, userID))

				rec := httptest.NewRecorder()
				handlerFunc.ServeHTTP(rec, req)
			}
		})
	}
}
