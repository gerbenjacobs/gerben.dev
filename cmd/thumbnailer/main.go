package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gerbenjacobs/resize"
)

const thumbExtension = "_thumb"

func main() {
	dir := flag.String("dir", "", "directory with images that need to be turned into thumbnails")
	recur := flag.Bool("r", false, "recursively look through subdirectories")
	width := flag.Uint("w", 300, "width of the thumbnail")
	flag.Parse()

	if *dir == "" {
		log.Fatalf("the `dir` flag is required to operate")
	}

	fmt.Printf("searching through %q with recursion set to %v\n", *dir, *recur)

	// walk through dir to find photos
	var dirs []string
	var photos []string
	err := filepath.WalkDir(*dir,
		func(path string, info os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				dirs = append(dirs, path)
				// if we're not recursing, skip all subdirectories
				// once the original directory is done
				if path != *dir && !*recur {
					return filepath.SkipAll
				}
				return nil
			}
			photos = append(photos, path)
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	// turn photos into thumbnails
	thumbnails := findMissingThumbnails(photos)

	for _, t := range thumbnails {
		file, err := os.Open(t)
		if err != nil {
			log.Fatalf("failed to open file %q: %v", t, err)
		}
		ext := strings.ToLower(filepath.Ext(t))
		var img image.Image
		switch ext {
		case ".jpg", ".jpeg", ".webp":
			img, err = jpeg.Decode(file)
		case ".png":
			img, err = png.Decode(file)
		default:
			log.Printf("unsupported file type %q\n", t)
			continue
		}
		if err != nil {
			log.Fatalf("failed to decode image %q: %v", t, err)
		}
		m := resize.Resize(*width, 0, img, resize.Lanczos3)
		tf := fileWithoutExtension(t) + thumbExtension + ext
		out, err := os.Create(tf)
		if err != nil {
			log.Fatalf("failed to create thumbnail file %q: %v", t, err)
		}
		defer out.Close()

		// write new image to file
		err = jpeg.Encode(out, m, nil)
		if err != nil {
			log.Fatalf("failed to encode thumbnail image %q: %v", t, err)
		}

		fmt.Printf("created thumbnail %q\n", tf)
	}
}

func findMissingThumbnails(files []string) []string {
	thumbs := make(map[string]struct {
		File    string
		Thumbed bool
	})

	// store all files in the map, thumbnails are set to 'false' by default
	for _, f := range files {
		thumbs[fileWithoutExtension(f)] = struct {
			File    string
			Thumbed bool
		}{File: f, Thumbed: !fileIsThumbnail(f)}
	}

	// find out who needs a thumbnail
	for fx, t := range thumbs {
		if !t.Thumbed {
			// doesn't need to be thumbnailed
			continue
		}

		// look if we have a thumbnail version
		if _, ok := thumbs[fx+thumbExtension]; ok {
			thumbs[fx] = struct {
				File    string
				Thumbed bool
			}{File: t.File, Thumbed: false}
		}
	}

	var thumbFiles []string
	for _, t := range thumbs {
		if t.Thumbed {
			thumbFiles = append(thumbFiles, t.File)
		}
	}

	return thumbFiles
}

func fileWithoutExtension(filePath string) string {
	return strings.TrimSuffix(filePath, filepath.Ext(filePath))
}

func fileIsThumbnail(filePath string) bool {
	return strings.HasSuffix(fileWithoutExtension(filePath), thumbExtension)
}
