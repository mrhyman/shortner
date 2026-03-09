package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/api"
	"github.com/mrhyman/shortner/internal/auth"
	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler/http"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
)

func Example_getUserLinksHandler() {
	ce, _ := auth.NewCookieEncoder("secret")
	userID := "user123"
	userHash, _ := ce.EncodeUserID(userID)

	ctx := context.WithValue(context.Background(), model.UserIDKey, userHash)

	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService("http://localhost:8080", repo)
	h := handler.New(*svc)

	// Setup test data
	store.Store(ctx, model.Link{
		UUID:        uuid.New(),
		ShortURL:    "http://localhost:8080/a1",
		OriginalURL: "https://example.com/1",
		UserID:      userHash,
	})
	store.Store(ctx, model.Link{
		UUID:        uuid.New(),
		ShortURL:    "http://localhost:8080/a2",
		OriginalURL: "https://example.com/2",
		UserID:      userHash,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.GetUserLinksHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	fmt.Printf("%d\n", res.StatusCode)

	var links []api.UserLinksResponse
	if err := json.NewDecoder(res.Body).Decode(&links); err == nil {
		for _, link := range links {
			fmt.Printf("%s %s\n", link.ShortURL, link.OriginalURL)
		}
	}

	// Output:
	// 200
	// http://localhost:8080/a1 https://example.com/1
	// http://localhost:8080/a2 https://example.com/2
}

func TestHandler_GetUserLinksHandler(t *testing.T) {
	type testCase struct {
		name           string
		method         string
		userID         string
		setupStore     func(*storage.MemoryStorage, string)
		expectedStatus int
		expectedLinks  []api.UserLinksResponse
	}

	ce, _ := auth.NewCookieEncoder(config.DefaultHashKey)
	userID := uuid.New().String()

	cases := []testCase{
		{
			name:   "Happy path",
			method: http.MethodGet,
			userID: userID,
			setupStore: func(s *storage.MemoryStorage, userID string) {
				s.Store(context.Background(), model.Link{
					UUID:        uuid.New(),
					ShortURL:    config.DefaultBaseURL + "/a1",
					OriginalURL: "https://example.com/1",
					UserID:      userID,
				})
				s.Store(context.Background(), model.Link{
					UUID:        uuid.New(),
					ShortURL:    config.DefaultBaseURL + "/a2",
					OriginalURL: "https://example.com/2",
					UserID:      userID,
				})
			},
			expectedStatus: http.StatusOK,
			expectedLinks: []api.UserLinksResponse{
				{ShortURL: config.DefaultBaseURL + "/a1", OriginalURL: "https://example.com/1"},
				{ShortURL: config.DefaultBaseURL + "/a2", OriginalURL: "https://example.com/2"},
			},
		},
		{
			name:           "No content",
			method:         http.MethodGet,
			userID:         "user456",
			setupStore:     func(s *storage.MemoryStorage, userID string) {},
			expectedStatus: http.StatusNoContent,
			expectedLinks:  nil,
		},
		{
			name:           "Wrong method",
			method:         http.MethodPost,
			userID:         "user123",
			setupStore:     func(s *storage.MemoryStorage, userID string) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedLinks:  nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userHash, _ := ce.EncodeUserID(tc.userID)
			ctx := context.WithValue(context.Background(), model.UserIDKey, userHash)

			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService(config.DefaultBaseURL, repo)
			h := handler.New(*svc)

			tc.setupStore(store, userHash)

			req := httptest.NewRequest(tc.method, "/api/user/urls", nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			// act
			h.GetUserLinksHandler(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			// assert
			if res.StatusCode != tc.expectedStatus {
				t.Fatalf("[%s] expected status %d, got %d", tc.name, tc.expectedStatus, res.StatusCode)
			}

			if tc.expectedStatus != http.StatusOK {
				return
			}

			var gotLinks []api.UserLinksResponse
			if err := json.NewDecoder(res.Body).Decode(&gotLinks); err != nil {
				t.Fatalf("[%s] failed to decode response: %v", tc.name, err)
			}

			if len(gotLinks) != len(tc.expectedLinks) {
				t.Fatalf("[%s] expected %d links, got %d", tc.name, len(tc.expectedLinks), len(gotLinks))
			}

			for i, l := range gotLinks {
				if l != tc.expectedLinks[i] {
					t.Errorf("[%s] link %d mismatch: expected %+v, got %+v", tc.name, i, tc.expectedLinks[i], l)
				}
			}
		})
	}
}
