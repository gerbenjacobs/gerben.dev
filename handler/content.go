package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"path"
	"strings"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
)

func (h *Handler) posts(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/posts.gohtml")...))

	// get posts
	kindyType := local.KindyTypePost
	entries, err := internal.GetKindyByType(kindyType)
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

func (h *Handler) photos(w http.ResponseWriter, r *http.Request) {
	funcs := map[string]any{
		"hasSuffix": func(s template.HTML, suffix string) bool {
			return strings.HasSuffix(string(s), suffix)
		},
	}
	t := template.Must(template.New(path.Base(layoutFiles[0])).Funcs(funcs).ParseFiles(append(layoutFiles, "static/views/photos.gohtml")...))

	// get posts
	kindyType := local.KindyTypePhoto
	entries, err := internal.GetKindyByType(kindyType)
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
			Title:       "Photos",
			Description: "My photos some by digital camera, others by phone, some pictures from Instagram and some because I just feel like it!",
		},
		Entries: entries,
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}
