package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JamesClonk/go-todotxt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/perhenrik/timesheet-txt/file"
	"github.com/perhenrik/timesheet-txt/model"
	"github.com/perhenrik/timesheet-txt/report"
	"github.com/perhenrik/timesheet-txt/util"
)

var (
	bgColor      = lipgloss.Color("#1F2933")
	panelBgColor = lipgloss.Color("#273444")
	borderColor  = lipgloss.Color("#486581")
	textColor    = lipgloss.Color("#E5E7EB")
	mutedColor   = lipgloss.Color("#9AA5B1")
	accentColor  = lipgloss.Color("#F59E0B")
	successColor = lipgloss.Color("#22C55E")
	errorColor   = lipgloss.Color("#F97316")
)

type tuiRefreshMsg struct {
	reportText    string
	runningStatus string
	projects      []string
	tasks         []string
	err           error
}

type tuiModel struct {
	dateInput    textinput.Model
	periodInput  textinput.Model
	projectInput textinput.Model
	taskInput    textinput.Model

	focusIndex int
	reportType string

	reportText  string
	statusText  string
	runningText string

	projects     []string
	tasks        []string
	projectIndex int
	taskIndex    int

	width  int
	height int
	ready  bool
}

func runTUI() {
	program := tea.NewProgram(newTUIModel(), tea.WithAltScreen())
	_, err := program.Run()
	util.Check(err)
}

func newTUIModel() tuiModel {
	dateInput := textinput.New()
	dateInput.Placeholder = "YYYY-MM-DD"
	dateInput.SetValue(time.Now().UTC().Format("2006-01-02"))
	dateInput.CharLimit = 10
	dateInput.Width = 12

	periodInput := textinput.New()
	periodInput.Placeholder = "5d"
	periodInput.SetValue("5d")
	periodInput.CharLimit = 8
	periodInput.Width = 10

	projectInput := textinput.New()
	projectInput.Placeholder = "project"
	projectInput.CharLimit = 30
	projectInput.Width = 20

	taskInput := textinput.New()
	taskInput.Placeholder = "task"
	taskInput.CharLimit = 30
	taskInput.Width = 20

	m := tuiModel{
		dateInput:    dateInput,
		periodInput:  periodInput,
		projectInput: projectInput,
		taskInput:    taskInput,
		reportType:   "simple",
		statusText:   "Report view loaded. Focus project/task and press Enter to start stopwatch.",
		runningText:  "idle",
	}

	m.setFocus(0)
	return m
}

func (m tuiModel) Init() tea.Cmd {
	return refreshTUICmd(m.dateInput.Value(), m.periodInput.Value(), m.reportType)
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.setFocus((m.focusIndex + 1) % 4)
			return m, nil
		case "shift+tab", "backtab":
			next := m.focusIndex - 1
			if next < 0 {
				next = 3
			}
			m.setFocus(next)
			return m, nil
		case "enter":
			if m.focusIndex <= 1 {
				return m, refreshTUICmd(m.dateInput.Value(), m.periodInput.Value(), m.reportType)
			}

			statusText, err := m.startStopwatchFromInputs()
			if err != nil {
				m.statusText = "Start failed: " + err.Error()
				return m, nil
			}
			m.statusText = statusText
			return m, refreshTUICmd(m.dateInput.Value(), m.periodInput.Value(), m.reportType)
		case "ctrl+r", "r":
			if key == "r" && m.focusIndex >= 2 {
				break
			}
			return m, refreshTUICmd(m.dateInput.Value(), m.periodInput.Value(), m.reportType)
		case "ctrl+t", "t":
			if key == "t" && m.focusIndex >= 2 {
				break
			}
			if m.reportType == "simple" {
				m.reportType = "summary"
			} else {
				m.reportType = "simple"
			}
			m.statusText = "Switched report type to " + m.reportType + "."
			return m, refreshTUICmd(m.dateInput.Value(), m.periodInput.Value(), m.reportType)
		case "ctrl+x", "x":
			if key == "x" && m.focusIndex >= 2 {
				break
			}
			statusText, err := stopStopwatchFromTUI()
			if err != nil {
				m.statusText = "Stop failed: " + err.Error()
				return m, nil
			}
			m.statusText = statusText
			return m, refreshTUICmd(m.dateInput.Value(), m.periodInput.Value(), m.reportType)
		case "ctrl+n", "n":
			if key == "n" && m.focusIndex < 2 {
				break
			}
			m.applyNextOption()
			return m, nil
		case "ctrl+p", "p":
			if key == "p" && m.focusIndex < 2 {
				break
			}
			m.applyPreviousOption()
			return m, nil
		}

		var cmd tea.Cmd
		switch m.focusIndex {
		case 0:
			m.dateInput, cmd = m.dateInput.Update(msg)
		case 1:
			m.periodInput, cmd = m.periodInput.Update(msg)
		case 2:
			m.projectInput, cmd = m.projectInput.Update(msg)
		case 3:
			m.taskInput, cmd = m.taskInput.Update(msg)
		}

		return m, cmd

	case tuiRefreshMsg:
		if msg.err != nil {
			m.reportText = ""
			m.runningText = "unknown"
			m.statusText = "Refresh failed: " + msg.err.Error()
			return m, nil
		}

		m.reportText = msg.reportText
		m.runningText = msg.runningStatus
		m.projects = msg.projects
		m.tasks = msg.tasks

		if m.projectInput.Value() == "" && len(m.projects) > 0 {
			m.projectIndex = 0
			m.projectInput.SetValue(m.projects[m.projectIndex])
		}
		if m.taskInput.Value() == "" && len(m.tasks) > 0 {
			m.taskIndex = 0
			m.taskInput.SetValue(m.tasks[m.taskIndex])
		}
		return m, nil
	}

	return m, nil
}

