package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
)

func (h *Handler) sitemap(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/sitemap.html")...))

	type pageData struct {
		Metadata internal.Metadata
		Posts    []local.Kindy
		Photos   []local.Kindy
		Notes    []local.Kindy
	}

	posts, err := GetKindyByType("posts")
	if err != nil {
		slog.Error("failed to load kindy posts", "error", err)
	}
	photos, err := GetKindyByType("photos")
	if err != nil {
		slog.Error("failed to load kindy photos", "error", err)
	}
	notes, err := GetKindyByType("notes")
	if err != nil {
		slog.Error("failed to load kindy notes", "error", err)
	}

	data := pageData{
		Metadata: internal.Metadata{Title: "Sitemap", Description: "A HTML version of my sitemap"},
		Posts:    posts,
		Photos:   photos,
		Notes:    notes,
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}
