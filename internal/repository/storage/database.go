package storage

import (
	"github.com/jmoiron/sqlx"
	"github.com/mrhyman/shortner/internal/model"
)

type DBStorage struct {
	db *sqlx.DB
}

func NewDBStorage(db *sqlx.DB) (*DBStorage, error) {
	return &DBStorage{db: db}, nil
}

func (ds *DBStorage) Store(link model.Link) error {
	// TODO
	return nil
}

func (ds *DBStorage) GetByID(id string) (string, error) {
	// TODO

	return "", nil
}

func (ds *DBStorage) Ping() error {
	return ds.db.Ping()
}
