package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/youtube/v3"
)

var videoMetadataParts = []string{
	"snippet",
	"status",
	"paidProductPlacementDetails",
	"recordingDetails",
}

// VideoMetadata is the owner-visible video metadata safe to return to callers.
// It deliberately excludes credentials, raw API responses, and unrelated fields.
type VideoMetadata struct {
	ID                      string   `json:"id"`
	ChannelID               string   `json:"channel_id"`
	Title                   string   `json:"title"`
	Description             string   `json:"description"`
	Tags                    []string `json:"tags"`
	CategoryID              string   `json:"category_id"`
	DefaultLanguage         string   `json:"default_language"`
	DefaultAudioLanguage    string   `json:"default_audio_language"`
	PrivacyStatus           string   `json:"privacy_status"`
	PublishAt               string   `json:"publish_at,omitempty"`
	License                 string   `json:"license"`
	Embeddable              bool     `json:"embeddable"`
	PublicStatsViewable     bool     `json:"public_stats_viewable"`
	MadeForKids             bool     `json:"made_for_kids"`
	SelfDeclaredMadeForKids *bool    `json:"self_declared_made_for_kids,omitempty"`
	ContainsSyntheticMedia  *bool    `json:"contains_synthetic_media,omitempty"`
	HasPaidProductPlacement bool     `json:"has_paid_product_placement"`
	RecordingDate           string   `json:"recording_date,omitempty"`
}

// VideoMetadataPatch contains the writable metadata supported by this release.
// Pointer fields preserve the difference between an omitted value and explicit false.
type VideoMetadataPatch struct {
	SelfDeclaredMadeForKids *bool `json:"self_declared_made_for_kids,omitempty"`
	ContainsSyntheticMedia  *bool `json:"contains_synthetic_media,omitempty"`
}

func (p VideoMetadataPatch) Empty() bool {
	return p.SelfDeclaredMadeForKids == nil && p.ContainsSyntheticMedia == nil
}

// GetVideo reads owner-visible metadata and verifies that the video belongs to
// expectedChannelID before returning a sanitized result.
func (c *Core) GetVideo(
	ctx context.Context,
	videoID string,
	expectedChannelID string,
	token *oauth2.Token,
) (*VideoMetadata, error) {
	video, err := c.getOwnedVideo(ctx, videoID, expectedChannelID, token)
	if err != nil {
		return nil, err
	}

	return videoMetadataFromAPI(video), nil
}

// UpdateVideoMetadata reads and verifies the current video, merges only the
// supplied patch into a clean writable status payload, updates it, and reads it
// again so callers receive the live result.
func (c *Core) UpdateVideoMetadata(
	ctx context.Context,
	videoID string,
	expectedChannelID string,
	patch VideoMetadataPatch,
	token *oauth2.Token,
) (*VideoMetadata, error) {
	if patch.Empty() {
		return nil, fmt.Errorf("video metadata patch must contain at least one field")
	}

	current, err := c.getOwnedVideo(ctx, videoID, expectedChannelID, token)
	if err != nil {
		return nil, err
	}
	if current.video.Status == nil {
		return nil, fmt.Errorf("video %s has no status metadata", videoID)
	}

	status := writableVideoStatus(current)
	if patch.SelfDeclaredMadeForKids != nil {
		status.SelfDeclaredMadeForKids = *patch.SelfDeclaredMadeForKids
		status.ForceSendFields = appendForceSendField(
			status.ForceSendFields,
			"SelfDeclaredMadeForKids",
		)
	}
	if patch.ContainsSyntheticMedia != nil {
		status.ContainsSyntheticMedia = *patch.ContainsSyntheticMedia
		status.ForceSendFields = appendForceSendField(
			status.ForceSendFields,
			"ContainsSyntheticMedia",
		)
	}
	if current.video.Etag == "" {
		return nil, fmt.Errorf("video %s has no ETag; refusing an unsafe metadata update", videoID)
	}

	service, err := c.Service(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}
	call := service.Videos.Update([]string{"status"}, &youtube.Video{
		Id:     videoID,
		Status: status,
	})
	call.Header().Set("If-Match", current.video.Etag)
	_, err = call.Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusPreconditionFailed {
			return nil, fmt.Errorf(
				"video %s changed after it was read; metadata update aborted to avoid overwriting concurrent changes",
				videoID,
			)
		}
		return nil, fmt.Errorf("failed to update video %s metadata: %w", videoID, err)
	}

	updated, err := c.getOwnedVideo(ctx, videoID, expectedChannelID, token)
	if err != nil {
		return nil, fmt.Errorf("video metadata updated but verification read failed: %w", err)
	}

	return videoMetadataFromAPI(updated), nil
}

