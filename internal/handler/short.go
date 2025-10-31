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
		log.With("err", model.ErrInvalidRequestParams.Error())
		http.Error(res, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "text/plain") {
		log.With("err", model.ErrInvalidRequestHeaders.Error())
		http.Error(res, model.ErrInvalidRequestHeaders.Error(), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.With("err", model.ErrInvalidURL.Error(), "trace", err.Error())
		http.Error(res, model.ErrInvalidURL.Error(), http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(body))

	shortURL, err := h.svc.Shorten(req.Context(), originalURL)
	if err != nil {
		log.With("err", model.ErrEnvParsing.Error(), "trace", err.Error())
		switch {
		case errors.Is(err, model.ErrInvalidURL):
			http.Error(res, err.Error(), http.StatusBadRequest)
		case errors.Is(err, model.ErrShortLinkGeneration) || errors.Is(err, model.ErrShortenError):
			http.Error(res, err.Error(), http.StatusInternalServerError)
		default:
			http.Error(res, model.ErrWentWrong.Error(), http.StatusInternalServerError)
		}
		return
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, shortURL)
}
