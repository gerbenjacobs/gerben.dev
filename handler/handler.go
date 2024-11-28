package handler

import (
	"html/template"
	"net/http"

	"github.com/gerbenjacobs/gerben.dev/internal"
)

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
type Dependencies struct{}

// New creates a new handler given a set of dependencies
func New(dependencies Dependencies) *Handler {
	h := &Handler{
		Dependencies: dependencies,
	}

	r := http.NewServeMux()

	r.Handle("GET /images/", http.StripPrefix("/images/", http.FileServer(http.Dir("static/images"))))
	r.Handle("GET /css/", http.StripPrefix("/css/", http.FileServer(http.Dir("static/css"))))
	r.HandleFunc("GET /robots.txt", singlePage("static/robots.txt"))
	r.HandleFunc("GET /humans.txt", singlePage("static/humans.txt"))

	r.HandleFunc("GET /{$}", h.singlePageLayout("static/views/index.html"))
	r.HandleFunc("GET /joke", h.singlePageLayout("content/single/joke.html"))
	r.HandleFunc("GET /changelog", h.singlePageLayout("content/single/changelog.html"))
	r.HandleFunc("GET /sitemap", h.singlePageLayout("content/single/sitemap.html"))

	r.HandleFunc("GET /tags/{tag}", h.tags)
	r.HandleFunc("GET /notes/{file}", h.Kindy)
	r.HandleFunc("GET /posts/{file}", h.Kindy)
	r.HandleFunc("GET /likes/{file}", h.Kindy)
	r.HandleFunc("GET /replies/{file}", h.Kindy)

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

func (h *Handler) singlePageLayout(fileName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.ParseFiles(append(layoutFiles, fileName)...))

		if err := t.Execute(w, nil); err != nil {
			http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *Handler) tags(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/tags.html")...))

	if err := t.Execute(w, r.PathValue("tag")); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}
