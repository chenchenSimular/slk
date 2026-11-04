package ui

import (
	"strings"
	"testing"
)

func TestUnreadsCommandTogglesSidebarFilter(t *testing.T) {
	a := NewApp()
	_ = executeCommand(a, "unreads")
	if !a.sidebar.UnreadOnly() {
		t.Fatal(":unreads did not turn the sidebar's unread-only filter on")
	}
	if out := statusbarText(a); !strings.Contains(out, "Sidebar: unreads only") {
		t.Errorf("toast = %q", out)
	}
	_ = executeCommand(a, "unreads")
	if a.sidebar.UnreadOnly() {
		t.Fatal("second :unreads left the filter on")
	}
	if out := statusbarText(a); !strings.Contains(out, "Sidebar: all conversations") {
		t.Errorf("toast = %q", out)
	}
}

func TestFKeyTogglesSidebarUnreadOnly(t *testing.T) {
	a := newTestApp(t, withSize(120, 30))
	_ = handleNormalMode(a, keyPress('f'))
	if !a.sidebar.UnreadOnly() {
		t.Fatal("f did not turn the unread-only filter on")
	}
	_ = handleNormalMode(a, keyPress('f'))
	if a.sidebar.UnreadOnly() {
		t.Fatal("second f left the filter on")
	}
}

func TestSidebarStyleAppliesUnreadOnlyAndHideApps(t *testing.T) {
	a := newTestApp(t)
	a.SetSidebarStyle(SidebarStyle{UnreadOnly: true, HideApps: true})
	if !a.sidebar.UnreadOnly() || !a.sidebar.HideApps() {
		t.Fatalf("unread_only=%v hide_apps=%v, want both on", a.sidebar.UnreadOnly(), a.sidebar.HideApps())
	}
}
