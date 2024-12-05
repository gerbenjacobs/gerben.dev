# gerben.dev

My personal tech-related webpage.

## Contributing

If you find a spelling error or want to add something cool, feel free to contribute!

## Operations

### Images

#### Thumbnails

By default, we should put photos in `content/kindy/data/photos`.

Images that are not meant to be shown on the /photos page should live `content/kindy/data/images`.

Generate thumbnails for images in a directory, by running the following command:

```shell
go run cmd/thumbnailer/main.go -r -dir ./content/kindy/data/photos
```

```
  -dir string
        directory with images that need to be turned into thumbnails
  -r    recursively look through subdirectories
  -w uint
        width of the thumbnail (default 300)
```

#### Resizing photos

(TODO)

Ideally we don't store photos larger than 1280x720. To resize photos, run the following command:

```shell
go run cmd/resizer/main.go
```
