package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler/http"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
)

func Example_pingHandler() {
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService("http://localhost:8080", repo)
	h := handler.New(*svc)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	h.PingHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	fmt.Printf("%d\n", res.StatusCode)
	fmt.Printf("%s\n", rec.Body.String())

	// Output:
	// 200
	// pong
}

func TestPingHandler(t *testing.T) {

	baseURL := config.DefaultBaseURL

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success GET request",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   "pong",
		},
		{
			name:           "Wrong POST request",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   model.ErrInvalidRequestParams.Error() + "\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// --- arrange ---
			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService(baseURL, repo)
			h := handler.New(*svc)

			req := httptest.NewRequest(tc.method, "/ping", strings.NewReader(""))
			w := httptest.NewRecorder()

			// --- act ---
			h.PingHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			body := w.Body.String()

			// --- assert ---
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			if body != tc.expectedBody {
				t.Errorf("expected body %q, got %q", tc.expectedBody, body)
			}
		})
	}
}
