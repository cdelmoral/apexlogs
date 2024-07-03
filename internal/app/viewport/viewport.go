package viewport

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	emptyMsg   = "Select an apex log to see the content"
	loadingMsg = "Loading selected apex log..."
)

type Viewport = viewport.Model

type Model struct {
	Viewport
	spinner     spinner.Model
	isFocused   bool
	isEmpty     bool
	showSpinner bool
}

func New(width, height int) Model {
	return Model{
		Viewport: viewport.New(width, height),
	}
}

func (a Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd
	default:
		if a.isFocused {
			var cmd tea.Cmd
			a.Viewport, cmd = a.Viewport.Update(msg)
			return a, cmd

		}
	}

	return a, nil
}

func (a Model) View() string {
	if !a.isEmpty && !a.showSpinner {
		return a.Viewport.View()
	}

	w, h := a.Width, a.Height
	if sw := a.Style.GetWidth(); sw != 0 {
		w = min(w, sw)
	}
	if sh := a.Style.GetHeight(); sh != 0 {
		h = min(h, sh)
	}

	style := lipgloss.NewStyle().
		Width(w).
		Height(h).
		Align(lipgloss.Center, lipgloss.Center)

	if a.showSpinner {
		s := fmt.Sprintf("%s %s", a.spinner.View(), loadingMsg)
		return style.Render(s)
	}

	return style.Render(emptyMsg)
}

func (a *Model) SetContent(s string) {
	a.isEmpty = s == ""
	a.Viewport.SetContent(s)
}

func (a *Model) StartSpinner() tea.Cmd {
	a.showSpinner = true
	a.spinner = spinner.New()
	return a.spinner.Tick
}

func (a *Model) StopSpinner() {
	a.showSpinner = false
	a.spinner = spinner.New()
}

func (a *Model) Focus() {
	a.isFocused = true
}

func (a *Model) Blur() {
	a.isFocused = false
}
