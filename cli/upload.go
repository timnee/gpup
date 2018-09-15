package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/int128/gpup/photos"
)

func (c *CLI) upload(ctx context.Context) error {
	if len(c.Paths) == 0 {
		return fmt.Errorf("Nothing to upload")
	}
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}
	mediaItems, err := findMediaItems(c.Paths, client)
	if err != nil {
		return err
	}
	if len(mediaItems) == 0 {
		return fmt.Errorf("Nothing to upload in %s", strings.Join(c.Paths, ", "))
	}
	log.Printf("The following %d items will be uploaded:", len(mediaItems))
	for i, mediaItem := range mediaItems {
		fmt.Printf("%3d: %s\n", i+1, mediaItem)
	}

	service, err := photos.New(client)
	if err != nil {
		return err
	}
	switch {
	case c.AlbumTitle != "":
		return service.AddToAlbum(ctx, c.AlbumTitle, mediaItems)
	case c.NewAlbum != "":
		_, err = service.CreateAlbum(ctx, c.NewAlbum, mediaItems)
		return err
	default:
		return service.AddToLibrary(ctx, mediaItems)
	}
}

func findMediaItems(args []string, client *http.Client) ([]photos.MediaItem, error) {
	mediaItems := make([]photos.MediaItem, 0)
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://"):
			r, err := http.NewRequest("GET", arg, nil)
			if err != nil {
				return nil, fmt.Errorf("Could not parse URL: %s", err)
			}
			mediaItems = append(mediaItems, &photos.HTTPMediaItem{
				Client:  client,
				Request: r,
			})
		default:
			if err := filepath.Walk(arg, func(name string, info os.FileInfo, err error) error {
				switch {
				case err != nil:
					return err
				case info.Mode().IsRegular():
					mediaItems = append(mediaItems, photos.FileMediaItem(name))
					return nil
				default:
					return nil
				}
			}); err != nil {
				return nil, fmt.Errorf("Error while finding files in %s: %s", arg, err)
			}
		}
	}
	return mediaItems, nil
}
