package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Configuration & Persistence ---

const persistenceDir = "D:\\.zenith"

type Task struct {
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

func ensureDir() {
	if _, err := os.Stat(persistenceDir); os.IsNotExist(err) {
		_ = os.MkdirAll(persistenceDir, 0755)
	}
}

func getFilename(d time.Time) string {
	return filepath.Join(persistenceDir, fmt.Sprintf("tasks_%s.json", d.Format("2006-01-02")))
}

func loadTasks(d time.Time) []Task {
	ensureDir()
	data, err := os.ReadFile(getFilename(d))
	if err != nil {
		return []Task{}
	}
	var tasks []Task
	_ = json.Unmarshal(data, &tasks)
	return tasks
}

func saveTasks(d time.Time, tasks []Task) {
	ensureDir()
	data, _ := json.MarshalIndent(tasks, "", "  ")
	_ = os.WriteFile(getFilename(d), data, 0644)
}

// --- UI Styling ---

var (
	accentColor = lipgloss.Color("99")
	grayColor   = lipgloss.Color("245")
	redColor    = lipgloss.Color("196")

	headerStyle = lipgloss.NewStyle().
			Background(accentColor).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			Bold(true)

	dateStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	cursorCol = lipgloss.NewStyle().Width(3)
	checkCol  = lipgloss.NewStyle().Width(3)

	helpKeyStyle   = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	helpValueStyle = lipgloss.NewStyle().Foreground(grayColor)
	grayTextStyle  = lipgloss.NewStyle().Foreground(grayColor)
)

// --- Model ---

type state int

const (
	viewState state = iota
	inputState
	editState
	searchState
	gotoDateState
	helpState
)

type model struct {
	tasks        []Task
	cursor       int
	page         int
	selectedDate time.Time

	textInput   textinput.Model
	searchInput textinput.Model
	dateInput   textinput.Model

	state  state
	width  int
	height int
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = " Task description..."

	si := textinput.New()
	si.Placeholder = " Search..."

	di := textinput.New()
	di.Placeholder = " YYYY-MM-DD"
	di.CharLimit = 10

	now := time.Now()
	m := model{
		tasks:        loadTasks(now),
		selectedDate: now,
		textInput:    ti,
		searchInput:  si,
		dateInput:    di,
		state:        viewState,
	}
	m.sortTasks()
	return m
}

func (m *model) sortTasks() {
	sort.SliceStable(m.tasks, func(i, j int) bool {
		if m.tasks[i].Completed != m.tasks[j].Completed {
			return !m.tasks[i].Completed
		}
		return m.tasks[i].CreatedAt.Before(m.tasks[j].CreatedAt)
	})
}

// --- Helpers ---

func (m model) filteredTasks() []Task {
	if m.state != searchState || m.searchInput.Value() == "" {
		return m.tasks
	}

	q := strings.ToLower(m.searchInput.Value())
	var out []Task
	for _, t := range m.tasks {
		if strings.Contains(strings.ToLower(t.Title), q) {
			out = append(out, t)
		}
	}
	return out
}

func (m model) pageSize() int {
	ps := m.height - 10
	if ps < 1 {
		return 1
	}
	return ps
}

func (m model) totalPages() int {
	items := len(m.filteredTasks())
	ps := m.pageSize()
	if items == 0 {
		return 1
	}
	return (items + ps - 1) / ps
}

func (m model) pagedTasks() []Task {
	ps := m.pageSize()
	items := m.filteredTasks()

	start := m.page * ps
	end := start + ps

	if start >= len(items) {
		return nil
	}
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func (m *model) clampCursor() {
	pageTasks := m.pagedTasks()
	if len(pageTasks) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(pageTasks) {
		m.cursor = len(pageTasks) - 1
	}
}

func (m model) realIndex() int {
	paged := m.pagedTasks()
	if len(paged) == 0 || m.cursor >= len(paged) {
		return -1
	}
	target := paged[m.cursor]
	for i, t := range m.tasks {
		if t.CreatedAt.Equal(target.CreatedAt) && t.Title == target.Title {
			return i
		}
	}
	return -1
}

// --- Update ---

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.clampCursor()

	case tea.KeyMsg:
		// --- HELP MODE ---
		if m.state == helpState {
			m.state = viewState
			return m, nil
		}

		// --- GO TO DATE MODE ---
		if m.state == gotoDateState {
			switch msg.String() {
			case "enter":
				if d, err := time.Parse("2006-01-02", m.dateInput.Value()); err == nil {
					m.selectedDate = d
					m.tasks = loadTasks(d)
					m.sortTasks()
					m.page, m.cursor = 0, 0
				}
				m.dateInput.SetValue("")
				m.state = viewState
			case "esc":
				m.dateInput.SetValue("")
				m.state = viewState
			default:
				m.dateInput, cmd = m.dateInput.Update(msg)
				return m, cmd
			}
			return m, nil
		}

		// --- SEARCH MODE ---
		if m.state == searchState {
			switch msg.String() {
			case "esc", "enter":
				m.state = viewState
				if msg.String() == "esc" {
					m.searchInput.SetValue("")
				}
				m.page, m.cursor = 0, 0
			default:
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.page, m.cursor = 0, 0
				return m, cmd
			}
			return m, nil
		}

		// --- INPUT / EDIT MODE ---
		if m.state == inputState || m.state == editState {
			switch msg.String() {
			case "enter":
				if m.textInput.Value() != "" {
					if m.state == editState {
						idx := m.realIndex()
						if idx >= 0 {
							m.tasks[idx].Title = m.textInput.Value()
						}
					} else {
						m.tasks = append(m.tasks, Task{
							Title:     m.textInput.Value(),
							CreatedAt: time.Now(),
						})
					}
					m.sortTasks()
					saveTasks(m.selectedDate, m.tasks)
					m.textInput.SetValue("")
					m.state = viewState
				}
			case "esc":
				m.state = viewState
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			return m, nil
		}

		// --- VIEW MODE ---
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "?":
			m.state = helpState

		case "g":
			m.state = gotoDateState
			m.dateInput.SetValue("")
			m.dateInput.Focus()

		case "/":
			m.state = searchState
			m.searchInput.Focus()
			m.page, m.cursor = 0, 0

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else if m.page > 0 {
				m.page--
				m.cursor = m.pageSize() - 1
			}

		case "down", "j":
			if m.cursor < len(m.pagedTasks())-1 {
				m.cursor++
			} else if m.page < m.totalPages()-1 {
				m.page++
				m.cursor = 0
			}

		case "left", "h":
			m.selectedDate = m.selectedDate.AddDate(0, 0, -1)
			m.tasks = loadTasks(m.selectedDate)
			m.sortTasks()
			m.page, m.cursor = 0, 0

		case "right", "l":
			m.selectedDate = m.selectedDate.AddDate(0, 0, 1)
			m.tasks = loadTasks(m.selectedDate)
			m.sortTasks()
			m.page, m.cursor = 0, 0

		case "t":
			m.selectedDate = time.Now()
			m.tasks = loadTasks(m.selectedDate)
			m.sortTasks()
			m.page, m.cursor = 0, 0

		case "n":
			m.state = inputState
			m.textInput.SetValue("")
			m.textInput.Focus()

		case "e":
			if len(m.pagedTasks()) > 0 {
				m.state = editState
				m.textInput.SetValue(m.pagedTasks()[m.cursor].Title)
				m.textInput.Focus()
			}

		case " ":
			idx := m.realIndex()
			if idx >= 0 {
				m.tasks[idx].Completed = !m.tasks[idx].Completed
				m.sortTasks()
				saveTasks(m.selectedDate, m.tasks)
			}

		case "d":
			idx := m.realIndex()
			if idx >= 0 {
				m.tasks = append(m.tasks[:idx], m.tasks[idx+1:]...)
				saveTasks(m.selectedDate, m.tasks)
				m.clampCursor()
			}
		}
	}

	m.clampCursor()
	return m, nil
}

