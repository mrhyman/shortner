package handler_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bxcodec/faker/v4"
	"github.com/google/uuid"

	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/middleware"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
)

func Example_shortenEndpoint() {
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService("http://localhost:8080", repo)

	// Мокаем генератор
	svc.IDGenerator = func() (string, error) {
		return "test1234", nil
	}

	h := handler.New(*svc)
	handlerFunc := middleware.WithAuth("secret")(h.ShortLinkHandler)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")

	rec := httptest.NewRecorder()
	handlerFunc.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	fmt.Printf("%d\n", res.StatusCode)
	fmt.Printf("%s\n", strings.TrimSpace(string(body)))

	// Output:
	// 201
	// http://localhost:8080/test1234
}

func TestHandler_ShortLinkHandler(t *testing.T) {
	type testCase struct {
		name           string
		method         string
		contentType    string
		originalURL    string
		expectedStatus int
		expectInStore  bool
	}

	ctx := context.Background()
	baseURL := config.DefaultBaseURL

	cases := []testCase{
		{
			name:           "Happy path",
			method:         http.MethodPost,
			contentType:    "text/plain",
			originalURL:    faker.URL(),
			expectedStatus: http.StatusCreated,
			expectInStore:  true,
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			contentType:    "text/plain",
			originalURL:    faker.URL(),
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid content type",
			method:         http.MethodPost,
			contentType:    "application/json",
			originalURL:    faker.URL(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty body",
			method:         http.MethodPost,
			contentType:    "text/plain",
			originalURL:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing content type",
			method:         http.MethodPost,
			contentType:    "",
			originalURL:    faker.URL(),
			expectedStatus: http.StatusCreated,
			expectInStore:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService(baseURL, repo)
			h := handler.New(*svc)
			handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.ShortLinkHandler)

			req := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.originalURL))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			// act
			rec := httptest.NewRecorder()
			handlerFunc.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			// assert
			if res.StatusCode != tc.expectedStatus {
				t.Fatalf("[%s] expected %d, got %d", tc.name, tc.expectedStatus, res.StatusCode)
			}

			if tc.expectInStore {
				body, _ := io.ReadAll(res.Body)
				shortURL := strings.TrimSpace(string(body))
				err := store.Store(ctx, model.Link{
					UUID:        uuid.New(),
					ShortURL:    shortURL,
					OriginalURL: tc.originalURL,
				})
				if err != nil {
					t.Fatalf("[%s] short_url %q not found in store", tc.name, shortURL)
				}

				v, _ := store.GetByShortURL(ctx, shortURL)
				if v.OriginalURL != tc.originalURL {
					t.Errorf("[%s] expected %q, got %q", tc.name, tc.originalURL, v.OriginalURL)
				}
			}
		})
	}
}
