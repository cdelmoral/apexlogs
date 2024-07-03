package viewport

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

const emptyMsg = "Select an apex log to see the content"

type Viewport = viewport.Model

type Model struct {
	Viewport
	isEmpty bool
}

func New(width, height int) Model {
	return Model{
		Viewport: viewport.New(width, height),
	}
}

func (a Model) View() string {
	if a.isEmpty {
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

		return style.Render(emptyMsg)
	}

	return a.Viewport.View()
}

func (a *Model) SetContent(s string) {
	a.isEmpty = s == ""
	a.Viewport.SetContent(s)
}
