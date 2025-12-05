package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrhyman/shortner/internal/auth"
	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/middleware"
	"github.com/mrhyman/shortner/internal/model"
)

func TestWithAuth(t *testing.T) {
	secret := config.DefaultHashKey
	ce, _ := auth.NewCookieEncoder(secret)

	// простая конечная handler-функция для проверки контекста
	nextHandler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		userID, ok := req.Context().Value(model.UserIDKey).(string)
		if !ok || userID == "" {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(userID))
	})

	cases := []struct {
		name            string
		setupCookie     func() *http.Cookie
		expectedStatus  int
		expectNewCookie bool
	}{
		{
			name:            "No cookie -> create new",
			setupCookie:     func() *http.Cookie { return nil },
			expectedStatus:  http.StatusOK,
			expectNewCookie: true,
		},
		{
			name: "Valid cookie -> use existing",
			setupCookie: func() *http.Cookie {
				val, _ := ce.EncodeUserID("existing-user")
				return &http.Cookie{Name: "X-USER-ID", Value: val}
			},
			expectedStatus:  http.StatusOK,
			expectNewCookie: false,
		},
		{
			name: "Invalid cookie -> create new",
			setupCookie: func() *http.Cookie {
				return &http.Cookie{Name: "X-USER-ID", Value: "nothexdata"}
			},
			expectedStatus:  http.StatusOK,
			expectNewCookie: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if c := tc.setupCookie(); c != nil {
				req.AddCookie(c)
			}
			rec := httptest.NewRecorder()

			handlerFunc := middleware.WithAuth(secret)(nextHandler)

			handlerFunc(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedStatus {
				t.Fatalf("[%s] expected status %d, got %d", tc.name, tc.expectedStatus, res.StatusCode)
			}

			cookies := res.Cookies()
			hasCookie := false
			for _, c := range cookies {
				if c.Name == "X-USER-ID" {
					hasCookie = true
				}
			}

			if hasCookie != tc.expectNewCookie {
				t.Fatalf("[%s] expected new cookie=%v, got=%v", tc.name, tc.expectNewCookie, hasCookie)
			}
		})
	}
}
