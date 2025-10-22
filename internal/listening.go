package internal

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

func GetListeningData(downloadFresh bool) (*gofeed.Feed, error) {
	feedUrl := "https://lfm.xiffy.nl/theonewithout"
	cacheFile := ".cache/listening.xml"

	// check if cache file exists and is not older than 10 minutes
	expiry := 10 * time.Minute
	if !downloadFresh {
		expiry = 0
	}
	b, err := GetCache(cacheFile, expiry)
	if err != nil {
		slog.Warn("downloading new listening feed")
		resp, err := http.Get(feedUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to get feed url: %w", err)
		}
		b, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read feed: %w", err)
		}
		SetCache(cacheFile, b)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}
	return feed, nil
}
