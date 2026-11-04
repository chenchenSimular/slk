// internal/ui/mute.go
//
// :mute / :unmute: mute or unmute the conversation on screen on Slack's
// servers (users.prefs.setNotifications), so every client follows.
package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/gammons/slk/internal/ids"
	"github.com/gammons/slk/internal/ui/sidebar"
)

func cmdMute(a *App, _ commandArgs) tea.Cmd   { return a.setMuteOnScreen(true) }
func cmdUnmute(a *App, _ commandArgs) tea.Cmd { return a.setMuteOnScreen(false) }

// setMuteOnScreen asks Slack to mute or unmute conversationOnScreen.
// The answer arrives as ChannelMuteSetMsg (reduceMute).
func (a *App) setMuteOnScreen(muted bool) tea.Cmd {
	channelID := a.conversationOnScreen()
	if channelID == "" {
		return toastWithClear(a, "No conversation open", 2*time.Second)
	}
	cmd := teaCmd(a.channels.SetMuted(a.activeTeamID, ids.ChannelID(channelID), muted))
	if cmd == nil {
		return toastWithClear(a, "Muting is not available", 2*time.Second)
	}
	verb := "Unmuting "
	if muted {
		verb = "Muting "
	}
	return tea.Batch(toastWithClear(a, verb+a.conversationLabel(channelID)+"…", 5*time.Second), cmd)
}

// conversationLabel names a sidebar conversation for a toast: "#name"
// for a channel, the name for a DM, the ID when the sidebar lacks it.
func (a *App) conversationLabel(channelID string) string {
	if it, ok := a.sidebarItem(channelID); ok && it.Name != "" {
		if it.Type == "channel" || it.Type == "private" {
			return "#" + it.Name
		}
		return it.Name
	}
	return channelID
}

// sidebarItem returns the active workspace's sidebar item for channelID.
func (a *App) sidebarItem(channelID string) (sidebar.ChannelItem, bool) {
	for _, it := range a.sidebar.Items() {
		if it.ID == channelID {
			return it, true
		}
	}
	return sidebar.ChannelItem{}, false
}

// reduceMute applies a mute change Slack accepted to the sidebar at
// once, without waiting for the pref_change echo. The echo is what
// updates the workspace's mute store and every other copy of the flag
// (cmd/slk OnPrefChange); it then rebuilds the sidebar with the same
// value.
var reduceMute reducerFunc = func(a *App, msg tea.Msg) (tea.Cmd, bool) {
	m, ok := msg.(ChannelMuteSetMsg)
	if !ok {
		return nil, false
	}
	label := m.ChannelID
	if m.TeamID == a.activeTeamID {
		label = a.conversationLabel(m.ChannelID)
	}
	if m.Err != nil {
		verb := "unmute "
		if m.Muted {
			verb = "mute "
		}
		return toastWithClear(a, "Could not "+verb+label+": "+m.Err.Error(), 4*time.Second), true
	}
	if m.TeamID == a.activeTeamID {
		if it, ok := a.sidebarItem(m.ChannelID); ok && it.IsMuted != m.Muted {
			it.IsMuted = m.Muted
			a.sidebar.UpsertItem(it)
		}
	}
	text := "Unmuted " + label
	if m.Muted {
		text = "Muted " + label
	}
	return toastWithClear(a, text, 3*time.Second), true
}
