package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mrhyman/shortner/api"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

func (h *HTTPHandler) ShortenBatchHandler(res http.ResponseWriter, req *http.Request) {
	log := logger.FromContext(req.Context())

	if req.Method != http.MethodPost {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(res, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	var sbr []api.ShortenBatchRequest
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&sbr); err != nil {
		log.With("err", err.Error()).Warn()
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(sbr) == 0 {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode([]struct{}{})
		return
	}

	links, err := h.svc.ShortenBatch(req.Context(), sbr)
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

	var resp = make([]api.ShortenBatchResponse, 0, len(links))
	for _, l := range links {
		resp = append(resp, api.ShortenBatchResponse{
			CorrelationID: l.CorrelationID,
			ShortURL:      l.ShortURL,
		})
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
