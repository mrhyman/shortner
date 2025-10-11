package handler

import (
	"net/http"
	"strings"

	"github.com/mrhyman/shortner/internal/model"
)

func (h *HttpHandler) ExpandHandler(res http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/")
	if id == "" {
		http.Error(res, model.ErrInvalidLinkId.Error(), http.StatusBadRequest)
		return
	}

	originalURL, err := h.Repo.GetByID(h.Ctx, id)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
