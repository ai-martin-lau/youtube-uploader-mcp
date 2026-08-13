package tests

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/therewardstore/httpmatter"
)

func (s *YumSuite) mockVerifiedChannel() {
	s.mock.Add("verify_channel_request", "verify_channel_response").Respond(nil)
}

func (s *YumSuite) TestUploadVideo() {
	s.mockVerifiedChannel()
	reqAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Equal("https://youtube.googleapis.com/upload/youtube/v3/videos?alt=json&part=snippet&part=status&prettyPrint=false&uploadType=multipart", req.URL.String())
		s.Equal("Bearer mock-access-token", req.Header.Get("Authorization"))

		mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		s.NoError(err)
		s.True(strings.HasPrefix(mediaType, "multipart/"))
		boundary := params["boundary"]

		mr := multipart.NewReader(req.Body, boundary)

		// Part 1: JSON Metadata
		p1, err := mr.NextPart()
		s.NoError(err)
		s.Equal("application/json", p1.Header.Get("Content-Type"))

		b, err := io.ReadAll(p1)
		s.NoError(err)

		expectedJSON := `{
			"snippet": {
				"title": "mock-title",
				"description": "mock-description",
				"tags": ["mock-tag1", "mock-tag2"],
				"categoryId": "mock-category-id"
			},
			"status": {
				"privacyStatus": "unlisted"
			}
		}`
		s.JSONEq(expectedJSON, string(b))

		// Part 2: Video Content
		p2, err := mr.NextPart()
		s.NoError(err)
		content, err := io.ReadAll(p2)
		s.NoError(err)
		s.NotEmpty(content)

		return 0
	}

	s.mock.Add("upload_video_request", "upload_video_response").Respond(
		httpmatter.RequestResponse(reqAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "upload_video",
			"arguments": mcp.Params{
				"channel_id":  "mock-channel-id",
				"file_path":   "./data/videos/video_1.mp4",
				"description": "mock-description",
				"title":       "mock-title",
				"tags":        "mock-tag1,mock-tag2",
				"category_id": "mock-category-id",
				"status":      "unlisted",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `{"id":"video_id_12345","path":"./data/videos/video_1.mp4","title":"mock-title","description":"mock-description","tags":["mock-tag1","mock-tag2"],"category_id":"mock-category-id","language":"","privacy_status":"unlisted"}`)
}

func (s *YumSuite) TestUploadVideoExplicitFalseControls() {
	s.mockVerifiedChannel()
	reqAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Equal("https://youtube.googleapis.com/upload/youtube/v3/videos?alt=json&notifySubscribers=false&part=snippet&part=status&prettyPrint=false&uploadType=multipart", req.URL.String())

		mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		s.NoError(err)
		s.True(strings.HasPrefix(mediaType, "multipart/"))

		mr := multipart.NewReader(req.Body, params["boundary"])
		metadata, err := mr.NextPart()
		s.NoError(err)
		body, err := io.ReadAll(metadata)
		s.NoError(err)

		s.JSONEq(`{
			"snippet": {
				"title": "mock-title",
				"description": "mock-description",
				"tags": ["mock-tag1", "mock-tag2"],
				"categoryId": "mock-category-id"
			},
			"status": {
				"privacyStatus": "private",
				"selfDeclaredMadeForKids": false,
				"containsSyntheticMedia": false
			}
		}`, string(body))
		return 0
	}

	s.mock.Add("upload_video_controls_request", "upload_video_response").Respond(
		httpmatter.RequestResponse(reqAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "upload_video",
			"arguments": mcp.Params{
				"channel_id":               "mock-channel-id",
				"file_path":                "./data/videos/video_1.mp4",
				"description":              "mock-description",
				"title":                    "mock-title",
				"tags":                     "mock-tag1,mock-tag2",
				"category_id":              "mock-category-id",
				"status":                   "private",
				"made_for_kids":            false,
				"contains_synthetic_media": false,
				"notify_subscribers":       false,
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)
	s.Contains(text.Text, `"made_for_kids":false`)
	s.Contains(text.Text, `"contains_synthetic_media":false`)
	s.Contains(text.Text, `"notify_subscribers":false`)
}

func (s *YumSuite) TestUploadVideoWithLanguage() {
	s.mockVerifiedChannel()
	reqAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Equal("https://youtube.googleapis.com/upload/youtube/v3/videos?alt=json&part=snippet&part=status&prettyPrint=false&uploadType=multipart", req.URL.String())

		mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		s.NoError(err)
		s.True(strings.HasPrefix(mediaType, "multipart/"))
		boundary := params["boundary"]

		mr := multipart.NewReader(req.Body, boundary)

		// Part 1: JSON Metadata
		p1, err := mr.NextPart()
		s.NoError(err)
		b, err := io.ReadAll(p1)
		s.NoError(err)

		expectedJSON := `{
			"snippet": {
				"title": "mock-title",
				"description": "mock-description",
				"tags": ["mock-tag1", "mock-tag2"],
				"categoryId": "mock-category-id",
				"defaultLanguage": "fr",
				"defaultAudioLanguage": "fr"
			},
			"status": {
				"privacyStatus": "unlisted"
			}
		}`
		s.JSONEq(expectedJSON, string(b))

		return 0
	}

	s.mock.Add("upload_video_request", "upload_video_response").Respond(
		httpmatter.RequestResponse(reqAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "upload_video",
			"arguments": mcp.Params{
				"channel_id":     "mock-channel-id",
				"file_path":      "./data/videos/video_1.mp4",
				"description":    "mock-description",
				"title":          "mock-title",
				"tags":           "mock-tag1,mock-tag2",
				"category_id":    "mock-category-id",
				"video_language": "fr",
				"status":         "unlisted",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `"language":"fr"`)
}

func (s *YumSuite) TestUploadScheduledVideo() {
	s.mockVerifiedChannel()
	reqAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Equal("https://youtube.googleapis.com/upload/youtube/v3/videos?alt=json&part=snippet&part=status&prettyPrint=false&uploadType=multipart", req.URL.String())
		s.Equal("Bearer mock-access-token", req.Header.Get("Authorization"))

		mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		s.NoError(err)
		s.True(strings.HasPrefix(mediaType, "multipart/"))
		boundary := params["boundary"]

		mr := multipart.NewReader(req.Body, boundary)

		// Part 1: JSON Metadata
		p1, err := mr.NextPart()
		s.NoError(err)
		b, err := io.ReadAll(p1)
		s.NoError(err)

		expectedJSON := `{
			"snippet": {
				"title": "mock-title",
				"description": "mock-description",
				"tags": ["mock-tag1", "mock-tag2"],
				"categoryId": "mock-category-id"
			},
			"status": {
				"privacyStatus": "private",
				"publishAt": "2026-01-20T12:00:00Z"
			}
		}`
		s.JSONEq(expectedJSON, string(b))

		return 0
	}

	s.mock.Add("upload_video_request", "upload_video_response").Respond(
		httpmatter.RequestResponse(reqAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "upload_video",
			"arguments": mcp.Params{
				"channel_id":  "mock-channel-id",
				"file_path":   "./data/videos/video_1.mp4",
				"description": "mock-description",
				"title":       "mock-title",
				"tags":        "mock-tag1,mock-tag2",
				"category_id": "mock-category-id",
				"status":      "public",
				"publish_at":  "2026-01-20T12:00:00Z",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `"publish_at":"2026-01-20T12:00:00Z"`)
	s.Contains(text.Text, `"privacy_status":"private"`)
}

func (s *YumSuite) TestGetVideoSuccess() {
	s.mockVerifiedChannel()
	requestAssert := func(req *http.Request) int {
		s.Equal("GET", req.Method)
		s.Equal("Bearer mock-access-token", req.Header.Get("Authorization"))
		return 0
	}
	s.mock.Add("get_video_request", "get_video_response").Respond(
		httpmatter.RequestResponse(requestAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "get_video",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
				"video_id":   "video_id_12345",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)
	s.JSONEq(`{
		"id":"video_id_12345",
		"channel_id":"mock-channel-id",
		"title":"Mock video title",
		"description":"Mock video description",
		"tags":["bass","play-along"],
		"category_id":"10",
		"default_language":"en",
		"default_audio_language":"en",
		"privacy_status":"private",
		"publish_at":"2026-08-15T22:00:00Z",
		"license":"youtube",
		"embeddable":true,
		"public_stats_viewable":false,
		"made_for_kids":false,
		"self_declared_made_for_kids":false,
		"contains_synthetic_media":false,
		"has_paid_product_placement":false,
		"recording_date":"2026-08-13T00:00:00Z"
	}`, text.Text)
}

func (s *YumSuite) TestGetVideoRejectsOwnerMismatch() {
	s.mockVerifiedChannel()
	s.mock.Add("get_video_request", "get_video_owner_mismatch_response").Respond(nil)
	s.mock.Init()

	result, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "get_video",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
				"video_id":   "video_id_12345",
			},
		}).
		Call(s.Ctx(), 1, true)
	s.NoError(err)
	text, ok := result.Content[0].(mcp.TextContent)
	s.True(ok)
	s.Contains(text.Text, "video video_id_12345 belongs to channel other-channel-id, not requested channel mock-channel-id")
}

func (s *YumSuite) TestUpdateVideoMetadataRejectsEmptyPatch() {
	result, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video_metadata",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
				"video_id":   "video_id_12345",
			},
		}).
		Call(s.Ctx(), 1, true)
	s.NoError(err)
	text, ok := result.Content[0].(mcp.TextContent)
	s.True(ok)
	s.Equal("at least one update parameter (self_declared_made_for_kids or contains_synthetic_media) must be provided", text.Text)
}

