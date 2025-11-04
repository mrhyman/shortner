package model

import (
	"github.com/google/uuid"
)

type Link struct {
	UUID        uuid.UUID
	ShortURL    string
	OriginalURL string
}

func NewLink(
	linkID uuid.UUID, originalURL string, shortURL string,
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
		UUID:        linkID,
		OriginalURL: originalURL,
		ShortURL:    shortURL,
	}

	return link, nil
}
