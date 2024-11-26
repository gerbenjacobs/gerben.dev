package handler

import (
	"net/http"

	"github.com/gerbenjacobs/gerben.dev/internal"
)

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

	r.Handle("GET /images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	r.Handle("GET /css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	r.HandleFunc("GET /robots.txt", singlePage("robots.txt"))
	r.HandleFunc("GET /humans.txt", singlePage("humans.txt"))

	r.HandleFunc("GET /{$}", h.Homepage)
	r.HandleFunc("POST /api/anon", h.ApiAnonPost)

	r.HandleFunc("GET /joke", singlePage("content/single/joke.html"))

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
