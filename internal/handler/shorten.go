package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mrhyman/shortner/api"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

func (h *HTTPHandler) ShortenHandler(res http.ResponseWriter, req *http.Request) {
	log := logger.FromContext(req.Context())

	if req.Method != http.MethodPost {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(res, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	var r api.ShortenRequest
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&r); err != nil {
		log.With("err", err.Error()).Warn()
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.svc.Shorten(req.Context(), r.URL)

	if err != nil {
		var existsErr *model.AlreadyExistsError

		switch {
		case errors.As(err, &existsErr):
			resp := api.ShortenResponse{
				Result: strings.TrimSuffix(existsErr.ShortURL, "\n"),
			}

			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusConflict)

			log.With("err", existsErr.Error()).Error()
			enc := json.NewEncoder(res)
			if err := enc.Encode(resp); err != nil {
				log.With("err", err.Error())
				http.Error(res, model.ErrResponseEncoding.Error(), http.StatusBadRequest)
			}
			return

		case errors.Is(err, model.ErrInvalidURL):
			log.With("err", err.Error()).Warn()
			http.Error(res, err.Error(), http.StatusBadRequest)
			return

		case errors.Is(err, model.ErrShortLinkGeneration) || errors.Is(err, model.ErrShortenError):
			log.With("err", err.Error()).Error()
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return

		default:
			log.With("err", err.Error()).Error()
			http.Error(res, model.ErrWentWrong.Error(), http.StatusInternalServerError)
			return
		}
	}

	resp := api.ShortenResponse{
		Result: shortURL,
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(res)
	if err := enc.Encode(resp); err != nil {
		log.With("err", err.Error())
		http.Error(res, model.ErrResponseEncoding.Error(), http.StatusBadRequest)
		return
	}
}
