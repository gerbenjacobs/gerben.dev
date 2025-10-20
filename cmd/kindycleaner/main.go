package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	local "github.com/gerbenjacobs/gerben.dev"
)

func main() {
	dir := flag.String("dir", "", "directory with images that need to be turned into thumbnails")
	flag.Parse()

	if *dir == "" {
		log.Fatalf("the `dir` flag is required to operate")
	}

	println("Removing Author struct in directory:", *dir)

	filepath.WalkDir(*dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var kindy local.Kindy
		if err := json.Unmarshal(b, &kindy); err != nil {
			return err
		}

		// Remove Author block
		if kindy.Author == nil {
			return nil
		}
		kindy.Author = nil

		updatedContent, err := json.MarshalIndent(kindy, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(path, updatedContent, 0644); err != nil {
			return err
		}

		fmt.Println("Updated ", path)

		return nil
	})
}
