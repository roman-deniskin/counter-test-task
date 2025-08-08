package router

import (
	"net/http"

	"counter-test-task/internal/handler"
)

func New(h *handler.Handler) http.Handler {
	mux := http.NewServeMux()

	// /counter/{id} — только GET
	mux.HandleFunc("/counter/", h.Counter)

	// /stats/{id} — только POST
	mux.HandleFunc("/stats/", h.Stats)

	return mux
}
