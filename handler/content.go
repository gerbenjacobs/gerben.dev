package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
)

const PhotosPerPage = 12

func (h *Handler) indexPage(w http.ResponseWriter, r *http.Request) {
	pageFile := "static/views/index.gohtml"
	t := template.Must(template.ParseFiles(append(layoutFiles, pageFile, "static/views/partials/timeline-paginated.gohtml")...))

	// get posts
	kindyType := local.KindyTypePost
	posts, err := internal.GetKindyCacheByType(kindyType)
	if err != nil {
		slog.Error("failed to load entries", "type", kindyType, "error", err)
		http.Error(w, "failed to load entries: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// get photos
	kindyType = local.KindyTypePhoto
	photos, err := internal.GetKindyCacheByType(kindyType)
	if err != nil {
		slog.Error("failed to load entries", "type", kindyType, "error", err)
		http.Error(w, "failed to load entries: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// get timeline entries
	entries := internal.GetTimelineData(time.Now(), nil, true, true, true, true)

	author, _ := getAuthor() // goddamn, what a hack..
	type pageData struct {
		Metadata internal.Metadata
		Author   *local.KindyAuthor
		NewSince string
		Posts    []local.Kindy
		Photos   []local.Kindy
		Entries  []local.Kindy
	}
	data := pageData{
		Metadata: internal.Metadata{
			Env:         Env,
			Description: "Welcome to my personal website. Here you can find my blog posts and photos.",
			Image:       "/images/opengraph.png",
			Permalink:   "",
			SourceLink:  codeSourcePath + pageFile,
		},
		Author:   author,
		NewSince: "", // disables auto scrolling
		Posts:    kindyLimit(posts, 3),
		Photos:   kindyLimit(photos, 12),
		Entries:  kindyLimit(entries, 10),
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) posts(w http.ResponseWriter, r *http.Request) {
	pageFile := "static/views/posts.gohtml"
	t := template.Must(template.ParseFiles(append(layoutFiles, pageFile)...))

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
			SourceLink:  codeSourcePath + pageFile,
		},
		Entries: entries,
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) photos(w http.ResponseWriter, r *http.Request) {
	pageFile := "static/views/photos.gohtml"
	funcs := map[string]any{
		"hasSuffix": func(s template.HTML, suffix string) bool {
			return strings.HasSuffix(string(s), suffix)
		},
	}
	t := template.Must(template.New(path.Base(layoutFiles[0])).
		Funcs(funcs).
		ParseFiles(append(layoutFiles, pageFile, "static/views/partials/photos-paginated.gohtml")...))

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
	if nextPage*PhotosPerPage >= totalEntries {
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
			SourceLink:  codeSourcePath + pageFile,
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
