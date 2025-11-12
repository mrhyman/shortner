package main

import (
	"context"
	"time"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/server"
	"github.com/mrhyman/shortner/internal/service"
)

func main() {
	ctx := context.Background()
	log := logger.New()
	logger.WithinContext(ctx, log)

	defer log.Sync()

	cfg := config.Load(ctx)

	db, err := sqlx.Open("postgres", cfg.DBDSN)
	if err != nil {
		log.With("err", err).Fatal()
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)

	storage, err := storage.NewFileStorage(cfg.StoragePath)
	if err != nil {
		log.With("err", err.Error()).Fatal()
	}

	repo := repository.NewURLRepository(storage)
	svc := service.NewURLService(cfg.BaseURL, repo)
	h := handler.New(*svc)
	s := server.New(cfg.ServerAddress, *h)

	s.Start(ctx)
}
