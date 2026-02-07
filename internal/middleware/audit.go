// Package middleware реализует промежуточное ПО для обработки HTTP запросов.
package middleware

import (
	"net/http"
	"time"

	"github.com/mrhyman/shortner/internal/observer"
)

// WithAudit это middleware для аудита действий пользователей.
// Он отправляет события в издатель (publisher) для дальнейшей обработки и наблюдения.
func WithAudit(pub *observer.Publisher) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		if pub == nil {
			return next
		}

		return func(res http.ResponseWriter, req *http.Request) {
			start := time.Now()

			next.ServeHTTP(res, req)

			userID, _ := req.Context().Value("userID").(string)

			pub.Notify(observer.Event{
				TS:     start,
				Action: observer.ActionShorten,
				UserID: userID,
				URL:    req.URL.String(),
			})
		}
	}
}
