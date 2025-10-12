package handler

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mrhyman/shortner/internal/model"
)

// TODO: refactor while sprint 4
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

func (h *HTTPHandler) ShortLinkHandler(res http.ResponseWriter, req *http.Request) {
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
		http.Error(res, model.ErrInvalidURL.Error(), http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(res, model.ErrInvalidURL.Error(), http.StatusBadRequest)
		return
	}

	id, err := generateShortID(8)
	if err != nil {
		http.Error(res, model.ErrInvalidRequestHeaders.Error(), http.StatusBadRequest)
		return
	}
	h.Repo.Store(req.Context(), id, originalURL)

	base, _ := url.Parse(h.BaseShortURL) 
	short, _ := url.Parse(id)
	shortURL := base.ResolveReference(short).String()

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, shortURL)
}
