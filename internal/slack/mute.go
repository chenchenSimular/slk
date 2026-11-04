package slackclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/slack-go/slack"

	"github.com/gammons/slk/internal/debuglog"
	"github.com/gammons/slk/internal/slackhttp"
)

// MuteRejectedError is Slack's refusal of a users.prefs.setNotifications
// call: ok=false with this error code.
type MuteRejectedError struct {
	Code string
}

func (e *MuteRejectedError) Error() string {
	return "users.prefs.setNotifications: " + e.Code
}

// setChannelNotificationReason is the _x_reason the web client sends
// when a channel's notification override changes from its channel menu.
const setChannelNotificationReason = "prefs-store/setMultiChannelNotificationOverride"

// muteForm builds the users.prefs.setNotifications form that mutes or
// unmutes one conversation for the current user.
//
// The endpoint is undocumented. The shape is the web client's
// multi-pref form (global=false, channel_ids, a JSON prefs array,
// sync=false), as read from the web client's preference adapters by
// open-source clients (3kh0/super-platinum's channel-mute request).
// The older single-pref form (name=muted, value=true, channel_id=…)
// is answered with invalid_arguments by current workspaces, so it is
// not used.
//
// Slack echoes the change as a pref_change event for
// all_notifications_prefs, which MuteStore.ApplyPrefChange consumes, so
// other clients and slk's own sidebar stay in step.
func muteForm(channelID string, muted bool) url.Values {
	value := "false"
	if muted {
		value = "true"
	}
	prefs, _ := json.Marshal([]map[string]string{{"name": "muted", "value": value}})
	return url.Values{
		"global":      {"false"},
		"channel_ids": {channelID},
		"prefs":       {string(prefs)},
		"sync":        {"false"},
	}
}

// SetChannelMuted mutes (muted=true) or unmutes one conversation for the
// authenticated user, server-side, so the change reaches every client.
//
// No retry: the call is idempotent, but a rate-limited user asked once
// and gets told so (*slack.RateLimitedError). Slack's refusal is
// *MuteRejectedError.
func (c *Client) SetChannelMuted(ctx context.Context, channelID string, muted bool) error {
	if channelID == "" {
		return errors.New("users.prefs.setNotifications: empty channel")
	}
	ctx = slackhttp.WithReason(ctx, setChannelNotificationReason)
	debuglog.General("users.prefs.setNotifications: channel=%s muted=%v", channelID, muted)
	raw, err := c.postForm(ctx, "users.prefs.setNotifications", muteForm(channelID, muted))
	if err == nil {
		err = parseSetNotifications(raw)
	}
	if err != nil {
		debuglog.General("users.prefs.setNotifications: error=%v", err)
		var rl *rateLimitError
		if errors.As(err, &rl) {
			return &slack.RateLimitedError{RetryAfter: rl.retryAfter}
		}
		return err
	}
	return nil
}

// parseSetNotifications reads users.prefs.setNotifications' answer.
func parseSetNotifications(raw []byte) error {
	var resp struct {
		OK    *bool  `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("parsing users.prefs.setNotifications: %w (body=%s)", err, truncateForLog(raw))
	}
	if resp.OK == nil {
		return fmt.Errorf("parsing users.prefs.setNotifications: no ok field (body=%s)", truncateForLog(raw))
	}
	if !*resp.OK {
		return &MuteRejectedError{Code: resp.Error}
	}
	return nil
}
