// Package handler реализует HTTP обработчики для сервиса сокращения URL.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mrhyman/shortner/api"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

// GetUserLinksHandler обрабатывает запросы на получение всех ссылок пользователя.
// Принимает GET запрос и возвращает список всех сокращенных ссылок пользователя в формате JSON.
func (h *HTTPHandler) GetUserLinksHandler(res http.ResponseWriter, req *http.Request) {
	log := logger.Get()

	if req.Method != http.MethodGet {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(res, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	userID, ok := req.Context().Value(model.UserIDKey).(string)
	if !ok {
		http.Error(res, model.ErrUnknownUser.Error(), http.StatusInternalServerError)
		return
	}

	links, err := h.svc.GetUserLinks(req.Context(), userID)
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