// --- View ---

func (m model) helpView() string {
	keys := [][]string{
		{"j/k", "Move cursor Up/Down"},
		{"h/l", "Previous/Next Day"},
		{"t", "Jump to Today"},
		{"n", "New Task"},
		{"e", "Edit Task"},
		{"Space", "Toggle Complete"},
		{"d", "Delete Task"},
		{"/", "Search Tasks"},
		{"g", "Go to Date"},
		{"q", "Quit"},
	}

	var s strings.Builder
	s.WriteString(headerStyle.Render(" ZENITH HELP ") + "\n\n")

	// for _, k := range keys {
	// 	row := fmt.Sprintf("%10s  %s\n", helpKeyStyle.Render(k[0]), helpValueStyle.Render(k[1]))
	// 	s.WriteString(row)
	// }
	//
	for _, k := range keys {
	    // Render the styled strings first
	    key := helpKeyStyle.Render(k[0])
	    val := helpValueStyle.Render(k[1])

	    // Place the key in a box of 15 cells, aligned to the right
	    keyCol := lipgloss.NewStyle().Width(10).Align(lipgloss.Left).Render(key)

	    // Concatenate with a gap
	    row := fmt.Sprintf(" %s  %s\n", keyCol, val)
	    s.WriteString(row)
	}

	s.WriteString("\n" + grayTextStyle.Render("Press any key to return..."))
	return s.String()
}

