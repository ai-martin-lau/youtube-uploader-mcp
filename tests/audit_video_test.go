package tests

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/therewardstore/httpmatter"
)

func (s *YumSuite) TestAuditVideoRequiresAtLeastOneExpectation() {
	result, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "audit_video",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
				"video_id":   "video_id_12345",
			},
		}).
		Call(s.Ctx(), 1, true)
	s.NoError(err)
	text, ok := result.Content[0].(mcp.TextContent)
	s.True(ok)
	s.Equal("at least one expected_* audit parameter must be provided", text.Text)
}

func (s *YumSuite) TestAuditVideoPassesReadableChecksAndMarksStudioOnlyChecksUnverifiable() {
	s.mockVerifiedChannel()
	readAssert := func(req *http.Request) int {
		s.Equal(http.MethodGet, req.Method)
		s.Equal("Bearer mock-access-token", req.Header.Get("Authorization"))
		return 0
	}
	s.mock.Add("get_video_request", "get_video_without_recording_response").Respond(
		httpmatter.RequestResponse(readAssert))
	s.mock.Add("list_captions_request", "list_captions_empty_response").Respond(
		httpmatter.RequestResponse(readAssert))
	s.mock.Add("list_playlist_snapshot_request", "list_playlist_snapshot_response").Respond(
		httpmatter.RequestResponse(readAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "audit_video",
			"arguments": mcp.Params{
				"channel_id":                           "mock-channel-id",
				"video_id":                             "video_id_12345",
				"expected_self_declared_made_for_kids": false,
				"expected_contains_synthetic_media":    false,
				"expected_has_paid_product_placement":  false,
				"expected_default_language":            "en",
				"expected_default_audio_language":      "en",
				"expected_category_id":                 "10",
				"expected_privacy_status":              "private",
				"expected_publish_at":                  "2026-08-15T22:00:00Z",
				"expected_license":                     "youtube",
				"expected_embeddable":                  true,
				"expected_public_stats_viewable":       true,
				"expect_recording_date_absent":         true,
				"expect_recording_location_absent":     true,
				"expected_caption_track_count":         0,
				"expected_playlist_contents": map[string][]string{
					"playlist_id_12345": {"video_id_00001", "video_id_00002", "video_id_12345"},
				},
				"expected_notify_subscribers": false,
				"expected_automatic_chapters": false,
				"expected_comment_moderation": "strict",
				"expected_cards_present":      true,
				"expected_end_screen_present": true,
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	var result struct {
		Overall string `json:"overall"`
		Counts  struct {
			Pass         int `json:"pass"`
			Fail         int `json:"fail"`
			Unverifiable int `json:"unverifiable"`
		} `json:"counts"`
		Checks []struct {
			Field  string `json:"field"`
			Status string `json:"status"`
			Reason string `json:"reason,omitempty"`
		} `json:"checks"`
	}
	s.NoError(json.Unmarshal([]byte(text.Text), &result))
	s.Equal("partial", result.Overall)
	s.Equal(15, result.Counts.Pass)
	s.Zero(result.Counts.Fail)
	s.Equal(5, result.Counts.Unverifiable)

	statuses := make(map[string]string, len(result.Checks))
	for _, check := range result.Checks {
		statuses[check.Field] = check.Status
	}
	s.Equal("pass", statuses["self_declared_made_for_kids"])
	s.Equal("pass", statuses["caption_track_count"])
	s.Equal("pass", statuses["playlist_contents:playlist_id_12345"])
	s.Equal("unverifiable", statuses["notify_subscribers"])
	s.Equal("unverifiable", statuses["end_screen_present"])
}

func (s *YumSuite) TestAuditVideoFailsDuplicatePlaylistMembership() {
	s.mockVerifiedChannel()
	s.mock.Add("get_video_request", "get_video_without_recording_response").Respond(nil)
	s.mock.Add("list_playlist_items_request", "list_playlist_items_duplicate_response").Respond(nil)
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "audit_video",
			"arguments": mcp.Params{
				"channel_id":            "mock-channel-id",
				"video_id":              "video_id_12345",
				"expected_playlist_ids": []string{"playlist_id_12345"},
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)
	s.Contains(text.Text, `"overall":"fail"`)
	s.Contains(text.Text, `"reason":"duplicate_playlist_entries"`)
	s.Contains(text.Text, `"count":2`)
}

func (s *YumSuite) TestAuditVideoFailsMissingPlaylistMembership() {
	s.mockVerifiedChannel()
	s.mock.Add("get_video_request", "get_video_without_recording_response").Respond(nil)
	s.mock.Add("list_playlist_items_request", "list_playlist_items_empty_response").Respond(nil)
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "audit_video",
			"arguments": mcp.Params{
				"channel_id":            "mock-channel-id",
				"video_id":              "video_id_12345",
				"expected_playlist_ids": []string{"playlist_id_12345"},
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)
	s.Contains(text.Text, `"overall":"fail"`)
	s.Contains(text.Text, `"reason":"video_not_in_playlist"`)
	s.Contains(text.Text, `"count":0`)
}

func (s *YumSuite) TestAuditVideoFailsPlaylistOrderMismatch() {
	s.mockVerifiedChannel()
	s.mock.Add("get_video_request", "get_video_without_recording_response").Respond(nil)
	s.mock.Add("list_playlist_snapshot_request", "list_playlist_snapshot_response").Respond(nil)
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "audit_video",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
				"video_id":   "video_id_12345",
				"expected_playlist_contents": map[string][]string{
					"playlist_id_12345": {"video_id_12345", "video_id_00001", "video_id_00002"},
				},
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)
	s.Contains(text.Text, `"overall":"fail"`)
	s.Contains(text.Text, `"reason":"playlist_member_count_or_order_mismatch"`)
	s.Contains(text.Text, `"video_id":"video_id_12345","position":2`)
}

func (s *YumSuite) TestAuditVideoPassesPlaylistOrderAcrossPages() {
	s.mockVerifiedChannel()
	s.mock.Add("get_video_request", "get_video_without_recording_response").Respond(nil)
	s.mock.Add("list_playlist_snapshot_page1_request", "list_playlist_snapshot_page1_response").Respond(nil)
	s.mock.Add("list_playlist_snapshot_page2_request", "list_playlist_snapshot_page2_response").Respond(nil)
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "audit_video",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
				"video_id":   "video_id_12345",
				"expected_playlist_contents": map[string][]string{
					"playlist_id_12345": {"video_id_00001", "video_id_00002", "video_id_12345"},
				},
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)
	s.Contains(text.Text, `"overall":"pass"`)
	s.Contains(text.Text, `"count":3`)
	s.Contains(text.Text, `"video_id":"video_id_12345","position":2`)
}

func (s *YumSuite) TestAuditVideoMarksCaptionAndPlaylistReadErrorsUnverifiable() {
	s.mockVerifiedChannel()
	s.mock.Add("get_video_request", "get_video_without_recording_response").Respond(nil)
	s.mock.Add("list_captions_request", "list_captions_error_response").Respond(nil)
	s.mock.Add("list_playlist_items_request", "list_playlist_items_error_response").Respond(nil)
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "audit_video",
			"arguments": mcp.Params{
				"channel_id":                   "mock-channel-id",
				"video_id":                     "video_id_12345",
				"expected_caption_track_count": 0,
				"expected_playlist_ids":        []string{"playlist_id_12345"},
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)
	s.Contains(text.Text, `"overall":"partial"`)
	s.Contains(text.Text, `"pass":0,"fail":0,"unverifiable":2`)
	s.Equal(2, strings.Count(text.Text, `"reason":"api_read_failed"`))
}

func (s *YumSuite) TestAuditVideoFailsExplicitMismatchAndUnsetDeclaration() {
	s.mockVerifiedChannel()
	s.mock.Add("get_video_request", "get_video_without_synthetic_response").Respond(nil)
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "audit_video",
			"arguments": mcp.Params{
				"channel_id":                           "mock-channel-id",
				"video_id":                             "video_id_12345",
				"expected_self_declared_made_for_kids": false,
				"expected_contains_synthetic_media":    false,
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	var result struct {
		Overall string `json:"overall"`
		Counts  struct {
			Pass         int `json:"pass"`
			Fail         int `json:"fail"`
			Unverifiable int `json:"unverifiable"`
		} `json:"counts"`
		Checks []struct {
			Field  string `json:"field"`
			Status string `json:"status"`
			Reason string `json:"reason,omitempty"`
		} `json:"checks"`
	}
	s.NoError(json.Unmarshal([]byte(text.Text), &result))
	s.Equal("fail", result.Overall)
	s.Zero(result.Counts.Pass)
	s.Equal(2, result.Counts.Fail)
	s.Zero(result.Counts.Unverifiable)
	s.Equal("value_mismatch", result.Checks[0].Reason)
	s.Equal("declaration_unset", result.Checks[1].Reason)
}
