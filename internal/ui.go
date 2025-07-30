package internal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UI struct {
	device      *Device
	cursor      int
	message     string
	inputMode   bool
	inputPrompt string
	textInput   textinput.Model
	deviceReady bool
}

// Messages for async operations
type (
	deviceCheckMsg struct{ ready bool }
	scanResultMsg  struct {
		devices []string
		err     error
	}
	pairResultMsg   struct{ err error }
	actionResultMsg struct {
		message string
		err     error
	}
)

func NewUI(device *Device) *UI {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 3
	ti.Width = 20

	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9933"))        // Light orange
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF99FF")) // Electric pink
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))     // Cyan cursor

	return &UI{
		device:    device,
		textInput: ti,
	}
}

func (ui UI) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, textinput.Blink)

	// Load config and check device status
	if err := ui.device.LoadConfig(); err == nil {
		cmds = append(cmds, ui.checkDeviceStatus())
	}

	return tea.Batch(cmds...)
}

func (ui UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ui.inputMode {
		return ui.updateInput(msg)
	}
	return ui.updateMenu(msg)
}

func (ui UI) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			value := ui.textInput.Value()
			ui.inputMode = false
			ui.textInput.SetValue("")
			return ui, ui.handleBrightnessInput(value)
		case "esc":
			ui.inputMode = false
			ui.textInput.SetValue("")
			return ui, nil
		case "ctrl+c":
			return ui, tea.Quit
		}
	}

	var cmd tea.Cmd
	ui.textInput, cmd = ui.textInput.Update(msg)
	return ui, cmd
}

func (ui UI) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case deviceCheckMsg:
		ui.deviceReady = msg.ready
		if msg.ready {
			ui.message = successStyle.Render("Device connected")
		}
		return ui, nil

	case scanResultMsg:
		if msg.err != nil {
			ui.message = errorStyle.Render(fmt.Sprintf("Scan failed: %v", msg.err))
		} else if len(msg.devices) > 0 {
			ui.device.SetDevice(msg.devices[0])
			ui.message = successStyle.Render(fmt.Sprintf("Found %d device(s)", len(msg.devices)))
		} else {
			ui.message = errorStyle.Render("No devices found")
		}
		return ui, nil

	case pairResultMsg:
		if msg.err != nil {
			ui.message = errorStyle.Render(fmt.Sprintf("Pairing failed: %v", msg.err))
		} else {
			ui.deviceReady = true
			ui.message = successStyle.Render("Successfully paired with device")
		}
		return ui, nil

	case actionResultMsg:
		if msg.err != nil {
			ui.message = errorStyle.Render(fmt.Sprintf("Action failed: %v", msg.err))
		} else {
			ui.message = successStyle.Render(msg.message)
		}
		return ui, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return ui, tea.Quit
		case "s":
			if !ui.deviceReady {
				return ui, ui.handleScan()
			}
		case "p":
			if !ui.deviceReady && ui.device.GetDeviceIP() != "" {
				return ui, ui.handlePair()
			}
		case "o":
			if ui.deviceReady {
				return ui, ui.handleTurnOn()
			}
		case "x":
			if ui.deviceReady {
				return ui, ui.handleTurnOff()
			}
		case "b":
			if ui.deviceReady {
				ui.inputMode = true
				ui.inputPrompt = "Enter brightness (0-100)"
				ui.textInput.Placeholder = "0-100"
				return ui, textinput.Blink
			}
		case "up", "k":
			if ui.cursor > 0 {
				ui.cursor--
			}
		case "down", "j":
			choices := ui.getMenuChoices()
			if ui.cursor < len(choices)-1 {
				ui.cursor++
			}
		case "enter":
			return ui.handleMenuSelect()
		}
	}

	return ui, nil
}

func (ui UI) getMenuChoices() []string {
	if ui.deviceReady {
		return []string{"[o] Turn On", "[x] Turn Off", "[b] Brightness", "[q] Quit"}
	}
	return []string{"[s] Scan Devices", "[p] Pair Device", "[q] Quit"}
}

