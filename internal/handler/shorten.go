package handler

import (
	"encoding/json"
	"errors"
	"net/http"

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
		log.With("err", model.ErrInvalidRequestParams.Error()).Error()
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortURL, err := h.svc.Shorten(req.Context(), r.URL)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidURL):
			log.With("err", model.ErrEnvParsing.Error()).Warn()
			http.Error(res, err.Error(), http.StatusBadRequest)
		case errors.Is(err, model.ErrShortLinkGeneration) || errors.Is(err, model.ErrShortenError):
			log.With("err", model.ErrEnvParsing.Error()).Error()
			http.Error(res, err.Error(), http.StatusInternalServerError)
		default:
			log.With("err", model.ErrEnvParsing.Error()).Error()
			http.Error(res, model.ErrWentWrong.Error(), http.StatusInternalServerError)
		}
		return
	}

	resp := api.ShortenResponse{
		Result: shortURL,
	}

	res.Header().Set("Content-Type", "application/json")

	res.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(res)
	if err := enc.Encode(resp); err != nil {
		log.With("err", model.ErrResponseEncoding.Error(), "trace", err.Error())
		http.Error(res, model.ErrResponseEncoding.Error(), http.StatusInternalServerError)
		return
	}
}
