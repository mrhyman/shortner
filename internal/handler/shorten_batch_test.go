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
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
)

func TestShortenBatchHandler(t *testing.T) {
	type testCase struct {
		name           string
		method         string
		body           any
		wantStatusCode int
		wantResp       any
	}

	baseURL := config.DefaultBaseURL

	tests := []testCase{
		{
			name:           "invalid method",
			method:         http.MethodGet,
			body:           nil,
			wantStatusCode: http.StatusMethodNotAllowed,
			wantResp:       nil,
		},
		{
			name:           "empty body",
			method:         http.MethodPost,
			body:           nil,
			wantStatusCode: http.StatusBadRequest,
			wantResp:       nil,
		},
		{
			name:           "empty slice",
			method:         http.MethodPost,
			body:           []api.ShortenBatchRequest{},
			wantStatusCode: http.StatusOK,
			wantResp:       []struct{}{},
		},
		// замокать бд
		// {
		// 	name:   "valid requests",
		// 	method: http.MethodPost,
		// 	body: []api.ShortenBatchRequest{
		// 		{CorrelationID: "1", OriginalURL: "http://example.com/1"},
		// 		{CorrelationID: "2", OriginalURL: "http://example.com/2"},
		// 	},
		// 	wantStatusCode: http.StatusCreated,
		// 	wantResp: []api.ShortenBatchResponse{
		// 		{CorrelationID: "1", ShortURL: baseURL + "1"},
		// 		{CorrelationID: "2", ShortURL: baseURL + "2"},
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService(baseURL, repo)
			h := handler.New(*svc)

			var bodyBytes []byte
			if tt.body != nil {
				var err error
				bodyBytes, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("failed to marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/shorten/batch", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.ShortenBatchHandler(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatusCode {
				t.Errorf("status code: got %v, want %v", res.StatusCode, tt.wantStatusCode)
			}

			if tt.wantResp != nil {
				switch tt.wantResp.(type) {
				case string:
					got := string(bodyBytes)
					if got != tt.wantResp.(string) {
						t.Errorf("response: got %q, want %q", got, tt.wantResp.(string))
					}
				default:
					var got []api.ShortenBatchResponse
					if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
						t.Fatalf("failed to decode response: %v", err)
					}

					if !equal(got, tt.wantResp) {
						t.Errorf("response: got %+v, want %+v", got, tt.wantResp)
					}
				}
			}
		})
	}
}

// простой helper для проверки равенства слайсов
func equal(a, b interface{}) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return bytes.Equal(aj, bj)
}
