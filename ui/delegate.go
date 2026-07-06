package ui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// NewItemDelegate customizes selected-row styling and maps Enter to an execute
// message so list interactions remain centralized in the model.
func NewItemDelegate(keys *DelegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	accent := lipgloss.Color("#a07902")
	fg := lipgloss.Color("#eab308")

	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(fg).
		BorderForeground(accent)

	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(accent).
		BorderForeground(accent)

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch {
			case key.Matches(msg, keys.Choose):
				return func() tea.Msg {
					return ExecuteSelectedCommandMsg{}
				}
			}
		}

		return nil
	}

	help := []key.Binding{keys.Choose}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}
