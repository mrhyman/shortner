package storage

type Storage interface {
	GetByID(id string) (string, error)
	Store(id, original string) error
}
