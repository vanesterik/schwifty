package ui

import (
	"strings"
	"sync"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	Styles        Styles
	Width, Height int
	Once          *sync.Once
	List          list.Model
	Keys          *ListKeyMap
	DelegateKeys  *DelegateKeyMap
	CommandToRun  string
}

type ExecuteSelectedCommandMsg struct{}

type titledItem interface {
	Title() string
}

type runnableItem interface {
	Runnable() bool
}

// Init requests initial terminal background information so styles can be
// refreshed consistently after startup.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
	)
}

// UpdateListProperties recalculates list dimensions and title styling so the UI
// stays aligned with current window size and theme.
func (m *Model) UpdateListProperties() {
	h, v := m.Styles.App.GetFrameSize()
	m.List.SetSize(m.Width-h, m.Height-v)

	m.Styles = NewStyles()
	m.List.Styles.Title = m.Styles.Title
}

// Update handles resize/theme/key events and captures execute intent so command
// execution can happen after the TUI exits.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.UpdateListProperties()
		return m, nil

	case tea.WindowSizeMsg:
		m.Width, m.Height = msg.Width, msg.Height
		m.UpdateListProperties()
		return m, nil

	case ExecuteSelectedCommandMsg:
		selected := m.List.SelectedItem()
		item, ok := selected.(titledItem)
		if !ok {
			return m, tea.Quit
		}

		r, ok := selected.(runnableItem)
		if ok && !r.Runnable() {
			return m, tea.Quit
		}

		m.CommandToRun = strings.TrimSpace(item.Title())
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.List.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.Keys.ToggleTitleBar):
			v := !m.List.ShowTitle()
			m.List.SetShowTitle(v)
			m.List.SetShowFilter(v)
			m.List.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, m.Keys.ToggleStatusBar):
			m.List.SetShowStatusBar(!m.List.ShowStatusBar())
			return m, nil

		case key.Matches(msg, m.Keys.TogglePagination):
			m.List.SetShowPagination(!m.List.ShowPagination())
			return m, nil

		case key.Matches(msg, m.Keys.ToggleHelpMenu):
			m.List.SetShowHelp(!m.List.ShowHelp())
			return m, nil
		}
	}

	newListModel, cmd := m.List.Update(msg)
	m.List = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the full application in alt-screen mode so the interface stays
// isolated from normal terminal output.
func (m Model) View() tea.View {
	v := tea.NewView(m.Styles.App.Render(m.List.View()))
	v.AltScreen = true
	return v
}
