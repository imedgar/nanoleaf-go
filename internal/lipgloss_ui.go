package internal

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type LipglossUI struct{}

// NewLipglossUI creates a new LipglossUI instance.
func NewLipglossUI() *LipglossUI {
	return &LipglossUI{}
}

var (
	// Header styling
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			MarginBottom(0)

	// List styling
	listStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	// Selected item styling
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("57")).
			Bold(true)

	// Normal item styling
	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// List styling
	messageStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	// Success styling
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))
		// Error styling
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))
)

// RenderHeader renders the application header with a title and device readiness status.
func (ui *LipglossUI) RenderHeader(title, ip string, deviceReady bool) string {
	var status string
	if deviceReady {
		status = successStyle.Render("[O]", ip)
	} else {
		status = errorStyle.Render("[X]", ip)
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, headerStyle.Render(title), " ", status)
}

// RenderMenu renders the menu with choices and highlights the selected item.
func (ui *LipglossUI) RenderMenu(choices []string, cursor int) string {
	var menuItems []string
	for i, choice := range choices {
		prefix := "  "
		if cursor == i {
			prefix = "▶ "
			menuItems = append(menuItems, selectedStyle.Render(prefix+choice))
		} else {
			menuItems = append(menuItems, normalStyle.Render(prefix+choice))
		}
	}
	return listStyle.Render(strings.Join(menuItems, "\n"))
}

// RenderLog renders the log message.
func (ui *LipglossUI) RenderLog(message string) string {
	return messageStyle.Render(message)
}

// GetSelectedStyle returns the lipgloss style for selected items.
func (ui *LipglossUI) GetSelectedStyle() lipgloss.Style {
	return selectedStyle
}

// GetNormalStyle returns the lipgloss style for normal items.
func (ui *LipglossUI) GetNormalStyle() lipgloss.Style {
	return normalStyle
}

// GetSuccessStyle returns the lipgloss style for success messages.
func (ui *LipglossUI) GetSuccessStyle() lipgloss.Style {
	return successStyle
}

// GetErrorStyle returns the lipgloss style for error messages.
func (ui *LipglossUI) GetErrorStyle() lipgloss.Style {
	return errorStyle
}
