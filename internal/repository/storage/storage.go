package storage

import "github.com/mrhyman/shortner/internal/model"

type Storage interface {
	GetByID(id string) (string, error)
	Store(link model.Link) error
	Ping() error
}
