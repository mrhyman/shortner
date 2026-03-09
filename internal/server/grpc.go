package server

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"

	"github.com/mrhyman/shortner/internal/auth"
	"github.com/mrhyman/shortner/internal/config"
	grpcHandler "github.com/mrhyman/shortner/internal/handler/grpc"
	"github.com/mrhyman/shortner/internal/logger"
	grpcMiddleware "github.com/mrhyman/shortner/internal/middleware/grpc"
	"github.com/mrhyman/shortner/internal/service"
	pb "github.com/mrhyman/shortner/proto"
)

type GRPCServer struct {
	Instance *grpc.Server
	Listener net.Listener
	Config   config.AppConfig
}

func NewGRPC(cfg config.AppConfig, svc *service.URLService) (*GRPCServer, error) {
	grpcAddr := cfg.GRPCAddress

	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC listener: %w", err)
	}

	authService := auth.NewAuthService(cfg.HashKey)
	interceptors := setupGRPCInterceptors(authService)

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptors.unary...),
		grpc.ChainStreamInterceptor(interceptors.stream...),
	}

	if cfg.EnableHTTPS {
		if err := ensureCertificates(cfg.CertFile, cfg.KeyFile); err != nil {
			return nil, fmt.Errorf("failed to ensure certificates: %w", err)
		}

		creds, err := credentials.NewServerTLSFromFile(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	grpcServer := grpc.NewServer(opts...)

	shortenerHandler := grpcHandler.NewShortenerHandler(svc)
	pb.RegisterShortenerServiceServer(grpcServer, shortenerHandler)

	return &GRPCServer{
		Instance: grpcServer,
		Listener: listener,
		Config:   cfg,
	}, nil
}

func (s *GRPCServer) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		protocol := "gRPC"
		if s.Config.EnableHTTPS {
			protocol = "gRPC with TLS"
		}
		logger.Get().Infof("listening on %s with %s", s.Listener.Addr().String(), protocol)

		if err := s.Instance.Serve(s.Listener); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *GRPCServer) Shutdown(ctx context.Context) error {
	logger.Get().Info("Gracefully stopping gRPC server...")

	stopped := make(chan struct{})
	go func() {
		s.Instance.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		return nil
	case <-ctx.Done():
		s.Instance.Stop() // Force stop
		return ctx.Err()
	}
}

type grpcInterceptors struct {
	unary  []grpc.UnaryServerInterceptor
	stream []grpc.StreamServerInterceptor
}

func setupGRPCInterceptors(authService grpcMiddleware.AuthService) *grpcInterceptors {
	log := logger.Get()

	recoveryOpts := []grpcrecovery.Option{
		grpcrecovery.WithRecoveryHandler(func(p any) error {
			log.With("panic", p).Error("gRPC panic recovered")
			return status.Errorf(codes.Internal, "internal server error")
		}),
	}
	loggingInterceptor := grpcMiddleware.NewLoggingInterceptor()
	authInterceptor := grpcMiddleware.NewAuthInterceptor(authService)

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		grpcrecovery.UnaryServerInterceptor(recoveryOpts...),
		loggingInterceptor.Unary(),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		grpcrecovery.StreamServerInterceptor(recoveryOpts...),
		loggingInterceptor.Stream(),
	}

	unaryInterceptors = append(unaryInterceptors, authInterceptor.Unary())
	streamInterceptors = append(streamInterceptors, authInterceptor.Stream())

	return &grpcInterceptors{
		unary:  unaryInterceptors,
		stream: streamInterceptors,
	}
}
