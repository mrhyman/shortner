package handler

import (
	"net/http"
	"strings"

	"github.com/mrhyman/shortner/internal/model"
)

func (h *HTTPHandler) ExpandHandler(res http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/")
	if id == "" {
		http.Error(res, model.ErrInvalidLinkID.Error(), http.StatusBadRequest)
		return
	}

	originalURL, err := h.Repo.GetByID(req.Context(), id)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
