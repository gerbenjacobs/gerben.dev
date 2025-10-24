package handler

import (
	"encoding/json"
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
	// for HTML content
	if strings.HasSuffix(string(kind.Content), ".html") {
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
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	p := r.Form.Get("permalink")
	rawKindy := r.Form.Get("raw")

	err = os.WriteFile("content/kindy"+p+".json", []byte(rawKindy), 0644)
	if err != nil {
		http.Error(w, "failed to write file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, p, http.StatusSeeOther)
}
