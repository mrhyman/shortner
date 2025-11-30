package storage

import (
	"context"

	"github.com/mrhyman/shortner/internal/model"
)

type Storage interface {
	GetByShortURL(ctx context.Context, shortURL string) (*model.Link, error)
	GetByUserID(ctx context.Context, userID string) ([]model.Link, error)
	Store(ctx context.Context, link model.Link) error
	StoreBatch(ctx context.Context, links []model.Link) error
	DeleteUserLinksByID(ctx context.Context, links []string) error
	Ping() error
	Close() error
}
