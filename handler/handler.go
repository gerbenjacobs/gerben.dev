package handler

import (
	"html/template"
	"net/http"

	"github.com/gerbenjacobs/gerben.dev/internal"
)

// layoutFiles are all the files required to create the site w.r.t. template/define
var layoutFiles = []string{
	"static/views/baseLayout.html",
	"static/views/partials/navbar.html",
	"static/views/partials/aside-hcard.html",
}

// Handler is your dependency container
type Handler struct {
	mux http.Handler
	Dependencies
}

// Dependencies contains all the dependencies your application and its services require
type Dependencies struct {
	SecretKey string
}

// New creates a new handler given a set of dependencies
func New(dependencies Dependencies) *Handler {
	h := &Handler{
		Dependencies: dependencies,
	}

	r := http.NewServeMux()

	// static
	r.Handle("GET /images/", http.StripPrefix("/images/", http.FileServer(http.Dir("static/images"))))
	r.Handle("GET /css/", http.StripPrefix("/css/", http.FileServer(http.Dir("static/css"))))
	r.HandleFunc("GET /robots.txt", singlePage("static/robots.txt"))
	r.HandleFunc("GET /humans.txt", singlePage("static/humans.txt"))

	// Pages
	r.HandleFunc("GET /{$}", h.singlePageLayout("static/views/index.html", internal.Metadata{}))
	r.HandleFunc("GET /joke", h.singlePageLayout(
		"content/single/joke.html",
		internal.Metadata{
			Title:       "fed.brid.gy test",
			Description: "Test page to see how fed.brid.gy posts content, copied from https://fed.brid.gy/docs#web-how-post",
		},
	))
	r.HandleFunc("GET /changelog", h.singlePageLayout(
		"content/single/changelog.html",
		internal.Metadata{
			Title:       "Changelog",
			Description: "This page explains all the (structural) changes that happened to this site.",
		},
	))
	r.HandleFunc("GET /sitemap", h.sitemap)
	r.HandleFunc("GET /tags/{tag}", h.tags)

	// Kindy endpoints
	r.HandleFunc("GET /notes/{file}", Kindy)
	r.HandleFunc("GET /posts/{file}", Kindy)
	r.HandleFunc("GET /likes/{file}", Kindy)
	r.HandleFunc("GET /replies/{file}", Kindy)
	r.HandleFunc("GET /photos/{file}", Kindy)
	r.Handle("GET /kd/", http.StripPrefix("/kd/", http.FileServer(http.Dir("content/kindy/data"))))
	r.HandleFunc("/kindy", internal.BasicAuth(kindyEditor, h.SecretKey))

	h.mux = internal.LogWriter(r)
	return h
}

// ServeHTTP makes sure Handler implements the http.Handler interface
// this keeps the underlying mux private
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func singlePage(fileName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, fileName)
	}
}

func (h *Handler) singlePageLayout(fileName string, metadata internal.Metadata) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.ParseFiles(append(layoutFiles, fileName)...))

		type pageData struct {
			Metadata internal.Metadata
		}
		if err := t.Execute(w, pageData{metadata}); err != nil {
			http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *Handler) tags(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/tags.html")...))

	type pageData struct {
		Metadata internal.Metadata
		Tag      string
	}
	data := pageData{
		Metadata: internal.Metadata{
			Title:       r.PathValue("tag") + " | Tags",
			Description: "All content on gerben.de for the term: " + r.PathValue("tag"),
		},
		Tag: r.PathValue("tag"),
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}
