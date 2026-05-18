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
	bgColor      lipgloss.TerminalColor = lipgloss.NoColor{}
	panelBgColor lipgloss.TerminalColor = lipgloss.NoColor{}
	borderColor  lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#7A7A7A", Dark: "#5A5A5A"}
	textColor    lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#111111", Dark: "#E5E7EB"}
	mutedColor   lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#5F6368", Dark: "#9AA5B1"}
	accentColor  lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#A24B00", Dark: "#F59E0B"}
	successColor lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#0A7A24", Dark: "#22C55E"}
	errorColor   lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#B63A00", Dark: "#F97316"}
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

	manualDateInput    textinput.Model
	manualProjectInput textinput.Model
	manualTaskInput    textinput.Model
	manualHoursInput   textinput.Model

	focusIndex int
	reportType string

	reportText  string
	statusText  string
	runningText string

	projects     []string
	tasks        []string
	projectIndex int
	taskIndex    int

	manualProjectIndex int
	manualTaskIndex    int

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

	manualDateInput := textinput.New()
	manualDateInput.Placeholder = "YYYY-MM-DD"
	manualDateInput.SetValue(time.Now().UTC().Format("2006-01-02"))
	manualDateInput.CharLimit = 10
	manualDateInput.Width = 12

	manualProjectInput := textinput.New()
	manualProjectInput.Placeholder = "project"
	manualProjectInput.CharLimit = 30
	manualProjectInput.Width = 20

	manualTaskInput := textinput.New()
	manualTaskInput.Placeholder = "task"
	manualTaskInput.CharLimit = 30
	manualTaskInput.Width = 20

	manualHoursInput := textinput.New()
	manualHoursInput.Placeholder = "1.5"
	manualHoursInput.CharLimit = 8
	manualHoursInput.Width = 8

	m := tuiModel{
		dateInput:          dateInput,
		periodInput:        periodInput,
		projectInput:       projectInput,
		taskInput:          taskInput,
		manualDateInput:    manualDateInput,
		manualProjectInput: manualProjectInput,
		manualTaskInput:    manualTaskInput,
		manualHoursInput:   manualHoursInput,
		reportType:         "simple",
		statusText:         "Report view loaded. Use Enter in stopwatch or manual fields.",
		runningText:        "idle",
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
			m.setFocus((m.focusIndex + 1) % 8)
			return m, nil
		case "down":
			if m.canCycleFocusedOptions() {
				m.applyNextOption()
			} else {
				m.setFocus((m.focusIndex + 1) % 8)
			}
			return m, nil
		case "up":
			if m.canCycleFocusedOptions() {
				m.applyPreviousOption()
			} else {
				next := m.focusIndex - 1
				if next < 0 {
					next = 7
				}
				m.setFocus(next)
			}
			return m, nil
		case "shift+tab", "backtab":
			next := m.focusIndex - 1
			if next < 0 {
				next = 7
			}
			m.setFocus(next)
			return m, nil
		case "enter":
			if m.focusIndex <= 1 {
				return m, m.refreshReportCmd("Report updated.")
			}
			if m.focusIndex >= 4 {
				statusText, err := m.addManualEntryFromInputs()
				if err != nil {
					m.statusText = "Manual add failed: " + err.Error()
					return m, nil
				}
				return m, m.refreshReportCmd(statusText)
			}

			statusText, err := m.startStopwatchFromInputs()
			if err != nil {
				m.statusText = "Start failed: " + err.Error()
				return m, nil
			}
			return m, m.refreshReportCmd(statusText)
		case "ctrl+r", "r":
			if key == "r" && m.focusIndex >= 2 {
				break
			}
			return m, m.refreshReportCmd("Report updated.")
		case "ctrl+t", "t":
			if key == "t" && m.focusIndex >= 2 {
				break
			}
			if m.reportType == "simple" {
				m.reportType = "summary"
			} else {
				m.reportType = "simple"
			}
			return m, m.refreshReportCmd("Switched report type to " + m.reportType + ".")
		case "ctrl+x", "x":
			if key == "x" && m.focusIndex >= 2 {
				break
			}
			statusText, err := stopStopwatchFromTUI()
			if err != nil {
				m.statusText = "Stop failed: " + err.Error()
				return m, nil
			}
			return m, m.refreshReportCmd(statusText)
		case "ctrl+n", "n":
			if key == "n" && !m.canCycleFocusedOptions() {
				break
			}
			m.applyNextOption()
			return m, nil
		case "ctrl+p", "p":
			if key == "p" && !m.canCycleFocusedOptions() {
				break
			}
			m.applyPreviousOption()
			return m, nil
		case "m":
			if m.focusIndex < 2 {
				m.setFocus(4)
				return m, nil
			}
		}

		var cmd tea.Cmd
		switch m.focusIndex {
		case 0:
			cmd = m.updateDateInput(msg)
		case 1:
			m.periodInput, cmd = m.periodInput.Update(msg)
		case 2:
			m.projectInput, cmd = m.projectInput.Update(msg)
		case 3:
			m.taskInput, cmd = m.taskInput.Update(msg)
		case 4:
			cmd = m.updateManualDateInput(msg)
		case 5:
			m.manualProjectInput, cmd = m.manualProjectInput.Update(msg)
		case 6:
			m.manualTaskInput, cmd = m.manualTaskInput.Update(msg)
		case 7:
			m.manualHoursInput, cmd = m.manualHoursInput.Update(msg)
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
		if m.manualProjectInput.Value() == "" && len(m.projects) > 0 {
			m.manualProjectIndex = 0
			m.manualProjectInput.SetValue(m.projects[m.manualProjectIndex])
		}
		if m.manualTaskInput.Value() == "" && len(m.tasks) > 0 {
			m.manualTaskIndex = 0
			m.manualTaskInput.SetValue(m.tasks[m.manualTaskIndex])
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

	footer := m.renderFooter()

	bodyHeight := m.height - lipgloss.Height(footer)
	if bodyHeight < 8 {
		bodyHeight = 8
	}

	body := m.renderBody(bodyHeight)
	content := lipgloss.JoinVertical(lipgloss.Left, body, footer)

	return lipgloss.NewStyle().Background(bgColor).Foreground(textColor).Render(content)
}

func (m tuiModel) renderValidationHintLine() string {
	dateText, dateIsError := dateValidationHint(m.dateInput.Value())
	periodText, periodIsError := periodValidationHint(m.periodInput.Value())

	dateStyle := lipgloss.NewStyle().Foreground(mutedColor)
	periodStyle := lipgloss.NewStyle().Foreground(mutedColor)
	if dateIsError {
		dateStyle = lipgloss.NewStyle().Foreground(errorColor)
	}
	if periodIsError {
		periodStyle = lipgloss.NewStyle().Foreground(errorColor)
	}

	datePrefix := strings.Repeat(" ", len("Date: "))
	periodPrefix := strings.Repeat(" ", len("Period: "))

	left := datePrefix + dateStyle.Render(dateText)
	right := periodPrefix + periodStyle.Render(periodText)
	return left + "   " + right
}

func (m tuiModel) renderBody(height int) string {
	leftWidth := 42
	if m.width < 120 {
		leftWidth = 36
	}
	rightWidth := m.width - leftWidth
	if rightWidth < 30 {
		rightWidth = 30
		leftWidth = m.width - rightWidth
	}

	left := m.renderLeftPane(leftWidth, height)
	right := m.renderRightPane(rightWidth, height)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m tuiModel) renderLeftPane(width int, height int) string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(panelBgColor).
		Padding(1, 1)

	innerWidth := width - panelStyle.GetHorizontalBorderSize()
	innerHeight := height - panelStyle.GetVerticalBorderSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	projectField := m.renderInput("Project", m.projectInput, m.focusIndex == 2)
	taskField := m.renderInput("Task", m.taskInput, m.focusIndex == 3)
	manualDateField := m.renderInput("Date", m.manualDateInput, m.focusIndex == 4)
	manualProjectField := m.renderInput("Project", m.manualProjectInput, m.focusIndex == 5)
	manualTaskField := m.renderInput("Task", m.manualTaskInput, m.focusIndex == 6)
	manualHoursField := m.renderInput("Hours", m.manualHoursInput, m.focusIndex == 7)

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
		lipgloss.NewStyle().Foreground(mutedColor).Render("Running:"),
		lipgloss.NewStyle().Foreground(textColor).Render(m.runningText),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(accentColor).Render("Manual Entry"),
		lipgloss.NewStyle().Foreground(mutedColor).Render("Press Enter in manual fields to add hours."),
		manualDateField,
		manualProjectField,
		manualTaskField,
		manualHoursField,
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

	return panelStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(strings.Join(content, "\n"))
}

