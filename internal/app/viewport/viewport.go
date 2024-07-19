package viewport

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
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
	viewportStyle  lipgloss.Style
	textInputStyle lipgloss.Style
	content        string
	Viewport
	textInput       textinput.Model
	spinner         spinner.Model
	isFocused       bool
	isEmpty         bool
	showSpinner     bool
	showFilter      bool
	containerHeight int
}

func New(width, height int) Model {
	m := Model{
		Viewport: viewport.New(width, height),
		viewportStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(baseColor),
		showFilter: false,
	}
	m.textInput = textinput.New()
	m.textInput.Placeholder = "Search apex log..."
	m.textInputStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(baseColor).
		Height(1).
		MaxHeight(4)
	m.SetWidth(width)
	m.SetHeight(height)
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Slash):
			m.showFilter = true
			m.textInput.Focus()
			m.SetHeight(m.containerHeight)
			return m, textinput.Blink

		case key.Matches(msg, keys.Enter):
			if m.showFilter {
				m.Viewport.SetContent(m.filterContent(m.textInput.Value()))
				m.textInput.Blur()
				return m, nil
			}
		case key.Matches(msg, keys.Esc):
			m.showFilter = false
			m.textInput.Blur()
			m.SetHeight(m.containerHeight)
			return m, nil
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	if m.isFocused {
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.isEmpty && !m.showSpinner {
		v := m.viewportStyle.Render(m.Viewport.View())
		ti := m.textInputStyle.Render(m.textInput.View())
		if m.showFilter {
			return lipgloss.JoinVertical(lipgloss.Left, v, ti)
		} else {
			return v
		}
	}

	style := m.centerMsgStyle()

	if m.showSpinner {
		s := fmt.Sprintf("%s %s", m.spinner.View(), loadingMsg)
		return m.viewportStyle.Render(style.Render(s))
	}

	return m.viewportStyle.Render(style.Render(emptyMsg))
}

func (m *Model) SetContent(s string) {
	m.content = s
	m.isEmpty = s == ""
	m.Viewport.SetContent(s)
}

func (m *Model) StartSpinner() tea.Cmd {
	m.showSpinner = true
	m.spinner = spinner.New()
	return m.spinner.Tick
}

func (m *Model) StopSpinner() {
	m.showSpinner = false
	m.spinner = spinner.New()
}

func (m *Model) Focus() {
	m.isFocused = true
	m.viewportStyle = m.viewportStyle.BorderForeground(focusedColor)
	m.textInputStyle = m.textInputStyle.BorderForeground(focusedColor)
	if m.showFilter {
		m.textInput.Focus()
	}
}

func (m *Model) Blur() {
	m.isFocused = false
	m.viewportStyle = m.viewportStyle.BorderForeground(baseColor)
	m.textInputStyle = m.textInputStyle.BorderForeground(baseColor)
	m.textInput.Blur()
}

func (m *Model) SetWidth(w int) {
	m.Width = w - 2
	m.viewportStyle = m.viewportStyle.Width(w - 2).MaxWidth(w)
	m.textInput.Width = w - 5
	m.textInputStyle = m.textInputStyle.Width(w - 2).MaxWidth(w)
}

func (m *Model) SetHeight(h int) {
	m.containerHeight = h
	hm := h - 3
	hti := m.getTextInputWidth()
	m.Height = hm - hti
	m.viewportStyle = m.viewportStyle.Height(hm - hti).MaxHeight(h - hti)
}

func (m Model) getTextInputWidth() int {
	if m.showFilter {
		return 3
	}
	return 0
}

func (m Model) centerMsgStyle() lipgloss.Style {
	w, h := m.Width, m.Height
	if sw := m.Style.GetWidth(); sw != 0 {
		w = min(w, sw)
	}
	if sh := m.Style.GetHeight(); sh != 0 {
		h = min(h, sh)
	}

	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Align(lipgloss.Center, lipgloss.Center)
}

func (m Model) filterContent(s string) string {
	var b strings.Builder
	lines := strings.Split(m.content, "\n")
	for _, line := range lines {
		if strings.Contains(line, s) {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return b.String()
}
