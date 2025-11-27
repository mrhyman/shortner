package model

import (
	"github.com/google/uuid"
)

type Link struct {
	UUID          uuid.UUID `db:"uuid"`
	ShortURL      string    `db:"short_url"`
	OriginalURL   string    `db:"original_url"`
	CorrelationID string    `db:"correlation_id"`
	UserID        string    `db:"user_id"`
}

func NewLink(
	linkID uuid.UUID, originalURL string, shortURL string, correlationID string, userID string,
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
	}

	return link, nil
}
