package storage_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository/storage"
)

func setupMockDB(t *testing.T) (*storage.DBStorage, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbStorage := &storage.DBStorage{DB: sqlxDB}

	cleanup := func() {
		db.Close()
	}

	return dbStorage, mock, cleanup
}

func TestDBStorage_Store(t *testing.T) {
	tests := []struct {
		name    string
		link    model.Link
		mockFn  func(mock sqlmock.Sqlmock, link model.Link)
		wantErr bool
	}{
		{
			name: "store valid link",
			link: model.Link{
				UUID:        uuid.New(),
				ShortURL:    "http://localhost:8080/abc123",
				OriginalURL: "https://example.com",
				UserID:      "user1",
			},
			mockFn: func(mock sqlmock.Sqlmock, link model.Link) {
				mock.ExpectExec("INSERT INTO links").
					WithArgs(link.UUID, link.ShortURL, link.OriginalURL, link.UserID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "store duplicate link",
			link: model.Link{
				UUID:        uuid.New(),
				ShortURL:    "http://localhost:8080/abc123",
				OriginalURL: "https://example.com",
				UserID:      "user1",
			},
			mockFn: func(mock sqlmock.Sqlmock, link model.Link) {
				mock.ExpectExec("INSERT INTO links").
					WithArgs(link.UUID, link.ShortURL, link.OriginalURL, link.UserID).
					WillReturnError(&pq.Error{Code: "23505", Constraint: "idx_links_original_url"})

				mock.ExpectQuery("SELECT (.+) FROM links WHERE original_url").
					WithArgs(link.OriginalURL).
					WillReturnRows(sqlmock.NewRows([]string{"uuid", "short_url", "original_url", "user_id"}).
						AddRow(link.UUID, link.ShortURL, link.OriginalURL, link.UserID))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			storage, mock, cleanup := setupMockDB(t)
			defer cleanup()
			tt.mockFn(mock, tt.link)
			ctx := context.Background()

			// act
			err := storage.Store(ctx, tt.link)

			// assert
			if (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestDBStorage_StoreBatch(t *testing.T) {
	tests := []struct {
		name    string
		links   []model.Link
		mockFn  func(mock sqlmock.Sqlmock, links []model.Link)
		wantErr bool
	}{
		{
			name: "store multiple links",
			links: []model.Link{
				{
					UUID:          uuid.New(),
					ShortURL:      "http://localhost:8080/link1",
					OriginalURL:   "https://example1.com",
					CorrelationID: "corr1",
					UserID:        "user1",
				},
				{
					UUID:          uuid.New(),
					ShortURL:      "http://localhost:8080/link2",
					OriginalURL:   "https://example2.com",
					CorrelationID: "corr2",
					UserID:        "user1",
				},
			},
			mockFn: func(mock sqlmock.Sqlmock, links []model.Link) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO links")
				for _, link := range links {
					mock.ExpectExec("INSERT INTO links").
						WithArgs(link.UUID, link.ShortURL, link.OriginalURL, link.CorrelationID, link.UserID).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:  "store empty batch",
			links: []model.Link{},
			mockFn: func(mock sqlmock.Sqlmock, links []model.Link) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO links")
				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			storage, mock, cleanup := setupMockDB(t)
			defer cleanup()
			tt.mockFn(mock, tt.links)
			ctx := context.Background()

			// act
			err := storage.StoreBatch(ctx, tt.links)

			// assert
			if (err != nil) != tt.wantErr {
				t.Errorf("StoreBatch() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestDBStorage_GetByShortURL(t *testing.T) {
	tests := []struct {
		name     string
		shortURL string
		mockFn   func(mock sqlmock.Sqlmock, shortURL string)
		wantErr  bool
	}{
		{
			name:     "get existing link",
			shortURL: "http://localhost:8080/abc123",
			mockFn: func(mock sqlmock.Sqlmock, shortURL string) {
				rows := sqlmock.NewRows([]string{"uuid", "short_url", "original_url", "user_id", "is_deleted"}).
					AddRow(uuid.New(), shortURL, "https://example.com", "user1", false)
				mock.ExpectQuery("SELECT (.+) FROM links WHERE short_url LIKE").
					WithArgs(shortURL).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:     "get non-existent link",
			shortURL: "http://localhost:8080/notfound",
			mockFn: func(mock sqlmock.Sqlmock, shortURL string) {
				mock.ExpectQuery("SELECT (.+) FROM links WHERE short_url LIKE").
					WithArgs(shortURL).
					WillReturnError(sqlmock.ErrCancelled)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			storage, mock, cleanup := setupMockDB(t)
			defer cleanup()
			tt.mockFn(mock, tt.shortURL)
			ctx := context.Background()

			// act
			link, err := storage.GetByShortURL(ctx, tt.shortURL)

			// assert
			if tt.wantErr {
				if err == nil {
					t.Error("GetByShortURL() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GetByShortURL() unexpected error: %v", err)
				}
				if link == nil {
					t.Error("GetByShortURL() returned nil link")
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestDBStorage_GetByOriginalURL(t *testing.T) {
	tests := []struct {
		name        string
		originalURL string
		mockFn      func(mock sqlmock.Sqlmock, originalURL string)
		wantErr     bool
	}{
		{
			name:        "get existing link",
			originalURL: "https://example.com",
			mockFn: func(mock sqlmock.Sqlmock, originalURL string) {
				rows := sqlmock.NewRows([]string{"uuid", "short_url", "original_url", "user_id", "is_deleted"}).
					AddRow(uuid.New(), "http://localhost:8080/abc123", originalURL, "user1", false)
				mock.ExpectQuery("SELECT (.+) FROM links WHERE original_url").
					WithArgs(originalURL).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			storage, mock, cleanup := setupMockDB(t)
			defer cleanup()
			tt.mockFn(mock, tt.originalURL)
			ctx := context.Background()

			// act
			link, err := storage.GetByOriginalURL(ctx, tt.originalURL)

			// assert
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByOriginalURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && link == nil {
				t.Error("GetByOriginalURL() returned nil link")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestDBStorage_GetByUserID(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		mockFn    func(mock sqlmock.Sqlmock, userID string)
		wantCount int
		wantErr   bool
	}{
		{
			name:   "get links for user with multiple links",
			userID: "user1",
			mockFn: func(mock sqlmock.Sqlmock, userID string) {
				rows := sqlmock.NewRows([]string{"uuid", "short_url", "original_url", "user_id", "is_deleted"}).
					AddRow(uuid.New(), "http://localhost:8080/link1", "https://example1.com", userID, false).
					AddRow(uuid.New(), "http://localhost:8080/link2", "https://example2.com", userID, false)
				mock.ExpectQuery("SELECT (.+) FROM links WHERE user_id").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:   "get links for user with no links",
			userID: "user1",
			mockFn: func(mock sqlmock.Sqlmock, userID string) {
				rows := sqlmock.NewRows([]string{"uuid", "short_url", "original_url", "user_id", "is_deleted"})
				mock.ExpectQuery("SELECT (.+) FROM links WHERE user_id").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			storage, mock, cleanup := setupMockDB(t)
			defer cleanup()
			tt.mockFn(mock, tt.userID)
			ctx := context.Background()

			// act
			links, err := storage.GetByUserID(ctx, tt.userID)

			// assert
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByUserID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(links) != tt.wantCount {
				t.Errorf("GetByUserID() got %d links, want %d", len(links), tt.wantCount)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestDBStorage_DeleteUserLinksByID(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		links   []string
		mockFn  func(mock sqlmock.Sqlmock, userID string, links []string)
		wantErr bool
		errType error
	}{
		{
			name:   "delete user links",
			userID: "user1",
			links:  []string{"http://localhost:8080/link1", "http://localhost:8080/link2"},
			mockFn: func(mock sqlmock.Sqlmock, userID string, links []string) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE links SET is_deleted")
				mock.ExpectExec("UPDATE links SET is_deleted").
					WithArgs(pq.Array(links), userID).
					WillReturnResult(sqlmock.NewResult(0, 2))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:    "delete without userID in context",
			userID:  "",
			links:   []string{"http://localhost:8080/link1"},
			mockFn:  func(mock sqlmock.Sqlmock, userID string, links []string) {},
			wantErr: true,
			errType: model.ErrUnknownUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			storage, mock, cleanup := setupMockDB(t)
			defer cleanup()
			tt.mockFn(mock, tt.userID, tt.links)

			var ctx context.Context
			if tt.userID != "" {
				ctx = context.WithValue(context.Background(), model.UserIDKey, tt.userID)
			} else {
				ctx = context.Background()
			}

			// act
			err := storage.DeleteUserLinksByID(ctx, tt.links)

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
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestDBStorage_Ping(t *testing.T) {
	// arrange
	storage, mock, cleanup := setupMockDB(t)
	defer cleanup()
	mock.ExpectPing()

	// act
	err := storage.Ping()

	// assert
	if err != nil {
		t.Errorf("Ping() unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDBStorage_Close(t *testing.T) {
	// arrange
	storage, mock, cleanup := setupMockDB(t)
	defer cleanup()
	mock.ExpectClose()

	// act
	err := storage.Close()

	// assert
	if err != nil {
		t.Errorf("Close() unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