func (m tuiModel) renderRightPane(width int, height int) string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(panelBgColor).
		Padding(1, 1)

	innerWidth := width - panelStyle.GetHorizontalBorderSize()
	innerHeight := height - panelStyle.GetVerticalBorderSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(accentColor).Render("Report")
	filters := []string{
		m.renderInput("Date", m.dateInput, m.focusIndex == 0),
		m.renderInput("Period", m.periodInput, m.focusIndex == 1),
		lipgloss.NewStyle().Foreground(mutedColor).Render("Type:") + " " + lipgloss.NewStyle().Bold(true).Foreground(accentColor).Render(strings.ToUpper(m.reportType)),
	}
	filterLine := lipgloss.NewStyle().Foreground(textColor).Render(strings.Join(filters, "   "))
	hintLine := m.renderValidationHintLine()
	topSection := lipgloss.JoinVertical(lipgloss.Left, title, filterLine, hintLine, "")

	reportBody := m.reportText
	if reportBody == "" {
		reportBody = "(no report rows for selected period)"
	}

	reportHeight := innerHeight - lipgloss.Height(topSection)
	if reportHeight < 3 {
		reportHeight = 3
	}
	reportBody = fitTextBlock(reportBody, innerWidth, reportHeight)

	content := lipgloss.JoinVertical(lipgloss.Left,
		topSection,
		lipgloss.NewStyle().Foreground(textColor).Render(reportBody),
	)

	return panelStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(content)
}

