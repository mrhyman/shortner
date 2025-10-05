package handler

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mrhyman/shortner/internal/model"
)

func generateShortID(n int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}

func (h *Handler) ShortLinkHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, model.ErrInvalidRequestParams.Error(), http.StatusBadRequest)
		return
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "text/plain") {
		http.Error(res, model.ErrInvalidRequestHeaders.Error(), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, model.ErrInvalidUrl.Error(), http.StatusBadRequest)
		return
	}
	
	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(res, model.ErrInvalidUrl.Error(), http.StatusBadRequest)
		return
	}

	id, err := generateShortID(8)
	if err != nil {
		http.Error(res, model.ErrInvalidRequestHeaders.Error(), http.StatusBadRequest)
		return
	}
	h.Store.Set(id, originalURL)

	shortURL := fmt.Sprintf("%s/%s", h.Base, id)

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, shortURL)
}
