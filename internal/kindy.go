package internal

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	local "github.com/gerbenjacobs/gerben.dev"
)

func GetKindyByType(kindyType local.KindyType) (entries []local.Kindy, err error) {
	contentPath := local.KindyContentPath + kindyType.URL()

	err = filepath.WalkDir(contentPath, func(path string, f os.DirEntry, walkErr error) error {
		// Do folder walking, file reading and JSON unmarshalling
		if !strings.HasSuffix(f.Name(), ".json") {
			// skip all non .json files
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var tmp local.Kindy
		if err := json.Unmarshal(b, &tmp); err != nil {
			return err
		}
		entries = append(entries, tmp)
		return nil
	})
	if err != nil {
		slog.Error("failed to walk kindy content", "error", err)
		return nil, err
	}

	// Sort the entries on published date
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].PublishedAt.After(entries[j].PublishedAt)
	})

	return entries, nil
}

// GetKindyPaths loads multiple Kindy entries from given file paths
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

func GetKindyNeighbours(t local.KindyType, slug string) (prev string, next string, err error) {
	var entries []local.Kindy
	if t.IsTimelineType() {
		likes, _ := GetKindyCacheByType(local.KindyTypeLike)
		entries = append(entries, likes...)
		replies, _ := GetKindyCacheByType(local.KindyTypeReplies)
		entries = append(entries, replies...)
		reposts, _ := GetKindyCacheByType(local.KindyTypeRepost)
		entries = append(entries, reposts...)
		notes, _ := GetKindyCacheByType(local.KindyTypeNote)
		entries = append(entries, notes...)
		// Sort the entries on published date
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].PublishedAt.After(entries[j].PublishedAt)
		})
	} else {
		entries, err = GetKindyCacheByType(t)
		if err != nil {
			return "", "", err
		}
	}

	var index int
	for i, entry := range entries {
		if entry.Slug == slug {
			index = i
			break
		}
	}

	if index > 0 {
		prev = entries[index-1].Permalink
	}
	if index < len(entries)-1 {
		next = entries[index+1].Permalink
	}

	return prev, next, nil
}