func (m tuiModel) renderFooter() string {
	help := []string{
		"tab/shift+tab: focus",
		"up/down: focus or cycle",
		"enter: refresh/start/add",
		"m: jump manual entry",
		"r: refresh",
		"t: type",
		"x: stop",
		"n/p: cycle options",
		"q: quit",
	}

	footerStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(panelBgColor).
		Foreground(mutedColor)

	innerWidth := m.width - footerStyle.GetHorizontalBorderSize()
	if innerWidth < 1 {
		innerWidth = 1
	}

	return footerStyle.
		Width(innerWidth).
		Render(strings.Join(help, "   "))
}

func (m tuiModel) renderInput(label string, input textinput.Model, focused bool) string {
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
	if focused {
		labelStyle = lipgloss.NewStyle().Bold(true).Foreground(accentColor)
	}
	return labelStyle.Render(label+":") + " " + input.View()
}

func (m *tuiModel) refreshReportCmd(statusOnSuccess string) tea.Cmd {
	if err := validateDateInput(m.dateInput.Value()); err != nil {
		m.statusText = "Date must be valid (YYYY-MM-DD, also accepts YYYY-M-D)."
		return nil
	}
	if err := validatePeriodInput(m.periodInput.Value()); err != nil {
		m.statusText = "Period must be valid (examples: 5d, 8h, 30m, 2w)."
		return nil
	}
	if statusOnSuccess != "" {
		m.statusText = statusOnSuccess
	}

	return refreshTUICmd(m.dateInput.Value(), m.periodInput.Value(), m.reportType)
}

func (m *tuiModel) updateDateInput(msg tea.KeyMsg) tea.Cmd {
	digits := extractDateDigits(m.dateInput.Value())

	if msg.Type == tea.KeyRunes {
		handled := false
		for _, r := range msg.Runes {
			if r == '-' {
				handled = true
				continue
			}
			if r < '0' || r > '9' {
				continue
			}
			handled = true

			candidate := normalizeDateDigits(digits + string(r))
			if isPotentialDateDigits(candidate) {
				digits = candidate
			}
		}
		if handled {
			m.setDateFromDigits(digits)
			return nil
		}
		m.dateInput, _ = m.dateInput.Update(msg)
		m.setDateFromDigits(extractDateDigits(m.dateInput.Value()))
		return nil
	}

	switch msg.String() {
	case "backspace", "ctrl+h", "delete":
		if len(digits) > 0 {
			digits = digits[:len(digits)-1]
		}
		m.setDateFromDigits(digits)
	case "ctrl+u":
		m.setDateFromDigits("")
	default:
		var cmd tea.Cmd
		m.dateInput, cmd = m.dateInput.Update(msg)
		m.setDateFromDigits(extractDateDigits(m.dateInput.Value()))
		return cmd
	}

	return nil
}

