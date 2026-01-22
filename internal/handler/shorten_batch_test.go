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
	"github.com/mrhyman/shortner/internal/middleware"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
	"github.com/stretchr/testify/assert"
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
			name:           "Invalid method",
			method:         http.MethodGet,
			body:           nil,
			wantStatusCode: http.StatusMethodNotAllowed,
			wantResp:       nil,
		},
		{
			name:           "Empty body",
			method:         http.MethodPost,
			body:           nil,
			wantStatusCode: http.StatusBadRequest,
			wantResp:       nil,
		},
		{
			name:           "Empty slice",
			method:         http.MethodPost,
			body:           []api.ShortenBatchRequest{},
			wantStatusCode: http.StatusOK,
			wantResp:       []struct{}{},
		},
		{
			name:   "Happy path",
			method: http.MethodPost,
			body: []api.ShortenBatchRequest{
				{CorrelationID: "1", OriginalURL: "http://example.com/1"},
				{CorrelationID: "2", OriginalURL: "http://example.com/2"},
			},
			wantStatusCode: http.StatusCreated,
			wantResp: []api.ShortenBatchResponse{
				{CorrelationID: "1"},
				{CorrelationID: "2"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService(baseURL, repo)
			h := handler.New(*svc)
			handlerFunc := middleware.WithAuth(config.DefaultHashKey)(h.ShortenBatchHandler)

			var bodyBytes []byte
			if tc.body != nil {
				var err error
				bodyBytes, err = json.Marshal(tc.body)
				if err != nil {
					t.Fatalf("failed to marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(tc.method, "/shorten/batch", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			handlerFunc.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.wantStatusCode {
				t.Errorf("status code: got %v, want %v", res.StatusCode, tc.wantStatusCode)
			}

			if tc.wantResp != nil {
				switch tc.wantResp.(type) {
				case string:
					got := string(bodyBytes)
					if got != tc.wantResp.(string) {
						t.Errorf("response: got %q, want %q", got, tc.wantResp.(string))
					}
				default:
					var got []api.ShortenBatchResponse
					if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
						t.Fatalf("failed to decode response: %v", err)
					}

					wantResp, _ := tc.wantResp.([]api.ShortenBatchResponse)
					if len(got) != len(wantResp) {
						t.Errorf("response length: got %d items, want %d", len(got), len(wantResp))
					}

					for i := range wantResp {
						assert.Equal(t, wantResp[i].CorrelationID, got[i].CorrelationID)
						assert.NotEmpty(t, got[i].ShortURL)
					}
				}
			}
		})
	}
}
