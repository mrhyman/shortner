package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"net/http"
	_ "net/http/pprof"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"

	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler/http"
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

	store := initStorage(cfg)
	defer store.Close()

	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(cfg.BaseURL, repo)
	h := handler.New(*svc)

	// HTTP server
	httpServer, cleanup, err := server.New(cfg, *h)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	// gRPC server
	grpcServer, err := server.NewGRPC(cfg, svc)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	// pprof server
	g.Go(func() error {
		pprofServer := &http.Server{Addr: ":9090"}
		fmt.Println("pprof server started on :9090")

		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
			defer cancel()
			if err := pprofServer.Shutdown(shutdownCtx); err != nil {
				logger.Get().With("err", err.Error()).Error("pprof server shutdown error")
			}
		}()

		if err := pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	// HTTP server
	g.Go(func() error {
		return httpServer.Start(ctx)
	})

	// gRPC server
	g.Go(func() error {
		return grpcServer.Start(ctx)
	})

	logger.Get().Info("Application started")

	if err := g.Wait(); err != nil {
		logger.Get().With("err", err.Error()).Error("Server error")
	}

	logger.Get().Info("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Get().With("err", err.Error()).Error("HTTP shutdown error")
	}

	// Shutdown gRPC server
	if err := grpcServer.Shutdown(shutdownCtx); err != nil {
		logger.Get().With("err", err.Error()).Error("gRPC shutdown error")
	}

	logger.Get().Info("Application stopped")
}

func initStorage(cfg config.AppConfig) storage.Storage {
	var store storage.Storage
	var err error

	log := logger.Get()

	switch cfg.StorageMode {
	case config.StorageDB:
		store, err = storage.NewDBStorage(cfg.DBDSN)
		dbStore, ok := store.(*storage.DBStorage)
		if !ok {
			log.Fatal("failed to cast storage to *DBStorage")
		}
		if err != nil {
			log.With("err", err.Error()).Fatal()
		}

		err = dbStore.MigrateUp("migrations", cfg.DBDSN)
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