func (m *tuiModel) updateManualDateInput(msg tea.KeyMsg) tea.Cmd {
	digits := extractDateDigits(m.manualDateInput.Value())

	if msg.Type == tea.KeyRunes {
		handled := false
		for _, r := range msg.Runes {
			if r == '-' {
				handled = true
				continue
			}
			if r < '0' || r > '9' {
				continue
			}
			handled = true

			candidate := normalizeDateDigits(digits + string(r))
			if isPotentialDateDigits(candidate) {
				digits = candidate
			}
		}
		if handled {
			m.setManualDateFromDigits(digits)
			return nil
		}
		m.manualDateInput, _ = m.manualDateInput.Update(msg)
		m.setManualDateFromDigits(extractDateDigits(m.manualDateInput.Value()))
		return nil
	}

	switch msg.String() {
	case "backspace", "ctrl+h", "delete":
		if len(digits) > 0 {
			digits = digits[:len(digits)-1]
		}
		m.setManualDateFromDigits(digits)
	case "ctrl+u":
		m.setManualDateFromDigits("")
	default:
		var cmd tea.Cmd
		m.manualDateInput, cmd = m.manualDateInput.Update(msg)
		m.setManualDateFromDigits(extractDateDigits(m.manualDateInput.Value()))
		return cmd
	}

	return nil
}

func (m *tuiModel) setDateFromDigits(digits string) {
	if len(digits) > 8 {
		digits = digits[:8]
	}

	m.dateInput.SetValue(formatDateDigits(digits))
	m.dateInput.CursorEnd()
}

func (m *tuiModel) setManualDateFromDigits(digits string) {
	if len(digits) > 8 {
		digits = digits[:8]
	}

	m.manualDateInput.SetValue(formatDateDigits(digits))
	m.manualDateInput.CursorEnd()
}

func extractDateDigits(value string) string {
	b := strings.Builder{}
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}

	digits := b.String()
	if len(digits) > 8 {
		return digits[:8]
	}

	return digits
}

func normalizeDateDigits(digits string) string {
	if len(digits) <= 4 {
		return digits
	}

	if len(digits) == 5 {
		monthTens := digits[4]
		if monthTens > '1' {
			return digits[:4] + "0" + string(monthTens)
		}
	}

	if len(digits) == 7 {
		dayTens := digits[6]
		if dayTens > '3' {
			return digits[:6] + "0" + string(dayTens)
		}
	}

	if len(digits) > 8 {
		return digits[:8]
	}

	return digits
}

func formatDateDigits(digits string) string {
	if len(digits) < 4 {
		return digits
	}
	if len(digits) == 4 {
		return digits + "-"
	}
	if len(digits) < 6 {
		return digits[:4] + "-" + digits[4:]
	}
	if len(digits) == 6 {
		return digits[:4] + "-" + digits[4:6] + "-"
	}
	return digits[:4] + "-" + digits[4:6] + "-" + digits[6:]
}

func validateDateInput(value string) error {
	_, err := parseFlexibleDate(value)
	return err
}

func validatePeriodInput(value string) error {
	_, err := parseReportPeriod(value)
	return err
}

func dateValidationHint(value string) (text string, isError bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "YYYY-MM-DD (also YYYY-M-D)", false
	}

	if validateDateInput(trimmed) == nil {
		return "looks good", false
	}

	if isPotentialDateInput(trimmed) {
		return "continue typing", false
	}

	return "invalid date", true
}

func periodValidationHint(value string) (text string, isError bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "default 5d", false
	}

	if validatePeriodInput(trimmed) == nil {
		return "looks good", false
	}

	return "use 5d, 8h, 30m, 2w", true
}

func isPotentialDateInput(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return true
	}

	for _, r := range trimmed {
		if (r < '0' || r > '9') && r != '-' {
			return false
		}
	}

	if strings.Contains(trimmed, "-") {
		parts := strings.Split(trimmed, "-")
		if len(parts) > 3 {
			return false
		}

		yearPart := parts[0]
		if len(yearPart) > 4 {
			return false
		}
		if !isAllDigits(yearPart) {
			return false
		}

		if len(parts) >= 2 {
			monthPart := parts[1]
			if len(monthPart) > 2 {
				return false
			}
			if monthPart != "" && !isAllDigits(monthPart) {
				return false
			}
			if len(monthPart) == 2 {
				month, err := strconv.Atoi(monthPart)
				if err != nil || month < 1 || month > 12 {
					return false
				}
			}
		}

		if len(parts) == 3 {
			dayPart := parts[2]
			if len(dayPart) > 2 {
				return false
			}
			if dayPart != "" && !isAllDigits(dayPart) {
				return false
			}
			if len(dayPart) == 2 {
				day, err := strconv.Atoi(dayPart)
				if err != nil || day < 1 || day > 31 {
					return false
				}
			}
		}

		if len(parts) == 3 && len(yearPart) == 4 && len(parts[1]) == 2 && len(parts[2]) == 2 {
			year, _ := strconv.Atoi(yearPart)
			month, _ := strconv.Atoi(parts[1])
			day, _ := strconv.Atoi(parts[2])
			candidate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			if candidate.Year() != year || int(candidate.Month()) != month || candidate.Day() != day {
				return false
			}
		}

		return true
	}

	digits := extractDateDigits(trimmed)
	return len(digits) <= 8 && isPotentialDateDigits(digits)
}

