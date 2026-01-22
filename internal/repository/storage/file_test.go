package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository/storage"
)

func TestNewFileStorage(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() string
		wantErr bool
	}{
		{
			name: "create new file storage",
			setup: func() string {
				return filepath.Join(t.TempDir(), "test.json")
			},
			wantErr: false,
		},
		{
			name: "load existing file storage",
			setup: func() string {
				path := filepath.Join(t.TempDir(), "existing.json")
				os.WriteFile(path, []byte(`[{"uuid":"d5c7df47-3d3e-4c83-8255-cc93b1359558","short_url":"http://test","original_url":"https://example.com","user_id":"user1"}]`), 0644)
				return path
			},
			wantErr: false,
		},
		{
			name: "create storage in nested directory",
			setup: func() string {
				return filepath.Join(t.TempDir(), "nested", "dir", "test.json")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			path := tt.setup()

			// act
			fs, err := storage.NewFileStorage(path)

			// assert
			if tt.wantErr {
				if err == nil {
					t.Error("NewFileStorage() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewFileStorage() unexpected error: %v", err)
				return
			}

			if fs == nil {
				t.Error("NewFileStorage() returned nil storage")
			}
		})
	}
}

func TestFileStorage_Store(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			path := filepath.Join(t.TempDir(), "test.json")
			fs, err := storage.NewFileStorage(path)
			if err != nil {
				t.Fatalf("NewFileStorage() error: %v", err)
			}
			ctx := context.Background()

			// act
			err = fs.Store(ctx, tt.link)

			// assert
			if (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			stored, err := fs.GetByShortURL(ctx, tt.link.ShortURL)
			if err != nil {
				t.Errorf("Store() link was not saved: %v", err)
				return
			}

			if stored.ShortURL != tt.link.ShortURL {
				t.Errorf("Store() saved link mismatch, got %v, want %v", stored.ShortURL, tt.link.ShortURL)
			}

			// Проверяем, что файл существует
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Error("Store() file was not created")
			}
		})
	}
}

func TestFileStorage_StoreBatch(t *testing.T) {
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
			path := filepath.Join(t.TempDir(), "test.json")
			fs, err := storage.NewFileStorage(path)
			if err != nil {
				t.Fatalf("NewFileStorage() error: %v", err)
			}
			ctx := context.Background()

			// act
			err = fs.StoreBatch(ctx, tt.links)

			// assert
			if (err != nil) != tt.wantErr {
				t.Errorf("StoreBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, link := range tt.links {
				stored, err := fs.GetByShortURL(ctx, link.ShortURL)
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

func TestFileStorage_GetByShortURL(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(fs *storage.FileStorage)
		shortURL string
		wantErr  bool
		errType  error
	}{
		{
			name: "get existing link",
			setup: func(fs *storage.FileStorage) {
				fs.Store(context.Background(), model.Link{
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
			setup:    func(fs *storage.FileStorage) {},
			shortURL: "http://localhost:8080/notfound",
			wantErr:  true,
			errType:  model.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			path := filepath.Join(t.TempDir(), "test.json")
			fs, err := storage.NewFileStorage(path)
			if err != nil {
				t.Fatalf("NewFileStorage() error: %v", err)
			}
			ctx := context.Background()
			tt.setup(fs)

			// act
			link, err := fs.GetByShortURL(ctx, tt.shortURL)

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

func TestFileStorage_GetByUserID(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(fs *storage.FileStorage)
		userID    string
		wantCount int
	}{
		{
			name: "get links for user with multiple links",
			setup: func(fs *storage.FileStorage) {
				fs.Store(context.Background(), model.Link{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/link1",
					OriginalURL: "https://example1.com",
					UserID:      "user1",
				})
				fs.Store(context.Background(), model.Link{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/link2",
					OriginalURL: "https://example2.com",
					UserID:      "user1",
				})
				fs.Store(context.Background(), model.Link{
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
			setup:     func(fs *storage.FileStorage) {},
			userID:    "user1",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			path := filepath.Join(t.TempDir(), "test.json")
			fs, err := storage.NewFileStorage(path)
			if err != nil {
				t.Fatalf("NewFileStorage() error: %v", err)
			}
			ctx := context.Background()
			tt.setup(fs)

			// act
			links, err := fs.GetByUserID(ctx, tt.userID)

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

func TestFileStorage_DeleteUserLinksByID(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(fs *storage.FileStorage)
		userID  string
		links   []string
		wantErr bool
		errType error
	}{
		{
			name: "delete user links",
			setup: func(fs *storage.FileStorage) {
				fs.Store(context.Background(), model.Link{
					UUID:        uuid.New(),
					ShortURL:    "http://localhost:8080/link1",
					OriginalURL: "https://example1.com",
					UserID:      "user1",
					IsDeleted:   false,
				})
				fs.Store(context.Background(), model.Link{
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
			setup: func(fs *storage.FileStorage) {
				fs.Store(context.Background(), model.Link{
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
			setup:   func(fs *storage.FileStorage) {},
			userID:  "",
			links:   []string{"http://localhost:8080/link1"},
			wantErr: true,
			errType: model.ErrUnknownUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			path := filepath.Join(t.TempDir(), "test.json")
			fs, err := storage.NewFileStorage(path)
			if err != nil {
				t.Fatalf("NewFileStorage() error: %v", err)
			}
			tt.setup(fs)

			var ctx context.Context
			if tt.userID != "" {
				ctx = context.WithValue(context.Background(), model.UserIDKey, tt.userID)
			} else {
				ctx = context.Background()
			}

			// act
			err = fs.DeleteUserLinksByID(ctx, tt.links)

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
				link, err := fs.GetByShortURL(ctx, linkURL)
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

func TestFileStorage_Ping(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*storage.FileStorage, error)
		wantErr bool
	}{
		{
			name: "ping existing file",
			setup: func() (*storage.FileStorage, error) {
				path := filepath.Join(t.TempDir(), "test.json")
				return storage.NewFileStorage(path)
			},
			wantErr: false,
		},
		{
			name: "ping deleted file",
			setup: func() (*storage.FileStorage, error) {
				path := filepath.Join(t.TempDir(), "deleted.json")
				fs, err := storage.NewFileStorage(path)
				if err != nil {
					return nil, err
				}
				os.Remove(path)
				return fs, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			fs, _ := tt.setup()

			// act
			err := fs.Ping()

			// assert
			if (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileStorage_Close(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "test.json")
	fs, err := storage.NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error: %v", err)
	}

	// act
	err = fs.Close()

	// assert
	if err != nil {
		t.Errorf("Close() unexpected error: %v", err)
	}
}

func TestFileStorage_Persistence(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "test.json")

	fs1, err := storage.NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error: %v", err)
	}

	link := model.Link{
		UUID:        uuid.New(),
		ShortURL:    "http://localhost:8080/persist",
		OriginalURL: "https://example.com",
		UserID:      "user1",
	}

	// act
	err = fs1.Store(context.Background(), link)
	if err != nil {
		t.Fatalf("Store() error: %v", err)
	}

	fs2, err := storage.NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() second time error: %v", err)
	}

	// assert
	stored, err := fs2.GetByShortURL(context.Background(), link.ShortURL)
	if err != nil {
		t.Errorf("Persistence test failed: link not found after reload: %v", err)
		return
	}

	if stored.ShortURL != link.ShortURL {
		t.Errorf("Persistence test failed: got %v, want %v", stored.ShortURL, link.ShortURL)
	}
}
