package handler

import (
	"database/sql"
	"fmt"
	"net/http"
)

type Handler struct {
	DB *sql.DB
}

func New(db *sql.DB) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) Hello(w http.ResponseWriter, r *http.Request) {
	status := "Connected to database"
	if err := h.DB.Ping(); err != nil {
		status = "DB ping error: " + err.Error()
	}
	fmt.Fprintf(w, "Hello world!<br>Status: %s", status)
}