func (s *YumSuite) TestUpdateVideoMetadataReadMergeUpdateRead() {
	s.mockVerifiedChannel()
	readAssert := func(req *http.Request) int {
		s.Equal("GET", req.Method)
		return 0
	}
	updateAssert := func(req *http.Request) int {
		s.Equal("PUT", req.Method)
		s.Equal("Bearer mock-access-token", req.Header.Get("Authorization"))
		s.Equal("etag-before-update", req.Header.Get("If-Match"))
		body, err := io.ReadAll(req.Body)
		s.NoError(err)
		s.JSONEq(`{
			"id":"video_id_12345",
			"status":{
				"privacyStatus":"private",
				"publishAt":"2026-08-15T22:00:00Z",
				"license":"youtube",
				"embeddable":true,
				"publicStatsViewable":false,
				"selfDeclaredMadeForKids":false,
				"containsSyntheticMedia":false
			}
		}`, string(body))
		return 0
	}

	s.mock.Add("get_video_request", "get_video_before_update_response").Respond(
		httpmatter.RequestResponse(readAssert))
	s.mock.Add("get_video_request", "get_video_after_update_response").Respond(
		httpmatter.RequestResponse(readAssert))
	s.mock.Add("update_video_metadata_request", "update_video_metadata_response").Respond(
		httpmatter.RequestResponse(updateAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video_metadata",
			"arguments": mcp.Params{
				"channel_id":                  "mock-channel-id",
				"video_id":                    "video_id_12345",
				"self_declared_made_for_kids": false,
				"contains_synthetic_media":    false,
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)
	s.Contains(text.Text, `"privacy_status":"private"`)
	s.Contains(text.Text, `"publish_at":"2026-08-15T22:00:00Z"`)
	s.Contains(text.Text, `"public_stats_viewable":false`)
	s.Contains(text.Text, `"self_declared_made_for_kids":false`)
	s.Contains(text.Text, `"contains_synthetic_media":false`)
}

func (s *YumSuite) TestUpdateVideoMetadataDoesNotInventAbsentSyntheticDeclaration() {
	s.mockVerifiedChannel()
	readAssert := func(req *http.Request) int {
		s.Equal("GET", req.Method)
		return 0
	}
	updateAssert := func(req *http.Request) int {
		s.Equal("PUT", req.Method)
		s.Equal("etag-without-synthetic", req.Header.Get("If-Match"))
		body, err := io.ReadAll(req.Body)
		s.NoError(err)
		s.NotContains(string(body), "containsSyntheticMedia")
		s.JSONEq(`{
			"id":"video_id_12345",
			"status":{
				"privacyStatus":"private",
				"publishAt":"2026-08-15T22:00:00Z",
				"license":"youtube",
				"embeddable":true,
				"publicStatsViewable":false,
				"selfDeclaredMadeForKids":false
			}
		}`, string(body))
		return 0
	}

	s.mock.Add("get_video_request", "get_video_without_synthetic_response").Respond(
		httpmatter.RequestResponse(readAssert))
	s.mock.Add("get_video_request", "get_video_after_update_response").Respond(
		httpmatter.RequestResponse(readAssert))
	s.mock.Add("update_video_metadata_request", "update_video_metadata_response").Respond(
		httpmatter.RequestResponse(updateAssert))
	s.mock.Init()

	_, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video_metadata",
			"arguments": mcp.Params{
				"channel_id":                  "mock-channel-id",
				"video_id":                    "video_id_12345",
				"self_declared_made_for_kids": false,
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)
}

func (s *YumSuite) TestUpdateVideoMetadataStopsAfterPreconditionFailure() {
	s.mockVerifiedChannel()
	s.mock.Add("get_video_request", "get_video_before_update_response").Respond(nil)
	s.mock.Add(
		"update_video_metadata_request",
		"update_video_metadata_precondition_failed_response",
	).Respond(nil)
	s.mock.Init()

	result, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video_metadata",
			"arguments": mcp.Params{
				"channel_id":               "mock-channel-id",
				"video_id":                 "video_id_12345",
				"contains_synthetic_media": false,
			},
		}).
		Call(s.Ctx(), 1, true)
	s.NoError(err)
	text, ok := result.Content[0].(mcp.TextContent)
	s.True(ok)
	s.Equal(
		"failed to update video metadata: video video_id_12345 changed after it was read; metadata update aborted to avoid overwriting concurrent changes",
		text.Text,
	)
}

