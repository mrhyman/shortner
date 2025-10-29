package main

import (
	"context"

	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/server"
	"github.com/mrhyman/shortner/internal/service"
	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

func main() {
	ctx := context.Background()
	l := logger.New()
	logger.WithinContext(ctx, l)

	defer l.Sync()

	cfg := config.Load(ctx)
	store := repository.NewLocalStore()
	repo := repository.NewLocalURLRepository(store)
	svc := service.NewURLService(cfg.BaseURL, repo)
	h := handler.New(*svc)
	s := server.New(cfg.ServerAddress, *h)

	s.Start(ctx)
}
