package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"golang.org/x/oauth2"
)

// VideoAuditFacts contains only owner-visible fields that the YouTube Data API
// can reliably read back. Pointer values preserve the distinction between an
// explicit false/empty value and a field that the API did not return.
type VideoAuditFacts struct {
	VideoID                      string
	ChannelID                    string
	Title                        string
	SelfDeclaredMadeForKids      *bool
	ContainsSyntheticMedia       *bool
	HasPaidProductPlacement      *bool
	DefaultLanguage              *string
	DefaultAudioLanguage         *string
	CategoryID                   *string
	PrivacyStatus                *string
	PublishAt                    *string
	License                      *string
	Embeddable                   *bool
	PublicStatsViewable          *bool
	RecordingDate                string
	RecordingLocationDescription string
	RecordingLocation            *RecordingLocation
}

type RecordingLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

type CaptionAuditItem struct {
	ID            string `json:"id"`
	Language      string `json:"language,omitempty"`
	Name          string `json:"name,omitempty"`
	TrackKind     string `json:"track_kind,omitempty"`
	Status        string `json:"status,omitempty"`
	FailureReason string `json:"failure_reason,omitempty"`
	IsDraft       bool   `json:"is_draft"`
	IsAutoSynced  bool   `json:"is_auto_synced"`
}

type PlaylistMembership struct {
	PlaylistID string   `json:"playlist_id"`
	Present    bool     `json:"present"`
	Count      int      `json:"count"`
	ItemIDs    []string `json:"item_ids"`
	Positions  []int64  `json:"positions"`
}

type PlaylistSnapshotItem struct {
	ItemID   string `json:"item_id"`
	VideoID  string `json:"video_id"`
	Position int64  `json:"position"`
}

type PlaylistSnapshot struct {
	PlaylistID string                 `json:"playlist_id"`
	Count      int                    `json:"count"`
	Items      []PlaylistSnapshotItem `json:"items"`
}

func (c *Core) GetVideoAuditFacts(
	ctx context.Context,
	videoID string,
	expectedChannelID string,
	token *oauth2.Token,
) (*VideoAuditFacts, error) {
	owned, err := c.getOwnedVideo(ctx, videoID, expectedChannelID, token)
	if err != nil {
		return nil, err
	}

	video := owned.video
	facts := &VideoAuditFacts{
		VideoID:   video.Id,
		ChannelID: expectedChannelID,
	}
	if video.Snippet != nil {
		facts.Title = video.Snippet.Title
		facts.DefaultLanguage = stringIfPresent(
			owned.presence.Snippet,
			"defaultLanguage",
			video.Snippet.DefaultLanguage,
		)
		facts.DefaultAudioLanguage = stringIfPresent(
			owned.presence.Snippet,
			"defaultAudioLanguage",
			video.Snippet.DefaultAudioLanguage,
		)
		facts.CategoryID = stringIfPresent(
			owned.presence.Snippet,
			"categoryId",
			video.Snippet.CategoryId,
		)
	}
	if video.Status != nil {
		facts.PrivacyStatus = stringIfPresent(
			owned.presence.Status,
			"privacyStatus",
			video.Status.PrivacyStatus,
		)
		facts.PublishAt = stringIfPresent(
			owned.presence.Status,
			"publishAt",
			video.Status.PublishAt,
		)
		facts.License = stringIfPresent(
			owned.presence.Status,
			"license",
			video.Status.License,
		)
		facts.Embeddable = boolIfPresent(
			owned.presence.Status,
			"embeddable",
			video.Status.Embeddable,
		)
		facts.PublicStatsViewable = boolIfPresent(
			owned.presence.Status,
			"publicStatsViewable",
			video.Status.PublicStatsViewable,
		)
		facts.SelfDeclaredMadeForKids = boolIfPresent(
			owned.presence.Status,
			"selfDeclaredMadeForKids",
			video.Status.SelfDeclaredMadeForKids,
		)
		facts.ContainsSyntheticMedia = boolIfPresent(
			owned.presence.Status,
			"containsSyntheticMedia",
			video.Status.ContainsSyntheticMedia,
		)
	}
	if video.PaidProductPlacementDetails != nil {
		facts.HasPaidProductPlacement = boolIfPresent(
			owned.presence.PaidProductPlacementDetails,
			"hasPaidProductPlacement",
			video.PaidProductPlacementDetails.HasPaidProductPlacement,
		)
	}
	if video.RecordingDetails != nil {
		facts.RecordingDate = video.RecordingDetails.RecordingDate
		facts.RecordingLocationDescription = video.RecordingDetails.LocationDescription
		if video.RecordingDetails.Location != nil {
			facts.RecordingLocation = &RecordingLocation{
				Latitude:  video.RecordingDetails.Location.Latitude,
				Longitude: video.RecordingDetails.Location.Longitude,
				Altitude:  video.RecordingDetails.Location.Altitude,
			}
		}
	}

	return facts, nil
}

