package auth

import (
	"errors"

	"github.com/google/uuid"
)

// AuthService предоставляет методы для работы с авторизацией.
type AuthService struct {
	encoder *CookieEncoder
}

// NewAuthService создает новый экземпляр AuthService.
func NewAuthService(secret string) *AuthService {
	encoder, err := NewCookieEncoder(secret)
	if err != nil {
		panic("failed to create cookie encoder: " + err.Error())
	}

	return &AuthService{
		encoder: encoder,
	}
}

// GenerateToken генерирует новый токен для нового пользователя.
func (a *AuthService) GenerateToken() (string, error) {
	userID := uuid.New().String()
	return a.encoder.EncodeUserID(userID)
}

// ValidateToken валидирует токен и возвращает userID.
func (a *AuthService) ValidateToken(token string) (string, error) {
	if token == "" {
		return "", errors.New("empty token")
	}

	userID, err := a.encoder.DecodeValue(token)
	if err != nil {
		return "", err
	}

	if userID == "" {
		return "", errors.New("invalid token: empty userID")
	}

	return userID, nil
}
