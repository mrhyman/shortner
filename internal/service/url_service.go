// Package service реализует бизнес-логику для сервиса сокращения URL.
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/api"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository"
)

// URLService предоставляет методы для работы с сокращенными URL.
type URLService struct {
	base string
	repo repository.URLRepository
}

// NewURLService создает новый экземпляр URLService.
func NewURLService(base string, repo repository.URLRepository) *URLService {
	return &URLService{base: base, repo: repo}
}

// GenerateShortID генерирует случайный короткий идентификатор для URL.
func GenerateShortID() (string, error) {
	const length = 8
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// Expand раскрывает сокращенный URL в оригинальный URL.
// Возвращает оригинальный URL или ошибку, если сокращенный URL не найден или недействителен.
func (s *URLService) Expand(ctx context.Context, shortURL string) (string, error) {
	if shortURL == "" {
		return "", model.ErrInvalidLinkID
	}

	base, err := url.Parse(s.base)
	if err != nil {
		return "", model.ErrInvalidURL
	}

	short, err := url.Parse(shortURL)
	if err != nil {
		return "", model.ErrInvalidURL
	}

	rr := base.ResolveReference(short).String()

	link, err := s.repo.GetByShortURL(ctx, rr)
	if err != nil {
		return "", model.ErrNotFound
	}

	if link.IsDeleted {
		return "", model.ErrLinkIsGone
	}

	return link.OriginalURL, nil
}

// Shorten сокращает оригинальный URL в короткий URL.
// Возвращает сокращенный URL или ошибку, если возникла проблема при сокращении.
func (s *URLService) Shorten(ctx context.Context, originalURL string) (string, error) {
	originalURL = strings.TrimSpace(originalURL)
	if originalURL == "" {
		return "", model.ErrInvalidURL
	}

	parsedURL, err := url.ParseRequestURI(originalURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return "", model.ErrInvalidURL
	}

	shortID, err := GenerateShortID()
	if err != nil {
		return "", model.ErrShortLinkGeneration
	}

	shortURL := s.base + "/" + shortID

	if err := s.repo.Store(ctx, shortURL, originalURL); err != nil {
		var existsErr *model.AlreadyExistsError
		if errors.As(err, &existsErr) {
			return "", err
		}
		return "", model.ErrShortenError
	}

	return shortURL, nil
}

// ShortenBatch сокращает множество URL в пакетном режиме.
// Возвращает массив сокращенных URL или ошибку, если возникла проблема при сокращении.
func (s *URLService) ShortenBatch(ctx context.Context, batch []api.ShortenBatchRequest) ([]api.ShortenBatchResponse, error) {
	if len(batch) == 0 {
		return []api.ShortenBatchResponse{}, nil
	}

	// Получаем userID из контекста
	userID := ""
	if uid, ok := ctx.Value(model.UserIDKey).(string); ok {
		userID = uid
	}

	links := make([]model.Link, 0, len(batch))
	response := make([]api.ShortenBatchResponse, 0, len(batch))

	for _, item := range batch {
		// Валидация URL
		originalURL := strings.TrimSpace(item.OriginalURL)
		if originalURL == "" {
			return nil, model.ErrInvalidURL
		}

		parsedURL, err := url.ParseRequestURI(originalURL)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			return nil, model.ErrInvalidURL
		}

		// Генерируем короткий ID
		shortID, err := GenerateShortID()
		if err != nil {
			return nil, model.ErrShortLinkGeneration
		}

		shortURL := s.base + "/" + shortID

		links = append(links, model.Link{
			UUID:        uuid.New(),
			ShortURL:    shortURL,
			OriginalURL: originalURL,
			UserID:      userID,
		})

		response = append(response, api.ShortenBatchResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	// Сохраняем batch
	if err := s.repo.StoreBatch(ctx, links); err != nil {
		return nil, model.ErrShortenError
	}

	return response, nil
}

// GetUserLinks получает все ссылки пользователя по его идентификатору.
// Возвращает массив ссылок пользователя или ошибку, если возникла проблема при получении.
func (s *URLService) GetUserLinks(ctx context.Context, userID string) ([]model.Link, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// DeleteUserLinksByID асинхронно удаляет ссылки пользователя по их идентификаторам.
// Принимает массив идентификаторов ссылок для удаления.
func (s *URLService) DeleteUserLinksByID(ctx context.Context, links []string) {
	jobs := make(chan []string)
	const batchSize = 100
	const workers = 4

	ctx = context.WithoutCancel(ctx)
	log := logger.Get()

	formatedLinks := make([]string, len(links))
	for i, l := range links {
		formatedLinks[i] = fmt.Sprintf("%s/%s", s.base, l)
	}

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for batch := range jobs {
				if err := s.repo.DeleteUserLinksByID(ctx, batch); err != nil {
					log.With("err", err.Error()).Warn()
				}
			}
		}()
	}

	for start := 0; start < len(formatedLinks); start += batchSize {
		end := min(start+batchSize, len(formatedLinks))
		jobs <- formatedLinks[start:end]
	}

	close(jobs)
	wg.Wait()
}

// Ping проверяет работоспособность сервиса и его зависимостей.
// Возвращает nil, если сервис работает корректно, или ошибку в противном случае.
func (s *URLService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
