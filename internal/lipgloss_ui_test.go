package internal

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewLipglossUI(t *testing.T) {
	ui := NewLipglossUI()
	if ui == nil {
		t.Error("Expected LipglossUI to be initialized, but it was nil")
	}
}

func TestLipglossUI_RenderHeader(t *testing.T) {
	tests := []struct {
		title       string
		ip          string
		deviceReady bool
		wantContain string
	}{
		{"Test Title", "127.0.0.1", true, "[O]"},
		{"Test Title", "127.0.0.1", false, "[X]"},
	}

	for _, tc := range tests {
		ui := NewLipglossUI()
		got := ui.RenderHeader(tc.title, tc.ip, tc.deviceReady)
		if !strings.Contains(got, tc.wantContain) {
			t.Errorf("RenderHeader(%q, %q, %v) = %q, want to contain %q", tc.title, tc.ip, tc.deviceReady, got, tc.wantContain)
		}
	}
}

func TestLipglossUI_RenderMenu(t *testing.T) {
	choices := []string{"one", "two", "three"}
	cursor := 1
	ui := NewLipglossUI()
	got := ui.RenderMenu(choices, cursor)

	if !strings.Contains(got, "▶ two") {
		t.Errorf("RenderMenu() = %q, want to contain %q", got, "▶ two")
	}
}

func TestLipglossUI_RenderLog(t *testing.T) {
	ui := NewLipglossUI()
	message := "This is a log message"
	got := ui.RenderLog(message)
	if !strings.Contains(got, message) {
		t.Errorf("RenderLog() = %q, want to contain %q", got, message)
	}
}

func TestLipglossUI_GetSelectedStyle(t *testing.T) {
	ui := NewLipglossUI()
	style := ui.GetSelectedStyle()
	if style.GetBold() != true {
		t.Error("Expected selected style to be bold")
	}
}

func TestLipglossUI_GetNormalStyle(t *testing.T) {
	ui := NewLipglossUI()
	style := ui.GetNormalStyle()
	if style.GetForeground() != lipgloss.Color("252") {
		t.Error("Expected normal style to have a specific foreground color")
	}
}

func TestLipglossUI_GetSuccessStyle(t *testing.T) {
	ui := NewLipglossUI()
	style := ui.GetSuccessStyle()
	if style.GetForeground() != lipgloss.Color("2") {
		t.Error("Expected success style to have a specific foreground color")
	}
}

func TestLipglossUI_GetErrorStyle(t *testing.T) {
	ui := NewLipglossUI()
	style := ui.GetErrorStyle()
	if style.GetForeground() != lipgloss.Color("1") {
		t.Error("Expected error style to have a specific foreground color")
	}
}
