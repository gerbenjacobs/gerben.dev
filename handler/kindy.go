package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Kindy(w http.ResponseWriter, r *http.Request) {
	redirects := map[string]string{
		"/posts/20241128-bringing-the-indieweb": "/posts/bringing-the-indieweb",
		"/posts/20241204-instagram-archive":     "/posts/instagram-archive",
	}
	for from, to := range redirects {
		if r.URL.Path == from {
			http.Redirect(w, r, to, http.StatusMovedPermanently)
			return
		}
	}

	t := template.Must(template.ParseFiles(append(layoutFiles, "static/views/kindy.gohtml")...))
	kindyFile := "content/kindy" + r.URL.Path + ".json"

	_, err := os.Stat(kindyFile)
	// Posts can live in subfolders, so we need to find them by permalink
	// and use the slug to find the json file
	if os.IsNotExist(err) && strings.HasPrefix(r.URL.Path, local.KindyURLPosts) {
		posts, err := internal.GetKindyCacheByType(local.KindyTypePost)
		if err != nil {
			slog.Error("failed to get posts", "error", err)
			// in the rare case that this happens, just ignore
		} else {
			for _, post := range posts {
				if r.URL.Path == post.Permalink {
					kindyFile = "content/kindy/posts/" + post.Slug + ".json"
					break
				}
			}
		}
	}

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
	content, err := internal.GetPostContent(string(kind.Content))
	if err != nil {
		slog.Error("failed to get post content", "file", r.URL.Path, "error", err)
		http.Error(w, "failed to load content", http.StatusInternalServerError)
		return
	}
	kind.Content = content

	type pageData struct {
		Metadata internal.Metadata
		local.Kindy
		RawKindy string
	}

	metadata := internal.Metadata{
		Env:         Env,
		Title:       internal.Titlify(kind.MustTitle()) + " | " + cases.Title(language.Und).String(string(kind.Type)),
		Description: internal.Descriptify(string(kind.MustDescription())),
		Permalink:   kind.Permalink,
		Kindy:       &kind,
		SourceLink:  codeSourcePath + kindyFile,
	}
	if kind.Image != "" {
		metadata.Image = kind.Image
	}
	if kind.Type == "photo" {
		metadata.Image = string(kind.Content)
	}
	data := pageData{
		Metadata: metadata,
		Kindy:    kind,
		RawKindy: string(b),
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "failed to execute template:"+err.Error(), http.StatusInternalServerError)
	}
}

func KindyUpdate(w http.ResponseWriter, r *http.Request) {
	if Env != "dev" {
		http.Error(w, "not allowed", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	t := local.KindyType(r.Form.Get("type"))
	slug := r.Form.Get("slug")
	rawKindy := r.Form.Get("raw")

	f := fmt.Sprintf("%s%s/%s.json", local.KindyContentPath, t.URL(), slug)
	err = os.WriteFile(f, []byte(rawKindy), 0644)
	if err != nil {
		http.Error(w, "failed to write file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var kind local.Kindy
	if err := json.Unmarshal([]byte(rawKindy), &kind); err != nil {
		http.Error(w, "failed to unmarshal kindy: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, kind.Permalink, http.StatusSeeOther)
}
