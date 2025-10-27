package main

import (
	"context"

	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/server"
	"github.com/mrhyman/shortner/internal/service"
)

func main() {
	cfg := config.Load()
	store := repository.NewLocalStore()
	repo := repository.NewLocalURLRepository(store)
	svc := service.NewURLService(cfg.BaseURL, repo)
	h := handler.New(*svc)
	s := server.New(cfg.ServerAddress, *h)

	s.Start(context.Background())
}
