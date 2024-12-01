package handler

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"

	local "github.com/gerbenjacobs/gerben.dev"
)

func Kindy(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/kindy.gohtml")...))

	// TODO: move kindy content creation to seperate service

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

	if err := t.Execute(w, kind); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}

func GetKindyByType(kindyType string) (entries []local.Kindy, err error) {
	contentPath := KindyContentPath + kindyType
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

	// Sort the entries on
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].PublishedAt.After(entries[j].PublishedAt)
	})

	return entries, nil
}
