package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
)

const PhotosPerPage = 56

func (h *Handler) posts(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/posts.gohtml")...))

	// get posts
	kindyType := local.KindyTypePost
	entries, err := internal.GetKindyCacheByType(kindyType)
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
			Env:         Env,
			Title:       "Posts",
			Description: "These are the posts written on gerben.dev",
			Permalink:   "/posts/",
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
	t := template.Must(template.New(path.Base(layoutFiles[0])).
		Funcs(funcs).
		ParseFiles(append(layoutFiles, "static/views/photos.gohtml", "static/views/partials/photos-paginated.gohtml")...))

	// get posts
	kindyType := local.KindyTypePhoto
	entries, err := internal.GetKindyCacheByType(kindyType)
	if err != nil {
		slog.Error("failed to load entries", "type", kindyType, "error", err)
		http.Error(w, "failed to load entries: "+err.Error(), http.StatusInternalServerError)
		return
	}
	totalEntries := len(entries)

	// paginate
	page := 0
	if r.URL.Query().Get("page") != "" {
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}
	nextPage := page + 1
	lastEntry := nextPage * PhotosPerPage
	if nextPage*PhotosPerPage > totalEntries {
		nextPage = 0
		lastEntry = totalEntries
	}
	entries = entries[page*PhotosPerPage : lastEntry]

	type pageData struct {
		Metadata     internal.Metadata
		TotalEntries int
		NextPage     int
		Entries      []local.Kindy
	}
	data := pageData{
		Metadata: internal.Metadata{
			Env:         Env,
			Title:       "Photos",
			Description: "My photos some by digital camera, others by phone, some pictures from Instagram and some because I just feel like it!",
			Permalink:   "/photos/",
			Image:       string(entries[0].Content), // use latest photo as og:image
		},
		NextPage:     nextPage,
		TotalEntries: totalEntries,
		Entries:      entries,
	}

	// if HTMX call, we return partials only
	isHX := r.Header.Get("HX-Request") //r.URL.Query().Get("HX-Request") to test
	if isHX == "true" {
		if err := t.ExecuteTemplate(w, "photos-paginated", data); err != nil {
			http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}
