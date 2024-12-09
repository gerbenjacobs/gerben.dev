package handler

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path"
	"sort"
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

	// TODO: move kindy content creation to separate service

	b, err := os.ReadFile("content/kindy" + r.URL.Path + ".json")
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
	if string(kind.Content) == r.URL.Path+".html" {
		b, err = os.ReadFile("content/kindy" + string(kind.Content))
		if err != nil {
			slog.Error("failed to read content file", "file", r.URL.Path, "content", kind.Content, "error", err)
			http.Error(w, "entry not found", http.StatusNotFound)
			return
		}
		kind.Content = template.HTML(b)
	}

	type pageData struct {
		Metadata internal.Metadata
		local.Kindy
	}

	metadata := internal.Metadata{
		Title:       internal.Titlify(kind.MustTitle()) + " | " + cases.Title(language.Und).String(string(kind.Type)),
		Description: internal.Descriptify(string(kind.MustDescription())),
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

func GetKindyByType(kindyType string) (entries []local.Kindy, err error) {
	contentPath := local.KindyContentPath + kindyType
	files, err := os.ReadDir(contentPath)
	if err != nil {
		return nil, err
	}

	// Do folder walking, file reading and JSON unmarshalling
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			// skip all non .json files
			continue
		}

		b, err := os.ReadFile(contentPath + "/" + f.Name())
		if err != nil {
			return nil, err
		}

		var tmp local.Kindy
		if err := json.Unmarshal(b, &tmp); err != nil {
			return nil, err
		}
		entries = append(entries, tmp)
	}

	// Sort the entries on published date
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].PublishedAt.After(entries[j].PublishedAt)
	})

	return entries, nil
}

func GetKindyPaths(paths []string) (entries []local.Kindy, err error) {
	for _, f := range paths {

		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}

		var tmp local.Kindy
		if err := json.Unmarshal(b, &tmp); err != nil {
			return nil, err
		}
		entries = append(entries, tmp)
	}

	// Sort the entries on
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].PublishedAt.After(entries[j].PublishedAt)
	})

	return entries, nil
}
