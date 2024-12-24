package handler

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/gerbenjacobs/gerben.dev/internal"
	"github.com/otiai10/opengraph/v2"
)

var tmpl *template.Template
var opengraphCache = ".cache/opengraph/"
var opengraphTemplate = `
<blockquote>
	<div>
		<p><img src="{{.Favicon.URL}}" alt="{{.Title}}" class="timeline-author" loading="lazy"> <b>{{.Title}}</b></p>
		<p>{{.Description}}</p>
	</div>
	{{range .Image}}
	<figure><img src="{{.URL}}" alt="{{.Alt}}" loading="lazy"><figcaption>{{.Alt}}</figcaption></figure>
	{{end}}
	<cite>&mdash; <a href="{{.URL}}">{{.Title}}</a></cite>
</blockquote>
`

func init() {
	tmpl = template.Must(template.New("opengraph").Parse(opengraphTemplate))
}

func (h *Handler) apiOpenGraph(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}

	cacheFile := fmt.Sprintf("%s%x.json", opengraphCache, md5.Sum([]byte(url)))
	b, err := internal.GetCache(cacheFile, 0)
	if err != nil {
		slog.Info("downloading new opengraph data", "url", url)
		// fetch fresh data
		ogp, err := opengraph.Fetch(url)
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
	absErr := og.ToAbs()
	if err != nil || absErr != nil {
		http.Error(w, fmt.Sprintf("failed to unmarshal opengraph data: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, og); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
	}
}
