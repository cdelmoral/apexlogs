package app

import (
	"fmt"
	"log"
	"os"
	"time"

	sf "github.com/cdelmoral/apexlogs/internal/salesforce"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const defaultDebugLevelName = "SFDC_DevConsole"

var columns = []table.Column{
	{Title: "Operation", Width: 10},
	{Title: "Status", Width: 10},
	{Title: "Start time", Width: 20},
	{Title: "Id", Width: 18},
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	MarginRight(2)

var focusedStyle = baseStyle.BorderForeground(lipgloss.Color("255"))

var headerStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	BorderBottom(true).
	Bold(false)

var selectedStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("229")).
	Background(lipgloss.Color("57")).
	Bold(false)

type TraceFlagNotFoundError struct {
	s string
}

func (t *TraceFlagNotFoundError) Error() string {
	return t.s
}

type selectApexLogMsg struct {
	id string
}

// TODO: Create init message struct?
type apexLogsMsg struct {
	salesforceClient *sf.Client
	logs             []sf.ApexLog
}

type apexLogBodyMsg struct {
	body string
}

type keyMap struct {
	quit         key.Binding
	enter        key.Binding
	tab          key.Binding
	help         key.Binding
	showTable    bool
	showViewport bool
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.tab, k.help, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	var ks [][]key.Binding
	tk := table.DefaultKeyMap()
	if k.showTable {
		ks = append(ks, []key.Binding{k.enter, tk.LineUp, tk.LineDown, tk.PageUp, tk.PageDown, tk.HalfPageUp, tk.HalfPageDown, tk.GotoTop, tk.GotoBottom})
	}
	if k.showViewport {
		vk := viewport.DefaultKeyMap()
		ks = append(ks, []key.Binding{vk.PageDown, vk.PageUp, vk.HalfPageUp, vk.HalfPageDown, vk.Down, vk.Up})
	}
	ks = append(ks, []key.Binding{k.tab, k.help, k.quit})
	return ks
}

var keys = keyMap{
	quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open selected log"),
	),
	tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch focus"),
	),
	help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}

type model struct {
	keys             keyMap
	help             help.Model
	selectedLogId    string
	logBody          string
	salesforceClient *sf.Client
	viewport         viewport.Model
	table            table.Model
	ready            bool
	quitting         bool
	height           int
}

func newModel() model {
	t := table.New(table.WithColumns(columns), table.WithFocused(true), table.WithHeight(10))
	s := table.DefaultStyles()
	s.Header = headerStyle.Inherit(s.Header)
	s.Selected = selectedStyle.Inherit(s.Selected)
	t.SetStyles(s)
	t.Focus()

	keys.showTable = true
	keys.showViewport = false

	return model{table: t, keys: keys, help: help.New()}
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		userInfo, err := sf.GetDefaultUserInfo()
		if err != nil {
			log.Fatalf("error getting default dx user: %s", err)
		}
		orgInfo := sf.ScratchOrgInfo{
			AccessToken: userInfo.AccessToken,
			InstanceUrl: userInfo.InstanceUrl,
			ApiVersion:  "61.0",
			Alias:       userInfo.Alias,
		}

		client := sf.NewClient(orgInfo)
		debugLevelId := initSalesforceDebugLog(client)
		initSalesforceTraceFlag(client, userInfo.Id, debugLevelId)

		apexLogsQuery := sf.SelectApexLogs()
		apexLogs, err := sf.DoQuery[sf.ApexLog](client, apexLogsQuery)
		if err != nil {
			log.Fatalf("error getting apex logs: %s", err)
		}

		return apexLogsMsg{logs: apexLogs.Records, salesforceClient: client}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.tab):
			m.switchFocus()
			return m, nil
		case key.Matches(msg, m.keys.enter):
			if m.table.Focused() {
				return m, m.selectApexLog
			}
		case key.Matches(msg, m.keys.help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}
	case apexLogsMsg:
		m.table.SetRows(marshalLogs(msg.logs))
		m.salesforceClient = msg.salesforceClient
		return m, nil
	case selectApexLogMsg:
		m.selectedLogId = msg.id
		return m, m.fetchApexLog
	case apexLogBodyMsg:
		m.switchFocus()
		m.viewport.SetContent(msg.body)
		return m, nil
	case tea.WindowSizeMsg:
		// TODO: Figure out how to substract border dynamically
		// I think I have to use a function called GetFrameSize
		// _, v := baseStyle.GetFrameSize()
		m.height = msg.Height - 10
		m.help.Width = msg.Width

		if !m.ready {
			m.viewport = viewport.New(msg.Width-10, msg.Height-10)
			m.viewport.HighPerformanceRendering = false
			m.viewport.SetContent(m.logBody)
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}

		m.ready = true
		return m, nil
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	if !m.table.Focused() {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var v string
	if m.table.Focused() {
		v = lipgloss.JoinHorizontal(lipgloss.Top, focusedStyle.Height(m.height-10).Render(m.table.View()), m.viewport.View())
	} else {
		v = lipgloss.JoinHorizontal(lipgloss.Top, baseStyle.Height(m.height-10).Render(m.table.View()), m.viewport.View())
	}

	helpView := m.help.View(m.keys)

	return lipgloss.JoinVertical(lipgloss.Left, v, helpView)
}

