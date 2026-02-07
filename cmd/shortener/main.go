package main

import (
	"context"
	"fmt"
	"log"

	"net/http"
	_ "net/http/pprof"

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

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

	ctx := context.Background()
	if err := logger.Init(); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	cfg := config.Load(ctx)

	store := initStorage(ctx, cfg)
	defer store.Close()

	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(cfg.BaseURL, repo)
	h := handler.New(*svc)

	s, cleanup, err := server.New(cfg, *h)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	go func() {
		fmt.Println("pprof server started on :9090")
		// Если ваш основной сервер уже на :8080, используйте другой порт
		fmt.Println(http.ListenAndServe(":9090", nil))
	}()

	s.Start(ctx)
}

func initStorage(ctx context.Context, cfg config.AppConfig) storage.Storage {
	var store storage.Storage
	var err error

	log := logger.Get()

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

func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
