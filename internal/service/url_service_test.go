package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/api"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository"
	"github.com/mrhyman/shortner/internal/repository/storage"
	"github.com/mrhyman/shortner/internal/service"
)

func TestGenerateShortID(t *testing.T) {
	tests := []struct {
		name    string
		wantLen int
	}{
		{"generates short ID", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := service.GenerateShortID()

			if err != nil {
				t.Fatalf("GenerateShortID() unexpected error: %v", err)
			}

			if len(id) != tt.wantLen {
				t.Errorf("GenerateShortID() length = %d, want %d", len(id), tt.wantLen)
			}

			// Проверяем что содержит только hex символы
			for _, c := range id {
				if !((c >= 'a' && c <= 'f') || (c >= '0' && c <= '9')) {
					t.Errorf("GenerateShortID() contains invalid hex character: %c", c)
				}
			}
		})
	}
}

func TestGenerateShortID_Uniqueness(t *testing.T) {
	// Генерируем несколько ID и проверяем уникальность
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := service.GenerateShortID()
		if err != nil {
			t.Fatalf("GenerateShortID() error: %v", err)
		}
		if ids[id] {
			t.Errorf("GenerateShortID() generated duplicate: %s", id)
		}
		ids[id] = true
	}
}

func TestURLService_Shorten(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errType error
	}{
		{
			name:    "valid URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
			errType: model.ErrInvalidURL,
		},
		{
			name:    "invalid URL",
			url:     "not-a-url",
			wantErr: true,
			errType: model.ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService("http://localhost:8080", repo)

			shortURL, err := svc.Shorten(context.Background(), tt.url)

			if tt.wantErr {
				if err == nil {
					t.Error("Shorten() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Shorten() unexpected error: %v", err)
				return
			}

			if shortURL == "" {
				t.Error("Shorten() returned empty short URL")
			}

			if len(shortURL) < 10 { // "http://..." минимум
				t.Errorf("Shorten() returned too short URL: %s", shortURL)
			}
		})
	}
}

func TestURLService_Expand(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService("http://localhost:8080", repo)

	// Создаем короткую ссылку
	originalURL := "https://example.com/test"
	shortURL, err := svc.Shorten(ctx, originalURL)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	tests := []struct {
		name      string
		shortCode string
		wantURL   string
		wantErr   bool
	}{
		{
			name:      "existing short code",
			shortCode: shortURL[len("http://localhost:8080/"):],
			wantURL:   originalURL,
			wantErr:   false,
		},
		{
			name:      "non-existing short code",
			shortCode: "nonexistent",
			wantErr:   true,
		},
		{
			name:      "empty short code",
			shortCode: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := svc.Expand(ctx, tt.shortCode)

			if tt.wantErr {
				if err == nil {
					t.Error("Expand() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expand() unexpected error: %v", err)
				return
			}

			if url != tt.wantURL {
				t.Errorf("Expand() = %v, want %v", url, tt.wantURL)
			}
		})
	}
}

