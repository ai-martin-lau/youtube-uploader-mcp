package tool

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	auditPass         = "pass"
	auditFail         = "fail"
	auditUnverifiable = "unverifiable"
)

type AuditVideoTool struct {
	Core *core.Core
}

func (t *AuditVideoTool) Name() string {
	return "audit_video"
}

func (t *AuditVideoTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Read-only audit of an owned YouTube video against explicitly supplied expectations. API-readable fields are compared live; Studio-only or insert-only controls are reported as unverifiable, never silently passed."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithString("channel_id",
			mcp.Required(),
			mcp.Description("Expected authenticated owner channel ID"),
		),
		mcp.WithString("video_id",
			mcp.Required(),
			mcp.Description("Video ID to audit"),
		),
		mcp.WithBoolean("expected_self_declared_made_for_kids",
			mcp.Description("Expected explicit owner made-for-kids declaration"),
		),
		mcp.WithBoolean("expected_contains_synthetic_media",
			mcp.Description("Expected explicit altered or synthetic media declaration"),
		),
		mcp.WithBoolean("expected_has_paid_product_placement",
			mcp.Description("Expected paid product placement declaration"),
		),
		mcp.WithString("expected_default_language",
			mcp.Description("Expected video metadata language code, such as en"),
		),
		mcp.WithString("expected_default_audio_language",
			mcp.Description("Expected default audio language code, such as en"),
		),
		mcp.WithString("expected_category_id",
			mcp.Description("Expected YouTube video category ID"),
		),
		mcp.WithString("expected_privacy_status",
			mcp.Description("Expected live privacy status"),
			mcp.Enum("private", "unlisted", "public"),
		),
		mcp.WithString("expected_publish_at",
			mcp.Description("Expected scheduled publication time in RFC3339 format"),
		),
		mcp.WithString("expected_license",
			mcp.Description("Expected YouTube license value"),
			mcp.Enum("youtube", "creativeCommon"),
		),
		mcp.WithBoolean("expected_embeddable",
			mcp.Description("Expected embed permission"),
		),
		mcp.WithBoolean("expected_public_stats_viewable",
			mcp.Description("Expected public statistics visibility"),
		),
		mcp.WithBoolean("expect_recording_date_absent",
			mcp.Description("Whether recording date should be absent"),
		),
		mcp.WithBoolean("expect_recording_location_absent",
			mcp.Description("Whether recording location and description should be absent"),
		),
		mcp.WithNumber("expected_caption_track_count",
			mcp.Description("Expected total number of all caption resources returned by captions.list, including ASR, draft, syncing, and failed tracks"),
			mcp.Min(0),
			mcp.MultipleOf(1),
		),
		mcp.WithArray("expected_playlist_ids",
			mcp.Description("Playlist IDs that must each contain exactly one item for this video; use expected_playlist_contents to also assert full member count and order"),
			mcp.Items(map[string]any{"type": "string"}),
			mcp.MaxItems(20),
			mcp.UniqueItems(true),
		),
		mcp.WithObject("expected_playlist_contents",
			mcp.Description("Map of playlist ID to the exact ordered video ID list expected in that playlist"),
			mcp.MaxProperties(20),
			mcp.AdditionalProperties(map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"uniqueItems": true,
			}),
		),
		mcp.WithBoolean("expected_notify_subscribers",
			mcp.Description("Expected upload-time subscriber notification choice; the API cannot read this back after upload"),
		),
		mcp.WithBoolean("expected_automatic_chapters",
			mcp.Description("Expected Studio automatic chapters setting; not exposed for read-back by the Data API"),
		),
		mcp.WithBoolean("expected_automatic_places",
			mcp.Description("Expected Studio automatic places setting; not exposed for read-back by the Data API"),
		),
		mcp.WithBoolean("expected_automatic_concepts",
			mcp.Description("Expected Studio automatic concepts setting; not exposed for read-back by the Data API"),
		),
		mcp.WithString("expected_shorts_remixing",
			mcp.Description("Expected Shorts remixing policy; not exposed for read-back by the Data API"),
			mcp.Enum("allow_video_and_audio", "allow_audio_only", "disallow"),
		),
		mcp.WithString("expected_comment_moderation",
			mcp.Description("Internal policy vocabulary for the expected comment moderation setting; not exposed for read-back by the Data API"),
			mcp.Enum("strict", "basic", "hold_all", "none", "disabled"),
		),
		mcp.WithBoolean("expected_cards_present",
			mcp.Description("Whether cards should be present; cards are not exposed by the Data API"),
		),
		mcp.WithBoolean("expected_end_screen_present",
			mcp.Description("Whether an end screen should be present; end screens are not exposed by the Data API"),
		),
		mcp.WithString("expected_caption_certification",
			mcp.Description("Expected caption certification policy label from the caller's project; not exposed for read-back by the Data API"),
		),
	)
}