func (s *YumSuite) TestUpdateVideoMetadataRejectsTokenChannelMismatchBeforeWrite() {
	s.mock.Add("verify_channel_request", "verify_channel_mismatch_response").Respond(nil)
	s.mock.Init()

	result, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video_metadata",
			"arguments": mcp.Params{
				"channel_id":               "mock-channel-id",
				"video_id":                 "video_id_12345",
				"contains_synthetic_media": false,
			},
		}).
		Call(s.Ctx(), 1, true)
	s.NoError(err)
	text, ok := result.Content[0].(mcp.TextContent)
	s.True(ok)
	s.Equal(
		"authenticated token is bound to channel other-channel-id, not requested channel mock-channel-id",
		text.Text,
	)
}

func (s *YumSuite) TestUpdateVideoSuccess() {
	s.mockVerifiedChannel()
	playlistAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Contains(req.URL.String(), "youtube/v3/playlistItems")
		return 0
	}
	captionsAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Contains(req.URL.String(), "youtube/v3/captions")
		return 0
	}
	thumbnailAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Contains(req.URL.String(), "youtube/v3/thumbnails/set")
		return 0
	}

	s.mock.Add("add_playlist_request", "add_playlist_response").Respond(
		httpmatter.RequestResponse(playlistAssert))
	s.mock.Add("add_captions_request", "add_captions_response").Respond(
		httpmatter.RequestResponse(captionsAssert))
	s.mock.Add("set_thumbnail_request", "set_thumbnail_response").Respond(
		httpmatter.RequestResponse(thumbnailAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video",
			"arguments": mcp.Params{
				"channel_id":        "mock-channel-id",
				"video_id":          "video_id_12345",
				"playlist_id":       "playlist_id_12345",
				"subtitle_path":     "./data/subtitles.srt",
				"subtitle_language": "en",
				"thumbnail_path":    "./data/thumbnail.png",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `"video_id":"video_id_12345"`)
	s.Contains(text.Text, `"playlist_status":"success"`)
	s.Contains(text.Text, `"subtitles_status":"success"`)
	s.Contains(text.Text, `"thumbnail_status":"success"`)
}

func (s *YumSuite) TestUpdateVideoPartialFailure() {
	s.mockVerifiedChannel()
	playlistAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Contains(req.URL.String(), "youtube/v3/playlistItems")
		return 0
	}

	s.mock.Add("add_playlist_request", "add_playlist_response").Respond(
		httpmatter.RequestResponse(playlistAssert))
	s.mock.Add("add_captions_request", "error_response").Respond(nil)
	s.mock.Add("set_thumbnail_request", "error_response").Respond(nil)
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video",
			"arguments": mcp.Params{
				"channel_id":        "mock-channel-id",
				"video_id":          "video_id_12345",
				"playlist_id":       "playlist_id_12345",
				"subtitle_path":     "./data/subtitles.srt",
				"subtitle_language": "en",
				"thumbnail_path":    "./data/thumbnail.png",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `"video_id":"video_id_12345"`)
	s.Contains(text.Text, `"playlist_status":"success"`)
	s.Contains(text.Text, `"subtitles_status":"failed"`)
	s.Contains(text.Text, `"thumbnail_status":"failed"`)
	s.Contains(text.Text, `"errors"`)
}
