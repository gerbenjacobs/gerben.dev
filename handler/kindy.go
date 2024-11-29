package handler

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	local "github.com/gerbenjacobs/gerben.dev"
)

func Kindy(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/kindy.gohtml")...))

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
