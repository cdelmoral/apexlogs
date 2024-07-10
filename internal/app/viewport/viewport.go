package viewport

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	emptyMsg     = "Select an apex log to see the content"
	loadingMsg   = "Loading selected apex log..."
	focusedColor = lipgloss.Color("12")
	baseColor    = lipgloss.Color("7")
)

type Viewport = viewport.Model

type Model struct {
	viewportStyle lipgloss.Style
	Viewport
	spinner     spinner.Model
	isFocused   bool
	isEmpty     bool
	showSpinner bool
}

func New(width, height int) Model {
	m := Model{
		Viewport: viewport.New(width, height),
		viewportStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(baseColor),
	}
	m.SetWidth(width)
	m.SetHeight(height)
	return m
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
		return a.viewportStyle.Render(a.Viewport.View())
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
		return a.viewportStyle.Render(style.Render(s))
	}

	return a.viewportStyle.Render(style.Render(emptyMsg))
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
	a.viewportStyle = a.viewportStyle.BorderForeground(focusedColor)
}

func (a *Model) Blur() {
	a.isFocused = false
	a.viewportStyle = a.viewportStyle.BorderForeground(baseColor)
}

func (m *Model) SetWidth(w int) {
	m.Width = w - 2
	m.viewportStyle = m.viewportStyle.Width(w - 2).MaxWidth(w)
}

func (m *Model) SetHeight(h int) {
	m.Height = h - 3
	m.viewportStyle = m.viewportStyle.Height(h - 3).MaxHeight(h)
}
