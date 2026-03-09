package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithTrustedSubnet(t *testing.T) {
	// простая конечная handler-функция для проверки доступа
	nextHandler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("Access granted"))
	})

	cases := []struct {
		name           string
		trustedSubnet  string
		realIP         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "No trusted subnet configured -> forbidden",
			trustedSubnet:  "",
			realIP:         "192.168.1.1",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
		{
			name:           "Invalid CIDR -> forbidden",
			trustedSubnet:  "invalid-cidr",
			realIP:         "192.168.1.1",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
		{
			name:           "No X-Real-IP header -> forbidden",
			trustedSubnet:  "192.168.0.0/16",
			realIP:         "",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
		{
			name:           "Invalid IP format -> forbidden",
			trustedSubnet:  "192.168.0.0/16",
			realIP:         "not-an-ip",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
		{
			name:           "IP in trusted subnet -> allowed",
			trustedSubnet:  "192.168.0.0/16",
			realIP:         "192.168.1.1",
			expectedStatus: http.StatusOK,
			expectedBody:   "Access granted",
		},
		{
			name:           "IP NOT in trusted subnet -> forbidden",
			trustedSubnet:  "192.168.0.0/16",
			realIP:         "10.0.0.1",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
		{
			name:           "Localhost in /8 subnet -> allowed",
			trustedSubnet:  "127.0.0.0/8",
			realIP:         "127.0.0.1",
			expectedStatus: http.StatusOK,
			expectedBody:   "Access granted",
		},
		{
			name:           "Edge of subnet boundary -> allowed",
			trustedSubnet:  "192.168.0.0/24",
			realIP:         "192.168.0.255",
			expectedStatus: http.StatusOK,
			expectedBody:   "Access granted",
		},
		{
			name:           "Just outside subnet boundary -> forbidden",
			trustedSubnet:  "192.168.0.0/24",
			realIP:         "192.168.1.0",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
		{
			name:           "IPv6 in trusted subnet -> allowed",
			trustedSubnet:  "2001:db8::/32",
			realIP:         "2001:db8::1",
			expectedStatus: http.StatusOK,
			expectedBody:   "Access granted",
		},
		{
			name:           "IPv6 NOT in trusted subnet -> forbidden",
			trustedSubnet:  "2001:db8::/32",
			realIP:         "2001:db9::1",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)

			if tc.realIP != "" {
				req.Header.Set("X-Real-IP", tc.realIP)
			}

			rec := httptest.NewRecorder()

			handlerFunc := WithTrustedSubnet(tc.trustedSubnet)(nextHandler)

			handlerFunc(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedStatus {
				t.Fatalf("[%s] expected status %d, got %d", tc.name, tc.expectedStatus, res.StatusCode)
			}

			body := rec.Body.String()
			if body != tc.expectedBody {
				t.Fatalf("[%s] expected body %q, got %q", tc.name, tc.expectedBody, body)
			}
		})
	}
}
