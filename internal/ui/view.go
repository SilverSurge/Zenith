package ui

import (
	"fmt"
	"strings"
	"time"
	"zenith/internal/model"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) HelpView() string {
	keys := [][]string{
		{"j/k", "Move cursor Up/Down"},
		{"h/l", "Previous/Next Day"},
		{"t", "Jump to Today"},
		{"n", "New Task"},
		{"e", "Edit Task"},
		{" ", "Toggle Complete"},
		{"d", "Delete Task"},
		{"/", "Search Tasks"},
		{"g", "Go to Date"},
		{"q", "Quit"},
	}

	var s strings.Builder
	s.WriteString(HeaderStyle.Render(" ZENITH HELP ") + "\n\n")

	for _, k := range keys {
		// Render the styled strings first
		key := HelpKeyStyle.Render(k[0])
		val := HelpValueStyle.Render(k[1])

		// Place the key in a box of 10 cells, aligned to the left
		keyCol := lipgloss.NewStyle().Width(10).Align(lipgloss.Left).Render(key)

		// Concatenate with a gap
		row := fmt.Sprintf(" %s  %s\n", keyCol, val)
		s.WriteString(row)
	}

	s.WriteString("\n" + GrayTextStyle.Render("Press any key to return..."))
	return s.String()
}

func startOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func (m Model) View() string {
	if m.Width == 0 {
		return "Initializing..."
	}

	if m.State == HelpState {
		return m.Center(m.HelpView())
	}

	// --- Header ---
	header := HeaderStyle.Render(" ZENITH ")
	
	// Tabs
	taskTabStyle := lipgloss.NewStyle().Padding(0, 1)
	scriptTabStyle := lipgloss.NewStyle().Padding(0, 1)

	if m.ActiveTab == TaskTab {
		taskTabStyle = taskTabStyle.Background(GrayColor).Foreground(lipgloss.Color("230")).Bold(true)
	} else {
		scriptTabStyle = scriptTabStyle.Background(GrayColor).Foreground(lipgloss.Color("230")).Bold(true)
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Bottom, 
		taskTabStyle.Render(" Tasks "),
		scriptTabStyle.Render(" Scripts "),
	)
	
	topBar := lipgloss.JoinHorizontal(lipgloss.Center, header, "  ", tabs)

	// --- Main Content ---
	var content string
	var footer string

	if m.ActiveTab == TaskTab {
		content = m.viewTasks()
		footer = m.viewTaskFooter()
	} else {
		content = m.viewScripts()
		footer = m.viewScriptFooter()
	}

	main := topBar + "\n" + content + footer
	return m.Center(main)
}

func (m Model) viewTasks() string {
	var list strings.Builder
	list.WriteString("\n")

	dateInfo := DateStyle.Render(" " + m.SelectedDate.Format("Monday, 02 Jan 2006"))
	list.WriteString(dateInfo + "\n\n")

	today := startOfDay(time.Now())
	selected := startOfDay(m.SelectedDate)

	paged := m.PagedTasks()
	for i, t := range paged {
		cur := " "
		if i == m.Cursor && m.State == ViewState && m.ActiveTab == TaskTab {
			cur = lipgloss.NewStyle().Foreground(AccentColor).Render("❯")
		}

		icon := "🟪"
		if t.Completed {
			icon = "☑"
		}

		style := lipgloss.NewStyle()
		if t.Completed {
			style = style.Foreground(GrayColor).Strikethrough(true)
		} else if selected.Before(today) {
			style = style.Foreground(RedColor).Bold(true)
		}

		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			CursorCol.Render(cur),
			CheckCol.Render(icon),
			style.Render(t.Title),
		)
		list.WriteString(row + "\n")
	}

	// Pad empty space
	for i := len(paged); i < m.PageSize(); i++ {
		list.WriteString("\n")
	}
	return list.String()
}

func (m Model) PagedScripts() []model.Script {
	ps := m.PageSize()
	start := m.ScriptPage * ps
	end := start + ps

	if start >= len(m.Scripts) {
		return nil
	}
	if end > len(m.Scripts) {
		end = len(m.Scripts)
	}
	return m.Scripts[start:end]
}

func (m Model) viewScripts() string {
	var list strings.Builder
	list.WriteString("\n")
	list.WriteString(lipgloss.NewStyle().Bold(true).Foreground(AccentColor).Render(" AUTOMATION SCRIPTS") + "\n\n")

	paged := m.PagedScripts()
	for i, s := range paged {
		cur := " "
		if i == m.ScriptCursor && m.State == ViewState && m.ActiveTab == ScriptTab {
			cur = lipgloss.NewStyle().Foreground(AccentColor).Render("❯")
		}

		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			CursorCol.Render(cur),
			lipgloss.NewStyle().Width(20).Bold(true).Render(s.Name),
			lipgloss.NewStyle().Foreground(GrayColor).Render(s.Description),
		)
		list.WriteString(row + "\n")
	}
	
	// Pad empty space
	for i := len(paged); i < m.PageSize(); i++ {
		list.WriteString("\n")
	}

	return list.String()
}

func (m Model) viewTaskFooter() string {
	switch m.State {
	case SearchState:
		return "\n " + lipgloss.NewStyle().Foreground(AccentColor).Render("SEARCH:") + " " + m.SearchInput.View()
	case InputState, EditState:
		label := "NEW:"
		if m.State == EditState {
			label = "EDIT:"
		}
		return "\n " + lipgloss.NewStyle().Foreground(AccentColor).Render(label) + " " + m.TextInput.View()
	case GotoDateState:
		return "\n " + lipgloss.NewStyle().Foreground(AccentColor).Render("GO TO DATE:") + " " + m.DateInput.View()
	default:
		pageInfo := fmt.Sprintf(" Page %d / %d ", m.Page+1, m.TotalPages())
		return GrayTextStyle.Render("\n /: search • ?: help • Tab: Switch • " + pageInfo)
	}
}

func (m Model) ScriptTotalPages() int {
	items := len(m.Scripts)
	ps := m.PageSize()
	if items == 0 {
		return 1
	}
	return (items + ps - 1) / ps
}

func (m Model) viewScriptFooter() string {
	switch m.State {
	case RunScriptState:
		if len(m.ArgQueue) > 0 {
			argName := m.ArgQueue[0]
			return "\n " + lipgloss.NewStyle().Foreground(AccentColor).Render("ENTER "+argName+":") + " " + m.TextInput.View()
		}
		return ""
	case ScriptInputState:
		var label string
		switch m.ScriptInputStep {
		case 0:
			label = "NAME:"
		case 1:
			label = "COMMAND:"
		case 2:
			label = "DESCRIPTION:"
		}
		return "\n " + lipgloss.NewStyle().Foreground(AccentColor).Render(label) + " " + m.TextInput.View()
	default:
		pageInfo := fmt.Sprintf(" Page %d / %d ", m.ScriptPage+1, m.ScriptTotalPages())
		return GrayTextStyle.Render("\n Enter: Run • n: New • e: Edit • d: Delete • Tab: Switch • " + pageInfo)
	}
}

func (m Model) Center(content string) string {
	return lipgloss.Place(
		m.Width,
		m.Height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.NewStyle().
			Width(m.Width-10).
			Height(m.Height-6).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentColor).
			Padding(1).
			Render(content),
	)
}
