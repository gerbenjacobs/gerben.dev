package internal

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	local "github.com/gerbenjacobs/gerben.dev"
	"github.com/gerbenjacobs/resize"
)

const (
	MaxImageWidth  = 1280
	MaxThumbWidth  = 300
	ThumbExtension = "_thumb"
)

func HandleUploadedFile(file multipart.File, header *multipart.FileHeader) (*os.File, error) {
	img, fileType, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// resize image if it's too large
	if img.Bounds().Dx() > MaxImageWidth {
		img = resize.Resize(MaxImageWidth, 0, img, resize.Lanczos3)
	}

	dst, err := os.Create(local.KindyContentPath + "data/photos/" + header.Filename)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if err := encodeImage(dst, img, fileType); err != nil {
		return nil, err
	}

	// create thumbnail
	if err := createThumbnail(dst.Name(), img, fileType); err != nil {
		return nil, err
	}

	return dst, nil
}

func createThumbnail(fp string, img image.Image, fileType string) error {
	fpThumb := fileWithoutExtension(fp) + ThumbExtension + filepath.Ext(fp)
	dst, err := os.Create(fpThumb)
	if err != nil {
		return err
	}
	defer dst.Close()

	img = resize.Resize(MaxThumbWidth, 0, img, resize.Lanczos3)

	return encodeImage(dst, img, fileType)
}

func encodeImage(file *os.File, img image.Image, fileType string) error {
	switch fileType {
	case "jpg", "jpeg", "webp":
		return jpeg.Encode(file, img, nil)
	case "png":
		return png.Encode(file, img)
	default:
		return fmt.Errorf("unsupported file type %q", fileType)
	}
}

func fileWithoutExtension(filePath string) string {
	return strings.TrimSuffix(filePath, filepath.Ext(filePath))
}