func (m *model) switchFocus() {
	if m.table.Focused() {
		m.table.Blur()
		m.keys.showTable = false
		m.keys.showViewport = true
		// m.viewport.Focus()
	} else {
		m.table.Focus()
		m.keys.showTable = true
		m.keys.showViewport = false
		// m.viewport.Blur()
	}
}

func (m model) updateChildModels(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) fetchApexLog() tea.Msg {
	body, err := sf.GetSObjectBody(m.salesforceClient, "ApexLog", m.selectedLogId)
	if err != nil {
		log.Fatalf("error getting apex log: %v", err)
	}
	return apexLogBodyMsg{body: body}
}

func (m model) selectApexLog() tea.Msg {
	return selectApexLogMsg{id: m.table.SelectedRow()[3]}
}

func Start() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Println("Error running program: ", err)
		os.Exit(1)
	}
}

func initSalesforceDebugLog(client *sf.Client) string {
	debugLevelQuery := sf.SelectDebugLogByDeveloperName(defaultDebugLevelName)
	debugLevelResponse, err := sf.DoQuery[sf.DebugLevel](client, debugLevelQuery)
	if err != nil {
		log.Fatalf("error querying debug level record: %s", err)
	}

	if debugLevelResponse.TotalSize > 0 {
		return debugLevelResponse.Records[0].Id
	}

	postDebugLevelResponse, err := sf.PostSObject(client, "DebugLevel", sf.DebugLevel{})
	if err != nil {
		log.Fatalf("error sending new debug level record request: %s", err)
	}

	if !postDebugLevelResponse.Success {
		log.Fatalf("error creating debug level record: %s", err)
	}

	return postDebugLevelResponse.Id
}

func refreshSalesforceTraceFlag(client *sf.Client, userId string) error {
	traceFlagQuery := sf.SelectDebugLogTraceFlagByTracedId(userId)
	queryResult, err := sf.DoQuery[sf.TraceFlag](client, traceFlagQuery)
	if err != nil {
		return fmt.Errorf("error querying trace flag record: %s", err)
	}

	if queryResult.TotalSize == 0 {
		return &TraceFlagNotFoundError{"trace flag of type debug log not found"}
	}

	traceFlag := queryResult.Records[0]
	expirationDate, err := time.Parse(sf.DateTimeLayout, traceFlag.ExpirationDate)
	if err != nil {
		return fmt.Errorf("unexpected format found for trace flag expiration date: %s", traceFlag.ExpirationDate)
	}

	if expirationDate.Unix() < time.Now().Add(time.Minute*10).UTC().Unix() {
		patchPayload := map[string]string{"ExpirationDate": time.Now().Add(time.Minute * 30).UTC().Format(sf.DateTimeLayout)}
		err := sf.PatchSObject(client, "TraceFlag", traceFlag.Id, patchPayload)
		if err != nil {
			return fmt.Errorf("error sending request to update trace flag with id %s: %s", traceFlag.Id, err)
		}
	}

	time.AfterFunc(time.Minute*15, func() { refreshSalesforceTraceFlag(client, userId) })

	return nil
}

func initSalesforceTraceFlag(client *sf.Client, userId, debugLevelId string) {
	err := refreshSalesforceTraceFlag(client, userId)

	if _, ok := err.(*TraceFlagNotFoundError); ok {
		traceFlag := map[string]any{
			"TracedEntityId": userId,
			"DebugLevelId":   debugLevelId,
			"LogType":        "DEVELOPER_LOG",
			"ExpirationDate": time.Now().Add(time.Minute * 30).UTC().Format(sf.DateTimeLayout),
		}
		postResult, err := sf.PostSObject(client, "TraceFlag", traceFlag)
		if err != nil {
			log.Fatalf("error sending new trace flag request: %s", err)
		}

		if !postResult.Success {
			log.Fatalf("error creating trace flag record: %s", err)
		}

		time.AfterFunc(time.Minute*15, func() { refreshSalesforceTraceFlag(client, userId) })
	} else if err != nil {
		log.Fatalf("error refreshing trace flag: %s", err)
	}
}

func marshalLogs(logs []sf.ApexLog) []table.Row {
	rows := []table.Row{}
	for _, log := range logs {
		rows = append(rows, table.Row{log.Operation, log.Status, log.StartTime, log.ID})
	}
	return rows
}
