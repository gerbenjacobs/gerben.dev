package handler

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	app "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
	"github.com/otiai10/opengraph/v2"
)

var tmpl *template.Template
var opengraphCache = ".cache/opengraph/"
var opengraphTemplate = `
<blockquote>
	<div>
		<p>
			{{ if .Favicon.URL }}<img src="{{.Favicon.URL}}" alt="{{.Title}}" class="timeline-author" loading="lazy" onerror="this.style.display='none';">{{ end }}
			<b>{{.Title}}</b>
		</p>
		<p>{{.DescriptionHTML}}</p>
	</div>
	{{range .Image}}
	{{ if .URL }}
	<figure><img src="{{.URL}}" alt="{{or .Alt $.Title}}" loading="lazy"><figcaption>{{.Alt}}</figcaption></figure>
	{{end}}
	{{end}}
	<cite>&mdash; <a href="{{.URL}}">{{.Title}}</a></cite>
</blockquote>
`

func init() {
	tmpl = template.Must(template.New("opengraph").Parse(opengraphTemplate))
}

func (h *Handler) apiOpenGraph(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Query().Get("url")
	if u == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}

	cacheFile := fmt.Sprintf("%s%x.json", opengraphCache, md5.Sum([]byte(u)))
	b, err := internal.GetCache(cacheFile, 0)
	if err != nil {
		slog.Info("downloading new opengraph data", "url", u)
		// fetch fresh data
		ogp, err := internal.Opengraph(u)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch opengraph data: %v", err), http.StatusInternalServerError)
			return
		}
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(ogp); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode opengraph data: %v", err), http.StatusInternalServerError)
			return
		}
		b = buf.Bytes()
		if err := internal.SetCache(cacheFile, b); err != nil {
			http.Error(w, fmt.Sprintf("failed to cache opengraph data: %v", err), http.StatusInternalServerError)
			return
		}
	}

	var og opengraph.OpenGraph
	err = json.Unmarshal(b, &og)

	// remove empty images
	for i := 0; i < len(og.Image); i++ {
		if og.Image[i].URL == "" {
			og.Image = append(og.Image[:i], og.Image[i+1:]...)
			i--
		}
	}

	absErr := og.ToAbs()
	if err != nil || absErr != nil {
		http.Error(w, fmt.Sprintf("failed to unmarshal opengraph data: %v", err), http.StatusInternalServerError)
		return
	}

	// double check favicon for absolute URL
	if strings.HasPrefix(og.Favicon.URL, "/") || strings.HasPrefix(og.Favicon.URL, ":") {
		base, err := url.Parse(u)
		if err != nil {
			slog.Warn("failed to parse base URL for favicon", "url", u, "err", err)
		}
		og.Favicon.URL = base.Scheme + "://" + base.Host + og.Favicon.URL
	}

	data := struct {
		opengraph.OpenGraph
		DescriptionHTML template.HTML
	}{
		OpenGraph:       og,
		DescriptionHTML: template.HTML(app.MarkdownToHTML(og.Description)),
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
	}
}

func (h *Handler) apiNextPrevious(w http.ResponseWriter, r *http.Request) {
	// get slug from post
	type requestData struct {
		Type string `json:"type"`
		Slug string `json:"slug"`
	}
	var req requestData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("failed to decode request body", "err", err)
		http.Error(w, "failed to decode request body", http.StatusBadRequest)
		return
	}

	// validation
	if req.Slug == "" {
		http.Error(w, "missing slug in request body", http.StatusBadRequest)
		return
	}
	typedType := app.KindyType(req.Type)
	if !typedType.IsValid() {
		http.Error(w, "invalid type in request body", http.StatusBadRequest)
		return
	}

	slog.Warn("searching next/previous", "slug", req.Slug, "type", req.Type)

	prev, next, err := internal.GetKindyNeighbours(typedType, req.Slug)
	if err != nil {
		slog.Error("failed to get kindy neighbours", "err", err)
		http.Error(w, "failed to get neighbours", http.StatusInternalServerError)
		return
	}

	type responseData struct {
		Previous string `json:"previous,omitempty"`
		Next     string `json:"next,omitempty"`
	}
	if err := json.NewEncoder(w).Encode(responseData{
		Previous: prev,
		Next:     next,
	}); err != nil {
		slog.Error("failed to encode response data", "err", err)
	}
}
