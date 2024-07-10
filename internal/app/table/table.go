package table

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	emptyMsg     = "No logs found to display"
	loadingMsg   = "Loading apex logs..."
	focusedColor = lipgloss.Color("12")
	baseColor    = lipgloss.Color("7")
)

var headerStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	BorderBottom(true).
	Bold(false)

var selectedStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("229")).
	Background(lipgloss.Color("57")).
	Bold(false)

type Table = table.Model

type Model struct {
	style   lipgloss.Style
	cols    []table.Column
	spinner spinner.Model
	Table
	height      int
	width       int
	showSpinner bool
}

func New(opts ...table.Option) Model {
	t := table.New(opts...)
	s := table.DefaultStyles()
	s.Header = headerStyle.Inherit(s.Header)
	s.Selected = selectedStyle.Inherit(s.Selected)
	t.SetStyles(s)

	return Model{
		Table: t,
		style: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(focusedColor).
			MarginRight(1),
	}
}

func (a Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd
	default:
		var cmd tea.Cmd
		a.Table, cmd = a.Table.Update(msg)
		return a, cmd
	}
}

func (a Model) View() string {
	var v string

	if a.showSpinner {
		s := fmt.Sprintf("%s %s", a.spinner.View(), loadingMsg)
		v = lipgloss.JoinVertical(lipgloss.Left, a.Table.View(), s)
	} else if len(a.Rows()) == 0 {
		v = lipgloss.JoinVertical(lipgloss.Left, a.Table.View(), emptyMsg)
	} else {
		v = a.Table.View()
	}

	return a.style.Render(v)
}

func (a *Model) StartSpinner() tea.Cmd {
	a.Table.SetHeight(5)
	a.showSpinner = true
	a.spinner = spinner.New()
	return a.spinner.Tick
}

func (a *Model) StopSpinner() {
	a.SetHeight(a.height)
	a.showSpinner = false
	a.spinner = spinner.New()
}

func (a *Model) SetRows(rows []table.Row) {
	if len(rows) == 0 {
		a.Table.SetHeight(5)
	} else {
		a.SetHeight(a.height)
	}

	a.Table.SetRows(rows)
}

func (m *Model) SetHeight(h int) {
	m.height = h
	m.Table.SetHeight(h - 5)
	m.style = m.style.Height(h - 3).MaxHeight(h)
}

func (a *Model) SetWidth(w int) {
	a.width = w
	a.Table.SetWidth(w - 3)
	a.style = a.style.Width(w - 3).MaxWidth(w)
}

func (a *Model) Blur() {
	a.style = a.style.BorderForeground(baseColor)
	a.Table.Blur()
}

func (a *Model) Focus() {
	a.style = a.style.BorderForeground(focusedColor)
	a.Table.Focus()
}
