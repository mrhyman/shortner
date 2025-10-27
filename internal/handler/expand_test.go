package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/service"
)

func TestHandler_ExpandHandler(t *testing.T) {
	type testCase struct {
		name           string
		path           string
		setupStore     func(*repository.LocalStore)
		expectedStatus int
		expectedHeader string
	}

	cases := []testCase{
		{
			name: "Happy path",
			path: "/abc123",
			setupStore: func(s *repository.LocalStore) {
				s.Store("abc123", "https://example.com")
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: "https://example.com",
		},
		{
			name:           "Empty id",
			path:           "/",
			setupStore:     func(s *repository.LocalStore) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Id not found",
			path:           "/notfound",
			setupStore:     func(s *repository.LocalStore) {},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//arrange
			store := repository.NewLocalStore()
			repo := repository.NewLocalURLRepository(store)
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
