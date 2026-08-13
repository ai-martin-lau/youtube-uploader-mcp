package tool

import (
	"context"
	"encoding/json"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"github.com/mark3labs/mcp-go/mcp"
)

type UpdateVideoMetadataTool struct {
	Core *core.Core
}

func (t *UpdateVideoMetadataTool) Name() string {
	return "update_video_metadata"
}

func (t *UpdateVideoMetadataTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Safely update supported YouTube video declarations after verifying ownership. Only explicitly supplied fields are changed; existing status values are preserved."),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("channel_id",
			mcp.Required(),
			mcp.Description("Expected owner channel ID for the video"),
		),
		mcp.WithString("video_id",
			mcp.Required(),
			mcp.Description("The ID of the video to update"),
		),
		mcp.WithBoolean("self_declared_made_for_kids",
			mcp.Description("Optional owner declaration that the video is made for kids"),
		),
		mcp.WithBoolean("contains_synthetic_media",
			mcp.Description("Optional declaration that the video contains realistic altered or synthetic media"),
		),
	)
}

func (t *UpdateVideoMetadataTool) Handle(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	channelID := request.GetString("channel_id", "")
	videoID := request.GetString("video_id", "")
	if channelID == "" || videoID == "" {
		return mcp.NewToolResultError("channel_id and video_id are required"), nil
	}

	madeForKids, err := optionalBool(request, "self_declared_made_for_kids")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	containsSyntheticMedia, err := optionalBool(request, "contains_synthetic_media")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if madeForKids == nil && containsSyntheticMedia == nil {
		return mcp.NewToolResultError("at least one update parameter (self_declared_made_for_kids or contains_synthetic_media) must be provided"), nil
	}

	token, err := validChannelToken(ctx, t.Core, channelID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	video, err := t.Core.UpdateVideoMetadata(ctx, videoID, channelID, core.VideoMetadataPatch{
		SelfDeclaredMadeForKids: madeForKids,
		ContainsSyntheticMedia:  containsSyntheticMedia,
	}, token)
	if err != nil {
		return mcp.NewToolResultError("failed to update video metadata: " + err.Error()), nil
	}

	data, err := json.Marshal(video)
	if err != nil {
		return mcp.NewToolResultError("failed to marshal video metadata: " + err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}
