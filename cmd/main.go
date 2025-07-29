package main

import (
	"context"
	"fmt"
	"nanoleaf-go/internal"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuItem int

const (
	scan menuItem = iota
	pair
	turnOn
	turnOff
	brightness
	quit
)

var (
	fullChoices = []string{
		"[s] Scan Devices",
		"[p] Pair Device",
		"[o] Turn On",
		"[x] Turn Off",
		"[b] Brightness",
		"[q] Quit",
	}
	readyChoices = []string{
		"[o] Turn On",
		"[x] Turn Off",
		"[b] Brightness",
		"[q] Quit",
	}
)

type UI interface {
	RenderHeader(title, ip string, deviceReady bool) string
	RenderMenu(choices []string, cursor int) string
	RenderLog(message string) string
	GetSelectedStyle() lipgloss.Style
	GetNormalStyle() lipgloss.Style
	GetSuccessStyle() lipgloss.Style
	GetErrorStyle() lipgloss.Style
}

type (
	StatusMsg struct{}
	model     struct {
		cursor        int
		choices       []string
		message       string
		error         bool
		ip            string // first discovered device IP
		token         string // token from pairing
		app           *internal.NanoleafService
		ctx           context.Context
		cancel        context.CancelFunc
		deviceReady   bool
		ui            UI // UI interface
		inputMode     bool
		textInput     textinput.Model
		inputPrompt   string
		onInputAction func(string) tea.Cmd
	}
)

func initModel() model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 3
	ti.Width = 20

	httpClient := internal.NewDefaultHTTPClient()
	scanner := internal.NewNetworkScanner()
	configManager := internal.NewFileConfigManager()
	client := internal.NewAPIClient(httpClient)
	app := internal.NewNanoleafService(client, scanner, configManager)
	ui := internal.NewLipglossUI()
	ctx, cancel := context.WithCancel(context.Background())

	message := ""
	var ip, token string

	// Try to load existing configuration
	result := app.LoadConfiguration()
	var dInfo internal.ServiceResult
	if result.Success {
		if config, ok := result.Data.(internal.Config); ok {
			ip = config.IP
			token = config.Token

			dInfo = app.GetDeviceInfo(ctx, ip, token)
			if dInfo.Success {
				message = ui.GetNormalStyle().Render("Device paired (config loaded)")
			}
		}
	}

	var currentChoices []string
	if dInfo.Success {
		currentChoices = readyChoices
	} else {
		currentChoices = fullChoices
	}

	return model{
		choices:     currentChoices,
		ip:          ip,
		token:       token,
		message:     message,
		app:         app,
		deviceReady: dInfo.Success,
		ctx:         ctx,
		cancel:      cancel,
		ui:          ui,
		textInput:   ti,
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
			return StatusMsg{}
		}),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.error = false

	if m.inputMode {
		return m.updateInput(msg)
	}

	return m.updateMenu(msg)
}

func (m *model) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			inputValue := m.textInput.Value()
			m.onInputAction(inputValue)
			m.textInput.Reset()
			m.inputMode = false
			return m, nil
		case "esc":
			m.inputMode = false
			return m, nil
		case "ctrl+c":
			m.cancel()
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusMsg:
		prevDeviceReady := m.deviceReady
		if m.ip != "" && m.token != "" {
			result := m.app.GetDeviceInfo(m.ctx, m.ip, m.token)
			m.deviceReady = result.Success
		}

		if m.deviceReady != prevDeviceReady {
			if m.deviceReady {
				m.choices = readyChoices
			} else {
				m.choices = fullChoices
			}
			m.cursor = 0 // Reset cursor when choices change
		}
		return m, tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
			return StatusMsg{}
		})
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancel() // Cancel any ongoing operations
			return m, tea.Quit
		case "s":
			return m, m.handleAction(scan)
		case "p":
			return m, m.handleAction(pair)
		case "o":
			return m, m.handleAction(turnOn)
		case "x":
			return m, m.handleAction(turnOff)
		case "b":
			return m, m.handleAction(brightness)
		case "q":
			return m, m.handleAction(quit)
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			// Determine the menuItem based on the selected choice string
			var selectedMenuItem menuItem
			switch m.choices[m.cursor] {
			case fullChoices[scan]:
				selectedMenuItem = scan
			case fullChoices[pair]:
				selectedMenuItem = pair
			case fullChoices[turnOn]:
				selectedMenuItem = turnOn
			case fullChoices[turnOff]:
				selectedMenuItem = turnOff
			case fullChoices[brightness]:
				selectedMenuItem = brightness
			case fullChoices[quit]:
				selectedMenuItem = quit
			}
			return m, m.handleAction(selectedMenuItem)
		}
	}

	return m, nil
}

