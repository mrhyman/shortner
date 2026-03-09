package grpc

import (
	"context"
	"strings"

	"github.com/mrhyman/shortner/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	ValidateToken(token string) (string, error)
	GenerateToken() (string, error)
}

type AuthInterceptor struct {
	authService AuthService
}

func NewAuthInterceptor(authService AuthService) *AuthInterceptor {
	return &AuthInterceptor{
		authService: authService,
	}
}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		newCtx, err := a.authorize(ctx)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

func (a *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		newCtx, err := a.authorize(ss.Context())
		if err != nil {
			return err
		}

		wrapped := &wrappedStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		return handler(srv, wrapped)
	}
}

func (a *AuthInterceptor) authorize(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		// Нет metadata - генерируем новый токен
		return a.generateNewToken(ctx)
	}

	// Проверяем authorization header
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return a.generateNewToken(ctx)
	}

	// Валидируем существующий токен
	token := strings.TrimPrefix(authHeaders[0], "Bearer ")
	userID, err := a.authService.ValidateToken(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return context.WithValue(ctx, model.UserIDKey, userID), nil
}

func (a *AuthInterceptor) generateNewToken(ctx context.Context) (context.Context, error) {
	token, err := a.authService.GenerateToken()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	userID, err := a.authService.ValidateToken(token)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to validate token")
	}

	ctx = context.WithValue(ctx, model.UserIDKey, userID)

	header := metadata.Pairs("authorization", "Bearer "+token)
	if err := grpc.SendHeader(ctx, header); err != nil {
		return nil, status.Error(codes.Internal, "failed to send header")
	}

	return ctx, nil
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
