package sidebar

import (
	"strings"
	"testing"

	"github.com/gammons/slk/internal/cache"
)

// filterFixture: one conversation of each kind, read or unread.
func filterFixture(t *testing.T) (*Model, map[string]cache.ReadState) {
	t.Helper()
	m := New([]ChannelItem{
		{ID: "C1", Name: "general", Type: "channel"},
		{ID: "C2", Name: "random", Type: "channel"},
		{ID: "C3", Name: "noisy", Type: "channel", IsMuted: true},
		{ID: "C4", Name: "pinged", Type: "channel", IsMuted: true},
		{ID: "D1", Name: "alice", Type: "dm"},
		{ID: "D2", Name: "bob", Type: "dm"},
		{ID: "A1", Name: "github", Type: "app"},
		{ID: "A2", Name: "pagerduty", Type: "app"},
		{ID: "A3", Name: "jira", Type: "app"},
	})
	state := map[string]cache.ReadState{
		"C1": {LastReadTS: "1.0", HasUnread: true},
		"C2": {LastReadTS: "1.0"},
		"C3": {LastReadTS: "1.0", HasUnread: true},
		"C4": {LastReadTS: "1.0", HasUnread: true, MentionCount: 1},
		"D1": {LastReadTS: "1.0", HasUnread: true},
		"D2": {LastReadTS: "1.0"},
		"A1": {LastReadTS: "1.0"},
		"A2": {LastReadTS: "1.0", HasUnread: true},
		"A3": {LastReadTS: "1.0"},
	}
	m.SetReadStateReader(func() map[string]cache.ReadState { return state })
	return &m, state
}

func TestUnreadOnly_ShowsOnlyUnreadAndActive(t *testing.T) {
	m, _ := filterFixture(t)
	m.SetActiveChannelID("C2")
	assertVisible(t, m, []string{"general", "random", "noisy", "pinged", "alice", "bob", "github"}, nil)

	m.SetUnreadOnly(true)
	if !m.UnreadOnly() {
		t.Fatal("UnreadOnly() = false after SetUnreadOnly(true)")
	}
	// random is read but active; pinged is muted but mentions you;
	// noisy is muted chatter.
	assertVisible(t, m,
		[]string{"general", "random", "pinged", "alice", "pagerduty"},
		[]string{"noisy", "bob", "github", "jira"})

	m.SetUnreadOnly(false)
	assertVisible(t, m, []string{"noisy", "bob", "github", "jira"}, nil)
}

// A conversation that turns unread shows up, and one read elsewhere
// goes, on the read-state change (Invalidate) alone.
func TestUnreadOnly_FollowsReadStateChanges(t *testing.T) {
	m, state := filterFixture(t)
	m.SetUnreadOnly(true)
	state["D2"] = cache.ReadState{LastReadTS: "2.0", HasUnread: true}
	state["C1"] = cache.ReadState{LastReadTS: "2.0"}
	m.Invalidate()
	assertVisible(t, m, []string{"bob"}, []string{"general"})
}

func TestUnreadOnly_LabelsTheSidebar(t *testing.T) {
	m, _ := filterFixture(t)
	if view := m.View(30, 30); strings.Contains(view, unreadOnlyLabel) {
		t.Fatalf("label shown with the filter off:\n%s", view)
	}
	m.SetUnreadOnly(true)
	if view := m.View(30, 30); !strings.Contains(view, unreadOnlyLabel) {
		t.Fatalf("no %q row with the filter on:\n%s", unreadOnlyLabel, view)
	}
	// The label fills the separator row below Threads/Activity, and is
	// truncated rather than wrapped: renderRow counts it as one line.
	for _, width := range []int{30, 12, 6} {
		m.View(30, width)
		sep := m.cacheRows[2]
		if sep.navIdx != -1 || sep.normal == "" {
			t.Fatalf("width %d: row 2 is not the labelled separator: %+v", width, sep)
		}
		if strings.Contains(sep.normal, "\n") {
			t.Errorf("width %d: label wrapped: %q", width, sep.normal)
		}
	}
}

func TestHideApps_HidesReadAppDMsOnly(t *testing.T) {
	m, _ := filterFixture(t)
	m.SetActiveChannelID("A3")
	m.SetHideApps(true)
	assertVisible(t, m,
		[]string{"general", "random", "noisy", "alice", "bob", "pagerduty", "jira"},
		[]string{"github"})

	m.SetHideApps(false)
	assertVisible(t, m, []string{"github"}, nil)
}

// hide_apps leaves the items themselves alone, so whatever reads
// AllItems (the finder's source is separate, but nothing should lose
// the row) still sees the hidden app DM.
func TestHideApps_KeepsItemsForOtherReaders(t *testing.T) {
	m, _ := filterFixture(t)
	m.SetHideApps(true)
	found := false
	for _, it := range m.AllItems() {
		if it.Name == "github" {
			found = true
		}
	}
	if !found {
		t.Error("hide_apps dropped github from AllItems")
	}
}
