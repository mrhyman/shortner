package storage_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository/storage"
)

func TestMemoryStorage_Store(t *testing.T) {
	tests := []struct {
		name    string
		link    model.Link
		wantErr bool
	}{
		{
			name: "store valid link",
			link: model.Link{
				UUID:        uuid.New(),
				ShortURL:    "http://localhost:8080/abc123",
				OriginalURL: "https://example.com",
				UserID:      "user1",
				IsDeleted:   false,
			},
			wantErr: false,
		},
		{
			name: "store link with empty fields",
			link: model.Link{
				UUID:     uuid.New(),
				ShortURL: "http://localhost:8080/empty",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			ms := storage.NewMemoryStorage()
			ctx := context.Background()

			// act
			err := ms.Store(ctx, tt.link)

			// assert
			if (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			stored, err := ms.GetByShortURL(ctx, tt.link.ShortURL)
			if err != nil {
				t.Errorf("Store() link was not saved: %v", err)
				return
			}

			if stored.ShortURL != tt.link.ShortURL {
				t.Errorf("Store() saved link mismatch, got %v, want %v", stored.ShortURL, tt.link.ShortURL)
			}
		})
	}
}

func TestMemoryStorage_StoreBatch(t *testing.T) {
	tests := []struct {
		name    string
		links   []model.Link
		wantErr bool
	}{
		{
			name: "store multiple links",
			links: []model.Link{
				{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/link1",
					OriginalURL: "https://example1.com",
					UserID:      "user1",
				},
				{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/link2",
					OriginalURL: "https://example2.com",
					UserID:      "user1",
				},
			},
			wantErr: false,
		},
		{
			name:    "store empty batch",
			links:   []model.Link{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			ms := storage.NewMemoryStorage()
			ctx := context.Background()

			// act
			err := ms.StoreBatch(ctx, tt.links)

			// assert
			if (err != nil) != tt.wantErr {
				t.Errorf("StoreBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, link := range tt.links {
				stored, err := ms.GetByShortURL(ctx, link.ShortURL)
				if err != nil {
					t.Errorf("StoreBatch() link %s was not saved: %v", link.ShortURL, err)
					continue
				}

				if stored.ShortURL != link.ShortURL {
					t.Errorf("StoreBatch() saved link mismatch, got %v, want %v", stored.ShortURL, link.ShortURL)
				}
			}
		})
	}
}

func TestMemoryStorage_GetByShortURL(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(ms *storage.MemoryStorage)
		shortURL string
		wantErr  bool
		errType  error
	}{
		{
			name: "get existing link",
			setup: func(ms *storage.MemoryStorage) {
				ms.Store(context.Background(), model.Link{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/exists",
					OriginalURL: "https://example.com",
					UserID:      "user1",
				})
			},
			shortURL: "http://localhost:8080/exists",
			wantErr:  false,
		},
		{
			name:     "get non-existent link",
			setup:    func(ms *storage.MemoryStorage) {},
			shortURL: "http://localhost:8080/notfound",
			wantErr:  true,
			errType:  model.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			ms := storage.NewMemoryStorage()
			ctx := context.Background()
			tt.setup(ms)

			// act
			link, err := ms.GetByShortURL(ctx, tt.shortURL)

			// assert
			if tt.wantErr {
				if err == nil {
					t.Error("GetByShortURL() expected error, got nil")
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("GetByShortURL() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Errorf("GetByShortURL() unexpected error: %v", err)
				return
			}

			if link.ShortURL != tt.shortURL {
				t.Errorf("GetByShortURL() got %v, want %v", link.ShortURL, tt.shortURL)
			}
		})
	}
}

func TestMemoryStorage_GetByUserID(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ms *storage.MemoryStorage)
		userID    string
		wantCount int
	}{
		{
			name: "get links for user with multiple links",
			setup: func(ms *storage.MemoryStorage) {
				ms.Store(context.Background(), model.Link{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/link1",
					OriginalURL: "https://example1.com",
					UserID:      "user1",
				})
				ms.Store(context.Background(), model.Link{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/link2",
					OriginalURL: "https://example2.com",
					UserID:      "user1",
				})
				ms.Store(context.Background(), model.Link{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/link3",
					OriginalURL: "https://example3.com",
					UserID:      "user2",
				})
			},
			userID:    "user1",
			wantCount: 2,
		},
		{
			name:      "get links for user with no links",
			setup:     func(ms *storage.MemoryStorage) {},
			userID:    "user1",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			ms := storage.NewMemoryStorage()
			ctx := context.Background()
			tt.setup(ms)

			// act
			links, err := ms.GetByUserID(ctx, tt.userID)

			// assert
			if err != nil {
				t.Errorf("GetByUserID() unexpected error: %v", err)
				return
			}

			if len(links) != tt.wantCount {
				t.Errorf("GetByUserID() got %d links, want %d", len(links), tt.wantCount)
			}

			for _, link := range links {
				if link.UserID != tt.userID {
					t.Errorf("GetByUserID() returned link with wrong userID: got %v, want %v", link.UserID, tt.userID)
				}
			}
		})
	}
}