func isAllDigits(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isPotentialDateDigits(digits string) bool {
	if len(digits) == 0 {
		return true
	}
	if len(digits) > 8 {
		return false
	}

	for _, r := range digits {
		if r < '0' || r > '9' {
			return false
		}
	}

	if len(digits) >= 5 {
		monthTens := digits[4]
		if monthTens != '0' && monthTens != '1' {
			return false
		}
	}

	if len(digits) >= 6 {
		month, err := strconv.Atoi(digits[4:6])
		if err != nil || month < 1 || month > 12 {
			return false
		}
	}

	if len(digits) >= 7 {
		dayTens := digits[6]
		if dayTens < '0' || dayTens > '3' {
			return false
		}
	}

	if len(digits) == 8 {
		day, err := strconv.Atoi(digits[6:8])
		if err != nil || day < 1 || day > 31 {
			return false
		}
		_, err = time.Parse("20060102", digits)
		if err != nil {
			return false
		}
	}

	return true
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
	m.manualDateInput.Blur()
	m.manualProjectInput.Blur()
	m.manualTaskInput.Blur()
	m.manualHoursInput.Blur()

	promptBlur := lipgloss.NewStyle().Foreground(mutedColor)
	promptFocus := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	textBlur := lipgloss.NewStyle().Foreground(textColor)
	textFocus := lipgloss.NewStyle().Foreground(accentColor)

	m.dateInput.PromptStyle = promptBlur
	m.periodInput.PromptStyle = promptBlur
	m.projectInput.PromptStyle = promptBlur
	m.taskInput.PromptStyle = promptBlur
	m.manualDateInput.PromptStyle = promptBlur
	m.manualProjectInput.PromptStyle = promptBlur
	m.manualTaskInput.PromptStyle = promptBlur
	m.manualHoursInput.PromptStyle = promptBlur
	m.dateInput.TextStyle = textBlur
	m.periodInput.TextStyle = textBlur
	m.projectInput.TextStyle = textBlur
	m.taskInput.TextStyle = textBlur
	m.manualDateInput.TextStyle = textBlur
	m.manualProjectInput.TextStyle = textBlur
	m.manualTaskInput.TextStyle = textBlur
	m.manualHoursInput.TextStyle = textBlur

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
	case 4:
		m.manualDateInput.Focus()
		m.manualDateInput.PromptStyle = promptFocus
		m.manualDateInput.TextStyle = textFocus
	case 5:
		m.manualProjectInput.Focus()
		m.manualProjectInput.PromptStyle = promptFocus
		m.manualProjectInput.TextStyle = textFocus
	case 6:
		m.manualTaskInput.Focus()
		m.manualTaskInput.PromptStyle = promptFocus
		m.manualTaskInput.TextStyle = textFocus
	case 7:
		m.manualHoursInput.Focus()
		m.manualHoursInput.PromptStyle = promptFocus
		m.manualHoursInput.TextStyle = textFocus
	}
}

