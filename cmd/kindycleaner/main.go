package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	local "github.com/gerbenjacobs/gerben.dev"
)

// removeAuthorStruct removes the Author struct from Kindy JSON files in the specified directory.
func removeAuthorStruct(path string, d fs.DirEntry, err error) error {
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

	fmt.Println("Removed author from ", path)
	return nil
}

// moveToYearlyFolder moves Kindy JSON files into folders based on their published year.
// TODO: It does not move the content yet
func moveToYearlyFolder(path string, d fs.DirEntry, err error) error {
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

	// check if already in yearly folder
	year := kindy.PublishedAt.Year()
	oldPath := path
	//fmt.Println(oldPath, strings.HasSuffix(filepath.Dir(oldPath), fmt.Sprintf("%d", year)))
	if strings.HasSuffix(filepath.Dir(oldPath), fmt.Sprintf("%d", year)) {
		return nil
	}

	//return nil

	// move file to yearly folder
	newPath := filepath.Join(filepath.Dir(path), fmt.Sprintf("%d", year), filepath.Base(path))
	kindy.Slug = fmt.Sprintf("%d/%s", year, kindy.Slug)

	updatedContent, err := json.MarshalIndent(kindy, "", "  ")
	if err != nil {
		return err
	}
	// create yearly folder if not exists
	if err := os.MkdirAll(filepath.Dir(newPath), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(newPath, updatedContent, 0644); err != nil {
		return err
	}
	// remove old file
	if err := os.Remove(oldPath); err != nil {
		return err
	}

	fmt.Println("Moved to yearly folder ", newPath)
	return nil
}

func main() {
	dir := flag.String("dir", "", "directory with images that need to be turned into thumbnails")
	flag.Parse()

	if *dir == "" {
		log.Fatalf("the `dir` flag is required to operate")
	}

	fmt.Println("Executing function in", *dir)
	err := filepath.WalkDir(*dir, func(path string, d fs.DirEntry, err error) error {
		// return removeAuthorStruct(path, d, err)
		return moveToYearlyFolder(path, d, err)
	})
	if err != nil {
		log.Fatalf("error walking the directory: %v", err)
	}
}
