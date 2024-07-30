package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Start creates a new tea program and runs it.
func Start() {
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program: ", err)
		os.Exit(1)
	}
}
