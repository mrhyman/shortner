package storage

import (
	"context"
	"sync"

	"github.com/mrhyman/shortner/internal/model"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	link map[string]model.Link
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		link: make(map[string]model.Link),
	}
}

func (ms *MemoryStorage) Store(ctx context.Context, link model.Link) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.link[link.ShortURL] = link
	return nil
}

func (ms *MemoryStorage) StoreBatch(ctx context.Context, ls []model.Link) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, link := range ls {
		ms.link[link.ShortURL] = link
	}
	return nil
}

func (ms *MemoryStorage) GetByShortURL(ctx context.Context, shortURL string) (*model.Link, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	v, ok := ms.link[shortURL]
	if !ok {
		return nil, model.ErrNotFound
	}

	return &v, nil
}

func (ms *MemoryStorage) GetByUserID(ctx context.Context, userID string) ([]model.Link, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var links []model.Link

	for _, link := range ms.link {
		if link.UserID == userID {
			links = append(links, link)
		}
	}

	return links, nil
}

func (ms *MemoryStorage) DeleteUserLinksByID(ctx context.Context, links []string) error {
	userID := ctx.Value(model.UserIDKey).(string)
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, l := range links {
		link, ok := ms.link[l]
		if !ok {
			continue
		}

		if link.UserID == userID {
			link.IsDeleted = true
			ms.link[l] = link
		}
	}

	return nil
}

func (ms *MemoryStorage) Ping() error {
	return nil
}

func (ms *MemoryStorage) Close() error {
	return nil
}
