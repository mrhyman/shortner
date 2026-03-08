package middleware_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bxcodec/faker/v4"
	"github.com/mrhyman/shortner/api"
	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler/http"
	"github.com/mrhyman/shortner/internal/middleware/http"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
)

func gzipData(t *testing.T, data string) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(data)); err != nil {
		t.Fatalf("gzip write error: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close error: %v", err)
	}
	return &buf
}

func gunzipData(t *testing.T, data []byte) string {
	t.Helper()
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("create gzip reader error: %v", err)
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read gzip error: %v", err)
	}
	return string(out)
}

func TestWithGzip(t *testing.T) {
	type testCase struct {
		name              string
		requestBody       string
		requestCompressed bool
		expectCompressed  bool
	}

	baseURL := config.DefaultBaseURL

	cases := []testCase{
		{
			name:              "No encoding",
			requestBody:       fmt.Sprintf(`{"url": "%s"}`, faker.URL()),
			requestCompressed: false,
			expectCompressed:  false,
		},
		{
			name:              "Full encoding",
			requestBody:       fmt.Sprintf(`{"url": "%s"}`, faker.URL()),
			requestCompressed: true,
			expectCompressed:  true,
		},
	}

	for _, tс := range cases {
		t.Run(tс.name, func(t *testing.T) {
			// arrange
			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService(baseURL, repo)
			h := handler.New(*svc)

			body := bytes.NewBufferString(tс.requestBody)
			if tс.requestCompressed {
				body = gzipData(t, tс.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
			if tс.requestCompressed {
				req.Header.Set("Content-Encoding", "gzip")
			}
			if tс.expectCompressed {
				req.Header.Set("Accept-Encoding", "gzip")
				req.Header.Set("Content-Type", "application/json")
			} else {
				req.Header.Set("Content-Type", "application/json")
			}

			// act
			rec := httptest.NewRecorder()

			mw := middleware.WithGzip(h.ShortenHandler)
			mw.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read error: %v", err)
			}

			// assert
			if tс.expectCompressed {
				if got := resp.Header.Get("Content-Encoding"); got != "gzip" {
					t.Errorf("gzip expected, а получено: %q", got)
				}

				decoded := gunzipData(t, respBody)

				var result api.ShortenResponse
				if err := json.Unmarshal([]byte(decoded), &result); err != nil {
					t.Fatalf("JSON decode error: %v", err)
				}
			} else {
				if resp.Header.Get("Content-Encoding") == "gzip" {
					t.Errorf("unexpected gzip")
				}

				var result api.ShortenResponse
				if err := json.Unmarshal(respBody, &result); err != nil {
					t.Fatalf("JSON decode error: %v", err)
				}
			}
		})
	}
}
