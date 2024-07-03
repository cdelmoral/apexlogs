package table

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

const emptyMsg = "No logs found to display"

type Table = table.Model

type Model struct {
	cols []table.Column
	Table
}

func New(opts ...table.Option) Model {
	return Model{
		Table: table.New(opts...),
	}
}

func (a Model) View() string {
	if len(a.Rows()) == 0 {
		return lipgloss.JoinVertical(lipgloss.Center, a.Table.View(), emptyMsg)
	}

	return a.Table.View()
}
