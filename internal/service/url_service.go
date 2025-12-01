package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/api"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/repository"
)

type URLService struct {
	base string
	repo repository.URLRepository
}

func NewURLService(base string, repo repository.URLRepository) *URLService {
	return &URLService{base: base, repo: repo}
}

// TODO: refactor while sprint 4
func generateShortID(n int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}

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

func (s *URLService) Shorten(ctx context.Context, originalURL string) (string, error) {
	if originalURL == "" {
		return "", model.ErrInvalidURL
	}

	id, err := generateShortID(8)
	if err != nil {
		return "", model.ErrShortLinkGeneration
	}

	base, err := url.Parse(s.base)
	if err != nil {
		return "", model.ErrInvalidURL
	}

	short, err := url.Parse(id)
	if err != nil {
		return "", model.ErrInvalidURL
	}

	rr := base.ResolveReference(short).String()

	if err := s.repo.Store(ctx, rr, originalURL); err != nil {
		return "", err
	}

	return rr, nil
}

func (s *URLService) ShortenBatch(ctx context.Context, batch []api.ShortenBatchRequest) ([]model.Link, error) {
	links := make([]model.Link, 0, len(batch))

	for _, item := range batch {
		shortID, err := generateShortID(8)
		if err != nil {
			return nil, model.ErrShortLinkGeneration
		}

		link, err := model.NewLink(
			uuid.New(),
			item.OriginalURL,
			fmt.Sprintf("%s/%s", s.base, shortID),
			item.CorrelationID,
			ctx.Value(model.UserIDKey).(string),
			false,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, *link)
	}

	if err := s.repo.StoreBatch(ctx, links); err != nil {
		return nil, model.ErrShortenError
	}

	return links, nil
}

func (s *URLService) GerUserLinks(ctx context.Context, userID string) ([]model.Link, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *URLService) DeleteUserLinksByID(ctx context.Context, links []string) {
	jobs := make(chan []string)
	const batchSize = 100
	const workers = 4

	ctx = context.WithoutCancel(ctx)
	log := logger.FromContext(ctx)

	formatedLinks := make([]string, len(links))
	for i, l := range links {
		formatedLinks[i] = fmt.Sprintf("%s/%s", s.base, l)
	}

	for range workers {
		go func() {
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
}

func (s *URLService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
