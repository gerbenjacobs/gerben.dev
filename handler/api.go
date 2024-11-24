package handler

import (
	"log/slog"
	"net/http"

	"github.com/davecgh/go-spew/spew"
)

func (h *Handler) ApiAnonPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		slog.Error("failed to parse form for api/anon", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := spew.Sdump(r.PostForm)
	w.Write([]byte(resp))
}
