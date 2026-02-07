package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
)

func Example_expandHandler() {
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService("http://localhost:8080", repo)
	h := handler.New(*svc)

	// Setup test data
	ctx := context.Background()
	store.Store(ctx, model.Link{
		UUID:        uuid.New(),
		ShortURL:    "http://localhost:8080/abc123",
		OriginalURL: "https://example.com",
	})

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rec := httptest.NewRecorder()

	h.ExpandHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	fmt.Printf("%d\n", res.StatusCode)
	fmt.Printf("%s\n", res.Header.Get("Location"))

	// Output:
	// 307
	// https://example.com
}

func TestHandler_ExpandHandler(t *testing.T) {
	type testCase struct {
		name           string
		path           string
		setupStore     func(*storage.MemoryStorage)
		expectedStatus int
		expectedHeader string
	}

	baseURL := config.DefaultBaseURL
	ctx := context.Background()

	cases := []testCase{
		{
			name: "Happy path",
			path: "/abc123",
			setupStore: func(s *storage.MemoryStorage) {
				s.Store(ctx, model.Link{
					UUID:        uuid.New(),
					ShortURL:    baseURL + "/abc123",
					OriginalURL: "https://example.com",
				})
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: "https://example.com",
		},
		{
			name:           "Empty id",
			path:           "/",
			setupStore:     func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Id not found",
			path:           "/notfound",
			setupStore:     func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//arrange
			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService(config.DefaultBaseURL, repo)
			h := handler.New(*svc)
			tc.setupStore(store)

			// act
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			h.ExpandHandler(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			// assert
			if res.StatusCode != tc.expectedStatus {
				t.Fatalf("[%s] expected %d, got %d", tc.name, tc.expectedStatus, res.StatusCode)
			}

			if tc.expectedHeader != "" {
				location := res.Header.Get("Location")
				if location != tc.expectedHeader {
					t.Errorf("[%s] expected Location %q, got %q", tc.name, tc.expectedHeader, location)
				}
			}
		})
	}
}
