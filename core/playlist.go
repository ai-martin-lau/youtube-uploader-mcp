package core

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

// AddVideoToPlaylist adds the given video ID to the specified playlist.
func (c *Core) AddVideoToPlaylist(ctx context.Context, playlistID, videoID string, token *oauth2.Token) error {
	if playlistID == "" {
		return fmt.Errorf("playlist ID must be provided")
	}
	if videoID == "" {
		return fmt.Errorf("video ID must be provided")
	}
	if token == nil {
		return fmt.Errorf("token must be provided")
	}

	service, err := c.Service(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to create YouTube service: %w", err)
	}

	resource := &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistID,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: videoID,
			},
		},
	}

	call := service.PlaylistItems.Insert([]string{"snippet"}, resource)
	if _, err := call.Do(); err != nil {
		return fmt.Errorf("failed to insert playlist item: %w", err)
	}

	return nil
}

