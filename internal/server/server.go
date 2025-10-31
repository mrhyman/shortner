package server

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/middleware"
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
	logger.FromContext(ctx).Infof("listening on %s", s.Instance.Addr)
	if err := s.Instance.ListenAndServe(); err != nil {
		logger.FromContext(ctx).With("trace", err.Error())
		os.Exit(1)
	}
}

func SetupMux(h *handler.HTTPHandler) http.Handler {
	r := chi.NewRouter()

	r.Post("/", middleware.WithLogging(h.ShortLinkHandler))
	r.Get("/{id}", middleware.WithLogging(h.ExpandHandler))
	r.Post("/api/shorten", middleware.WithLogging(h.ShortenHandler))

	return r
}
