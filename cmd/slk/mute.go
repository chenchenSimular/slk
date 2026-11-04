package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gammons/slk/internal/ids"
	"github.com/gammons/slk/internal/ui"
)

// setMutedTimeout bounds one users.prefs.setNotifications call.
const setMutedTimeout = 10 * time.Second

// channelMuter is the part of *slackclient.Client muting needs.
type channelMuter interface {
	SetChannelMuted(ctx context.Context, channelID string, muted bool) error
}

// setMuted mutes or unmutes channelID through client, the client of
// workspace teamID, and reports the result for the UI. client is nil when
// that workspace is gone (switched away or signed out since the command).
//
// Nothing here touches the workspace's MuteStore: Slack echoes the change
// as a pref_change for all_notifications_prefs, and OnPrefChange applies
// it on the event goroutine, where wctx.Channels is owned. Writing the
// store here first would make that echo look like no change and leave
// wctx.Channels stale.
func setMuted(ctx context.Context, client channelMuter, teamID string, channelID ids.ChannelID, muted bool) ui.ChannelMuteSetMsg {
	msg := ui.ChannelMuteSetMsg{TeamID: teamID, ChannelID: string(channelID), Muted: muted}
	if client == nil {
		msg.Err = errors.New("workspace unavailable")
		return msg
	}
	ctx, cancel := context.WithTimeout(ctx, setMutedTimeout)
	defer cancel()
	if err := client.SetChannelMuted(ctx, string(channelID), muted); err != nil {
		log.Printf("Warning: failed to set muted=%v on %s: %v", muted, channelID, err)
		msg.Err = err
	}
	return msg
}
