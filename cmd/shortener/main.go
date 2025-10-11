package main

import (
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/server"
)

const (
	port = "8080"
)

func main() {
	store := repository.NewLocalStore()
	repo := repository.NewLocalURLRepository(store)
	h := handler.New(repo)
	s := server.New(port, *h)

	s.Start()
}
