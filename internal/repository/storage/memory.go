package storage

import (
	"context"
	"sync"

	"github.com/mrhyman/shortner/internal/model"
)

type MemoryStorage struct {
	mu  sync.RWMutex
	url map[string]model.Link
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		url: make(map[string]model.Link),
	}
}

func (ms *MemoryStorage) Store(ctx context.Context, link model.Link) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.url[link.ShortURL] = link
	return nil
}

func (ms *MemoryStorage) StoreBatch(ctx context.Context, ls []model.Link) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, link := range ls {
		ms.url[link.ShortURL] = link
	}
	return nil
}

func (ms *MemoryStorage) GetByShortURL(ctx context.Context, shortURL string) (*model.Link, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	v, ok := ms.url[shortURL]
	if !ok {
		return nil, model.ErrNotFound
	}

	return &v, nil
}

func (ms *MemoryStorage) Ping() error {
	return nil
}

func (ms *MemoryStorage) Close() error {
	return nil
}
