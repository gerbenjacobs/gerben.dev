# gerben.dev

My personal tech-related webpage.

## Contributing

If you find a spelling error or want to add something cool, feel free to contribute!

## Operations

### Images

#### Instagram Archiver

To archive Instagram images, run the following command:

```shell
go run cmd/igarchiver/main.go -workdir /path/to/your/instagram/archive -outputdir content/kindy/data/photos/media
```

```plaintext
  -outputdir string
        Output directory i.e. your Kindy content folder
  -postsfile string
        JSON file containing your posts, as seen from workdir (default "/your_instagram_activity/content/posts_1.json")
  -workdir string
        Working directory of your Instagram archive
```

Where the `workdir` is the root directory where you extracted your Instagram archive.

#### Thumbnails

By default, we should put photos in `content/kindy/data/photos`.

Images that are not meant to be shown on the /photos page should live `content/kindy/data/images`.

Generate thumbnails for images in a directory, by running the following command:

```shell
go run cmd/thumbnailer/main.go -r -dir ./content/kindy/data/photos
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

#### Resizing photos

(TODO)

Ideally we don't store photos larger than 1280x720. To resize photos, run the following command:

```shell
go run cmd/resizer/main.go
```