func (ui UI) handleMenuSelect() (tea.Model, tea.Cmd) {
	choices := ui.getMenuChoices()
	if ui.cursor >= len(choices) {
		return ui, nil
	}

	selected := choices[ui.cursor]
	switch selected {
	case "[s] Scan Devices":
		return ui, ui.handleScan()
	case "[p] Pair Device":
		return ui, ui.handlePair()
	case "[o] Turn On":
		return ui, ui.handleTurnOn()
	case "[x] Turn Off":
		return ui, ui.handleTurnOff()
	case "[b] Brightness":
		ui.inputMode = true
		ui.inputPrompt = "Enter brightness (0-100)"
		ui.textInput.Placeholder = "0-100"
		return ui, textinput.Blink
	case "[q] Quit":
		return ui, tea.Quit
	}
	return ui, nil
}

// Action handlers
func (ui UI) checkDeviceStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := ui.device.createContext()
		defer cancel()
		ready := ui.device.IsDeviceReady(ctx)
		return deviceCheckMsg{ready: ready}
	}
}

func (ui UI) handleScan() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := ui.device.createContext()
		defer cancel()
		devices, err := ui.device.ScanForDevices(ctx)
		return scanResultMsg{devices: devices, err: err}
	}
}

func (ui UI) handlePair() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := ui.device.createContext()
		defer cancel()
		err := ui.device.PairDevice(ctx)
		return pairResultMsg{err: err}
	}
}

func (ui UI) handleTurnOn() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := ui.device.createContext()
		defer cancel()
		err := ui.device.TurnOn(ctx)
		return actionResultMsg{message: "Device turned on", err: err}
	}
}

func (ui UI) handleTurnOff() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := ui.device.createContext()
		defer cancel()
		err := ui.device.TurnOff(ctx)
		return actionResultMsg{message: "Device turned off", err: err}
	}
}

func (ui UI) handleBrightnessInput(value string) tea.Cmd {
	brightness, err := strconv.Atoi(value)
	if err != nil {
		return func() tea.Msg {
			return actionResultMsg{err: fmt.Errorf("brightness must be a number (0-100)")}
		}
	}

	return func() tea.Msg {
		ctx, cancel := ui.device.createContext()
		defer cancel()
		err := ui.device.SetBrightness(ctx, brightness)
		return actionResultMsg{message: fmt.Sprintf("Brightness set to %d", brightness), err: err}
	}
}

func (ui UI) View() string {
	// Title box
	status := "Not Connected"
	if ui.deviceReady {
		status = fmt.Sprintf("Connected to %s", ui.device.GetDeviceIP())
	}
	titleContent := fmt.Sprintf("Nanoleaf Controller / %s", status)
	titleBox := titleBoxStyle.Render(titleContent)

	// Menu
	choices := ui.getMenuChoices()
	menuItems := make([]string, len(choices))
	for i, choice := range choices {
		if i == ui.cursor {
			menuItems[i] = selectedStyle.Render(choice)
		} else {
			menuItems[i] = textStyle.Render(choice)
		}
	}

	// Separator line
	separator := separatorStyle.Render(strings.Repeat("─", 46)) // 46 chars to fit within 50 width box

	// Log/Input content
	var logContent string
	if ui.inputMode {
		prompt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9933")).Render(ui.inputPrompt)        // Light orange
		cancelText := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9933")).Render("(esc to cancel)") // Light orange
		logContent = fmt.Sprintf("%s\n%s\n%s", prompt, ui.textInput.View(), cancelText)
	} else {
		logContent = ui.message
	}

	// Separator line
	titleSeparator := separatorStyle.Render(strings.Repeat("─", 46)) // 46 chars to fit within 50 width box

	// Combine menu and log in single box
	combinedContent := lipgloss.JoinVertical(lipgloss.Left,
		titleSeparator,
		"",
		lipgloss.JoinVertical(lipgloss.Left, menuItems...),
		"",
		separator,
		"",
		logContent,
	)

	mainBox := menuStyle.Render(combinedContent)

	// Stack title box directly on main box (no spacing)
	return lipgloss.JoinVertical(lipgloss.Center, titleBox, mainBox)
}

// Styles
var (
	titleBoxStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF00FF")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF00FF")).
			BorderBottom(false).
			Padding(0, 2).
			Width(50).
			Align(lipgloss.Center)

	menuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FFFF")).
			BorderTop(false).
			Padding(1, 2).
			Width(50)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#d4d177"))

	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0080")) // Error red
	successStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF80")) // Success green
	textStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF99FF")) // Electric pink for default text
	separatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")) // Electric yellow for separators
)
