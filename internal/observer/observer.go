package observer

import (
	"sync"
	"time"
)

type Action string

const (
	ActionShorten Action = "shorten"
	ActionFollow  Action = "follow"
)

type Event struct {
	TS     time.Time `json:"ts"`
	Action Action    `json:"action"`
	UserID string    `json:"user_id"`
	URL    string    `json:"url"`
}

type Observer interface {
	OnRequest(event Event)
}

type Publisher struct {
	mu        sync.RWMutex
	observers []Observer
}

func NewPublisher() *Publisher {
	return &Publisher{
		observers: make([]Observer, 0),
	}
}

func (s *Publisher) Subscribe(o Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, o)
}

func (s *Publisher) Unsubscribe(o Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, obs := range s.observers {
		if obs == o {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			return
		}
	}
}

func (s *Publisher) Notify(event Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, o := range s.observers {
		go o.OnRequest(event)
	}
}
