package storage

import (
	"sync"

	"github.com/mrhyman/shortner/internal/model"
)

type MemoryStorage struct {
	mu  sync.RWMutex
	url map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		url: make(map[string]string),
	}
}

func (s *MemoryStorage) Store(link model.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.url[link.ShortURL] = link.OriginalURL
	return nil
}

func (s *MemoryStorage) GetByID(id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.url[id]
	if !ok {
		return "", model.ErrNotFound
	}

	return v, nil
}

func (s *MemoryStorage) Ping() error {
	return nil
}
