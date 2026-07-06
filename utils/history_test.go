package utils

import (
	"bufio"
	"strings"
	"testing"
)

func TestTrackCommandCountsAndOrder(t *testing.T) {
	seen := make(map[string]int)
	ordered := make([]countedCommand, 0)

	trackCommand("  ls\n", seen, &ordered)
	trackCommand("pwd", seen, &ordered)
	trackCommand("ls", seen, &ordered)
	trackCommand("\n", seen, &ordered)
	trackCommand("   ", seen, &ordered)

	expected := []countedCommand{
		{command: "ls", count: 2},
		{command: "pwd", count: 1},
	}

	if len(ordered) != len(expected) {
		t.Fatalf("len mismatch: got %d, want %d", len(ordered), len(expected))
	}

	for i := range expected {
		if ordered[i] != expected[i] {
			t.Fatalf("entry %d mismatch: got %+v, want %+v", i, ordered[i], expected[i])
		}
	}

	if seen["ls"] != 0 {
		t.Fatalf("seen index for ls = %d, want 0", seen["ls"])
	}
	if seen["pwd"] != 1 {
		t.Fatalf("seen index for pwd = %d, want 1", seen["pwd"])
	}
}

func TestStreamPlainHistory(t *testing.T) {
	t.Run("counts duplicates", func(t *testing.T) {
		seen := make(map[string]int)
		ordered := make([]countedCommand, 0)
		reader := bufio.NewReader(strings.NewReader("ls\npwd\nls\n"))

		if err := streamPlainHistory(reader, seen, &ordered); err != nil {
			t.Fatalf("streamPlainHistory returned error: %v", err)
		}

		expected := []countedCommand{{command: "ls", count: 2}, {command: "pwd", count: 1}}
		assertCountedCommandsEqual(t, ordered, expected)
	})

	t.Run("handles eof without trailing newline", func(t *testing.T) {
		seen := make(map[string]int)
		ordered := make([]countedCommand, 0)
		reader := bufio.NewReader(strings.NewReader("ls\npwd"))

		if err := streamPlainHistory(reader, seen, &ordered); err != nil {
			t.Fatalf("streamPlainHistory returned error: %v", err)
		}

		expected := []countedCommand{{command: "ls", count: 1}, {command: "pwd", count: 1}}
		assertCountedCommandsEqual(t, ordered, expected)
	})
}

func TestStreamZshHistory(t *testing.T) {
	seen := make(map[string]int)
	ordered := make([]countedCommand, 0)

	content := ": 1:0;git status\n" +
		": 2:0;echo hi \\\n" +
		"there\n" +
		"plain\n"

	reader := bufio.NewReader(strings.NewReader(content))
	if err := streamZshHistory(reader, seen, &ordered); err != nil {
		t.Fatalf("streamZshHistory returned error: %v", err)
	}

	expected := []countedCommand{
		{command: "git status", count: 1},
		{command: "echo hi  there", count: 1},
		{command: "plain", count: 1},
	}
	assertCountedCommandsEqual(t, ordered, expected)
}

func TestStreamFishHistory(t *testing.T) {
	seen := make(map[string]int)
	ordered := make([]countedCommand, 0)

	content := "- cmd: ls\n" +
		"- cmd: echo\\nhello\n" +
		"- cmd: echo\\\\world\n" +
		"other: ignore me\n"

	reader := bufio.NewReader(strings.NewReader(content))
	if err := streamFishHistory(reader, seen, &ordered); err != nil {
		t.Fatalf("streamFishHistory returned error: %v", err)
	}

	expected := []countedCommand{
		{command: "ls", count: 1},
		{command: "echo hello", count: 1},
		{command: "echo\\world", count: 1},
	}
	assertCountedCommandsEqual(t, ordered, expected)
}

func TestNormalizeCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "trim whitespace", input: "  git status  ", expected: "git status"},
		{name: "remove carriage returns", input: "echo hi\r", expected: "echo hi"},
		{name: "replace newlines", input: "echo\nhello", expected: "echo hello"},
		{name: "combined", input: " \r\nfoo\nbar\r ", expected: "foo bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeCommand(tt.input)
			if got != tt.expected {
				t.Fatalf("normalizeCommand(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	home := "/Users/tester"

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "tilde only", input: "~", expected: home},
		{name: "tilde slash", input: "~/projects", expected: "/Users/tester/projects"},
		{name: "absolute path unchanged", input: "/tmp/file", expected: "/tmp/file"},
		{name: "other user form unchanged", input: "~someone/file", expected: "~someone/file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandHome(tt.input, home)
			if got != tt.expected {
				t.Fatalf("expandHome(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestShellFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		fallback string
		expected string
	}{
		{name: "detect from path zsh", path: "/Users/me/.zsh_history", fallback: "", expected: "zsh"},
		{name: "detect from path bash", path: "/Users/me/.bash_history", fallback: "", expected: "bash"},
		{name: "detect from path fish", path: "/Users/me/.local/share/fish/fish_history", fallback: "", expected: "fish"},
		{name: "fallback used", path: "/tmp/history", fallback: "/bin/zsh", expected: "zsh"},
		{name: "fallback empty", path: "/tmp/history", fallback: "", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shellFromPath(tt.path, tt.fallback)
			if got != tt.expected {
				t.Fatalf("shellFromPath(%q, %q) = %q, want %q", tt.path, tt.fallback, got, tt.expected)
			}
		})
	}
}

func assertCountedCommandsEqual(t *testing.T, got []countedCommand, expected []countedCommand) {
	t.Helper()

	if len(got) != len(expected) {
		t.Fatalf("len mismatch: got %d, want %d", len(got), len(expected))
	}

	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("entry %d mismatch: got %+v, want %+v", i, got[i], expected[i])
		}
	}
}
