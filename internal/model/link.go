package model

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID        uuid.UUID
	URL       string
	ShortURL  string
	Name      string
	CreatedAt time.Time
}

func NewLink(
	linkID uuid.UUID, url string, shortURL string, name string, createdAt time.Time,
) (*Link, error) {
	created := createdAt.UTC()
	if uuid.Nil == linkID {
		return nil, ErrInvalidLinkID
	}

	if url == "" {
		return nil, ErrInvalidURL
	}

	if shortURL == "" {
		return nil, ErrInvalidURL
	}

	if name == "" {
		return nil, ErrInvalidName
	}

	if time.Time.IsZero(createdAt) {
		created = timeNowFn()
	}

	link := &Link{
		ID:        linkID,
		URL:       url,
		ShortURL:  shortURL,
		Name:      name,
		CreatedAt: created,
	}

	return link, nil
}
