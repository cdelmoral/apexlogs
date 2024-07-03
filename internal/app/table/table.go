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
		return lipgloss.JoinVertical(lipgloss.Center, a.Table.View(), s)
	}

	if len(a.Rows()) == 0 {
		return lipgloss.JoinVertical(lipgloss.Center, a.Table.View(), emptyMsg)
	}

	return a.Table.View()
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
