package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/internal/auth"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

func WithAuth(secret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			log := logger.FromContext(req.Context())

			ce, err := auth.NewCookieEncoder(secret)
			if err != nil {
				log.With("err", err.Error()).Warn()
				http.Error(res, model.ErrWentWrong.Error(), http.StatusInternalServerError)
				return
			}

			c, err := req.Cookie("X-USER-ID")

			if err == http.ErrNoCookie {
				userID := uuid.New().String()
				v, err := ce.EncodeUserID(userID)
				if err != nil {
					http.Error(res, model.ErrWentWrong.Error(), http.StatusInternalServerError)
					return
				}

				http.SetCookie(res, &http.Cookie{
					Name:     "X-USER-ID",
					Value:    v,
					Path:     "/",
					HttpOnly: true,
				})

				ctx := context.WithValue(req.Context(), model.UserIDKey, userID)
				next.ServeHTTP(res, req.WithContext(ctx))
				return
			}

			userID, err := ce.DecodeUserID(c)
			if err != nil {
				userID = uuid.New().String()
				val, err := ce.EncodeUserID(userID)
				if err != nil {
					http.Error(res, model.ErrWentWrong.Error(), http.StatusInternalServerError)
					return
				}

				http.SetCookie(res, &http.Cookie{
					Name:     "X-USER-ID",
					Value:    val,
					Path:     "/",
					HttpOnly: true,
				})

				ctx := context.WithValue(req.Context(), model.UserIDKey, userID)
				next.ServeHTTP(res, req.WithContext(ctx))
				return
			}

			if userID == "" {
				log.With("err", model.ErrUnknownUser.Error()).Warn()
				res.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(req.Context(), model.UserIDKey, userID)
			next.ServeHTTP(res, req.WithContext(ctx))
		}
	}
}
