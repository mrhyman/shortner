package handler_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/bxcodec/faker/v4"
	"github.com/google/uuid"

	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
)

func TestHandler_ShortLinkHandler(t *testing.T) {
	type testCase struct {
		name           string
		method         string
		contentType    string
		originalURL    string
		expectedStatus int
		expectInStore  bool
	}

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

			req := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.originalURL))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			// act
			rec := httptest.NewRecorder()
			h.ShortLinkHandler(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			// assert
			if res.StatusCode != tc.expectedStatus {
				t.Fatalf("[%s] expected %d, got %d", tc.name, tc.expectedStatus, res.StatusCode)
			}

			if tc.expectInStore {
				body, _ := io.ReadAll(res.Body)
				shortURL := strings.TrimSpace(string(body))
				u, err := url.Parse(shortURL)
				if err != nil {
					t.Fatalf("[%s] invalid short URL: %v", tc.name, err)
				}
				id := path.Base(u.Path)
				err = store.Store(model.Link{
					UUID: uuid.New(),
					ShortURL: shortURL,
					OriginalURL: tc.originalURL,
				})
				if err != nil {
					t.Fatalf("[%s] id %q not found in store", tc.name, id)
				}

				v, _ := store.GetByID(id)
				if v != tc.originalURL {
					t.Errorf("[%s] expected %q, got %q", tc.name, tc.originalURL, v)
				}
			}
		})
	}
}
