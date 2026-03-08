package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mrhyman/shortner/internal/config"
	handler "github.com/mrhyman/shortner/internal/handler/http"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/middleware/http"
	"github.com/mrhyman/shortner/internal/observer"
)

type Server struct {
	Instance *http.Server
	Config   config.AppConfig
}

func New(cfg config.AppConfig, h handler.HTTPHandler) (*Server, func() error, error) {
	pub, cleanup, err := observer.SetupObservers(cfg)
	if err != nil {
		return nil, nil, err
	}

	if cfg.EnableHTTPS {
		if err := ensureCertificates(cfg.CertFile, cfg.KeyFile); err != nil {
			return nil, nil, err
		}
	}

	return &Server{
		Instance: &http.Server{
			Addr:    cfg.ServerAddress,
			Handler: SetupMux(&h, cfg, pub),
		},
		Config: cfg,
	}, cleanup, nil
}

func (s *Server) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		if s.Config.EnableHTTPS {
			logger.Get().Infof("listening on %s with HTTPS", s.Instance.Addr)
			if err := s.Instance.ListenAndServeTLS(s.Config.CertFile, s.Config.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errChan <- err
			}
		} else {
			logger.Get().Infof("listening on %s", s.Instance.Addr)
			if err := s.Instance.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errChan <- err
			}
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	logger.Get().Info("Shutting down server...")
	return s.Instance.Shutdown(ctx)
}

func ensureCertificates(certFile, keyFile string) error {
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			return nil
		}
	}

	certDir := filepath.Dir(certFile)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA private key: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Shortner App"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour), // 1 year

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return fmt.Errorf("failed to get network interface addresses: %w", err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				template.IPAddresses = append(template.IPAddresses, ipnet.IP)
			}
		}
	}

	template.IPAddresses = append(template.IPAddresses, net.ParseIP("127.0.0.1"), net.ParseIP("::1"))

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to encode certificate to PEM: %w", err)
	}

	keyOut, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyOut.Close()

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("failed to encode private key to PEM: %w", err)
	}

	return nil
}

func SetupMux(h *handler.HTTPHandler, cfg config.AppConfig, pub *observer.Publisher) http.Handler {
	r := chi.NewRouter()
	dmw := DefaultMiddleware(cfg, pub)

	// buisness logic endpoints
	r.Post("/", dmw(h.ShortLinkHandler))
	r.Get("/{id}", dmw(h.ExpandHandler))
	r.Post("/api/shorten", dmw(h.ShortenHandler))
	r.Post("/api/shorten/batch", dmw(h.ShortenBatchHandler))
	r.Get("/api/user/urls", dmw(h.GetUserLinksHandler))
	r.Delete("/api/user/urls", dmw(h.DeleteUserLinksHandler))

	// service endpoints
	r.Get("/ping", middleware.WithLogging(h.PingHandler))

	// internal endpoints with trusted subnet check
	r.Get("/api/internal/stats", InternalMiddleware(cfg)(h.StatsHandler))

	return r
}

func DefaultMiddleware(cfg config.AppConfig, pub *observer.Publisher) func(http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.WithAudit(pub)(middleware.WithAuth(cfg.HashKey)(
			middleware.WithGzip(
				middleware.WithLogging(h),
			),
		))
	}
}

// InternalMiddleware применяет middleware для внутренних эндпоинтов
func InternalMiddleware(cfg config.AppConfig) func(http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.WithTrustedSubnet(cfg.TrustedSubnet)(
			middleware.WithLogging(h),
		)
	}
}
