package internal

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	local "github.com/gerbenjacobs/gerben.dev"
)

func GetKindyByType(kindyType local.KindyType) (entries []local.Kindy, err error) {
	contentPath := local.KindyContentPath + kindyType.URL()
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
