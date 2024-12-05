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
	dryrun := flag.Bool("dry-run", false, "do not create thumbnails, just list the files that would be created")
	flag.Parse()

	if *dir == "" {
		log.Fatalf("the `dir` flag is required to operate")
	}

	fmt.Printf("searching through %q with recursion set to %v\n", *dir, *recur)

	photos, err := findPhotos(*dir, *recur)
	if err != nil {
		log.Fatalf("error finding photos: %v", err)
	}

	thumbnails := findMissingThumbnails(photos)
	for _, t := range thumbnails {
		if *dryrun {
			fmt.Printf("would create thumbnail for %q\n", t)
			continue
		}
		if err := createThumbnail(t, *width); err != nil {
			log.Printf("error creating thumbnail for %q: %v", t, err)
		}
	}
}

func findPhotos(dir string, recur bool) ([]string, error) {
	var photos []string
	err := filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != dir && !recur {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			photos = append(photos, path)
		}
		return nil
	})
	return photos, err
}

func createThumbnail(filePath string, width uint) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %q: %v", filePath, err)
	}
	defer file.Close()

	img, err := decodeImage(file, filePath)
	if err != nil {
		return fmt.Errorf("failed to decode image %q: %v", filePath, err)
	}

	m := resize.Resize(width, 0, img, resize.Lanczos3)
	tf := fileWithoutExtension(filePath) + thumbExtension + filepath.Ext(filePath)
	out, err := os.Create(tf)
	if err != nil {
		return fmt.Errorf("failed to create thumbnail file %q: %v", filePath, err)
	}
	defer out.Close()

	if err := jpeg.Encode(out, m, nil); err != nil {
		return fmt.Errorf("failed to encode thumbnail image %q: %v", filePath, err)
	}

	fmt.Printf("created thumbnail %q\n", tf)
	return nil
}

func decodeImage(file *os.File, filePath string) (image.Image, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg", ".webp":
		return jpeg.Decode(file)
	case ".png":
		return png.Decode(file)
	default:
		return nil, fmt.Errorf("unsupported file type %q", filePath)
	}
}

func findMissingThumbnails(files []string) []string {
	thumbs := make(map[string]struct {
		File    string
		Thumbed bool
	})

	for _, f := range files {
		thumbs[fileWithoutExtension(f)] = struct {
			File    string
			Thumbed bool
		}{File: f, Thumbed: !fileIsThumbnail(f)}
	}

	for fx, t := range thumbs {
		if !t.Thumbed {
			continue
		}
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
