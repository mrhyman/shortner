package storage

import (
	"context"

	"github.com/mrhyman/shortner/internal/model"
)

type Storage interface {
	GetByID(ctx context.Context, id string) (string, error)
	Store(ctx context.Context, link model.Link) error
	Ping() error
	Close() error
}
