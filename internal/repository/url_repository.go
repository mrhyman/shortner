package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository/storage"
)

type URLRepository interface {
	GetByShortURL(ctx context.Context, shortURL string) (*model.Link, error)
	Store(ctx context.Context, url string, originalURL string) error
	StoreBatch(ctx context.Context, links []model.Link) error
	Ping(ctx context.Context) error
}

type URLRepo struct {
	storage storage.Storage
}

func NewURLRepository(s storage.Storage) *URLRepo {
	return &URLRepo{storage: s}
}

func (r *URLRepo) GetByShortURL(ctx context.Context, shortURL string) (*model.Link, error) {
	return r.storage.GetByShortURL(ctx, shortURL)
}

func (r *URLRepo) Store(ctx context.Context, shortURL string, originalURL string) error {
	link, err := model.NewLink(uuid.New(), originalURL, shortURL, "")
	if err != nil {
		return err
	}
	return r.storage.Store(ctx, *link)
}

func (r *URLRepo) StoreBatch(ctx context.Context, links []model.Link) error {
	return r.storage.StoreBatch(ctx, links)
}

func (r *URLRepo) Ping(ctx context.Context) error {
	return r.storage.Ping()
}
