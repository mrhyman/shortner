package handler

import (
	"context"

	"github.com/mrhyman/shortner/internal/repository"
)

type HttpHandler struct {
	Ctx  context.Context
	Repo repository.URLRepository
}

func New(
	repo repository.URLRepository,
) *HttpHandler {
	return &HttpHandler{
		Repo: repo,
	}
}
