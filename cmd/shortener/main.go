package main

import (
	"context"
	"flag"

	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/server"
	"github.com/mrhyman/shortner/internal/service"
)

var cfg config.AppConfig

func init() {
	flag.StringVar(&cfg.BaseURL, "a", "localhost:8080", "HTTP server address, e.g. localhost:8888")
	flag.StringVar(&cfg.BaseShortURL, "b", "http://localhost:8080", "Base URL for short links, e.g. http://localhost:8080/")
	flag.Parse()
}

func main() {
	store := repository.NewLocalStore()
	repo := repository.NewLocalURLRepository(store)
	svc := service.NewURLService(cfg.BaseShortURL, repo)
	h := handler.New(*svc)
	s := server.New(cfg.BaseURL, *h)

	s.Start(context.Background())
}
