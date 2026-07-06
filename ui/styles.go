package ui

import (
	"charm.land/lipgloss/v2"
)

type Styles struct {
	App   lipgloss.Style
	Title lipgloss.Style
}

// NewStyles returns base app and title styles so layout and branding stay
// consistent across redraws.
func NewStyles() Styles {
	return Styles{
		App: lipgloss.NewStyle().
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#eab308")).
			Padding(0, 1),
	}
}
