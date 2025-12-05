package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/middleware"
)

type Server struct {
	Instance *http.Server
}

func New(cfg config.AppConfig, h handler.HTTPHandler) *Server {
	return &Server{
		Instance: &http.Server{
			Addr:    cfg.ServerAddress,
			Handler: SetupMux(&h, cfg),
		},
	}
}

func (s *Server) Start(ctx context.Context) {
	logger.FromContext(ctx).Infof("listening on %s", s.Instance.Addr)
	if err := s.Instance.ListenAndServe(); err != nil {
		logger.FromContext(ctx).With("err", err.Error()).Fatal()
	}
}

func SetupMux(h *handler.HTTPHandler, cfg config.AppConfig) http.Handler {
	r := chi.NewRouter()
	dmw := DefaultMiddleware(cfg)

	// buisness logic endpoints
	r.Post("/", dmw(h.ShortLinkHandler))
	r.Get("/{id}", dmw(h.ExpandHandler))
	r.Post("/api/shorten", dmw(h.ShortenHandler))
	r.Post("/api/shorten/batch", dmw(h.ShortenBatchHandler))
	r.Get("/api/user/urls", dmw(h.GetUserLinksHandler))
	r.Delete("/api/user/urls", dmw(h.DeleteUserLinksHandler))

	// service endpoints
	r.Get("/ping", middleware.WithLogging(h.PingHandler))

	return r
}

func DefaultMiddleware(cfg config.AppConfig) func(http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.WithAuth(cfg.HashKey)(
			middleware.WithGzip(
				middleware.WithLogging(h),
			),
		)
	}
}
