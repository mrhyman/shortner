package handler

import (
	"github.com/mrhyman/shortner/internal/service"
)

type HTTPHandler struct {
	svc service.URLService
}

func New(
	svc service.URLService,

) *HTTPHandler {
	return &HTTPHandler{

		svc: svc,
	}
}
