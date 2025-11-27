package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mrhyman/shortner/api"
	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/middleware"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
)

func TestShortenHandler(t *testing.T) {
	type testCase struct {
		name           string
		method         string
		body           any
		wantStatusCode int
		wantContains   string
	}

	baseURL := config.DefaultBaseURL

	cases := []testCase{
		{
			name:           "Happy path",
			method:         http.MethodPost,
			body:           api.ShortenRequest{URL: "https://example.com"},
			wantStatusCode: http.StatusCreated,
			wantContains:   config.DefaultServerAddress,
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			body:           nil,
			wantStatusCode: http.StatusMethodNotAllowed,
			wantContains:   model.ErrInvalidRequestParams.Error(),
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			body:           `{"url": 123}`, // неправильный тип
			wantStatusCode: http.StatusInternalServerError,
			wantContains:   "",
		},
		{
			name:           "Empty URL",
			method:         http.MethodPost,
			body:           api.ShortenRequest{URL: ""},
			wantStatusCode: http.StatusBadRequest,
			wantContains:   model.ErrInvalidURL.Error(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService(baseURL, repo)
			h := handler.New(*svc)
			handlerFunc := middleware.WithAuth(h.ShortenHandler)

			var reqBody []byte
			switch v := tc.body.(type) {
			case string:
				reqBody = []byte(v)
			case nil:
				reqBody = nil
			default:
				b, _ := json.Marshal(v)
				reqBody = b
			}

			// act
			req := httptest.NewRequest(tc.method, "/api/shorten", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			handlerFunc.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			body := rec.Body.Bytes()

			// assert
			if tc.wantStatusCode == http.StatusCreated {
				var resp api.ShortenResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("[%s] invalid JSON: %v\nОтвет: %s", tc.name, err, string(body))
				}

				if !strings.HasPrefix(resp.Result, baseURL) {
					t.Errorf("[%s] expected с %q, got %q",
						tc.name, baseURL, resp.Result)
				}
			} else if tc.wantContains != "" && !strings.Contains(string(body), tc.wantContains) {
				t.Errorf("[%s] expected %q got: %s", tc.name, tc.wantContains, string(body))
			}
		})
	}
}
