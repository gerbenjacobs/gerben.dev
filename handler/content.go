package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
)

func (h *Handler) posts(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/content.gohtml")...))

	// get posts
	kindyType := KindyURLPosts
	entries, err := GetKindyByType(kindyType)
	if err != nil {
		slog.Error("failed to load entries", "type", kindyType, "error", err)
		http.Error(w, "failed to load entries: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type pageData struct {
		Metadata internal.Metadata
		Entries  []local.Kindy
	}
	data := pageData{
		Metadata: internal.Metadata{
			Title:       "Posts",
			Description: "These are the posts written on gerben.dev",
		},
		Entries: entries,
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}
