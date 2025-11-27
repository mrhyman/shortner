package middleware

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

func WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		secret := "qwerty12345"
		log := logger.FromContext(req.Context())

		c, err := req.Cookie("X-USER-ID")

		if err == http.ErrNoCookie {
			userID := uuid.New().String()
			val, err := encodeUserID(userID, secret)
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

		userID, err := decodeCookie(c, secret)
		if err != nil {
			userID = uuid.New().String()
			val, err := encodeUserID(userID, secret)
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

func decodeCookie(cookie *http.Cookie, secret string) (string, error) {
	key := sha256.Sum256([]byte(secret))
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", model.ErrCookieDecoding
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", model.ErrCookieDecoding
	}

	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		return "", model.ErrCookieDecoding
	}

	nonce := data[:aesgcm.NonceSize()]
	ciphertext := data[aesgcm.NonceSize():]

	plain, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", model.ErrCookieDecoding
	}

	return string(plain), nil
}

func encodeUserID(userID string, secret string) (string, error) {
	key := sha256.Sum256([]byte(secret))
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", model.ErrCookieEncoding
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", model.ErrCookieEncoding
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", model.ErrCookieEncoding
	}

	encrypted := aesgcm.Seal(nil, nonce, []byte(userID), nil)
	cookieValue := hex.EncodeToString(append(nonce, encrypted...))

	return cookieValue, nil
}
