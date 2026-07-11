package tool

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"github.com/mark3labs/mcp-go/mcp"
)

type UploadVideoTool struct {
	Core *core.Core
}

func (t *UploadVideoTool) Name() string {
	return "upload_video"
}

func (t *UploadVideoTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Upload a video to YouTube, taking advantages of AI to generate descriptions, title and tags"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the video file"),
		),
		mcp.WithString("channel_id",
			mcp.Required(),
			mcp.Description("Channel ID to upload the video to, if not provided, Agent should call tool channels to get the list of channels and ask the user to select one"),
		),
		mcp.WithString("description",
			mcp.Required(),
			mcp.Description("Description of the video, if not provided, Agent should generate a description based on the video content"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Title of the video, if not provided, Agent should generate a title based on the video description"),
		),
		mcp.WithString("tags",
			mcp.Required(),
			mcp.Description("Tags for the video, if not provided, Agent should generate tags based on the video description"),
		),
		mcp.WithString("category_id",
			mcp.Required(),
			mcp.Description("Category ID for the video, if not provided, Agent should generate a category based on the video description"),
		),
		mcp.WithString("video_language",
			mcp.Description("Language of the video's audio/content (ISO 639-1). Default is en (English)."),
		),
		mcp.WithString("status",
			mcp.Description("status of video, could be any of unlisted, public, private. Default is private"),
		),
		mcp.WithString("publish_at",
			mcp.Description("The date and time when the video is scheduled to publish. It can be set only if the privacy status of the video is private. The value is specified in ISO 8601 format (YYYY-MM-DDThh:mm:ss.sZ)."),
		),
		mcp.WithString("playlist_id",
			mcp.Description("Optional playlist ID to which the uploaded video should be added"),
		),
		mcp.WithString("subtitle_path",
			mcp.Description("Optional path to a subtitle file (e.g., .srt or .vtt) to attach to the uploaded video"),
		),
		mcp.WithString("subtitle_language",
			mcp.Description("Language code of the subtitle track (ISO 639-1). Default is en (English)."),
		),
		mcp.WithString("thumbnail_path",
			mcp.Description("Optional path to an image file to use as the video's custom thumbnail"),
		),
		mcp.WithBoolean("made_for_kids",
			mcp.Description("Whether the video is made exclusively for kids. Default is false"),
		),
	)
}

func (t *UploadVideoTool) Handle(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {

	filePath := request.GetString("file_path", "")
	description := request.GetString("description", "")
	title := request.GetString("title", "")
	tags := request.GetString("tags", "")
	categoryID := request.GetString("category_id", "")
	if filePath == "" || description == "" || title == "" || tags == "" || categoryID == "" {
		return mcp.NewToolResultError(
			"all fields are required: file_path, description, title, tags, category_id"), nil
	}
	channelId := request.GetString("channel_id", "")
	if channelId == "" {
		return mcp.NewToolResultError("channel_id is required to upload video"), nil
	}
	status := request.GetString("status", "private")
	madeForKids := request.GetBool("made_for_kids", false)
	playlistID := request.GetString("playlist_id", "")
	subtitlePath := request.GetString("subtitle_path", "")
	subtitleLanguage := request.GetString("subtitle_language", "en")
	videoLanguage := request.GetString("video_language", "en")
	thumbnailPath := request.GetString("thumbnail_path", "")

	channel, err := t.Core.GetChannelByID(channelId)
	if err != nil {
		return mcp.NewToolResultError("Failed to load token: " + err.Error()), nil
	}

	if channel == nil ||
		channel.Token == nil {
		return mcp.NewToolResultError("channel or token is nil, please authenticate first"), nil
	}

	if channel.Token.Expiry.IsZero() ||
		channel.Token.AccessToken == "" ||
		channel.Token.RefreshToken == "" {
		return mcp.NewToolResultError(
			"channel token is expired or malformed, please start authenticate"), nil
	}

	// Check if token is expiring (within 2 minutes)
	now := time.Now().In(channel.Token.Expiry.Location())
	if channel.Token.Expiry.Before(now.Add(2 * time.Minute)) {
		newToken, err := t.Core.RefreshAccessToken(channel.Token)
		if err != nil {
			return mcp.NewToolResultError(
				"token was expiring, Failed to refresh token: " + err.Error()), nil
		}
		channel.Token = newToken
		// Optionally save the refreshed token for future use
		err = t.Core.SaveChannel(channel)
		if err != nil {
			return mcp.NewToolResultError(
				"token was expiring, Failed to save refreshed token: " + err.Error()), nil
		}
	}

	video := &core.Video{
		Path:          filePath,
		Title:         title,
		Description:   description,
		Tags:          strings.Split(tags, ","),
		CategoryID:    categoryID,
		Language:      videoLanguage,
		PrivacyStatus: status,
		MadeForKids:   madeForKids,
		PublishAt:     request.GetString("publish_at", ""),
	}
	id, err := t.Core.UploadVideo(ctx, video, channel.Token)
	if err != nil {
		return mcp.NewToolResultError("Failed to upload video: " + err.Error()), nil
	}
	video.ID = id

	if playlistID != "" {
		if err := t.Core.AddVideoToPlaylist(ctx, playlistID, id, channel.Token); err != nil {
			return mcp.NewToolResultError("Failed to add video to playlist: " + err.Error()), nil
		}
	}

	if subtitlePath != "" {
		if err := t.Core.AddSubtitles(ctx, id, subtitlePath, subtitleLanguage, channel.Token); err != nil {
			return mcp.NewToolResultError("Failed to add subtitles: " + err.Error()), nil
		}
	}

	if thumbnailPath != "" {
		if err := t.Core.SetThumbnail(ctx, id, thumbnailPath, channel.Token); err != nil {
			return mcp.NewToolResultError("Failed to set thumbnail: " + err.Error()), nil
		}
	}

	bytes, err := json.Marshal(video)
	if err != nil {
		return mcp.NewToolResultError("failed to marshal the video: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(bytes)), nil
}