func (c *Core) getOwnedVideo(
	ctx context.Context,
	videoID string,
	expectedChannelID string,
	token *oauth2.Token,
) (*ownedVideo, error) {
	if videoID == "" {
		return nil, fmt.Errorf("video ID must be provided")
	}
	if expectedChannelID == "" {
		return nil, fmt.Errorf("expected channel ID must be provided")
	}

	if token == nil {
		return nil, fmt.Errorf("token must be provided")
	}
	if c.config == nil {
		return nil, fmt.Errorf("OAuth2 config is not initialized in Core")
	}

	query := url.Values{}
	query.Set("alt", "json")
	query.Set("id", videoID)
	for _, part := range videoMetadataParts {
		query.Add("part", part)
	}
	query.Set("prettyPrint", "false")
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://youtube.googleapis.com/youtube/v3/videos?"+query.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create video metadata request: %w", err)
	}

	response, err := c.config.Client(ctx, token).Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get video %s: %w", videoID, err)
	}
	defer response.Body.Close()
	if err := googleapi.CheckResponse(response); err != nil {
		return nil, fmt.Errorf("failed to get video %s: %w", videoID, err)
	}

	var payload struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode video %s metadata: %w", videoID, err)
	}
	if len(payload.Items) == 0 {
		return nil, fmt.Errorf("video %s was not found", videoID)
	}

	var video youtube.Video
	if err := json.Unmarshal(payload.Items[0], &video); err != nil {
		return nil, fmt.Errorf("failed to decode video %s resource: %w", videoID, err)
	}
	if video.Snippet == nil {
		return nil, fmt.Errorf("video %s has no snippet metadata", videoID)
	}
	if video.Snippet.ChannelId != expectedChannelID {
		return nil, fmt.Errorf(
			"video %s belongs to channel %s, not requested channel %s",
			videoID,
			video.Snippet.ChannelId,
			expectedChannelID,
		)
	}

	var presence videoFieldPresence
	if err := json.Unmarshal(payload.Items[0], &presence); err != nil {
		return nil, fmt.Errorf("failed to inspect video %s status fields: %w", videoID, err)
	}
	_, selfDeclaredMadeForKidsPresent := presence.Status["selfDeclaredMadeForKids"]
	_, containsSyntheticMediaPresent := presence.Status["containsSyntheticMedia"]

	return &ownedVideo{
		video:                          &video,
		selfDeclaredMadeForKidsPresent: selfDeclaredMadeForKidsPresent,
		containsSyntheticMediaPresent:  containsSyntheticMediaPresent,
		presence:                       presence,
	}, nil
}

type videoFieldPresence struct {
	Snippet                     map[string]json.RawMessage `json:"snippet"`
	Status                      map[string]json.RawMessage `json:"status"`
	PaidProductPlacementDetails map[string]json.RawMessage `json:"paidProductPlacementDetails"`
}

type ownedVideo struct {
	video                          *youtube.Video
	selfDeclaredMadeForKidsPresent bool
	containsSyntheticMediaPresent  bool
	presence                       videoFieldPresence
}

func writableVideoStatus(current *ownedVideo) *youtube.VideoStatus {
	apiStatus := current.video.Status
	status := &youtube.VideoStatus{
		PrivacyStatus:       apiStatus.PrivacyStatus,
		PublishAt:           apiStatus.PublishAt,
		License:             apiStatus.License,
		Embeddable:          apiStatus.Embeddable,
		PublicStatsViewable: apiStatus.PublicStatsViewable,
		ForceSendFields: []string{
			"Embeddable",
			"PublicStatsViewable",
		},
	}
	if current.selfDeclaredMadeForKidsPresent {
		status.SelfDeclaredMadeForKids = apiStatus.SelfDeclaredMadeForKids
		status.ForceSendFields = append(
			status.ForceSendFields,
			"SelfDeclaredMadeForKids",
		)
	}
	if current.containsSyntheticMediaPresent {
		status.ContainsSyntheticMedia = apiStatus.ContainsSyntheticMedia
		status.ForceSendFields = append(
			status.ForceSendFields,
			"ContainsSyntheticMedia",
		)
	}

	return status
}

func appendForceSendField(fields []string, field string) []string {
	for _, existing := range fields {
		if existing == field {
			return fields
		}
	}
	return append(fields, field)
}

func videoMetadataFromAPI(owned *ownedVideo) *VideoMetadata {
	video := owned.video
	metadata := &VideoMetadata{ID: video.Id}
	if video.Snippet != nil {
		metadata.ChannelID = video.Snippet.ChannelId
		metadata.Title = video.Snippet.Title
		metadata.Description = video.Snippet.Description
		metadata.Tags = video.Snippet.Tags
		metadata.CategoryID = video.Snippet.CategoryId
		metadata.DefaultLanguage = video.Snippet.DefaultLanguage
		metadata.DefaultAudioLanguage = video.Snippet.DefaultAudioLanguage
	}
	if video.Status != nil {
		metadata.PrivacyStatus = video.Status.PrivacyStatus
		metadata.PublishAt = video.Status.PublishAt
		metadata.License = video.Status.License
		metadata.Embeddable = video.Status.Embeddable
		metadata.PublicStatsViewable = video.Status.PublicStatsViewable
		metadata.MadeForKids = video.Status.MadeForKids
		if owned.selfDeclaredMadeForKidsPresent {
			value := video.Status.SelfDeclaredMadeForKids
			metadata.SelfDeclaredMadeForKids = &value
		}
		if owned.containsSyntheticMediaPresent {
			value := video.Status.ContainsSyntheticMedia
			metadata.ContainsSyntheticMedia = &value
		}
	}
	if video.PaidProductPlacementDetails != nil {
		metadata.HasPaidProductPlacement =
			video.PaidProductPlacementDetails.HasPaidProductPlacement
	}
	if video.RecordingDetails != nil {
		metadata.RecordingDate = video.RecordingDetails.RecordingDate
	}

	return metadata
}
