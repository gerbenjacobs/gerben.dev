package handler

import (
	"bytes"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
	"github.com/mmcdole/gofeed"
)

// layoutFiles are all the files required to create the site w.r.t. template/define
var layoutFiles = []string{
	"static/views/baseLayout.html",
	"static/views/partials/navbar.html",
	"static/views/partials/aside-hcard.html",
	"static/views/partials/tag-preview.gohtml",
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
	// r.HandleFunc("GET /.well-known/atproto-did", singlePage("static/did")) // disabled for now

	// Pages
	r.HandleFunc("GET /{$}", h.singlePageLayout("static/views/index.html", internal.Metadata{Image: "/images/opengraph.png"}))
	r.HandleFunc("GET /changelog", h.singlePageLayout("static/views/changelog.html",
		internal.Metadata{
			Title:       "Changelog",
			Description: "This page explains all the (structural) changes that happened to this site.",
			Permalink:   "/changelog",
		},
	))
	r.HandleFunc("GET /sitemap", h.sitemap)
	r.HandleFunc("GET /sitemap.xml", h.sitemapXML)
	r.HandleFunc("GET /tags/{tag}", h.tags)
	r.HandleFunc("GET /listening", h.listening)
	r.HandleFunc("GET /timeline", h.timeline)
	r.HandleFunc("GET /timeline.xml", h.timelineXML)

	// Kindy endpoints
	r.HandleFunc("GET /notes/{file}", Kindy)
	r.HandleFunc("GET /posts/{file}", Kindy)
	r.HandleFunc("GET /likes/{file}", Kindy)
	r.HandleFunc("GET /reposts/{file}", Kindy)
	r.HandleFunc("GET /replies/{file}", Kindy)
	r.HandleFunc("GET /photos/{file}", Kindy)
	r.Handle("GET /kd/", http.StripPrefix("/kd/", http.FileServer(http.Dir("content/kindy/data"))))
	r.HandleFunc("/kindy", internal.BasicAuth(kindyEditor, h.SecretKey))

	r.HandleFunc("GET /posts/{$}", h.posts)
	r.HandleFunc("GET /photos/{$}", h.photos)
	r.HandleFunc("GET /notes/{$}", redirect("/timeline"))
	r.HandleFunc("GET /likes/{$}", redirect("/timeline"))
	r.HandleFunc("GET /reposts/{$}", redirect("/timeline"))
	r.HandleFunc("GET /tags/{$}", redirect("/timeline"))

	// API stuff?
	r.HandleFunc("GET /api/opengraph", h.apiOpenGraph)

	h.mux = internal.LogWriter(r)
	return h
}

// ServeHTTP makes sure Handler implements the http.Handler interface
// this keeps the underlying mux private
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func redirect(url string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusFound)
	}
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

	// find all Kindy content with the tag
	tag := strings.ToLower(r.PathValue("tag"))
	tagInfo := internal.GetTag(tag)
	var paths []string
	for _, entry := range tagInfo.Entries {
		paths = append(paths, entry.KindyPath)
	}

	entries, err := internal.GetKindyPaths(paths)
	if err != nil {
		slog.Error("failed to load kindy entries", "error", err)
	}

	// get all tags
	allTags := internal.GetAllTags()

	type pageData struct {
		Metadata internal.Metadata
		Tag      string
		Entries  []local.Kindy
		AllTags  map[string]internal.TagInfo
	}
	data := pageData{
		Metadata: internal.Metadata{
			Title:       tag + " | Tags",
			Description: "All content on gerben.dev for the term: #" + tag,
			Permalink:   "/tags/" + tag,
		},
		Tag:     tag,
		Entries: entries,
		AllTags: allTags,
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) listening(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/listening.html")...))
	feedUrl := "https://lfm.xiffy.nl/theonewithout"
	cacheFile := ".cache/listening.xml"

	// check if cache file exists and is not older than 10 minutes
	b, err := internal.GetCache(cacheFile, 10*time.Minute)
	if err != nil {
		slog.Warn("downloading new listening feed")
		resp, err := http.Get(feedUrl)
		if err != nil {
			http.Error(w, "failed to get feed:"+err.Error(), http.StatusInternalServerError)
			return
		}
		b, err = io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "failed to read feed:"+err.Error(), http.StatusInternalServerError)
			return
		}
		internal.SetCache(cacheFile, b)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(bytes.NewReader(b))
	if err != nil {
		http.Error(w, "failed to parse feed:"+err.Error(), http.StatusInternalServerError)
		return
	}

	type pageData struct {
		Metadata internal.Metadata
		Feed     *gofeed.Feed
	}
	data := pageData{
		Metadata: internal.Metadata{
			Title:       "Listening",
			Description: "This page lists what I'm currently listening to.",
		},
		Feed: feed,
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}
