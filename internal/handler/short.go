package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

func (h *HTTPHandler) ShortLinkHandler(res http.ResponseWriter, req *http.Request) {
	log := logger.FromContext(req.Context())

	if req.Method != http.MethodPost {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(res, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "text/plain") {
		log.With("err", model.ErrInvalidRequestHeaders.Error()).Warn()
		http.Error(res, model.ErrInvalidRequestHeaders.Error(), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.With("err", model.ErrInvalidURL.Error()).Warn()
		http.Error(res, model.ErrInvalidURL.Error(), http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(body))

	shortURL, err := h.svc.Shorten(req.Context(), originalURL)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidURL):
			log.With("err", err.Error()).Warn()
			http.Error(res, err.Error(), http.StatusBadRequest)
		case errors.Is(err, model.ErrShortLinkGeneration) || errors.Is(err, model.ErrShortenError):
			log.With("err", err.Error()).Error()
			http.Error(res, err.Error(), http.StatusInternalServerError)
		default:
			log.With("err", err.Error()).Error()
			http.Error(res, model.ErrWentWrong.Error(), http.StatusInternalServerError)
		}
		return
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, shortURL)
}
