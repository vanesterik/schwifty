package utils

import "testing"

func TestNewItemAndAccessors(t *testing.T) {
	item := NewItem("git status", "3 times used")

	if item.Title() != "git status" {
		t.Fatalf("Title() = %q, want %q", item.Title(), "git status")
	}

	if item.Description() != "3 times used" {
		t.Fatalf("Description() = %q, want %q", item.Description(), "3 times used")
	}

	if item.FilterValue() != "git status" {
		t.Fatalf("FilterValue() = %q, want %q", item.FilterValue(), "git status")
	}

	if !item.Runnable() {
		t.Fatal("Runnable() = false, want true")
	}
}

func TestNewNoticeItemAndAccessors(t *testing.T) {
	item := NewNoticeItem("No shell history found", "missing .zsh_history")

	if item.Title() != "No shell history found" {
		t.Fatalf("Title() = %q, want %q", item.Title(), "No shell history found")
	}

	if item.Description() != "missing .zsh_history" {
		t.Fatalf("Description() = %q, want %q", item.Description(), "missing .zsh_history")
	}

	if item.FilterValue() != "No shell history found" {
		t.Fatalf("FilterValue() = %q, want %q", item.FilterValue(), "No shell history found")
	}

	if item.Runnable() {
		t.Fatal("Runnable() = true, want false")
	}
}
