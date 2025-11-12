package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository/storage"
)

type URLRepository interface {
	GetByID(ctx context.Context, id string) (string, error)
	Store(ctx context.Context, url string, originalURL string) error
	Ping(ctx context.Context) error
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
	link, err := model.NewLink( uuid.New(), originalURL, shortURL )
	if err != nil {
		return err
	}
	return r.storage.Store(*link)
}

func (r *URLRepo) Ping(ctx context.Context) error {
	return r.storage.Ping()
}


