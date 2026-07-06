package ui

import (
	"charm.land/bubbles/v2/key"
)

type ListKeyMap struct {
	ToggleTitleBar   key.Binding
	ToggleStatusBar  key.Binding
	TogglePagination key.Binding
	ToggleHelpMenu   key.Binding
}

// NewListKeyMap defines global list toggles so users can adjust layout and help
// visibility while browsing commands.
func NewListKeyMap() *ListKeyMap {
	return &ListKeyMap{
		ToggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		ToggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		TogglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		ToggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

type DelegateKeyMap struct {
	Choose key.Binding
}

// ShortHelp exposes compact delegate shortcuts so key hints fit in tight UI
// space.
func (d DelegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.Choose,
	}
}

// FullHelp exposes the complete delegate key grouping so the expanded help view
// can document all delegate actions.
func (d DelegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.Choose,
		},
	}
}

// NewDelegateKeyMap defines selection bindings so Enter behavior is explicit and
// reusable in delegate setup.
func NewDelegateKeyMap() *DelegateKeyMap {
	return &DelegateKeyMap{
		Choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
	}
}
