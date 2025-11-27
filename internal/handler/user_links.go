package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mrhyman/shortner/api"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

func (h *HTTPHandler) UserLinksHandler(res http.ResponseWriter, req *http.Request) {
	log := logger.FromContext(req.Context())

	if req.Method != http.MethodGet {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(res, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	links, err := h.svc.GerUserLinks(req.Context(), req.Context().Value(model.UserIDKey).(string))
	if err != nil {
		log.With("err", err.Error()).Error()
		http.Error(res, model.ErrWentWrong.Error(), http.StatusInternalServerError)
		return
	}

	if len(links) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	var resp = make([]api.UserLinksResponse, 0, len(links))
	for _, l := range links {
		resp = append(resp, api.UserLinksResponse{
			ShortURL:    l.ShortURL,
			OriginalURL: l.OriginalURL,
		})
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(res)
	if err := enc.Encode(resp); err != nil {
		log.With("err", err.Error())
		http.Error(res, model.ErrResponseEncoding.Error(), http.StatusBadRequest)
		return
	}
}
