package table

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	emptyMsg   = "No logs found to display"
	loadingMsg = "Loading apex logs..."
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
	cols    []table.Column
	spinner spinner.Model
	Table
	height      int
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
	if a.showSpinner {
		s := fmt.Sprintf("%s %s", a.spinner.View(), loadingMsg)
		s = lipgloss.NewStyle().Height(a.height - 5).Render(s)
		return lipgloss.JoinVertical(lipgloss.Center, a.Table.View(), s)
	}

	if len(a.Rows()) == 0 {
		msg := lipgloss.NewStyle().Height(a.height - 5).Render(emptyMsg)
		return lipgloss.JoinVertical(lipgloss.Center, a.Table.View(), msg)
	}

	return a.Table.View()
}

func (a *Model) StartSpinner() tea.Cmd {
	a.Table.SetHeight(5)
	a.showSpinner = true
	a.spinner = spinner.New()
	return a.spinner.Tick
}

func (a *Model) StopSpinner() {
	a.Table.SetHeight(a.height)
	a.showSpinner = false
	a.spinner = spinner.New()
}

func (a *Model) SetRows(rows []table.Row) {
	if len(rows) == 0 {
		a.Table.SetHeight(5)
	} else {
		a.Table.SetHeight(a.height)
	}

	a.Table.SetRows(rows)
}

func (a *Model) SetHeight(h int) {
	// Substract the header row and underline
	a.height = h - 2
}
