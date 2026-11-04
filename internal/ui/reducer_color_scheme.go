// internal/ui/reducer_color_scheme.go
//
// Color-scheme reducer for App.Update. Lets a theme setting of the form
// "light:<theme>,dark:<theme>" follow the terminal's light/dark scheme
// (styles/colorscheme.go):
//
//	colorSchemePollMsg       - time to ask the terminal for its background
//	                           color again (OSC 11), and schedule the next ask.
//	tea.BackgroundColorMsg   - the terminal's answer. When the scheme flips
//	                           and the applied theme is a light/dark spec,
//	                           re-apply it so the other side takes effect.
//
// Polling OSC 11 rather than subscribing to mode 2031 keeps no terminal mode
// set that would need undoing on every exit path; tmux answers OSC 11 from the
// attached client, so this works inside tmux too.
package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/gammons/slk/internal/ui/styles"
)

// colorSchemePollInterval is how often slk re-asks the terminal for its
// background color, i.e. the longest a light/dark switch takes to show.
const colorSchemePollInterval = 2 * time.Second

type colorSchemePollMsg struct{}

// startColorSchemeWatch asks for the background color now and starts the poll.
func startColorSchemeWatch() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, scheduleColorSchemePoll())
}

func scheduleColorSchemePoll() tea.Cmd {
	return tea.Tick(colorSchemePollInterval, func(time.Time) tea.Msg {
		return colorSchemePollMsg{}
	})
}

var reduceColorScheme reducerFunc = func(a *App, msg tea.Msg) (tea.Cmd, bool) {
	switch m := msg.(type) {
	case colorSchemePollMsg:
		// Only a light/dark spec needs the answer; keep the timer running
		// so a spec applied later (theme switcher, workspace switch) is
		// picked up without a restart.
		if styles.IsColorSchemeSpec(styles.CurrentSpec()) {
			return tea.Batch(tea.RequestBackgroundColor, scheduleColorSchemePoll()), true
		}
		return scheduleColorSchemePoll(), true
	case tea.BackgroundColorMsg:
		if m.Color == nil {
			return nil, true
		}
		if styles.SetDarkBackground(m.IsDark()) && styles.IsColorSchemeSpec(styles.CurrentSpec()) {
			a.applyTheme(styles.CurrentSpec())
		}
		return nil, true
	}
	return nil, false
}