func (m tuiModel) View() string {
	if !m.ready {
		return "Loading TUI..."
	}

	if m.width < 80 || m.height < 18 {
		msg := "Terminal too small for TUI. Resize to at least 80x18."
		return lipgloss.NewStyle().Foreground(errorColor).Padding(1, 2).Render(msg)
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	bodyHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if bodyHeight < 8 {
		bodyHeight = 8
	}

	body := m.renderBody(bodyHeight)
	content := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)

	return lipgloss.NewStyle().Background(bgColor).Foreground(textColor).Render(content)
}

func (m tuiModel) renderHeader() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(accentColor).Render("timesheet-txt - report-first TUI")
	filters := []string{
		m.renderInput("Date", m.dateInput, m.focusIndex == 0),
		m.renderInput("Period", m.periodInput, m.focusIndex == 1),
		lipgloss.NewStyle().Foreground(mutedColor).Render("Type:") + " " + lipgloss.NewStyle().Bold(true).Foreground(accentColor).Render(strings.ToUpper(m.reportType)),
	}

	line1 := lipgloss.NewStyle().Width(m.width - 2).Render(title)
	line2 := lipgloss.NewStyle().Width(m.width - 2).Foreground(textColor).Render(strings.Join(filters, "   "))

	return lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(panelBgColor).
		Width(m.width).
		Render(lipgloss.JoinVertical(lipgloss.Left, line1, line2))
}

func (m tuiModel) renderBody(height int) string {
	leftWidth := 42
	if m.width < 120 {
		leftWidth = 36
	}
	rightWidth := m.width - leftWidth - 1
	if rightWidth < 30 {
		rightWidth = 30
		leftWidth = m.width - rightWidth - 1
	}

	left := m.renderLeftPane(leftWidth, height)
	right := m.renderRightPane(rightWidth, height)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m tuiModel) renderLeftPane(width int, height int) string {
	projectField := m.renderInput("Project", m.projectInput, m.focusIndex == 2)
	taskField := m.renderInput("Task", m.taskInput, m.focusIndex == 3)

	projects := "- (none)"
	if len(m.projects) > 0 {
		projects = "- " + strings.Join(m.projects, "\n- ")
	}
	tasks := "- (none)"
	if len(m.tasks) > 0 {
		tasks = "- " + strings.Join(m.tasks, "\n- ")
	}

	statusStyle := lipgloss.NewStyle().Foreground(successColor)
	if strings.HasPrefix(m.statusText, "Start failed:") || strings.HasPrefix(m.statusText, "Stop failed:") || strings.HasPrefix(m.statusText, "Refresh failed:") {
		statusStyle = lipgloss.NewStyle().Foreground(errorColor)
	}

	content := []string{
		lipgloss.NewStyle().Bold(true).Foreground(accentColor).Render("Stopwatch"),
		lipgloss.NewStyle().Foreground(mutedColor).Render("Press Enter on Project/Task fields to start."),
		"",
		projectField,
		taskField,
		"",
		lipgloss.NewStyle().Foreground(mutedColor).Render("Running:"),
		lipgloss.NewStyle().Foreground(textColor).Render(m.runningText),
		"",
		lipgloss.NewStyle().Foreground(mutedColor).Render("Status:"),
		statusStyle.Render(m.statusText),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(accentColor).Render("Known Projects"),
		lipgloss.NewStyle().Foreground(textColor).Render(projects),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(accentColor).Render("Known Tasks"),
		lipgloss.NewStyle().Foreground(textColor).Render(tasks),
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(panelBgColor).
		Padding(1, 1).
		Width(width).
		Height(height).
		Render(strings.Join(content, "\n"))
}

func (m tuiModel) renderRightPane(width int, height int) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(accentColor).Render("Report")
	reportBody := strings.TrimSpace(m.reportText)
	if reportBody == "" {
		reportBody = "(no report rows for selected period)"
	}

	reportHeight := height - 4
	if reportHeight < 3 {
		reportHeight = 3
	}
	reportBody = fitTextBlock(reportBody, width-4, reportHeight)

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		lipgloss.NewStyle().Foreground(textColor).Render(reportBody),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(panelBgColor).
		Padding(1, 1).
		Width(width).
		Height(height).
		Render(content)
}

