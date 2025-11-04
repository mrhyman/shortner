package repository

import (
	"context"

	"github.com/mrhyman/shortner/internal/repository/storage"
)

type URLRepository interface {
	GetByID(ctx context.Context, id string) (string, error)
	Store(ctx context.Context, url string, originalURL string) error
}

type URLRepo struct {
	storage storage.Storage
}

func NewURLRepository(s storage.Storage) *URLRepo {
	return &URLRepo{storage: s}
}

func (r *URLRepo) GetByID(ctx context.Context, id string) (string, error) {
	return r.storage.GetByID(id)
}

func (r *URLRepo) Store(ctx context.Context, shortURL string, originalURL string) error {
	return r.storage.Store(shortURL, originalURL)
}
