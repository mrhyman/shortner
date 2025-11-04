package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

func (h *HTTPHandler) ExpandHandler(res http.ResponseWriter, req *http.Request) {
	log := logger.FromContext(req.Context())
	
	if req.Method != http.MethodGet {
		log.With("err", model.ErrEnvParsing.Error()).Warn()
		http.Error(res, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(req.URL.Path, "/")

	originalURL, err := h.svc.Expand(req.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidLinkID):
			log.With("err", model.ErrEnvParsing.Error()).Warn()
			http.Error(res, err.Error(), http.StatusBadRequest)
		case errors.Is(err, model.ErrNotFound):
			log.With("err", model.ErrEnvParsing.Error()).Error()
			http.Error(res, err.Error(), http.StatusInternalServerError)
		default:
			log.With("err", model.ErrEnvParsing.Error()).Error()
			http.Error(res, model.ErrWentWrong.Error(), http.StatusInternalServerError)
		}
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
