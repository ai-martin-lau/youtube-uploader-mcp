package core

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

// AddSubtitles uploads a subtitle file and attaches it to the given video ID.
// language should be an ISO 639-1 language code (e.g., "en" for English).
func (c *Core) AddSubtitles(ctx context.Context, videoID, subtitlePath, language string, token *oauth2.Token) error {
	if videoID == "" {
		return fmt.Errorf("video ID must be provided")
	}
	if subtitlePath == "" {
		return fmt.Errorf("subtitle path must be provided")
	}
	if language == "" {
		language = "en"
	}
	if token == nil {
		return fmt.Errorf("token must be provided")
	}

	file, err := os.Open(subtitlePath)
	if err != nil {
		return fmt.Errorf("failed to open subtitle file %s: %w", subtitlePath, err)
	}
	defer file.Close()

	service, err := c.Service(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to create YouTube service: %w", err)
	}

	resource := &youtube.Caption{
		Snippet: &youtube.CaptionSnippet{
			Language: language,
			Name:     fmt.Sprintf("Subtitles (%s)", language),
			VideoId:  videoID,
		},
	}

	call := service.Captions.Insert([]string{"snippet"}, resource)
	if _, err := call.Media(file).Do(); err != nil {
		return fmt.Errorf("failed to insert subtitles: %w", err)
	}

	return nil
}