func TestURLService_ShortenBatch(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService("http://localhost:8080", repo)

	tests := []struct {
		name    string
		batch   []api.ShortenBatchRequest
		wantLen int
		wantErr bool
	}{
		{
			name: "valid batch",
			batch: []api.ShortenBatchRequest{
				{CorrelationID: "1", OriginalURL: "https://example.com/1"},
				{CorrelationID: "2", OriginalURL: "https://example.com/2"},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "empty batch",
			batch:   []api.ShortenBatchRequest{},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "batch with invalid URL",
			batch: []api.ShortenBatchRequest{
				{CorrelationID: "1", OriginalURL: "invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.ShortenBatch(ctx, tt.batch)

			if tt.wantErr {
				if err == nil {
					t.Error("ShortenBatch() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ShortenBatch() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantLen {
				t.Errorf("ShortenBatch() returned %d items, want %d", len(result), tt.wantLen)
			}

			// Проверяем что correlation IDs совпадают
			for i, item := range result {
				if item.CorrelationID != tt.batch[i].CorrelationID {
					t.Errorf("Item %d: correlation ID = %s, want %s",
						i, item.CorrelationID, tt.batch[i].CorrelationID)
				}
			}
		})
	}
}

func TestURLService_GetUserLinks(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService("http://localhost:8080", repo)

	userID := "test-user-123"

	// Создаем несколько ссылок для пользователя
	_, _ = svc.Shorten(ctx, "https://example.com/1")
	_, _ = svc.Shorten(ctx, "https://example.com/2")

	tests := []struct {
		name    string
		userID  string
		wantMin int
		wantErr bool
	}{
		{
			name:    "user with links",
			userID:  userID,
			wantMin: 0,
			wantErr: false,
		},
		{
			name:    "user without links",
			userID:  "non-existent-user",
			wantMin: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			links, err := svc.GetUserLinks(ctx, tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Error("GetUserLinks() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetUserLinks() unexpected error: %v", err)
				return
			}

			if len(links) < tt.wantMin {
				t.Errorf("GetUserLinks() returned %d links, want at least %d",
					len(links), tt.wantMin)
			}
		})
	}
}

func TestURLService_Ping(t *testing.T) {
	tests := []struct {
		name      string
		wantErr   bool
		setupRepo func() repository.URLRepository
	}{
		{
			name:    "successful ping",
			wantErr: false,
			setupRepo: func() repository.URLRepository {
				store := storage.NewMemoryStorage()
				return repository.NewURLRepository(store)
			},
		},
		{
			name:    "ping error",
			wantErr: true,
			setupRepo: func() repository.URLRepository {
				// Используйте мок или storage, который возвращает ошибку при Ping
				return &mockFailingRepository{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			svc := service.NewURLService("http://localhost:8080", repo)

			err := svc.Ping(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Error("Ping() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Ping() unexpected error: %v", err)
			}
		})
	}
}

func TestURLService_DeleteUserLinksByID(t *testing.T) {
	tests := []struct {
		name         string
		setupLinks   func(store *storage.MemoryStorage) []string
		userID       string
		checkDeleted bool
	}{
		{
			name: "delete single link",
			setupLinks: func(store *storage.MemoryStorage) []string {
				link := model.Link{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/abc12345",
					OriginalURL: "https://example.com",
					UserID:      "user1",
					IsDeleted:   false,
				}
				store.Store(context.Background(), link)
				return []string{"abc12345"}
			},
			userID:       "user1",
			checkDeleted: true,
		},
		{
			name: "delete multiple links",
			setupLinks: func(store *storage.MemoryStorage) []string {
				ids := make([]string, 3)
				for i := 0; i < 3; i++ {
					shortID := fmt.Sprintf("link%d", i)
					link := model.Link{
						UUID:        uuid.New(),
						ShortURL:    fmt.Sprintf("http://localhost:8080/%s", shortID),
						OriginalURL: fmt.Sprintf("https://example%d.com", i),
						UserID:      "user1",
						IsDeleted:   false,
					}
					store.Store(context.Background(), link)
					ids[i] = shortID
				}
				return ids
			},
			userID:       "user1",
			checkDeleted: true,
		},
		{
			name: "delete empty list",
			setupLinks: func(store *storage.MemoryStorage) []string {
				return []string{}
			},
			userID:       "user1",
			checkDeleted: false,
		},
		{
			name: "delete non-existent links",
			setupLinks: func(store *storage.MemoryStorage) []string {
				link := model.Link{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/existing",
					OriginalURL: "https://example.com",
					UserID:      "user1",
					IsDeleted:   false,
				}
				store.Store(context.Background(), link)
				return []string{"nonexistent1", "nonexistent2"}
			},
			userID:       "user1",
			checkDeleted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			store := storage.NewMemoryStorage()
			repo := repository.NewURLRepository(store)
			svc := service.NewURLService("http://localhost:8080", repo)

			ctx := context.WithValue(context.Background(), model.UserIDKey, tt.userID)
			linksToDelete := tt.setupLinks(store)

			// act
			svc.DeleteUserLinksByID(ctx, linksToDelete)

			// wait till goroutines finished
			time.Sleep(300 * time.Millisecond)

			if !tt.checkDeleted {
				return
			}

			// assert
			for _, linkID := range linksToDelete {
				shortURL := fmt.Sprintf("http://localhost:8080/%s", linkID)
				link, err := store.GetByShortURL(ctx, shortURL)

				if err != nil {
					t.Errorf("DeleteUserLinksByID() link %s not found in storage: %v", linkID, err)
					continue
				}

				if !link.IsDeleted {
					t.Errorf("DeleteUserLinksByID() link %s was not marked as deleted", linkID)
				}
			}
		})
	}
}

func TestURLService_DeleteUserLinksByID_LargeBatch(t *testing.T) {
	// arrange
	store := storage.NewMemoryStorage()
	repo := repository.NewURLRepository(store)
	svc := service.NewURLService("http://localhost:8080", repo)

	ctx := context.WithValue(context.Background(), model.UserIDKey, "user1")

	linkIDs := make([]string, 150)
	for i := 0; i < 150; i++ {
		shortID := fmt.Sprintf("link%d", i)
		link := model.Link{
			UUID:        uuid.New(),
			ShortURL:    fmt.Sprintf("http://localhost:8080/%s", shortID),
			OriginalURL: fmt.Sprintf("https://example%d.com", i),
			UserID:      "user1",
			IsDeleted:   false,
		}
		store.Store(ctx, link)
		linkIDs[i] = shortID
	}

	// act
	svc.DeleteUserLinksByID(ctx, linkIDs)
	time.Sleep(500 * time.Millisecond)

	// assert
	deletedCount := 0
	for _, linkID := range linkIDs {
		shortURL := fmt.Sprintf("http://localhost:8080/%s", linkID)
		link, err := store.GetByShortURL(ctx, shortURL)
		if err == nil && link.IsDeleted {
			deletedCount++
		}
	}

	if deletedCount != 150 {
		t.Errorf("DeleteUserLinksByID() marked %d links as deleted, expected 150", deletedCount)
	}
}

type mockFailingRepository struct {
	repository.URLRepository
}

func (m *mockFailingRepository) Ping(ctx context.Context) error {
	return errors.New("connection failed")
}
