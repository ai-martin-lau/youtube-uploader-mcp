package core

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
)

// SetThumbnail uploads a custom thumbnail for the given video ID.
func (c *Core) SetThumbnail(ctx context.Context, videoID, thumbnailPath string, token *oauth2.Token) error {
	if videoID == "" {
		return fmt.Errorf("video ID must be provided")
	}
	if thumbnailPath == "" {
		return fmt.Errorf("thumbnail path must be provided")
	}
	if token == nil {
		return fmt.Errorf("token must be provided")
	}

	file, err := os.Open(thumbnailPath)
	if err != nil {
		return fmt.Errorf("failed to open thumbnail file %s: %w", thumbnailPath, err)
	}
	defer file.Close()

	service, err := c.Service(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to create YouTube service: %w", err)
	}

	call := service.Thumbnails.Set(videoID)
	if _, err := call.Media(file).Do(); err != nil {
		return fmt.Errorf("failed to set thumbnail: %w", err)
	}

	return nil
}

