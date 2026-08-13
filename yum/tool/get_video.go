package tool

import (
	"context"
	"encoding/json"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"github.com/mark3labs/mcp-go/mcp"
)

type GetVideoTool struct {
	Core *core.Core
}

func (t *GetVideoTool) Name() string {
	return "get_video"
}

func (t *GetVideoTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Read the current YouTube metadata and status for a video after verifying that it belongs to the expected authenticated channel."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("channel_id",
			mcp.Required(),
			mcp.Description("Expected owner channel ID for the video"),
		),
		mcp.WithString("video_id",
			mcp.Required(),
			mcp.Description("The ID of the video to inspect"),
		),
	)
}

func (t *GetVideoTool) Handle(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	channelID := request.GetString("channel_id", "")
	videoID := request.GetString("video_id", "")
	if channelID == "" || videoID == "" {
		return mcp.NewToolResultError("channel_id and video_id are required"), nil
	}

	token, err := validChannelToken(ctx, t.Core, channelID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	video, err := t.Core.GetVideo(ctx, videoID, channelID, token)
	if err != nil {
		return mcp.NewToolResultError("failed to get video: " + err.Error()), nil
	}

	data, err := json.Marshal(video)
	if err != nil {
		return mcp.NewToolResultError("failed to marshal video metadata: " + err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}
