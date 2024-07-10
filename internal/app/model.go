package app

import (
	"fmt"
	"log"
	"time"

	apptable "github.com/cdelmoral/apexlogs/internal/app/table"
	"github.com/cdelmoral/apexlogs/internal/app/viewport"
	sf "github.com/cdelmoral/apexlogs/internal/salesforce"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultDebugLevelName = "SFDC_DevConsole"
	datetimeLayout        = "02 Jan 15:04"
)

type TraceFlagNotFoundError struct {
	s string
}

func (t *TraceFlagNotFoundError) Error() string {
	return t.s
}

type startFetchingLogsMsg struct{}

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

type model struct {
	help             help.Model
	salesforceClient *sf.Client
	logBody          string
	selectedLogId    string
	keys             keyMap
	viewport         viewport.Model
	table            apptable.Model
	terminalHeight   int
	terminalWidth    int
	viewportReady    bool
	quitting         bool
}

func newModel() model {
	columns := []table.Column{
		{Title: "Start time", Width: 12},
		{Title: "Operation", Width: 10},
		{Title: "Status", Width: 8},
		{Title: "Id", Width: 5},
	}

	t := apptable.New(table.WithColumns(columns), table.WithFocused(true), table.WithHeight(10))
	t.Focus()

	keys.showTable = true
	keys.showViewport = false

	return model{
		table: t,
		keys:  keys,
		help:  help.New(),
	}
}

func (m model) Init() tea.Cmd {
	startSpinners := func() tea.Msg {
		return startFetchingLogsMsg{}
	}
	return tea.Sequence(startSpinners, initApexLogs)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.tab):
			m.switchFocus()
			m.resize()
			return m, nil
		case key.Matches(msg, m.keys.enter):
			if m.table.Focused() {
				return m, m.selectApexLog
			}
		case key.Matches(msg, m.keys.help):
			m.help.ShowAll = !m.help.ShowAll
			m.resize()
			return m, nil
		case key.Matches(msg, m.keys.refresh):
			if m.table.Focused() {
				m.table.SetRows([]table.Row{})
				cmds = append(cmds, m.table.StartSpinner())
				cmds = append(cmds, refreshApexLogsCmd(m.salesforceClient))
				return m, tea.Sequence(cmds...)
			}
		}
	case startFetchingLogsMsg:
		cmd = m.table.StartSpinner()
		m.viewport.SetContent("")
		return m, cmd
	case apexLogsMsg:
		m.table.StopSpinner()
		m.table.SetRows(marshalLogs(msg.logs))
		m.salesforceClient = msg.salesforceClient
		return m, nil
	case selectApexLogMsg:
		cmd = m.viewport.StartSpinner()
		cmds = append(cmds, cmd)
		m.selectedLogId = msg.id
		cmds = append(cmds, m.fetchApexLog)
		return m, tea.Sequence(cmds...)
	case apexLogBodyMsg:
		m.switchFocus()
		m.viewport.StopSpinner()
		m.viewport.SetContent(msg.body)
		return m, nil
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		m.resize()
		return m, nil
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	v := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.table.View(),
		m.viewport.View(),
	)
	helpView := lipgloss.NewStyle().MarginTop(0).Render(m.help.View(m.keys))

	return lipgloss.JoinVertical(lipgloss.Left, v, helpView)
}

func (m *model) switchFocus() {
	if m.table.Focused() {
		m.table.Blur()
		m.keys.showTable = false
		m.keys.showViewport = true
		m.viewport.Focus()
	} else {
		m.table.Focus()
		m.keys.showTable = true
		m.keys.showViewport = false
		m.viewport.Blur()
	}
}

func (m *model) resize() {
	helpView := m.help.View(m.keys)
	helpViewHeight := lipgloss.Height(helpView)

	m.help.Width = m.terminalWidth

	ht := m.terminalHeight - helpViewHeight
	wl := percentInt(m.terminalWidth, 40)
	wr := percentInt(m.terminalWidth, 60)

	m.table.SetWidth(wl)
	m.table.SetHeight(ht)

	if !m.viewportReady {
		m.viewport = viewport.New(wr, ht)
		m.viewport.HighPerformanceRendering = false
		m.viewport.SetContent(m.logBody)
	}
	m.viewportReady = true
	m.viewport.SetWidth(wr)
	m.viewport.SetHeight(ht)
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

func refreshApexLogs(client *sf.Client) tea.Msg {
	apexLogsQuery := sf.SelectApexLogs()
	apexLogs, err := sf.DoQuery[sf.ApexLog](client, apexLogsQuery)
	if err != nil {
		log.Fatalf("error getting apex logs: %s", err)
	}

	return apexLogsMsg{logs: apexLogs.Records, salesforceClient: client}
}

func refreshApexLogsCmd(client *sf.Client) tea.Cmd {
	return func() tea.Msg {
		return refreshApexLogs(client)
	}
}

func initApexLogs() tea.Msg {
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

	return refreshApexLogs(client)
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
		patchPayload := map[string]string{
			"ExpirationDate": time.Now().Add(time.Minute * 30).UTC().Format(sf.DateTimeLayout),
			"StartDate":      time.Now().UTC().Format(sf.DateTimeLayout),
		}
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
			"StartDate":      time.Now().UTC().Format(sf.DateTimeLayout),
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
				log.ID,
			},
		)
	}
	return rows
}

func percentInt(a, b int) int {
	return int(float64(a) * (float64(b) / 100))
}
