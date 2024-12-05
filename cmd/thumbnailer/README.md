# Thumbnailer

`thumbnailer` is a simple command line tool to generate thumbnails from images.

It can go recursively through a directory and generate thumbnails for all images it finds.

Thumbnails are stored with the same name as the original image, but with a `_thumb` suffix.

It has a configurable limit for `width` and keeps the aspect ratio of the original image.
As such it might not be a good fit if you have high/long images.

## Limitations

- Only supports JPEG (WEBP) and PNG images.
- Resizes based on width only, but keeps aspect ratio.
- Stores thumbnails in the same directory as the original image with a `_thumb` suffix.
- Use as-is, no warranties.

## Installing

While it's possible for me to provide pre-built binaries, I'd rather not.

Make sure you have [Go installed](https://go.dev/dl/) and run:

```shell
go install github.com/gerbenjacobs/gerben.dev/cmd/thumbnailer
```

This will install the `thumbnailer` command in your `$GOPATH/bin` directory.

## Usage

```shell
thumbnailer -dir /path/to/images -w 300 -r -dry-run
```

```plaintext
  -dir string
        directory with images that need to be turned into thumbnails
  -dry-run
        do not create thumbnails, just list the files that would be created
  -r    recursively look through subdirectories
  -w uint
        width of the thumbnail (default 300)
```

If you're happy with the results, remove the `-dry-run` flag to actually generate the thumbnails.

You can also specify the width of the thumbnail with the `-w` flag, default is 300 pixels.

If you only want to generate thumbnails for the directory you specified, remove the `-r` flag.
