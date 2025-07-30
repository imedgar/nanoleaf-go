package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"nanoleaf-go/internal"
)

func main() {
	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create device and UI
	device := internal.NewDevice()
	ui := internal.NewUI(device)

	program := tea.NewProgram(ui)

	// Handle graceful shutdown
	go func() {
		<-sigChan
		program.Quit()
	}()

	if _, err := program.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}