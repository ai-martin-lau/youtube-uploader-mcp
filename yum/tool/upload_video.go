package tool

import (
	"context"
	"encoding/json"
	"os"
	"strings"

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
		mcp.WithDescription("Upload a video to YouTube, taking advantages of AI to generate descriptions, title and tags. To update a video with additional properties like playlist, subtitles, thumbnail, use the update_video tool."),
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
			mcp.Description("Optional language of the video's audio/content (ISO 639-1). If not set, YouTube's default or auto-detection is used."),
		),
		mcp.WithString("status",
			mcp.Description("status of video, could be any of unlisted, public, private. Default is private"),
		),
		mcp.WithString("publish_at",
			mcp.Description("The date and time when the video is scheduled to publish. It can be set only if the privacy status of the video is private. The value is specified in ISO 8601 format (YYYY-MM-DDThh:mm:ss.sZ)."),
		),
		mcp.WithBoolean("made_for_kids",
			mcp.Description("Optional owner declaration that the video is made for kids. Omit this field to leave the declaration unspecified."),
		),
		mcp.WithBoolean("contains_synthetic_media",
			mcp.Description("Optional declaration that the video contains realistic altered or synthetic media. Omit this field to leave the declaration unspecified."),
		),
		mcp.WithBoolean("notify_subscribers",
			mcp.Description("Optional control for publishing the upload to subscribers' feeds and sending notifications. YouTube defaults to true when this field is omitted."),
		),
	)
}

func optionalBool(request mcp.CallToolRequest, key string) (*bool, error) {
	if _, ok := request.GetArguments()[key]; !ok {
		return nil, nil
	}

	value, err := request.RequireBool(key)
	if err != nil {
		return nil, err
	}
	return &value, nil
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
	if status != "public" && status != "private" && status != "unlisted" {
		return mcp.NewToolResultError("status must be one of: public, private, unlisted"), nil
	}
	madeForKids, err := optionalBool(request, "made_for_kids")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	containsSyntheticMedia, err := optionalBool(request, "contains_synthetic_media")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	notifySubscribers, err := optionalBool(request, "notify_subscribers")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	videoLanguage := request.GetString("video_language", "")

	// Fail-fast validation of the video file path
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError("video file path does not exist: " + filePath), nil
		}
		return mcp.NewToolResultError("failed to check video file path: " + err.Error()), nil
	}
	if info.IsDir() {
		return mcp.NewToolResultError("video file path is a directory: " + filePath), nil
	}

	token, err := validChannelToken(ctx, t.Core, channelId)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	video := &core.Video{
		Path:                   filePath,
		Title:                  title,
		Description:            description,
		Tags:                   strings.Split(tags, ","),
		CategoryID:             categoryID,
		Language:               videoLanguage,
		PrivacyStatus:          status,
		MadeForKids:            madeForKids,
		ContainsSyntheticMedia: containsSyntheticMedia,
		NotifySubscribers:      notifySubscribers,
		PublishAt:              request.GetString("publish_at", ""),
	}
	id, err := t.Core.UploadVideo(ctx, video, token)
	if err != nil {
		return mcp.NewToolResultError("Failed to upload video: " + err.Error()), nil
	}
	video.ID = id

	bytes, err := json.Marshal(video)
	if err != nil {
		return mcp.NewToolResultError("failed to marshal the video: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(bytes)), nil
}
