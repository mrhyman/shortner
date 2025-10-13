package repository

import (
	"sync"

	"github.com/mrhyman/shortner/internal/model"
)

type LocalStore struct {
	mu  sync.RWMutex
	url map[string]string
}

func NewLocalStore() *LocalStore {
	return &LocalStore{
		url: make(map[string]string),
	}
}

func (s *LocalStore) Store(id, original string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.url[id] = original
	return nil
}

func (s *LocalStore) GetByID(id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.url[id]
	if !ok {
		return "", model.ErrNotFound
	}

	return v, nil
}