func startOfDay(t time.Time) time.Time {
    year, month, day := t.Date()
    return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	if m.state == helpState {
		return m.center(m.helpView())
	}

	header := headerStyle.Render(" ZENITH ")
	dateInfo := dateStyle.Render(" " + m.selectedDate.Format("Monday, 02 Jan 2006"))
	topBar := lipgloss.JoinHorizontal(lipgloss.Center, header, dateInfo)

	var list strings.Builder
	list.WriteString("\n")

	// today := time.Now().Truncate(24 * time.Hour)
	// selected := m.selectedDate.Truncate(24 * time.Hour)

	today := startOfDay(time.Now())
	selected := startOfDay(m.selectedDate)

	paged := m.pagedTasks()
	for i, t := range paged {
		cur := " "
		if i == m.cursor && m.state == viewState {
			cur = lipgloss.NewStyle().Foreground(accentColor).Render("❯")
		}

		icon := "🟪"
		if t.Completed {
			icon = "☑"
		}

		style := lipgloss.NewStyle()
		if t.Completed {
			style = style.Foreground(grayColor).Strikethrough(true)
		} else if selected.Before(today) {
			style = style.Foreground(redColor).Bold(true)
		}

		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			cursorCol.Render(cur),
			checkCol.Render(icon),
			style.Render(t.Title),
		)
		list.WriteString(row + "\n")
	}

	// Pad empty space if list is short to keep footer fixed
	for i := len(paged); i < m.pageSize(); i++ {
		list.WriteString("\n")
	}

	var footer string
	switch m.state {
	case searchState:
		footer = "\n " + lipgloss.NewStyle().Foreground(accentColor).Render("SEARCH:") + " " + m.searchInput.View()
	case inputState, editState:
		label := "NEW:"
		if m.state == editState {
			label = "EDIT:"
		}
		footer = "\n " + lipgloss.NewStyle().Foreground(accentColor).Render(label) + " " + m.textInput.View()
	case gotoDateState:
		footer = "\n " + lipgloss.NewStyle().Foreground(accentColor).Render("GO TO DATE:") + " " + m.dateInput.View()
	default:
		pageInfo := fmt.Sprintf(" Page %d / %d ", m.page+1, m.totalPages())
		footer = grayTextStyle.Render(
			"\n /: search • ?: help • " + pageInfo,
		)
	}

	main := topBar + "\n" + list.String() + footer
	return m.center(main)
}

func (m model) center(content string) string {
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.NewStyle().
			Width(m.width-10).
			Height(m.height-6).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Padding(1).
			Render(content),
	)
}

func main() {
	if _, err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
