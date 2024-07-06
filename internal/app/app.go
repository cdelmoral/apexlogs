package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func Start() {
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program: ", err)
		os.Exit(1)
	}
}
