package handler

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Kindy(w http.ResponseWriter, r *http.Request) {
	funcs := map[string]any{
		"hasSuffix": func(s template.HTML, suffix string) bool {
			return strings.HasSuffix(string(s), suffix)
		},
	}
	t := template.Must(template.New(path.Base(layoutFiles[0])).Funcs(funcs).ParseFiles(append(layoutFiles, "static/views/kindy.gohtml")...))

	kindyFile := "content/kindy" + r.URL.Path + ".json"
	b, err := os.ReadFile(kindyFile)
	if err != nil {
		slog.Error("failed to read file", "file", r.URL.Path, "error", err)
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}

	var kind local.Kindy
	if err := json.Unmarshal(b, &kind); err != nil {
		slog.Error("failed to unmarshal kindy", "file", r.URL.Path, "error", err)
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}

	// if no author, set default
	if kind.Author == nil {
		author, err := getAuthor()
		if err != nil {
			slog.Error("failed to get author", "error", err)
			return
		}
		kind.Author = author
	}

	// if our content is inside another file, load it
	// for HTML content
	if string(kind.Content) == r.URL.Path+".html" {
		b, err = os.ReadFile("content/kindy" + string(kind.Content))
		if err != nil {
			slog.Error("failed to read content file", "file", r.URL.Path, "content", kind.Content, "error", err)
			http.Error(w, "entry not found", http.StatusNotFound)
			return
		}
		kind.Content = template.HTML(b)
	}
	// for markdown content
	if strings.HasSuffix(string(kind.Content), ".md") {
		b, err = os.ReadFile("content/kindy" + string(kind.Content))
		if err != nil {
			slog.Error("failed to read content file", "file", r.URL.Path, "content", kind.Content, "error", err)
			http.Error(w, "entry not found", http.StatusNotFound)
			return
		}
		kind.Content = template.HTML(local.MarkdownToHTML(string(b)))
	}

	type pageData struct {
		Metadata internal.Metadata
		local.Kindy
	}

	metadata := internal.Metadata{
		Env:         Env,
		Title:       internal.Titlify(kind.MustTitle()) + " | " + cases.Title(language.Und).String(string(kind.Type)),
		Description: internal.Descriptify(string(kind.MustDescription())),
		Permalink:   kind.Permalink,
		Kindy:       &kind,
		SourceLink:  codeSourcePath + kindyFile,
	}
	if kind.Type == "photo" {
		metadata.Image = string(kind.Content)
	}
	data := pageData{
		Metadata: metadata,
		Kindy:    kind,
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}
