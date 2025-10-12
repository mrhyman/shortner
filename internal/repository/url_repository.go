package repository

import (
	"context"
)

type URLRepository interface {
	GetByID(ctx context.Context, id string) (string, error)
	Store(ctx context.Context, url string, originalURL string) error
}

type LocalURLRepository struct {
	store *LocalStore
	ctx   context.Context
}

func NewLocalURLRepository(s *LocalStore) URLRepository {
	return &LocalURLRepository{store: s}
}

func (r *LocalURLRepository) GetByID(ctx context.Context, id string) (string, error) {
	return r.store.GetByID(id)
}

func (r *LocalURLRepository) Store(ctx context.Context, url string, originalURL string) error {
	return r.store.Store(url, originalURL)
}
