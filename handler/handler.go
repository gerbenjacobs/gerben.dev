package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"strings"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
	"github.com/mmcdole/gofeed"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	codeSourcePath = "https://github.com/gerbenjacobs/gerben.dev/blob/main/"
)

var Env string

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
func New(env string, dependencies Dependencies) *Handler {
	Env = env
	h := &Handler{
		Dependencies: dependencies,
	}

	r := http.NewServeMux()

	// static
	r.Handle("GET /images/", http.StripPrefix("/images/", http.FileServer(http.Dir("static/images"))))
	r.Handle("GET /css/", http.StripPrefix("/css/", http.FileServer(http.Dir("static/css"))))
	r.Handle("GET /js/", http.StripPrefix("/js/", http.FileServer(http.Dir("static/js"))))
	r.Handle("GET /robots.txt", otelhttp.WithRouteTag("/robots.txt", http.HandlerFunc(singlePage("static/robots.txt"))))
	r.Handle("GET /humans.txt", otelhttp.WithRouteTag("/humans.txt", http.HandlerFunc(singlePage("static/humans.txt"))))
	// r.HandleFunc("GET /.well-known/atproto-did", singlePage("static/did")) // disabled for now

	// Pages
	r.Handle("GET /{$}", otelhttp.WithRouteTag("/", http.HandlerFunc(h.indexPage)))
	r.Handle("GET /changelog", otelhttp.WithRouteTag("/changelog", http.HandlerFunc(h.singlePageLayout("static/views/changelog.html", internal.Metadata{
		Env: Env, Permalink: "/changelog", Title: "Changelog",
		Description: "This page explains all the (structural) changes that happened to this site.",
	},
	))))
	r.Handle("GET /collection", otelhttp.WithRouteTag("/collection", http.HandlerFunc(h.singlePageLayout("static/views/collection.html", internal.Metadata{
		Env: Env, Permalink: "/collection", Title: "My Collection", Image: "/kd/photos/PXL_20240911_124800628.jpg",
		Description: "Sometimes when my kids and I play, I collect things, achievements, and other stuff. This page lists them.",
	}))))
	r.Handle("GET /projects", otelhttp.WithRouteTag("/projects", http.HandlerFunc(h.singlePageLayout("static/views/projects.html", internal.Metadata{
		Env: Env, Permalink: "/projects", Title: "My projects", Image: "/kd/photos/PXL_20240924_105523447.jpg",
		Description: "These are some of the projects, tools and libraries I have worked on.",
	}))))

	r.Handle("GET /sitemap", otelhttp.WithRouteTag("/sitemap", http.HandlerFunc(h.sitemap)))
	r.Handle("GET /sitemap.xml", otelhttp.WithRouteTag("/sitemap.xml", http.HandlerFunc(h.sitemapXML)))
	r.Handle("GET /tags/{tag}", otelhttp.WithRouteTag("/tags/{tag}", http.HandlerFunc(h.tags)))
	r.Handle("GET /listening", otelhttp.WithRouteTag("/listening", http.HandlerFunc(h.listening)))
	r.Handle("GET /timeline", otelhttp.WithRouteTag("/timeline", http.HandlerFunc(h.timeline)))
	r.Handle("GET /previously", otelhttp.WithRouteTag("/previously", http.HandlerFunc(h.previously)))
	r.Handle("GET /poems", otelhttp.WithRouteTag("/poems", http.HandlerFunc(h.singlePageLayout("static/views/poems.html", internal.Metadata{
		Env: Env, Permalink: "/poems", Title: "Poems", Image: "",
		Description: "My poems, mostly from my teenage years..",
	}))))

	// XML/RSS feeds
	r.Handle("GET /timeline.xml", otelhttp.WithRouteTag("/timeline.xml", http.HandlerFunc(h.timelineXML)))
	r.Handle("GET /feed.xml", otelhttp.WithRouteTag("/feed.xml", http.HandlerFunc(h.postsXML)))
	r.Handle("GET /photos.xml", otelhttp.WithRouteTag("/photos.xml", http.HandlerFunc(h.photosXML)))

	// Kindy endpoints
	r.Handle("GET /notes/{file}", otelhttp.WithRouteTag("/notes/{file}", http.HandlerFunc(Kindy)))
	r.Handle("GET /posts/{file}", otelhttp.WithRouteTag("/posts/{file}", http.HandlerFunc(Kindy)))
	r.Handle("GET /likes/{file}", otelhttp.WithRouteTag("/likes/{file}", http.HandlerFunc(Kindy)))
	r.Handle("GET /reposts/{file}", otelhttp.WithRouteTag("/reposts/{file}", http.HandlerFunc(Kindy)))
	r.Handle("GET /replies/{file}", otelhttp.WithRouteTag("/replies/{file}", http.HandlerFunc(Kindy)))
	r.Handle("GET /photos/{file}", otelhttp.WithRouteTag("/photos/{file}", http.HandlerFunc(Kindy)))
	r.Handle("POST /kindy/update", otelhttp.WithRouteTag("/kindy/update", http.HandlerFunc(KindyUpdate)))
	r.Handle("GET /kd/", otelhttp.WithRouteTag("/kd/", http.StripPrefix("/kd/", http.FileServer(http.Dir("content/kindy/data")))))
	r.Handle("/kindy", otelhttp.WithRouteTag("/kindy", http.HandlerFunc(internal.BasicAuth(kindyEditor, h.SecretKey))))

	r.Handle("GET /posts/{$}", otelhttp.WithRouteTag("/posts/{$}", http.HandlerFunc(h.posts)))
	r.Handle("GET /photos/featured", otelhttp.WithRouteTag("/photos/featured", http.HandlerFunc(h.photos(true))))
	r.Handle("GET /photos/{$}", otelhttp.WithRouteTag("/photos/{$}", http.HandlerFunc(h.photos(false))))
	r.HandleFunc("GET /notes/{$}", redirect("/timeline"))
	r.HandleFunc("GET /likes/{$}", redirect("/timeline"))
	r.HandleFunc("GET /reposts/{$}", redirect("/timeline"))
	r.HandleFunc("GET /tags/{$}", redirect("/timeline"))

	// API endpoints
	r.Handle("POST /api/opengraph", otelhttp.WithRouteTag("/api/opengraph", http.HandlerFunc(h.apiOpenGraph)))
	r.Handle("POST /api/nextprevious", otelhttp.WithRouteTag("/api/nextprevious", http.HandlerFunc(h.apiNextPrevious)))
	r.Handle("POST /api/thumbsup", otelhttp.WithRouteTag("/api/thumbsup", http.HandlerFunc(h.apiThumbsUp)))
	r.Handle("GET /api/thumbsup/count", otelhttp.WithRouteTag("/api/thumbsup/count", http.HandlerFunc(h.apiThumbsUpCount)))

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

		metadata.SourceLink = codeSourcePath + fileName

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
			Env:         Env,
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
	pageFile := "static/views/listening.html"
	t := template.Must(template.ParseFiles(append(layoutFiles, pageFile)...))

	feed, err := internal.GetListeningData(true)
	if err != nil {
		http.Error(w, "failed to get listening data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	type pageData struct {
		Metadata internal.Metadata
		Feed     *gofeed.Feed
	}
	data := pageData{
		Metadata: internal.Metadata{
			Env:         Env,
			Title:       "Listening",
			Description: "This page lists what I'm currently listening to.",
			Permalink:   "/listening",
			SourceLink:  codeSourcePath + pageFile,
		},
		Feed: feed,
	}
	if feed != nil && len(feed.Items) > 0 && feed.Items[0].Image != nil {
		data.Metadata.Image = feed.Items[0].Image.URL
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}
