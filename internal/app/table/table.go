package table

import (
	"fmt"
	"time"

	sf "github.com/cdelmoral/apexlogs/internal/salesforce"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	emptyMsg       = "No logs found to display"
	loadingMsg     = "Loading apex logs..."
	focusedColor   = lipgloss.Color("12")
	baseColor      = lipgloss.Color("7")
	datetimeLayout = "02 Jan 15:04"
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

// Model wraps the [table.Model] type.
//
// It adds the following functionality:
//   - Loading spinner
//   - Empty state message
type Model struct {
	style   lipgloss.Style
	cols    []table.Column
	ids     []string
	spinner spinner.Model
	Table
	height      int
	width       int
	showSpinner bool
}

// New creates a new [Model].
// It receives a list of [table.Option].
func New(opts ...table.Option) Model {
	columns := []table.Column{
		{Title: "Start time", Width: 12},
		{Title: "Operation", Width: 10},
		{Title: "Status", Width: 10},
		{Title: "Log Size", Width: 8},
	}

	opts = append(opts, table.WithColumns(columns))

	t := table.New(opts...)
	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderBottom(true)
	s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))
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

func (a *Model) SetLogs(logs []sf.ApexLog) {
	ids, rows := marshalLogs(logs)
	a.ids = ids
	if len(rows) == 0 {
		a.Table.SetHeight(5)
	} else {
		a.SetHeight(a.height)
	}

	a.SetRows(rows)
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

func (m Model) SelectedLogId() string {
	i := m.Cursor()
	if len(m.ids) > i {
		return m.ids[i]
	}
	return ""
}

func marshalLogs(logs []sf.ApexLog) ([]string, []table.Row) {
	rows := make([]table.Row, 0, len(logs))
	ids := make([]string, 0, len(logs))
	for _, log := range logs {
		st, err := time.Parse(sf.DateTimeLayout, log.StartTime)
		if err != nil {
			st = time.Now()
		}

		rows = append(
			rows,
			table.Row{
				st.Format(datetimeLayout),
				log.Operation,
				log.Status,
				printSize(log.LogLength),
			},
		)
		ids = append(ids, log.ID)
	}
	return ids, rows
}

func printSize(size int) string {
	kb, mb := size/1024, float32(size)/(1024*1024)
	s := fmt.Sprintf("%d KB", kb)
	if kb > 999 {
		s = fmt.Sprintf("%.1f MB", mb)
	}
	return s
}
