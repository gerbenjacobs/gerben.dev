package main

import (
	"encoding/json"
	"flag"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/gerben.dev/internal"
)

func main() {
	dir := flag.String("dir", "content/kindy", "directory to search for Kindy JSON files")
	output := flag.String("o", "content/kindy/data/tags.json", "output directory for tags.json")
	flag.Parse()

	if *dir == "" {
		log.Fatalf("the `dir` flag is required to operate")
	}

	slog.Info("searching through dir", "dir", *dir)
	files, err := findKindyJSON(*dir)
	if err != nil {
		log.Fatalf("failed to find Kindy JSON files: %v", err)
	}
	slog.Info("found Kindy JSON files", "files", len(files))

	tags := map[string]internal.TagInfo{}
	for _, file := range files {
		t, pm, tg := extractTags(file)
		mergeTags(tags, t, pm, tg)
	}

	slog.Info("writing tags.json", "output", *output)
	b, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal tags: %v", err)
	}
	if err := os.WriteFile(*output, b, 0644); err != nil {
		log.Fatalf("failed to write tags.json: %v", err)
	}

	slog.Info("done, wrote tags.json", "tags", len(tags))
}

func findKindyJSON(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".json" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func extractTags(file string) (local.KindyType, string, []string) {
	b, err := os.ReadFile(file)
	if err != nil {
		slog.Error("failed to read file", "file", file, "error", err)
		return "", "", nil
	}

	var kind local.Kindy
	if err := json.Unmarshal(b, &kind); err != nil {
		slog.Warn("failed to unmarshal kindy", "file", file, "error", err)
		return "", "", nil
	}

	return kind.Type, kind.Permalink, kind.Tags
}

func mergeTags(tagMap map[string]internal.TagInfo, t local.KindyType, file string, tags []string) {
	for _, tag := range tags {
		tmp := tagMap[tag]
		tmp.Count++
		if tmp.Permalinks == nil {
			tmp.Permalinks = map[local.KindyType][]string{}
		}
		tmp.Permalinks[t] = append(tmp.Permalinks[t], file)
		tagMap[tag] = tmp
	}
}
