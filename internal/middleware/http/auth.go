// Package middleware реализует промежуточное ПО для обработки HTTP запросов.
package middleware

import (
	"context"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/internal/auth"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

var (
	cookieName = "X-USER-ID"
	cookiePool = sync.Pool{
		New: func() interface{} {
			return &http.Cookie{
				Name:     cookieName,
				Path:     "/",
				HttpOnly: true,
			}
		},
	}
)

// WithAuth это middleware для аутентификации пользователей.
// Он проверяет наличие cookie с идентификатором пользователя и устанавливает его в контекст запроса.
// Если cookie отсутствует, создается новый идентификатор пользователя и устанавливается cookie.
func WithAuth(secret string) func(http.HandlerFunc) http.HandlerFunc {
	ce, err := auth.NewCookieEncoder(secret)
	if err != nil {
		panic("failed to create cookie encoder: " + err.Error())
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			log := logger.Get()
			var userID string

			if c, err := req.Cookie(cookieName); err == nil {
				userID, err = ce.DecodeUserID(c)
				if err != nil {
					log.With("err", err.Error()).Warn()
				} else if userID != "" {
					ctx := context.WithValue(req.Context(), model.UserIDKey, userID)
					next.ServeHTTP(res, req.WithContext(ctx))
					return
				}
			}

			userID = uuid.New().String()
			encodedValue, err := ce.EncodeUserID(userID)
			if err != nil {
				log.With("err", err.Error()).Error(model.ErrCookieEncoding.Error())
				http.Error(res, model.ErrWentWrong.Error(), http.StatusInternalServerError)
				return
			}

			cookie := cookiePool.Get().(*http.Cookie)
			cookie.Value = encodedValue
			http.SetCookie(res, cookie)

			cookie.Value = ""
			cookiePool.Put(cookie)

			ctx := context.WithValue(req.Context(), model.UserIDKey, userID)
			next.ServeHTTP(res, req.WithContext(ctx))
		}
	}
}
