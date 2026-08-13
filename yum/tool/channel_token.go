package tool

import (
	"context"
	"fmt"
	"time"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"golang.org/x/oauth2"
)

func validChannelToken(ctx context.Context, c *core.Core, channelID string) (*oauth2.Token, error) {
	channel, err := c.GetChannelByID(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load token: %w", err)
	}
	if channel == nil || channel.Token == nil {
		return nil, fmt.Errorf("channel or token is nil, please authenticate first")
	}
	if channel.Token.Expiry.IsZero() || channel.Token.AccessToken == "" || channel.Token.RefreshToken == "" {
		return nil, fmt.Errorf("channel token is expired or malformed, please authenticate again")
	}

	now := time.Now().In(channel.Token.Expiry.Location())
	if channel.Token.Expiry.Before(now.Add(2 * time.Minute)) {
		newToken, err := c.RefreshAccessToken(channel.Token)
		if err != nil {
			return nil, fmt.Errorf("token was expiring, failed to refresh: %w", err)
		}
		channel.Token = newToken
		if err := c.SaveChannel(channel); err != nil {
			return nil, fmt.Errorf("token was expiring, failed to save refreshed token: %w", err)
		}
	}
	if err := c.VerifyTokenChannel(ctx, channelID, channel.Token); err != nil {
		return nil, err
	}

	return channel.Token, nil
}
