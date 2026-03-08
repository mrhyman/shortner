package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

func (h *HTTPHandler) ShortLinkHandler(res http.ResponseWriter, req *http.Request) {
	log := logger.Get()

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

	buf := bufferPool.Get()
	defer buf.Reset()

	if _, err := io.Copy(buf, req.Body); err != nil {
		log.With("err", err.Error()).Warn()
		http.Error(res, model.ErrInvalidURL.Error(), http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(buf.String())

	shortURL, err := h.svc.Shorten(req.Context(), originalURL)
	if err != nil {
		var existsErr *model.AlreadyExistsError

		switch {
		case errors.As(err, &existsErr):
			log.With("err", existsErr.Error()).Error()

			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusConflict)

			res.Write([]byte(existsErr.ShortURL))
			return

		case errors.Is(err, model.ErrInvalidURL):
			log.With("err", err.Error()).Warn()
			http.Error(res, err.Error(), http.StatusBadRequest)
			return

		case errors.Is(err, model.ErrShortLinkGeneration),
			errors.Is(err, model.ErrShortenError):
			log.With("err", err.Error()).Error()
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return

		default:
			log.With("err", err.Error()).Error()
			http.Error(res, model.ErrWentWrong.Error(), http.StatusInternalServerError)
			return
		}
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortURL))
}
