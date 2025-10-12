package main

import (
	"flag"

	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/server"
	"github.com/mrhyman/shortner/internal/config"
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
	h := handler.New(cfg.BaseShortURL, repo)
	s := server.New(cfg.BaseURL, *h)

	s.Start()
}