func TestMemoryStorage_DeleteUserLinksByID(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ms *storage.MemoryStorage)
		userID  string
		links   []string
		wantErr bool
		errType error
	}{
		{
			name: "delete user links",
			setup: func(ms *storage.MemoryStorage) {
				ms.Store(context.Background(), model.Link{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/link1",
					OriginalURL: "https://example1.com",
					UserID:      "user1",
					IsDeleted:   false,
				})
				ms.Store(context.Background(), model.Link{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/link2",
					OriginalURL: "https://example2.com",
					UserID:      "user1",
					IsDeleted:   false,
				})
			},
			userID:  "user1",
			links:   []string{"http://localhost:8080/link1", "http://localhost:8080/link2"},
			wantErr: false,
		},
		{
			name: "delete links of another user",
			setup: func(ms *storage.MemoryStorage) {
				ms.Store(context.Background(), model.Link{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/link1",
					OriginalURL: "https://example1.com",
					UserID:      "user2",
					IsDeleted:   false,
				})
			},
			userID:  "user1",
			links:   []string{"http://localhost:8080/link1"},
			wantErr: false,
		},
		{
			name:    "delete without userID in context",
			setup:   func(ms *storage.MemoryStorage) {},
			userID:  "",
			links:   []string{"http://localhost:8080/link1"},
			wantErr: true,
			errType: model.ErrUnknownUser,
		},
		{
			name:    "delete non-existent links",
			setup:   func(ms *storage.MemoryStorage) {},
			userID:  "user1",
			links:   []string{"http://localhost:8080/notfound"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			ms := storage.NewMemoryStorage()
			tt.setup(ms)

			var ctx context.Context
			if tt.userID != "" {
				ctx = context.WithValue(context.Background(), model.UserIDKey, tt.userID)
			} else {
				ctx = context.Background()
			}

			// act
			err := ms.DeleteUserLinksByID(ctx, tt.links)

			// assert
			if tt.wantErr {
				if err == nil {
					t.Error("DeleteUserLinksByID() expected error, got nil")
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("DeleteUserLinksByID() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Errorf("DeleteUserLinksByID() unexpected error: %v", err)
				return
			}

			for _, linkURL := range tt.links {
				link, err := ms.GetByShortURL(ctx, linkURL)
				if err == model.ErrNotFound {
					continue
				}
				if err != nil {
					t.Errorf("DeleteUserLinksByID() error checking link: %v", err)
					continue
				}

				if link.UserID == tt.userID && !link.IsDeleted {
					t.Errorf("DeleteUserLinksByID() link %s was not marked as deleted", linkURL)
				}
			}
		})
	}
}

func TestMemoryStorage_Ping(t *testing.T) {
	// arrange
	ms := storage.NewMemoryStorage()

	// act
	err := ms.Ping()

	// assert
	if err != nil {
		t.Errorf("Ping() unexpected error: %v", err)
	}
}

func TestMemoryStorage_Close(t *testing.T) {
	// arrange
	ms := storage.NewMemoryStorage()

	// act
	err := ms.Close()

	// assert
	if err != nil {
		t.Errorf("Close() unexpected error: %v", err)
	}
}
