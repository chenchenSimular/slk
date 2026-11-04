package styles

import "strings"

// A theme setting may name one theme per terminal color scheme instead of a
// single theme, in Ghostty's syntax:
//
//	theme = "light:GitHub Light,dark:GitHub Dark"
//
// Apply resolves such a spec against the scheme last reported through
// SetDarkBackground (dark until the terminal answers), and the App re-applies
// it when the terminal switches between light and dark.

// darkBackground is the terminal color scheme the last SetDarkBackground call
// reported. It starts dark: slk's default theme is dark, and a terminal that
// never answers the background query keeps today's behavior.
var darkBackground = true

// currentSpec is the theme setting the last Apply call received, before
// resolution: a plain theme name or a light/dark spec.
var currentSpec string

// SetDarkBackground records the terminal's color scheme and reports whether it
// changed.
func SetDarkBackground(dark bool) bool {
	if darkBackground == dark {
		return false
	}
	darkBackground = dark
	return true
}

// DarkBackground reports the terminal color scheme last recorded.
func DarkBackground() bool { return darkBackground }

// CurrentSpec returns the theme setting the last Apply call received, which is
// a light/dark spec when the applied theme follows the terminal.
func CurrentSpec() string { return currentSpec }

// IsColorSchemeSpec reports whether s names a theme per color scheme
// ("light:…", "dark:…", or both, comma separated).
func IsColorSchemeSpec(s string) bool {
	_, _, ok := parseColorSchemeSpec(s)
	return ok
}

// ResolveColorSchemeSpec returns the theme name s selects for the given
// scheme. A plain theme name is returned unchanged. A spec that names only one
// side uses that theme for both.
func ResolveColorSchemeSpec(s string, dark bool) string {
	light, darkName, ok := parseColorSchemeSpec(s)
	if !ok {
		return s
	}
	if dark {
		if darkName != "" {
			return darkName
		}
		return light
	}
	if light != "" {
		return light
	}
	return darkName
}

func parseColorSchemeSpec(s string) (light, dark string, ok bool) {
	for _, part := range strings.Split(s, ",") {
		key, value, found := strings.Cut(part, ":")
		if !found {
			return "", "", false
		}
		value = strings.TrimSpace(value)
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "light":
			light = value
		case "dark":
			dark = value
		default:
			return "", "", false
		}
	}
	return light, dark, light != "" || dark != ""
}
