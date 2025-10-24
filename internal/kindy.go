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
