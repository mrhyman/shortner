package repository

import "sync"

type Store struct {
	mu  sync.Mutex
	url map[string]string
}

func NewStore() *Store {
	return &Store{
		url: make(map[string]string),
	}
}

func (s *Store) Set(id, original string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.url[id] = original
}

func (s *Store) Get(id string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.url[id]
	return v, ok
}