type auditVideoArguments struct {
	ChannelID                       string              `json:"channel_id"`
	VideoID                         string              `json:"video_id"`
	ExpectedSelfDeclaredMadeForKids *bool               `json:"expected_self_declared_made_for_kids"`
	ExpectedContainsSyntheticMedia  *bool               `json:"expected_contains_synthetic_media"`
	ExpectedHasPaidProductPlacement *bool               `json:"expected_has_paid_product_placement"`
	ExpectedDefaultLanguage         *string             `json:"expected_default_language"`
	ExpectedDefaultAudioLanguage    *string             `json:"expected_default_audio_language"`
	ExpectedCategoryID              *string             `json:"expected_category_id"`
	ExpectedPrivacyStatus           *string             `json:"expected_privacy_status"`
	ExpectedPublishAt               *string             `json:"expected_publish_at"`
	ExpectedLicense                 *string             `json:"expected_license"`
	ExpectedEmbeddable              *bool               `json:"expected_embeddable"`
	ExpectedPublicStatsViewable     *bool               `json:"expected_public_stats_viewable"`
	ExpectRecordingDateAbsent       *bool               `json:"expect_recording_date_absent"`
	ExpectRecordingLocationAbsent   *bool               `json:"expect_recording_location_absent"`
	ExpectedCaptionTrackCount       *int                `json:"expected_caption_track_count"`
	ExpectedPlaylistIDs             []string            `json:"expected_playlist_ids"`
	ExpectedPlaylistContents        map[string][]string `json:"expected_playlist_contents"`
	ExpectedNotifySubscribers       *bool               `json:"expected_notify_subscribers"`
	ExpectedAutomaticChapters       *bool               `json:"expected_automatic_chapters"`
	ExpectedAutomaticPlaces         *bool               `json:"expected_automatic_places"`
	ExpectedAutomaticConcepts       *bool               `json:"expected_automatic_concepts"`
	ExpectedShortsRemixing          *string             `json:"expected_shorts_remixing"`
	ExpectedCommentModeration       *string             `json:"expected_comment_moderation"`
	ExpectedCardsPresent            *bool               `json:"expected_cards_present"`
	ExpectedEndScreenPresent        *bool               `json:"expected_end_screen_present"`
	ExpectedCaptionCertification    *string             `json:"expected_caption_certification"`
}

func (a auditVideoArguments) hasExpectation() bool {
	return a.ExpectedSelfDeclaredMadeForKids != nil ||
		a.ExpectedContainsSyntheticMedia != nil ||
		a.ExpectedHasPaidProductPlacement != nil ||
		a.ExpectedDefaultLanguage != nil ||
		a.ExpectedDefaultAudioLanguage != nil ||
		a.ExpectedCategoryID != nil ||
		a.ExpectedPrivacyStatus != nil ||
		a.ExpectedPublishAt != nil ||
		a.ExpectedLicense != nil ||
		a.ExpectedEmbeddable != nil ||
		a.ExpectedPublicStatsViewable != nil ||
		a.ExpectRecordingDateAbsent != nil ||
		a.ExpectRecordingLocationAbsent != nil ||
		a.ExpectedCaptionTrackCount != nil ||
		len(a.ExpectedPlaylistIDs) > 0 ||
		len(a.ExpectedPlaylistContents) > 0 ||
		a.ExpectedNotifySubscribers != nil ||
		a.ExpectedAutomaticChapters != nil ||
		a.ExpectedAutomaticPlaces != nil ||
		a.ExpectedAutomaticConcepts != nil ||
		a.ExpectedShortsRemixing != nil ||
		a.ExpectedCommentModeration != nil ||
		a.ExpectedCardsPresent != nil ||
		a.ExpectedEndScreenPresent != nil ||
		a.ExpectedCaptionCertification != nil
}

