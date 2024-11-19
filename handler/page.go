package handler

import (
	"html/template"
	"net/http"
)

func (h *Handler) Homepage(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("index.html"))

	if err := t.Execute(w, nil); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}
