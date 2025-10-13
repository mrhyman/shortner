package server

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mrhyman/shortner/internal/handler"
)

type Server struct {
	Instance *http.Server
}

func New(baseURL string, h handler.HTTPHandler) *Server {
	return &Server{
		Instance: &http.Server{
			Addr:    baseURL,
			Handler: SetupMux(&h),
		},
	}
}

func (s *Server) Start(ctx context.Context) {
	log.Printf("listening on %s", s.Instance.Addr)
	if err := s.Instance.ListenAndServe(); err != nil {
		slog.ErrorContext(ctx, "server start error", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func SetupMux(h *handler.HTTPHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))

	r.Post("/", h.ShortLinkHandler)
	r.Get("/{id}", h.ExpandHandler)

	return r
}
