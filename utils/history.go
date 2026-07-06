package utils

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type historySource struct {
	path  string
	shell string
}

type countedCommand struct {
	command string
	count   int
}

// LoadHistoryItems reads shell history, deduplicates commands, and annotates
// each item with usage count so the list reflects both recency and frequency.
func LoadHistoryItems() ([]Item, error) {
	source, err := detectHistorySource()
	if err != nil {
		return nil, err
	}

	// We pass these down to be populated during file streaming,
	// saving massive amounts of slice allocations.
	seen := make(map[string]int)
	var orderedCounts []countedCommand

	err = parseHistory(source, seen, &orderedCounts)
	if err != nil {
		return nil, err
	}

	// Reverse the orderedCounts so the newest commands are first
	// (Since history files are appended chronologically)
	items := make([]Item, 0, len(orderedCounts))
	for i := len(orderedCounts) - 1; i >= 0; i-- {
		entry := orderedCounts[i]

		// Optimize description formatting without closures or heavy fmt.Sprintf
		var desc string
		if entry.count == 1 {
			desc = "Used once"
		} else {
			desc = strconv.Itoa(entry.count) + " times used"
		}

		items = append(items, NewItem(entry.command, desc))
	}

	return items, nil
}

func parseHistory(source historySource, seen map[string]int, ordered *[]countedCommand) error {
	file, err := os.Open(source.path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use an increased buffer scanner to prevent crash on massive single-line entries
	reader := bufio.NewReader(file)

	switch source.shell {
	case "zsh":
		return streamZshHistory(reader, seen, ordered)
	case "fish":
		return streamFishHistory(reader, seen, ordered)
	default:
		return streamPlainHistory(reader, seen, ordered)
	}
}

// Helper logic to add/update tracked counts inline
func trackCommand(command string, seen map[string]int, ordered *[]countedCommand) {
	command = normalizeCommand(command)
	if command == "" {
		return
	}

	if idx, ok := seen[command]; ok {
		(*ordered)[idx].count++
	} else {
		seen[command] = len(*ordered)
		*ordered = append(*ordered, countedCommand{
			command: command,
			count:   1,
		})
	}
}

func streamPlainHistory(r *bufio.Reader, seen map[string]int, ordered *[]countedCommand) error {
	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}

		trackCommand(line, seen, ordered)

		if err == io.EOF {
			break
		}
	}
	return nil
}

func streamZshHistory(r *bufio.Reader, seen map[string]int, ordered *[]countedCommand) error {
	var builder strings.Builder

	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if err == io.EOF {
				break
			}
			continue
		}

		// Handle Zsh metadata strip
		if builder.Len() == 0 && strings.HasPrefix(trimmed, ": ") {
			if idx := strings.Index(trimmed, ";"); idx >= 0 && idx+1 < len(trimmed) {
				trimmed = trimmed[idx+1:]
			}
		}

		// Handle Zsh multiline backslash line continuations
		if strings.HasSuffix(trimmed, "\\") {
			builder.WriteString(strings.TrimSuffix(trimmed, "\\"))
			builder.WriteByte(' ')
		} else {
			builder.WriteString(trimmed)
			trackCommand(builder.String(), seen, ordered)
			builder.Reset()
		}

		if err == io.EOF {
			break
		}
	}
	return nil
}

func streamFishHistory(r *bufio.Reader, seen map[string]int, ordered *[]countedCommand) error {
	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}

		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- cmd:") {
			command := strings.TrimSpace(strings.TrimPrefix(trimmed, "- cmd:"))
			command = strings.ReplaceAll(command, "\\n", " ")
			command = strings.ReplaceAll(command, "\\\\", "\\")
			trackCommand(command, seen, ordered)
		}

		if err == io.EOF {
			break
		}
	}
	return nil
}

func normalizeCommand(command string) string {
	command = strings.TrimSpace(command)
	command = strings.ReplaceAll(command, "\r", "")
	command = strings.ReplaceAll(command, "\n", " ")
	return command
}

// detectHistorySource finds the first usable history file so the loader can
// support bash, zsh, and fish without requiring manual configuration.
func detectHistorySource() (historySource, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return historySource{}, err
	}

	var candidates []historySource

	if histfile := strings.TrimSpace(os.Getenv("HISTFILE")); histfile != "" {
		candidates = append(candidates, historySource{
			path:  expandHome(histfile, homeDir),
			shell: shellFromPath(histfile, os.Getenv("SHELL")),
		})
	}

	shellEnv := shellFromPath(os.Getenv("SHELL"), "")
	xdgDataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME"))
	if xdgDataHome == "" {
		xdgDataHome = filepath.Join(homeDir, ".local", "share")
	}

	defaults := []historySource{
		{path: filepath.Join(homeDir, ".zsh_history"), shell: "zsh"},
		{path: filepath.Join(homeDir, ".bash_history"), shell: "bash"},
		{path: filepath.Join(xdgDataHome, "fish", "fish_history"), shell: "fish"},
	}

	if shellEnv != "" {
		var prioritized []historySource
		var remainder []historySource
		for _, candidate := range defaults {
			if candidate.shell == shellEnv {
				prioritized = append(prioritized, candidate)
				continue
			}
			remainder = append(remainder, candidate)
		}
		candidates = append(candidates, prioritized...)
		candidates = append(candidates, remainder...)
	} else {
		candidates = append(candidates, defaults...)
	}

	seen := make(map[string]struct{})
	for _, candidate := range candidates {
		if candidate.path == "" {
			continue
		}
		if _, ok := seen[candidate.path]; ok {
			continue
		}
		seen[candidate.path] = struct{}{}

		info, err := os.Stat(candidate.path)
		if err != nil || info.IsDir() {
			continue
		}
		if candidate.shell == "" {
			candidate.shell = shellFromPath(candidate.path, os.Getenv("SHELL"))
		}
		if candidate.shell == "" {
			candidate.shell = "shell"
		}
		return candidate, nil
	}

	return historySource{}, errors.New("no supported shell history file found")
}

// expandHome expands leading ~ so environment-derived paths resolve correctly
// when checking history file existence.
func expandHome(path string, homeDir string) string {
	if path == "~" {
		return homeDir
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

// shellFromPath infers shell type from file or executable names so parser
// selection can work even with partial environment data.
func shellFromPath(path string, fallback string) string {
	base := strings.ToLower(filepath.Base(path))

	switch {
	case strings.Contains(base, "zsh"):
		return "zsh"
	case strings.Contains(base, "bash"):
		return "bash"
	case strings.Contains(base, "fish"):
		return "fish"
	}

	fallbackBase := strings.ToLower(filepath.Base(fallback))
	switch fallbackBase {
	case "zsh":
		return "zsh"
	case "bash":
		return "bash"
	case "fish":
		return "fish"
	default:
		return ""
	}
}
