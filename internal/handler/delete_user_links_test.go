package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestHandler_DeleteUserLinksHandler(t *testing.T) {
	type testCase struct {
		name           string
		method         string
		body           any
		setupStore     func(*storage.MemoryStorage)
		expectedStatus int
	}

	userHash, _ := middleware.EncodeUserID(uuid.New().String(), config.DefaultHashKey)
	ctx := context.Background()

	cases := []testCase{
		{
			name:   "Happy path",
			method: http.MethodDelete,
			body:   []string{"abc123", "def456"},
			setupStore: func(s *storage.MemoryStorage) {
				s.Store(ctx, model.Link{
					UUID:        uuid.New(),
					ShortURL:    config.DefaultBaseURL + "/abc123",
					OriginalURL: "https://example.com/1",
					UserID:      userHash,
				})
				s.Store(ctx, model.Link{
					UUID:        uuid.New(),
					ShortURL:    config.DefaultBaseURL + "/def456",
					OriginalURL: "https://example.com/2",
					UserID:      userHash,
				})
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "Wrong method",
			method:         http.MethodGet,
			body:           []string{"abc123"},
			setupStore:     func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Bad JSON body",
			method:         http.MethodDelete,
			body:           `{not valid json`,
			setupStore:     func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty list is OK",
			method:         http.MethodDelete,
			body:           []string{},
			setupStore:     func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusAccepted,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			// arrange

			ctx = context.WithValue(ctx, model.UserIDKey, userHash)

			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService(config.DefaultBaseURL, repo)
			h := handler.New(*svc)
			tc.setupStore(store)

			var reqBody []byte
			var err error

			switch b := tc.body.(type) {
			case string:
				reqBody = []byte(b)
			default:
				reqBody, err = json.Marshal(b)
				if err != nil {
					t.Fatalf("failed to marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(tc.method, "/api/user/urls", bytes.NewReader(reqBody))
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			// act
			h.DeleteUserLinksHandler(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			// assert
			if res.StatusCode != tc.expectedStatus {
				t.Fatalf("[%s] expected %d, got %d", tc.name, tc.expectedStatus, res.StatusCode)
			}
		})
	}
}
