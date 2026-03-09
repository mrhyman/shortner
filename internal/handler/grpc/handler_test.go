package grpc_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/mrhyman/shortner/internal/auth"
	"github.com/mrhyman/shortner/internal/config"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/server"
	"github.com/mrhyman/shortner/internal/service"
	grpcapi "github.com/mrhyman/shortner/proto"
)

func setupTestGRPCServer(t *testing.T) (*server.GRPCServer, *service.URLService, func()) {
	// Инициализация как в main.go
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService(config.DefaultBaseURL, repo)

	// Создаем конфиг для тестового сервера
	cfg := config.AppConfig{
		GRPCAddress: "127.0.0.1:50052", // используем другой порт для тестов
		HashKey:     "test-secret-key",
	}

	// Создаем gRPC сервер
	grpcServer, err := server.NewGRPC(cfg, svc)
	if err != nil {
		t.Fatal(err)
	}

	// Запускаем сервер в горутине
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := grpcServer.Start(ctx); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Даем серверу время запуститься
	time.Sleep(100 * time.Millisecond)

	// Cleanup функция
	cleanup := func() {
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		grpcServer.Shutdown(shutdownCtx)
		store.Close()
	}

	return grpcServer, svc, cleanup
}

func getTestClient(t *testing.T, addr string) grpcapi.ShortenerServiceClient {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		conn.Close()
	})

	return grpcapi.NewShortenerServiceClient(conn)
}

func TestGRPCHandler_ShortenURL(t *testing.T) {
	_, _, cleanup := setupTestGRPCServer(t)
	defer cleanup()

	client := getTestClient(t, "127.0.0.1:50052")
	ctx := context.Background()

	req := grpcapi.URLShortenRequest_builder{
		Url: "https://example.com",
	}.Build()

	resp, err := client.ShortenURL(ctx, req)
	if err != nil {
		t.Fatalf("ShortenURL failed: %v", err)
	}

	shortURL := resp.GetResult()
	if shortURL == "" {
		t.Error("expected non-empty short URL")
	}

	if !strings.Contains(shortURL, config.DefaultBaseURL) {
		t.Errorf("expected short URL to contain base URL %s, got %s", config.DefaultBaseURL, shortURL)
	}

	t.Logf("Short URL: %s", shortURL)
}

func TestGRPCHandler_ExpandURL(t *testing.T) {
	_, svc, cleanup := setupTestGRPCServer(t)
	defer cleanup()

	client := getTestClient(t, "127.0.0.1:50052")
	ctx := context.Background()

	// Сначала создаем короткую ссылку через сервис
	originalURL := "https://example.com"
	shortURL, err := svc.Shorten(ctx, originalURL)
	if err != nil {
		t.Fatal(err)
	}

	// Извлекаем ID
	parts := strings.Split(shortURL, "/")
	id := parts[len(parts)-1]

	// Тестируем раскрытие через gRPC
	req := grpcapi.URLExpandRequest_builder{
		Id: id,
	}.Build()

	resp, err := client.ExpandURL(ctx, req)
	if err != nil {
		t.Fatalf("ExpandURL failed: %v", err)
	}

	if resp.GetResult() != originalURL {
		t.Errorf("expected %s, got %s", originalURL, resp.GetResult())
	}

	t.Logf("Expanded URL: %s", resp.GetResult())
}

func TestGRPCHandler_ListUserURLs(t *testing.T) {
	_, svc, cleanup := setupTestGRPCServer(t)
	defer cleanup()

	client := getTestClient(t, "127.0.0.1:50052")

	// Создаем auth service для генерации токена
	authService := auth.NewAuthService("test-secret-key")
	token, err := authService.GenerateToken()
	if err != nil {
		t.Fatal(err)
	}

	// Получаем userID из токена
	userID, err := authService.ValidateToken(token)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем контекст с userID для сервиса
	ctxWithUser := context.WithValue(context.Background(), model.UserIDKey, userID)

	// Создаем несколько ссылок через сервис
	urls := []string{
		"https://example.com",
		"https://google.com",
		"https://github.com",
	}

	for _, url := range urls {
		shortURL, err := svc.Shorten(ctxWithUser, url)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Created short URL: %s", shortURL)
	}

	// Создаем контекст с authorization header для gRPC клиента
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctxWithToken := metadata.NewOutgoingContext(context.Background(), md)

	// Получаем список через gRPC
	resp, err := client.ListUserURLs(ctxWithToken, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("ListUserURLs failed: %v", err)
	}

	if len(resp.GetUrl()) != len(urls) {
		t.Errorf("expected %d URLs, got %d", len(urls), len(resp.GetUrl()))
	}

	for i, urlData := range resp.GetUrl() {
		t.Logf("%d. Short: %s, Original: %s", i+1, urlData.GetShortUrl(), urlData.GetOriginalUrl())
	}
}