func (m tuiModel) renderFooter() string {
	help := []string{
		"tab/shift+tab: focus",
		"enter: refresh or start",
		"r: refresh",
		"t: type",
		"x: stop",
		"n/p: cycle options",
		"q: quit",
	}

	return lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(panelBgColor).
		Foreground(mutedColor).
		Width(m.width).
		Render(strings.Join(help, "   "))
}

func (m tuiModel) renderInput(label string, input textinput.Model, focused bool) string {
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
	if focused {
		labelStyle = lipgloss.NewStyle().Bold(true).Foreground(accentColor)
	}
	return labelStyle.Render(label+":") + " " + input.View()
}

func fitTextBlock(text string, width int, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	lines := strings.Split(text, "\n")
	out := make([]string, 0, height)
	for _, line := range lines {
		if len(out) >= height {
			break
		}
		out = append(out, clipLine(line, width))
	}

	remaining := len(lines) - len(out)
	if remaining > 0 {
		notice := fmt.Sprintf("... (%d more lines)", remaining)
		if len(out) == height {
			out[height-1] = clipLine(notice, width)
		} else {
			out = append(out, clipLine(notice, width))
		}
	}

	return strings.Join(out, "\n")
}

func clipLine(line string, width int) string {
	r := []rune(line)
	if len(r) <= width {
		return line
	}
	if width <= 3 {
		return string(r[:width])
	}
	return string(r[:width-3]) + "..."
}

func (m *tuiModel) setFocus(index int) {
	m.focusIndex = index
	m.dateInput.Blur()
	m.periodInput.Blur()
	m.projectInput.Blur()
	m.taskInput.Blur()

	promptBlur := lipgloss.NewStyle().Foreground(mutedColor)
	promptFocus := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	textBlur := lipgloss.NewStyle().Foreground(textColor)
	textFocus := lipgloss.NewStyle().Foreground(accentColor)

	m.dateInput.PromptStyle = promptBlur
	m.periodInput.PromptStyle = promptBlur
	m.projectInput.PromptStyle = promptBlur
	m.taskInput.PromptStyle = promptBlur
	m.dateInput.TextStyle = textBlur
	m.periodInput.TextStyle = textBlur
	m.projectInput.TextStyle = textBlur
	m.taskInput.TextStyle = textBlur

	switch m.focusIndex {
	case 0:
		m.dateInput.Focus()
		m.dateInput.PromptStyle = promptFocus
		m.dateInput.TextStyle = textFocus
	case 1:
		m.periodInput.Focus()
		m.periodInput.PromptStyle = promptFocus
		m.periodInput.TextStyle = textFocus
	case 2:
		m.projectInput.Focus()
		m.projectInput.PromptStyle = promptFocus
		m.projectInput.TextStyle = textFocus
	case 3:
		m.taskInput.Focus()
		m.taskInput.PromptStyle = promptFocus
		m.taskInput.TextStyle = textFocus
	}
}

func (m *tuiModel) applyNextOption() {
	switch m.focusIndex {
	case 2:
		if len(m.projects) == 0 {
			return
		}
		m.projectIndex = (m.projectIndex + 1) % len(m.projects)
		m.projectInput.SetValue(m.projects[m.projectIndex])
	case 3:
		if len(m.tasks) == 0 {
			return
		}
		m.taskIndex = (m.taskIndex + 1) % len(m.tasks)
		m.taskInput.SetValue(m.tasks[m.taskIndex])
	}
}