func (m tuiModel) canCycleFocusedOptions() bool {
	return m.focusIndex == 2 || m.focusIndex == 3 || m.focusIndex == 5 || m.focusIndex == 6
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
	case 5:
		if len(m.projects) == 0 {
			return
		}
		m.manualProjectIndex = (m.manualProjectIndex + 1) % len(m.projects)
		m.manualProjectInput.SetValue(m.projects[m.manualProjectIndex])
	case 6:
		if len(m.tasks) == 0 {
			return
		}
		m.manualTaskIndex = (m.manualTaskIndex + 1) % len(m.tasks)
		m.manualTaskInput.SetValue(m.tasks[m.manualTaskIndex])
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
	case 5:
		if len(m.projects) == 0 {
			return
		}
		m.manualProjectIndex--
		if m.manualProjectIndex < 0 {
			m.manualProjectIndex = len(m.projects) - 1
		}
		m.manualProjectInput.SetValue(m.projects[m.manualProjectIndex])
	case 6:
		if len(m.tasks) == 0 {
			return
		}
		m.manualTaskIndex--
		if m.manualTaskIndex < 0 {
			m.manualTaskIndex = len(m.tasks) - 1
		}
		m.manualTaskInput.SetValue(m.tasks[m.manualTaskIndex])
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

func (m *tuiModel) addManualEntryFromInputs() (string, error) {
	projectName := sanitizeTagValue(m.manualProjectInput.Value())
	if projectName == "" {
		return "", fmt.Errorf("project is required")
	}

	taskName := sanitizeTagValue(m.manualTaskInput.Value())
	task, normalizedDate, hours, label, err := buildManualTask(
		strings.TrimSpace(m.manualDateInput.Value()),
		projectName,
		taskName,
		strings.TrimSpace(m.manualHoursInput.Value()),
	)
	if err != nil {
		return "", err
	}

	timesheetFile := file.TimesheetFile{Name: timesheetFilename}
	tasklist := timesheetFile.ReadFile()
	tasklist.AddTask(task)
	timesheetFile.WriteFile(tasklist)

	m.manualHoursInput.SetValue("")
	return fmt.Sprintf("Added %.2fh on %s to %s.", hours, normalizedDate, label), nil
}

func buildManualTask(manualDate string, projectName string, taskName string, hoursText string) (task *todotxt.Task, normalizedDate string, hours float64, label string, err error) {
	manualDay, err := parseFlexibleDate(manualDate)
	if err != nil {
		return nil, "", 0, "", fmt.Errorf("date is invalid")
	}
	manualDay = time.Date(manualDay.Year(), manualDay.Month(), manualDay.Day(), 0, 0, 0, 0, time.UTC)

	hours, err = strconv.ParseFloat(hoursText, 64)
	if err != nil || hours <= 0 {
		return nil, "", 0, "", fmt.Errorf("hours must be a positive number")
	}

	normalizedDate = manualDay.Format("2006-01-02")
	taskText := fmt.Sprintf("%s +%s", normalizedDate, projectName)
	if taskName != "" {
		taskText += " task:" + taskName
	}
	taskText += " hours:" + strconv.FormatFloat(hours, 'f', -1, 64)

	task, err = todotxt.ParseTask(taskText)
	if err != nil {
		return nil, "", 0, "", err
	}

	task.CreatedDate = manualDay
	task.Complete()
	task.CompletedDate = manualDay
	task.Original = task.String()

	label = projectName
	if taskName != "" {
		label += "." + taskName
	}

	return task, normalizedDate, hours, label, nil
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

	parsed, err := parseFlexibleDate(value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q (expected YYYY-MM-DD)", value)
	}

	return parsed, nil
}

func parseFlexibleDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}

	if strings.Contains(value, "-") {
		parts := strings.Split(value, "-")
		if len(parts) != 3 {
			return time.Time{}, fmt.Errorf("invalid date format")
		}

		yearPart := strings.TrimSpace(parts[0])
		monthPart := strings.TrimSpace(parts[1])
		dayPart := strings.TrimSpace(parts[2])

		if len(yearPart) != 4 || monthPart == "" || dayPart == "" {
			return time.Time{}, fmt.Errorf("incomplete date")
		}

		year, err := strconv.Atoi(yearPart)
		if err != nil {
			return time.Time{}, err
		}
		month, err := strconv.Atoi(monthPart)
		if err != nil || month < 1 || month > 12 {
			return time.Time{}, fmt.Errorf("invalid month")
		}
		day, err := strconv.Atoi(dayPart)
		if err != nil || day < 1 || day > 31 {
			return time.Time{}, fmt.Errorf("invalid day")
		}

		candidate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if candidate.Year() != year || int(candidate.Month()) != month || candidate.Day() != day {
			return time.Time{}, fmt.Errorf("invalid date")
		}

		return candidate, nil
	}

	digits := extractDateDigits(value)
	if len(digits) != 8 {
		return time.Time{}, fmt.Errorf("date must contain 8 digits")
	}

	if !isPotentialDateDigits(digits) {
		return time.Time{}, fmt.Errorf("date is invalid")
	}

	parsed, err := time.Parse("20060102", digits)
	if err != nil {
		return time.Time{}, err
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