func (c *Core) ListCaptions(
	ctx context.Context,
	videoID string,
	token *oauth2.Token,
) ([]CaptionAuditItem, error) {
	if videoID == "" {
		return nil, fmt.Errorf("video ID must be provided")
	}
	if token == nil {
		return nil, fmt.Errorf("token must be provided")
	}

	service, err := c.Service(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}
	response, err := service.Captions.List([]string{"id", "snippet"}, videoID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list captions for video %s: %w", videoID, err)
	}

	items := make([]CaptionAuditItem, 0, len(response.Items))
	for _, item := range response.Items {
		caption := CaptionAuditItem{ID: item.Id}
		if item.Snippet != nil {
			caption.Language = item.Snippet.Language
			caption.Name = item.Snippet.Name
			caption.TrackKind = item.Snippet.TrackKind
			caption.Status = item.Snippet.Status
			caption.FailureReason = item.Snippet.FailureReason
			caption.IsDraft = item.Snippet.IsDraft
			caption.IsAutoSynced = item.Snippet.IsAutoSynced
		}
		items = append(items, caption)
	}

	return items, nil
}

func (c *Core) GetPlaylistMembership(
	ctx context.Context,
	playlistID string,
	videoID string,
	token *oauth2.Token,
) (*PlaylistMembership, error) {
	if playlistID == "" {
		return nil, fmt.Errorf("playlist ID must be provided")
	}
	if videoID == "" {
		return nil, fmt.Errorf("video ID must be provided")
	}
	if token == nil {
		return nil, fmt.Errorf("token must be provided")
	}

	service, err := c.Service(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}
	membership := &PlaylistMembership{
		PlaylistID: playlistID,
		ItemIDs:    []string{},
		Positions:  []int64{},
	}
	pageToken := ""
	for {
		call := service.PlaylistItems.List([]string{"id", "snippet"}).
			PlaylistId(playlistID).
			VideoId(videoID).
			MaxResults(50)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		response, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf(
				"failed to inspect playlist %s for video %s: %w",
				playlistID,
				videoID,
				err,
			)
		}
		for _, item := range response.Items {
			membership.ItemIDs = append(membership.ItemIDs, item.Id)
			if item.Snippet != nil {
				membership.Positions = append(membership.Positions, item.Snippet.Position)
			}
		}
		if response.NextPageToken == "" {
			break
		}
		pageToken = response.NextPageToken
	}

	membership.Count = len(membership.ItemIDs)
	membership.Present = membership.Count > 0
	return membership, nil
}

func (c *Core) GetPlaylistSnapshot(
	ctx context.Context,
	playlistID string,
	token *oauth2.Token,
) (*PlaylistSnapshot, error) {
	if playlistID == "" {
		return nil, fmt.Errorf("playlist ID must be provided")
	}
	if token == nil {
		return nil, fmt.Errorf("token must be provided")
	}

	service, err := c.Service(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}
	snapshot := &PlaylistSnapshot{
		PlaylistID: playlistID,
		Items:      []PlaylistSnapshotItem{},
	}
	pageToken := ""
	for {
		call := service.PlaylistItems.List([]string{"id", "snippet"}).
			PlaylistId(playlistID).
			MaxResults(50)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		response, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to inspect playlist %s: %w", playlistID, err)
		}
		for _, item := range response.Items {
			snapshotItem := PlaylistSnapshotItem{ItemID: item.Id}
			if item.Snippet != nil {
				snapshotItem.Position = item.Snippet.Position
				if item.Snippet.ResourceId != nil {
					snapshotItem.VideoID = item.Snippet.ResourceId.VideoId
				}
			}
			snapshot.Items = append(snapshot.Items, snapshotItem)
		}
		if response.NextPageToken == "" {
			break
		}
		pageToken = response.NextPageToken
	}

	sort.SliceStable(snapshot.Items, func(i, j int) bool {
		return snapshot.Items[i].Position < snapshot.Items[j].Position
	})
	snapshot.Count = len(snapshot.Items)
	return snapshot, nil
}

func stringIfPresent(fields map[string]json.RawMessage, key string, value string) *string {
	if _, ok := fields[key]; !ok {
		return nil
	}
	copy := value
	return &copy
}

func boolIfPresent(fields map[string]json.RawMessage, key string, value bool) *bool {
	if _, ok := fields[key]; !ok {
		return nil
	}
	copy := value
	return &copy
}