func (m *tuiModel) applyPreviousOption() {
	switch m.focusIndex {
	case 2:
		if len(m.projects) == 0 {
			return
		}
		m.projectIndex--
		if m.projectIndex < 0 {
			m.projectIndex = len(m.projects) - 1
		}
		m.projectInput.SetValue(m.projects[m.projectIndex])
	case 3:
		if len(m.tasks) == 0 {
			return
		}
		m.taskIndex--
		if m.taskIndex < 0 {
			m.taskIndex = len(m.tasks) - 1
		}
		m.taskInput.SetValue(m.tasks[m.taskIndex])
	}
}

func (m tuiModel) startStopwatchFromInputs() (string, error) {
	projectName := sanitizeTagValue(m.projectInput.Value())
	if projectName == "" {
		return "", fmt.Errorf("project is required")
	}

	taskName := sanitizeTagValue(m.taskInput.Value())
	taskText := "+" + projectName
	if taskName != "" {
		taskText += " task:" + taskName
	}

	stoppedCount, startedTask, err := startStopwatchTask(taskText, time.Now().UTC())
	if err != nil {
		return "", err
	}

	if stoppedCount > 0 {
		return fmt.Sprintf("Stopped %d running stopwatch and started %s.", stoppedCount, startedTask), nil
	}

	return "Started stopwatch for " + startedTask + ".", nil
}

func stopStopwatchFromTUI() (string, error) {
	stoppedCount, totalHours, err := stopStopwatchTasks(time.Now().UTC())
	if err != nil {
		return "", err
	}

	if stoppedCount == 0 {
		return "No running stopwatch found.", nil
	}

	if stoppedCount == 1 {
		return fmt.Sprintf("Stopped stopwatch (%.2f hours).", totalHours), nil
	}

	return fmt.Sprintf("Stopped %d running stopwatches (%.2f hours total).", stoppedCount, totalHours), nil
}

func sanitizeTagValue(value string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(value)))
	if len(fields) == 0 {
		return ""
	}
	return strings.Join(fields, "-")
}

func refreshTUICmd(dateText string, periodText string, reportType string) tea.Cmd {
	return func() tea.Msg {
		tasklist := file.TimesheetFile{Name: timesheetFilename}.ReadFile()

		endTime, err := parseReportDate(dateText)
		if err != nil {
			return tuiRefreshMsg{err: err}
		}

		hours, err := parseReportPeriod(periodText)
		if err != nil {
			return tuiRefreshMsg{err: err}
		}

		reportItems := report.Create(tasklist, endTime, hours)
		reportText := report.Simple(reportItems)
		if reportType == "summary" {
			reportText = report.Summary(reportItems)
		}

		projects, tasks := projectAndTaskOptions(tasklist)
		return tuiRefreshMsg{
			reportText:    reportText,
			runningStatus: runningStatus(tasklist),
			projects:      projects,
			tasks:         tasks,
		}
	}
}

func parseReportDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Now().UTC(), nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q (expected YYYY-MM-DD)", value)
	}

	return parsed, nil
}

func parseReportPeriod(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "5d"
	}

	hours, err := model.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid period %q", value)
	}

	return hours, nil
}

func runningStatus(tasklist todotxt.TaskList) string {
	running := make([]string, 0)
	for _, task := range tasklist {
		if !isRunningTask(task) {
			continue
		}

		startedText := ""
		if startedRaw := task.AdditionalTags[stopwatchStartTagKey]; startedRaw != "" {
			startedUnix, err := strconv.ParseInt(startedRaw, 10, 64)
			if err == nil {
				startedText = time.Unix(startedUnix, 0).UTC().Format("2006-01-02 15:04")
			}
		}

		if startedText != "" {
			running = append(running, task.Task()+" (since "+startedText+" UTC)")
		} else {
			running = append(running, task.Task())
		}
	}

	if len(running) == 0 {
		return "idle"
	}

	return strings.Join(running, " | ")
}

func projectAndTaskOptions(tasklist todotxt.TaskList) ([]string, []string) {
	projectsMap := make(map[string]struct{})
	tasksMap := make(map[string]struct{})

	for _, task := range tasklist {
		if len(task.Projects) > 0 {
			projectsMap[task.Projects[0]] = struct{}{}
		}

		taskName := task.AdditionalTags["task"]
		if taskName != "" {
			tasksMap[taskName] = struct{}{}
		}
	}

	projects := make([]string, 0, len(projectsMap))
	for projectName := range projectsMap {
		projects = append(projects, projectName)
	}
	tasks := make([]string, 0, len(tasksMap))
	for taskName := range tasksMap {
		tasks = append(tasks, taskName)
	}

	sort.Strings(projects)
	sort.Strings(tasks)

	return projects, tasks
}