func (m *model) handleAction(item menuItem) tea.Cmd {
	switch item {
	case scan:
		if !m.deviceReady {
			scanCtx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
			defer cancel()

			result := m.app.ScanForDevices(scanCtx)
			if result.Success {
				if devices, ok := result.Data.([]string); ok && len(devices) > 0 {
					m.ip = devices[0]
					m.message = m.ui.GetSuccessStyle().Render(result.Message)
				} else {
					m.message = m.ui.GetErrorStyle().Render("No devices detected")
				}
			} else {
				m.message = m.ui.GetErrorStyle().Render(result.Message)
			}
		}
	case pair:
		if !m.deviceReady {
			if m.ip == "" {
				m.message = m.ui.GetNormalStyle().Render("No device found. Please scan first.")
			} else {
				pairCtx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
				defer cancel()

				result := m.app.PairWithDevice(pairCtx, m.ip)
				if result.Success {
					if token, ok := result.Data.(string); ok {
						m.token = token
					}
					m.message = m.ui.GetSuccessStyle().Render(result.Message)
				} else {
					m.message = m.ui.GetErrorStyle().Render(result.Message)
				}
			}
		}
	case turnOn:
		powerCtx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		result := m.app.SetDevicePower(powerCtx, m.ip, m.token, true)
		if result.Success {
			m.message = m.ui.GetSuccessStyle().Render(result.Message)
		} else {
			m.message = m.ui.GetErrorStyle().Render(result.Message)
		}
	case turnOff:
		powerCtx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		result := m.app.SetDevicePower(powerCtx, m.ip, m.token, false)
		if result.Success {
			m.message = m.ui.GetSuccessStyle().Render(result.Message)
		} else {
			m.message = m.ui.GetErrorStyle().Render(result.Message)
		}
	case brightness:
		m.inputMode = true
		m.inputPrompt = "Enter brightness (0-100)"
		m.textInput.Placeholder = "0-100"
		m.onInputAction = func(value string) tea.Cmd {
			brightnessCtx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
			defer cancel()

			b, err := strconv.Atoi(value)
			if err != nil {
				m.message = m.ui.GetErrorStyle().Render("brightness has to be numeric (0-100)")
				return nil
			}
			result := m.app.SetBrightness(brightnessCtx, m.ip, m.token, b)
			if result.Success {
				m.message = m.ui.GetSuccessStyle().Render(result.Message)
			} else {
				m.message = m.ui.GetErrorStyle().Render(result.Message)
			}
			return nil
		}
		return textinput.Blink
	case quit:
		m.cancel() // Cancel any ongoing operations
		return tea.Quit
	}
	return nil
}

func (m *model) View() string {
	header := m.ui.RenderHeader("Nanoleaf Controller", m.ip, m.deviceReady)
	menu := m.ui.RenderMenu(m.choices, m.cursor)
	var logBox string
	if m.inputMode {
		msg := fmt.Sprintf("%s\n%s\n(esc to back)", m.inputPrompt, m.textInput.View())
		logBox = m.ui.RenderLog(msg)
	} else {
		logBox = m.ui.RenderLog(m.message)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, menu, logBox)
}

func main() {
	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	model := initModel()
	p := tea.NewProgram(&model)

	// Handle graceful shutdown
	go func() {
		<-sigChan
		model.cancel()
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Cleanup
	model.cancel()
}
