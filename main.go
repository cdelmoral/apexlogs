package main

import (
	"fmt"
	"os"

	"github.com/cdelmoral/apexlogs/internal/app"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// TODO: Temporary log configuration
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	app.Start()
}
