package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"schwifty/ui"
	"schwifty/utils"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

// initialModel builds the list UI model with history-backed items so startup has
// ready-to-render state and keybindings.
func initialModel() ui.Model {
	// Create a new model with default styles and dimensions
	m := ui.Model{}
	m.Styles = ui.NewStyles()
	// Create key maps for list and delegate so the UI can respond to user input
	delegateKeys := ui.NewDelegateKeyMap()
	listKeys := ui.NewListKeyMap()
	// Load command history items from the user's shell history
	historyItems, err := utils.LoadHistoryItems()
	if err != nil {
		historyItems = []utils.Item{
			utils.NewNoticeItem("No shell history found", err.Error()),
		}
	}
	// Convert history items to list items for the Bubble Tea list component
	items := make([]list.Item, len(historyItems))
	for i, item := range historyItems {
		items[i] = item
	}
	// Create a delegate for the list that handles item selection and keybindings
	delegate := ui.NewItemDelegate(delegateKeys)
	commandList := list.New(items, delegate, 0, 0)
	commandList.Title = "SCHWIFTY"
	commandList.Styles.Title = m.Styles.Title
	commandList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.ToggleTitleBar,
			listKeys.ToggleStatusBar,
			listKeys.TogglePagination,
			listKeys.ToggleHelpMenu,
		}
	}
	// Set the initial model state with the list, keys, and delegate keys
	m.List = commandList
	m.Keys = listKeys
	m.DelegateKeys = delegateKeys

	return m
}

// main runs the TUI and then executes the chosen command in the user's shell so
// the UI can close before command output takes over the terminal.
func main() {
	// Run the TUI and capture the final model state after user interaction
	finalModel, err := tea.NewProgram(initialModel()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	// Type assert the final model to ui.Model to access the selected command
	m, ok := finalModel.(ui.Model)
	if !ok {
		return
	}
	// If no command was selected, exit without executing anything.
	command := strings.TrimSpace(m.CommandToRun)
	if command == "" {
		return
	}
	// Print the command to be executed for verbose feedback
	fmt.Printf("%s\n", command)
	// Execute the command in the user's shell
	shellPath := strings.TrimSpace(os.Getenv("SHELL"))
	if shellPath == "" {
		shellPath = "/bin/sh"
	}
	// Use the appropriate shell flags for executing a command
	args := []string{"-lc", command}
	if filepath.Base(shellPath) == "fish" {
		args = []string{"-c", command}
	}
	// Create the command and set its standard input/output to the terminal
	cmd := exec.Command(shellPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Run the command and handle any errors
	if err := cmd.Run(); err != nil {
		fmt.Println("Error executing command:", err)
		os.Exit(1)
	}
}
