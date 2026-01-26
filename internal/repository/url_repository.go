// Package repository предоставляет интерфейсы и реализации для работы с хранилищем данных.
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository/storage"
)

// URLRepository определяет интерфейс для работы с сокращенными URL в хранилище данных.
type URLRepository interface {
	// GetByShortURL получает полную информацию о ссылке по её сокращенному URL.
	GetByShortURL(ctx context.Context, shortURL string) (*model.Link, error)

	// Store сохраняет новую сокращенную ссылку в хранилище.
	Store(ctx context.Context, url string, originalURL string) error

	// StoreBatch сохраняет множество сокращенных ссылок в хранилище.
	StoreBatch(ctx context.Context, links []model.Link) error

	// GetByUserID получает все ссылки пользователя по его идентификатору.
	GetByUserID(ctx context.Context, userID string) ([]model.Link, error)

	// DeleteUserLinksByID удаляет ссылки пользователя по их идентификаторам.
	DeleteUserLinksByID(ctx context.Context, links []string) error

	// Ping проверяет работоспособность хранилища данных.
	Ping(ctx context.Context) error
}

// URLRepo реализует интерфейс URLRepository, используя различные типы хранилищ.
type URLRepo struct {
	storage storage.Storage
}

// NewURLRepository создает новый экземпляр URLRepo с указанным хранилищем.
func NewURLRepository(s storage.Storage) *URLRepo {
	return &URLRepo{storage: s}
}

// GetByShortURL получает полную информацию о ссылке по её сокращенному URL.
func (r *URLRepo) GetByShortURL(ctx context.Context, shortURL string) (*model.Link, error) {
	return r.storage.GetByShortURL(ctx, shortURL)
}

// GetByUserID получает все ссылки пользователя по его идентификатору.
func (r *URLRepo) GetByUserID(ctx context.Context, userID string) ([]model.Link, error) {
	return r.storage.GetByUserID(ctx, userID)
}

// DeleteUserLinksByID удаляет ссылки пользователя по их идентификаторам.
func (r *URLRepo) DeleteUserLinksByID(ctx context.Context, links []string) error {
	return r.storage.DeleteUserLinksByID(ctx, links)
}

// Store сохраняет новую сокращенную ссылку в хранилище.
func (r *URLRepo) Store(ctx context.Context, shortURL string, originalURL string) error {
	userID, _ := ctx.Value(model.UserIDKey).(string)

	link, err := model.NewLink(
		uuid.New(),
		originalURL,
		shortURL,
		"",
		userID,
		false,
	)
	if err != nil {
		return err
	}
	return r.storage.Store(ctx, *link)
}

// StoreBatch сохраняет множество сокращенных ссылок в хранилище.
func (r *URLRepo) StoreBatch(ctx context.Context, links []model.Link) error {
	return r.storage.StoreBatch(ctx, links)
}

// Ping проверяет работоспособность хранилища данных.
func (r *URLRepo) Ping(ctx context.Context) error {
	return r.storage.Ping()
}
