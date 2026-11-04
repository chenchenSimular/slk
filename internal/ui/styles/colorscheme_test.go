package styles

import (
	"testing"

	"github.com/gammons/slk/internal/core"
)

var zeroOverrides core.Theme

func TestResolveColorSchemeSpec(t *testing.T) {
	cases := []struct {
		spec       string
		dark, want string
	}{
		{"GitHub Dark", "dark", "GitHub Dark"},
		{"GitHub Dark", "light", "GitHub Dark"},
		{"light:GitHub Light,dark:GitHub Dark", "dark", "GitHub Dark"},
		{"light:GitHub Light,dark:GitHub Dark", "light", "GitHub Light"},
		{" Dark : Dracula , LIGHT: Solarized Light ", "light", "Solarized Light"},
		{"dark:Dracula", "light", "Dracula"},
		{"light:GitHub Light", "dark", "GitHub Light"},
		{"Theme: with colon", "dark", "Theme: with colon"},
	}
	for _, c := range cases {
		if got := ResolveColorSchemeSpec(c.spec, c.dark == "dark"); got != c.want {
			t.Errorf("ResolveColorSchemeSpec(%q, %s) = %q, want %q", c.spec, c.dark, got, c.want)
		}
	}
}

func TestIsColorSchemeSpec(t *testing.T) {
	for spec, want := range map[string]bool{
		"GitHub Dark":                          false,
		"":                                     false,
		"light:,dark:":                         false,
		"light:GitHub Light,dark:GitHub Dark":  true,
		"dark:Dracula":                         true,
		"light:GitHub Light,other:GitHub Dark": false,
	} {
		if got := IsColorSchemeSpec(spec); got != want {
			t.Errorf("IsColorSchemeSpec(%q) = %v, want %v", spec, got, want)
		}
	}
}

func TestApplyFollowsColorScheme(t *testing.T) {
	defer SetDarkBackground(DarkBackground())
	defer Apply(CurrentSpec(), zeroOverrides)

	SetDarkBackground(false)
	Apply("light:GitHub Light,dark:GitHub Dark", zeroOverrides)
	if CurrentTheme() != "GitHub Light" {
		t.Fatalf("light scheme applied %q, want GitHub Light", CurrentTheme())
	}
	if !SetDarkBackground(true) {
		t.Fatal("SetDarkBackground(true) reported no change")
	}
	Apply(CurrentSpec(), zeroOverrides)
	if CurrentTheme() != "GitHub Dark" {
		t.Fatalf("dark scheme applied %q, want GitHub Dark", CurrentTheme())
	}
}
