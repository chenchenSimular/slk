package ui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/gammons/slk/internal/core"
	"github.com/gammons/slk/internal/ids"
	"github.com/gammons/slk/internal/ui/sidebar"
)

type muteCall struct {
	teamID, channelID string
	muted             bool
}

// muteApp is an App in T1 on C1 (#general) whose SetMuted calls are
// recorded and answered with err.
func muteApp(t *testing.T, err error) (*App, *[]muteCall) {
	t.Helper()
	var calls []muteCall
	a := newTestApp(t,
		withSize(120, 30),
		withActiveTeam("T1"),
		withChannels(sidebar.ChannelItem{ID: "C1", Name: "general", Type: "channel"}),
		withActiveChannel("C1"),
		withChannelService(core.ChannelServiceFuncs{
			SetMuted: func(teamID string, channelID ids.ChannelID, muted bool) core.Cmd {
				calls = append(calls, muteCall{teamID, string(channelID), muted})
				return func() core.Msg {
					return ChannelMuteSetMsg{TeamID: teamID, ChannelID: string(channelID), Muted: muted, Err: err}
				}
			},
		}),
	)
	return a, &calls
}

// runMute runs a : command and feeds the service's answer back through
// Update, returning the App's sidebar item for C1.
func runMute(t *testing.T, a *App, line string) sidebar.ChannelItem {
	t.Helper()
	cmd := executeCommand(a, line)
	for _, m := range deliveredMsgs[ChannelMuteSetMsg](t, cmd) {
		a.Update(tea.Msg(m))
	}
	it, ok := a.sidebarItem("C1")
	if !ok {
		t.Fatal("C1 missing from the sidebar")
	}
	return it
}

func TestMuteCommandMutesTheConversationOnScreen(t *testing.T) {
	a, calls := muteApp(t, nil)
	it := runMute(t, a, "mute")
	if len(*calls) != 1 || (*calls)[0] != (muteCall{"T1", "C1", true}) {
		t.Fatalf("calls = %+v", *calls)
	}
	if !it.IsMuted {
		t.Error("sidebar item not muted after Slack accepted")
	}
	if out := statusbarText(a); !strings.Contains(out, "Muted #general") {
		t.Errorf("toast = %q", out)
	}

	it = runMute(t, a, "unmute")
	if (*calls)[1] != (muteCall{"T1", "C1", false}) {
		t.Fatalf("calls = %+v", *calls)
	}
	if it.IsMuted {
		t.Error("sidebar item still muted after :unmute")
	}
}

func TestMuteCommandFailureLeavesSidebarAlone(t *testing.T) {
	a, _ := muteApp(t, errors.New("users.prefs.setNotifications: invalid_arguments"))
	if it := runMute(t, a, "mute"); it.IsMuted {
		t.Error("sidebar item muted although Slack refused")
	}
	if out := statusbarText(a); !strings.Contains(out, "Could not mute #general") {
		t.Errorf("toast = %q", out)
	}
}

func TestMuteCommandWithNoConversationRefuses(t *testing.T) {
	a, calls := muteApp(t, nil)
	a.activeChannelID = ""
	_ = executeCommand(a, "mute")
	if len(*calls) != 0 {
		t.Fatalf("SetMuted called with no conversation: %+v", *calls)
	}
	if out := statusbarText(a); !strings.Contains(out, "No conversation open") {
		t.Errorf("toast = %q", out)
	}
}

// An answer for another workspace (the user switched meanwhile) must not
// mute the same-ID row of the one now on screen.
func TestMuteAnswerForAnotherWorkspaceIgnoresSidebar(t *testing.T) {
	a, _ := muteApp(t, nil)
	a.Update(ChannelMuteSetMsg{TeamID: "T2", ChannelID: "C1", Muted: true})
	if it, _ := a.sidebarItem("C1"); it.IsMuted {
		t.Error("T2's answer muted T1's row")
	}
}
