package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// /counter/{bannerID} (GET) — инкремент без тела ответа (204 No Content)
func (h *Handler) Counter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := idFromPath(r.URL.Path, "/counter/")
	if err != nil {
		http.Error(w, "bad banner id", http.StatusBadRequest)
		return
	}

	if err := h.Srv.IncClick(r.Context(), id); err != nil {
		http.Error(w, "increment failed", http.StatusInternalServerError)
		return
	}

	// Ничего не пишем в тело: только заголовки и 204
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

// /stats/{bannerID} (POST) — вернуть поминутную статистику (JSON)
type statsRequest struct {
	From string `json:"from"` // RFC3339
	To   string `json:"to"`   // RFC3339
}

type statsPoint struct {
	TS time.Time `json:"ts"`
	V  int       `json:"v"`
}

type statsResponse struct {
	Stats []statsPoint `json:"stats"`
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := idFromPath(r.URL.Path, "/stats/")
	if err != nil {
		http.Error(w, "bad banner id", http.StatusBadRequest)
		return
	}

	var req statsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}

	from, err := time.Parse(time.RFC3339, req.From)
	if err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}
	to, err := time.Parse(time.RFC3339, req.To)
	if err != nil {
		http.Error(w, "bad to: "+err.Error(), http.StatusBadRequest)
		return
	}
	if !to.After(from) {
		http.Error(w, "to must be after from", http.StatusBadRequest)
		return
	}

	points, err := h.Srv.GetStats(r.Context(), id, from, to)
	if err != nil {
		http.Error(w, "stats failed", http.StatusInternalServerError)
		return
	}

	out := statsResponse{Stats: make([]statsPoint, 0, len(points))}
	for _, p := range points {
		out.Stats = append(out.Stats, statsPoint{TS: p.TS, V: p.V})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// idFromPath извлекает числовой id из пути после prefix.
// Примеры: "/counter/42" при prefix "/counter/" вернёт 42.
func idFromPath(path, prefix string) (int64, error) {
	if !strings.HasPrefix(path, prefix) {
		return 0, errors.New("bad prefix")
	}
	rest := strings.TrimPrefix(path, prefix)
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		rest = rest[:i]
	}
	if rest == "" {
		return 0, fmt.Errorf("empty id")
	}
	id, err := strconv.ParseInt(rest, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}