type AuditCheck struct {
	Field    string `json:"field"`
	Status   string `json:"status"`
	Expected any    `json:"expected"`
	Actual   any    `json:"actual"`
	Reason   string `json:"reason,omitempty"`
}

type AuditCounts struct {
	Pass         int `json:"pass"`
	Fail         int `json:"fail"`
	Unverifiable int `json:"unverifiable"`
}

type AuditVideoResult struct {
	VideoID   string       `json:"video_id"`
	ChannelID string       `json:"channel_id"`
	Title     string       `json:"title"`
	Overall   string       `json:"overall"`
	Counts    AuditCounts  `json:"counts"`
	Checks    []AuditCheck `json:"checks"`
	Warnings  []string     `json:"warnings,omitempty"`
}

func (t *AuditVideoTool) Handle(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	var args auditVideoArguments
	if err := request.BindArguments(&args); err != nil {
		return mcp.NewToolResultError("invalid audit arguments: " + err.Error()), nil
	}
	if args.ChannelID == "" || args.VideoID == "" {
		return mcp.NewToolResultError("channel_id and video_id are required"), nil
	}
	if !args.hasExpectation() {
		return mcp.NewToolResultError("at least one expected_* audit parameter must be provided"), nil
	}
	if args.ExpectedPublishAt != nil {
		if _, err := time.Parse(time.RFC3339, *args.ExpectedPublishAt); err != nil {
			return mcp.NewToolResultError("expected_publish_at must be a valid RFC3339 timestamp"), nil
		}
	}

	token, err := validChannelToken(ctx, t.Core, args.ChannelID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	facts, err := t.Core.GetVideoAuditFacts(ctx, args.VideoID, args.ChannelID, token)
	if err != nil {
		return mcp.NewToolResultError("failed to audit video: " + err.Error()), nil
	}

	result := AuditVideoResult{
		VideoID:   facts.VideoID,
		ChannelID: facts.ChannelID,
		Title:     facts.Title,
		Checks:    []AuditCheck{},
	}
	addBoolCheck(&result, "self_declared_made_for_kids", args.ExpectedSelfDeclaredMadeForKids, facts.SelfDeclaredMadeForKids, true)
	addBoolCheck(&result, "contains_synthetic_media", args.ExpectedContainsSyntheticMedia, facts.ContainsSyntheticMedia, true)
	addBoolCheck(&result, "has_paid_product_placement", args.ExpectedHasPaidProductPlacement, facts.HasPaidProductPlacement, false)
	addStringCheck(&result, "default_language", args.ExpectedDefaultLanguage, facts.DefaultLanguage, true)
	addStringCheck(&result, "default_audio_language", args.ExpectedDefaultAudioLanguage, facts.DefaultAudioLanguage, true)
	addStringCheck(&result, "category_id", args.ExpectedCategoryID, facts.CategoryID, true)
	addStringCheck(&result, "privacy_status", args.ExpectedPrivacyStatus, facts.PrivacyStatus, true)
	addTimeCheck(&result, "publish_at", args.ExpectedPublishAt, facts.PublishAt)
	addStringCheck(&result, "license", args.ExpectedLicense, facts.License, true)
	addBoolCheck(&result, "embeddable", args.ExpectedEmbeddable, facts.Embeddable, false)
	addBoolCheck(&result, "public_stats_viewable", args.ExpectedPublicStatsViewable, facts.PublicStatsViewable, false)
	addPresenceCheck(&result, "recording_date_absent", args.ExpectRecordingDateAbsent, facts.RecordingDate != "")
	hasRecordingLocation := facts.RecordingLocation != nil || facts.RecordingLocationDescription != ""
	addPresenceCheck(&result, "recording_location_absent", args.ExpectRecordingLocationAbsent, hasRecordingLocation)

	if args.ExpectedCaptionTrackCount != nil {
		captions, err := t.Core.ListCaptions(ctx, args.VideoID, token)
		if err != nil {
			result.add(AuditCheck{
				Field: "caption_track_count", Status: auditUnverifiable,
				Expected: *args.ExpectedCaptionTrackCount, Actual: nil, Reason: "api_read_failed",
			})
			result.Warnings = append(result.Warnings, err.Error())
		} else {
			actual := struct {
				Count  int                     `json:"count"`
				Tracks []core.CaptionAuditItem `json:"tracks"`
			}{Count: len(captions), Tracks: captions}
			status, reason := compare(*args.ExpectedCaptionTrackCount, actual.Count)
			result.add(AuditCheck{Field: "caption_track_count", Status: status, Expected: *args.ExpectedCaptionTrackCount, Actual: actual, Reason: reason})
		}
	}

	for _, playlistID := range args.ExpectedPlaylistIDs {
		membership, err := t.Core.GetPlaylistMembership(ctx, playlistID, args.VideoID, token)
		field := "playlist:" + playlistID
		if err != nil {
			result.add(AuditCheck{Field: field, Status: auditUnverifiable, Expected: "exactly_one_item", Actual: nil, Reason: "api_read_failed"})
			result.Warnings = append(result.Warnings, err.Error())
			continue
		}
		switch membership.Count {
		case 1:
			result.add(AuditCheck{Field: field, Status: auditPass, Expected: "exactly_one_item", Actual: membership})
		case 0:
			result.add(AuditCheck{Field: field, Status: auditFail, Expected: "exactly_one_item", Actual: membership, Reason: "video_not_in_playlist"})
		default:
			result.add(AuditCheck{Field: field, Status: auditFail, Expected: "exactly_one_item", Actual: membership, Reason: "duplicate_playlist_entries"})
		}
	}

	playlistIDs := make([]string, 0, len(args.ExpectedPlaylistContents))
	for playlistID := range args.ExpectedPlaylistContents {
		playlistIDs = append(playlistIDs, playlistID)
	}
	sort.Strings(playlistIDs)
	for _, playlistID := range playlistIDs {
		expectedVideoIDs := args.ExpectedPlaylistContents[playlistID]
		field := "playlist_contents:" + playlistID
		snapshot, err := t.Core.GetPlaylistSnapshot(ctx, playlistID, token)
		if err != nil {
			result.add(AuditCheck{Field: field, Status: auditUnverifiable, Expected: expectedVideoIDs, Actual: nil, Reason: "api_read_failed"})
			result.Warnings = append(result.Warnings, err.Error())
			continue
		}
		actualVideoIDs := make([]string, 0, len(snapshot.Items))
		for _, item := range snapshot.Items {
			actualVideoIDs = append(actualVideoIDs, item.VideoID)
		}
		status, reason := compare(expectedVideoIDs, actualVideoIDs)
		if status == auditFail {
			reason = "playlist_member_count_or_order_mismatch"
		}
		result.add(AuditCheck{Field: field, Status: status, Expected: expectedVideoIDs, Actual: snapshot, Reason: reason})
	}

	addUnverifiableBool(&result, "notify_subscribers", args.ExpectedNotifySubscribers, "insert_parameter_not_persisted")
	addUnverifiableBool(&result, "automatic_chapters", args.ExpectedAutomaticChapters, "not_exposed_by_youtube_data_api")
	addUnverifiableBool(&result, "automatic_places", args.ExpectedAutomaticPlaces, "not_exposed_by_youtube_data_api")
	addUnverifiableBool(&result, "automatic_concepts", args.ExpectedAutomaticConcepts, "not_exposed_by_youtube_data_api")
	addUnverifiableString(&result, "shorts_remixing", args.ExpectedShortsRemixing)
	addUnverifiableString(&result, "comment_moderation", args.ExpectedCommentModeration)
	addUnverifiableBool(&result, "cards_present", args.ExpectedCardsPresent, "not_exposed_by_youtube_data_api")
	addUnverifiableBool(&result, "end_screen_present", args.ExpectedEndScreenPresent, "not_exposed_by_youtube_data_api")
	addUnverifiableString(&result, "caption_certification", args.ExpectedCaptionCertification)

	result.Overall = auditPass
	if result.Counts.Fail > 0 {
		result.Overall = auditFail
	} else if result.Counts.Unverifiable > 0 {
		result.Overall = "partial"
	}

	data, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError("failed to marshal audit result: " + err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func (r *AuditVideoResult) add(check AuditCheck) {
	r.Checks = append(r.Checks, check)
	switch check.Status {
	case auditPass:
		r.Counts.Pass++
	case auditFail:
		r.Counts.Fail++
	case auditUnverifiable:
		r.Counts.Unverifiable++
	}
}

func addBoolCheck(result *AuditVideoResult, field string, expected *bool, actual *bool, missingIsDeclarationFailure bool) {
	if expected == nil {
		return
	}
	if actual == nil {
		status := auditUnverifiable
		reason := "api_field_not_returned"
		if missingIsDeclarationFailure {
			status = auditFail
			reason = "declaration_unset"
		}
		result.add(AuditCheck{Field: field, Status: status, Expected: *expected, Actual: nil, Reason: reason})
		return
	}
	status, reason := compare(*expected, *actual)
	result.add(AuditCheck{Field: field, Status: status, Expected: *expected, Actual: *actual, Reason: reason})
}

func addStringCheck(result *AuditVideoResult, field string, expected *string, actual *string, missingIsFailure bool) {
	if expected == nil {
		return
	}
	if actual == nil || strings.TrimSpace(*actual) == "" {
		status := auditUnverifiable
		reason := "api_field_not_returned"
		if missingIsFailure {
			status = auditFail
			reason = "setting_unset"
		}
		result.add(AuditCheck{Field: field, Status: status, Expected: *expected, Actual: nil, Reason: reason})
		return
	}
	status, reason := compare(*expected, *actual)
	result.add(AuditCheck{Field: field, Status: status, Expected: *expected, Actual: *actual, Reason: reason})
}

func addTimeCheck(result *AuditVideoResult, field string, expected *string, actual *string) {
	if expected == nil {
		return
	}
	if actual == nil || *actual == "" {
		result.add(AuditCheck{Field: field, Status: auditFail, Expected: *expected, Actual: nil, Reason: "setting_unset"})
		return
	}
	expectedTime, _ := time.Parse(time.RFC3339, *expected)
	actualTime, err := time.Parse(time.RFC3339, *actual)
	if err != nil {
		result.add(AuditCheck{Field: field, Status: auditUnverifiable, Expected: *expected, Actual: *actual, Reason: "invalid_api_timestamp"})
		return
	}
	status := auditPass
	reason := ""
	if !expectedTime.Equal(actualTime) {
		status = auditFail
		reason = "value_mismatch"
	}
	result.add(AuditCheck{Field: field, Status: status, Expected: *expected, Actual: *actual, Reason: reason})
}

func addPresenceCheck(result *AuditVideoResult, field string, expectedAbsent *bool, actualPresent bool) {
	if expectedAbsent == nil {
		return
	}
	actualAbsent := !actualPresent
	status, reason := compare(*expectedAbsent, actualAbsent)
	result.add(AuditCheck{Field: field, Status: status, Expected: *expectedAbsent, Actual: actualAbsent, Reason: reason})
}

func addUnverifiableBool(result *AuditVideoResult, field string, expected *bool, reason string) {
	if expected == nil {
		return
	}
	result.add(AuditCheck{Field: field, Status: auditUnverifiable, Expected: *expected, Actual: nil, Reason: reason})
}

func addUnverifiableString(result *AuditVideoResult, field string, expected *string) {
	if expected == nil {
		return
	}
	result.add(AuditCheck{Field: field, Status: auditUnverifiable, Expected: *expected, Actual: nil, Reason: "not_exposed_by_youtube_data_api"})
}

func compare(expected any, actual any) (string, string) {
	if reflect.DeepEqual(expected, actual) {
		return auditPass, ""
	}
	return auditFail, "value_mismatch"
}
