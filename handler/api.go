package handler

import (
	"net/http"

	"github.com/davecgh/go-spew/spew"
)

func (h *Handler) ApiAnonPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := spew.Sdump(r.PostForm)
	w.Write([]byte(resp))
}
