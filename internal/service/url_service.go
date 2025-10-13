package service

import (
	"context"
	"crypto/rand"
	"net/url"

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

func (s *URLService) Expand(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", model.ErrInvalidLinkID
	}

	originalURL, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", model.ErrNotFound
	}

	return originalURL, nil
}

func (s *URLService) Shorten(ctx context.Context, originalURL string) (string, error) {
	if originalURL == "" {
		return "", model.ErrInvalidURL
	}

	id, err := generateShortID(8)
	if err != nil {
		return "", model.ErrShortLinkGeneration
	}

	if err := s.repo.Store(ctx, id, originalURL); err != nil {
		return "", model.ErrShortenError
	}

	base, err := url.Parse(s.base)
	if err != nil {
		return "", model.ErrInvalidURL
	}

	short, err := url.Parse(id)
	if err != nil {
		return "", model.ErrInvalidURL
	}

	return base.ResolveReference(short).String(), nil
}
