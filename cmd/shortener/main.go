package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mrhyman/shortner/internal/handler"
	"github.com/mrhyman/shortner/internal/repository"
)

const (
	addr = ":8080"
)

func main() {
	store := repository.NewStore()

	h := &handler.Handler{
		Store: store,
		Base:  fmt.Sprintf(`http://localhost%s`, addr),
	}

	mux := http.NewServeMux()

	mux.HandleFunc(`/`, h.ShortLinkHandler)
	mux.HandleFunc(`/{id}`, h.ExpandHandler)

	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
