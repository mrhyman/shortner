// Package model содержит модели данных, используемые в приложении.
package model

import (
	"github.com/google/uuid"
)

// Link представляет собой модель сокращенной ссылки.
type Link struct {
	// UUID уникальный идентификатор ссылки.
	UUID uuid.UUID `db:"uuid"`

	// ShortURL сокращенный URL.
	ShortURL string `db:"short_url"`

	// OriginalURL оригинальный URL.
	OriginalURL string `db:"original_url"`

	// CorrelationID идентификатор для корреляции запросов в пакетных операциях.
	CorrelationID string `db:"correlation_id"`

	// UserID идентификатор пользователя, которому принадлежит ссылка.
	UserID string `db:"user_id"`

	// IsDeleted флаг, указывающий, что ссылка была удалена.
	IsDeleted bool `db:"is_deleted"`
}

// NewLink создает новый экземпляр Link с указаными параметрами.
// Возвращает ошибку, если linkID равен нулю или если originalURL/shortURL пустые.
func NewLink(
	linkID uuid.UUID,
	originalURL string,
	shortURL string,
	correlationID string,
	userID string,
	isDeleted bool,
) (*Link, error) {
	if uuid.Nil == linkID {
		return nil, ErrInvalidLinkID
	}

	if originalURL == "" {
		return nil, ErrInvalidURL
	}

	if shortURL == "" {
		return nil, ErrInvalidURL
	}

	link := &Link{
		UUID:          linkID,
		OriginalURL:   originalURL,
		ShortURL:      shortURL,
		CorrelationID: correlationID,
		UserID:        userID,
		IsDeleted:     isDeleted,
	}

	return link, nil
}
