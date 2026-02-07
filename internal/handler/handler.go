// Package handler реализует HTTP обработчики для сервиса сокращения URL.
package handler

import (
	"github.com/mrhyman/shortner/internal/service"
)

// HTTPHandler обрабатывает HTTP запросы для сервиса сокращения URL.
type HTTPHandler struct {
	svc service.URLService
}

// New создает новый экземпляр HTTPHandler с предоставленным URLService.
func New(
	svc service.URLService,

) *HTTPHandler {
	return &HTTPHandler{
		svc: svc,
	}
}
