package handler

import (
	"net/http"

	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

func (h *HTTPHandler) PingHandler(res http.ResponseWriter, req *http.Request) {
	log := logger.Get()

	if req.Method != http.MethodGet {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(res, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	err := h.svc.Ping(req.Context())
	if err != nil {
		log.With("err", err.Error()).Error()
		http.Error(res, model.ErrStorageUnavailable.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("pong"))
}
