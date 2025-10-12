package handler

import (
	"github.com/mrhyman/shortner/internal/repository"
)

type HTTPHandler struct {
	BaseShortURL string
	Repo         repository.URLRepository
}

func New(
	baseShortURL string,
	repo repository.URLRepository,

) *HTTPHandler {
	return &HTTPHandler{
		BaseShortURL: baseShortURL,
		Repo:         repo,
	}
}
