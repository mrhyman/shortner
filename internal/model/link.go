package model

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	Id        uuid.UUID
	Url       string
	ShortUrl  string
	Name      string
	CreatedAt time.Time
}

func NewLink(
	linkId uuid.UUID, url string, shortUrl string, name string, createdAt time.Time,
) (*Link, error) {
	created := createdAt.UTC()
	if uuid.Nil == linkId {
		return nil, ErrInvalidLinkId
	}

	if url == "" {
		return nil, ErrInvalidUrl
	}

	if shortUrl == "" {
		return nil, ErrInvalidUrl
	}

	if name == "" {
		return nil, ErrInvalidName
	}

	if time.Time.IsZero(createdAt) {
		created = timeNowFn()
	}

	link := &Link{
		Id:        linkId,
		Url:       url,
		ShortUrl:  shortUrl,
		Name:      name,
		CreatedAt: created,
	}

	return link, nil
}
