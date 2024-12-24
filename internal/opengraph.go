package internal

import (
	"net/http"

	"github.com/otiai10/opengraph/v2"
)

func Opengraph(url string) (*opengraph.OpenGraph, error) {
	ogp := opengraph.New(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := ogp.Parse(resp.Body); err != nil {
		return nil, err
	}

	return ogp, nil
}
