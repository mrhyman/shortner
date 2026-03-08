// Package handler реализует HTTP обработчики для сервиса сокращения URL.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

// DeleteUserLinksHandler обрабатывает запросы на удаление ссылок пользователя.
// Принимает DELETE запрос с массивом идентификаторов ссылок для удаления.
// Асинхронно удаляет указанные ссылки и возвращает HTTP 202 Accepted.
func (h *HTTPHandler) DeleteUserLinksHandler(res http.ResponseWriter, req *http.Request) {
	log := logger.Get()

	if req.Method != http.MethodDelete {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(res, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	var links []string

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&links); err != nil {
		log.With("err", err.Error()).Warn()
		http.Error(res, model.ErrInvalidRequestParams.Error(), http.StatusBadRequest)
		return
	}

	h.svc.DeleteUserLinksByID(req.Context(), links)

	res.WriteHeader(http.StatusAccepted)
}
