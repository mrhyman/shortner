package main

import (
	"context"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

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

	store := initStorage(ctx, cfg)
	defer store.Close()

	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(cfg.BaseURL, repo)
	h := handler.New(*svc)
	s := server.New(cfg.ServerAddress, *h)

	s.Start(ctx)
}

func initStorage(ctx context.Context, cfg config.AppConfig) storage.Storage {
	var store storage.Storage
	var err error

	log := logger.FromContext(ctx)

	switch cfg.StorageMode {
	case config.StorageDB:
		store, err = storage.NewDBStorage(cfg.DBDSN)
		store, ok := store.(*storage.DBStorage)
		if !ok {
			log.Fatal("failed to cast storage to *DBStorage")
		}
		if err != nil {
			log.With("err", err.Error()).Fatal()
		}

		err = store.MigrateUp("migrations", cfg.DBDSN)
		if err != nil {
			log.With("err", err.Error()).Fatal()
		}
		log.Info("DB connection set. Migrations applied successfully")
	case config.StorageFile:
		store, err = storage.NewFileStorage(cfg.StoragePath)
		if err != nil {
			log.With("err", err.Error()).Fatal()
		}
		log.Info("Local file storage set and ready")
	default:
		store = storage.NewMemoryStorage()
	}

	return store
}
