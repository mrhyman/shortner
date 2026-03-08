package grpc

import (
	"context"
	"time"

	"github.com/mrhyman/shortner/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoggingInterceptor struct{}

func NewLoggingInterceptor() *LoggingInterceptor {
	return &LoggingInterceptor{}
}

// Unary возвращает unary server interceptor для логирования
func (l *LoggingInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		log := logger.Get()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			}
		}

		log.With(
			"method", info.FullMethod,
			"duration", duration.String(),
			"code", code.String(),
		).Info("gRPC request")

		return resp, err
	}
}

// Stream возвращает stream server interceptor для логирования
func (l *LoggingInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		log := logger.Get()

		err := handler(srv, ss)

		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			}
		}

		log.With(
			"method", info.FullMethod,
			"duration", duration.String(),
			"code", code.String(),
		).Info("gRPC stream request")

		return err
	}
}
