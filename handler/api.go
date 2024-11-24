package handler

import (
	"log/slog"
	"net/http"
)

func (h *Handler) ApiAnonPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		slog.Error("failed to parse form for api/anon", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("received data on api/anon", "data", r.PostForm, "headers", r.Header)
	w.WriteHeader(http.StatusNoContent)
}
