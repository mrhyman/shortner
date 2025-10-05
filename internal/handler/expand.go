package handler

import (
	"net/http"
	"strings"

	"github.com/mrhyman/shortner/internal/model"
)

func (h *Handler) ExpandHandler(res http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/")
	if id == "" {
		http.Error(res, model.ErrInvalidLinkId.Error(), http.StatusBadRequest)
		return
	}

	originalURL, ok := h.Store.Get(id)
	if !ok {
		http.Error(res, model.ErrNotFound.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
