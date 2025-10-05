package handler

type Store interface {
	Set(id, original string)
	Get(id string) (string, bool)
}

type Handler struct {
	Store Store
	Base  string
}
