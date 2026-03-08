// Package handler реализует HTTP обработчики для сервиса сокращения URL.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mrhyman/shortner/api"
	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

// StatsHandler обрабатывает запросы для получения статистики сервиса.
// Проверяет IP-адрес клиента на принадлежность к доверенной подсети.
// Возвращает количество сокращённых URL и пользователей в системе.
func (h *HTTPHandler) StatsHandler(res http.ResponseWriter, req *http.Request) {
	log := logger.Get()

	if req.Method != http.MethodGet {
		log.With("err", model.ErrInvalidRequestParams.Error()).Warn()
		http.Error(res, model.ErrInvalidRequestParams.Error(), http.StatusMethodNotAllowed)
		return
	}

	// Получение статистики
	urlsCount, usersCount, err := h.svc.GetStats(req.Context())
	if err != nil {
		log.With("err", err.Error()).Error()
		http.Error(res, model.ErrStorageUnavailable.Error(), http.StatusInternalServerError)
		return
	}

	// Формирование ответа
	response := api.StatsResponse{
		URLs:  urlsCount,
		Users: usersCount,
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(res).Encode(response); err != nil {
		log.With("err", err.Error()).Error()
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		return
	}
}
