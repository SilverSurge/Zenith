package ui

import (
	"strings"
	"time"
	"zenith/internal/model"
	"zenith/internal/repository"
	"zenith/internal/script"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width, m.Height = msg.Width, msg.Height
		m.ClampCursor()

	case tea.KeyMsg:
		// --- GLOBAL KEYS ---
		switch msg.String() {
		case "tab":
			if m.ActiveTab == TaskTab {
				m.ActiveTab = ScriptTab
			} else {
				m.ActiveTab = TaskTab
			}
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}

		// --- HELP MODE ---
		if m.State == HelpState {
			m.State = ViewState
			return m, nil
		}
		
		// --- RUN SCRIPT MODE (Arg Collection) ---
		if m.State == RunScriptState {
			switch msg.String() {
			case "enter":
				// Save the input for the current argument
				currentArg := m.ArgQueue[0]
				m.ScriptArgs[currentArg] = m.TextInput.Value()
				
				// Move to next arg
				m.ArgQueue = m.ArgQueue[1:]
				m.TextInput.SetValue("")
				
				if len(m.ArgQueue) == 0 {
					// All args collected, run the script
					finalCmd := script.ReplacePlaceholders(m.PendingScript.Command, m.ScriptArgs)
					_ = script.Run(finalCmd) // TODO: Handle error or output
					m.State = ViewState
					m.PendingScript = nil
				}
				return m, nil
			case "esc":
				m.State = ViewState
				m.PendingScript = nil
				return m, nil
			default:
				m.TextInput, cmd = m.TextInput.Update(msg)
				return m, cmd
			}
		}

		// --- GO TO DATE MODE ---
		if m.State == GotoDateState {
			switch msg.String() {
			case "enter":
				if d, err := time.Parse("2006-01-02", m.DateInput.Value()); err == nil {
					m.SelectedDate = d
					m.Tasks = repository.LoadTasks(d)
					m.SortTasks()
					m.Page, m.Cursor = 0, 0
				}
				m.DateInput.SetValue("")
				m.State = ViewState
			case "esc":
				m.DateInput.SetValue("")
				m.State = ViewState
			default:
				m.DateInput, cmd = m.DateInput.Update(msg)
				return m, cmd
			}
			return m, nil
		}

		// --- SEARCH MODE ---
		if m.State == SearchState {
			switch msg.String() {
			case "esc", "enter":
				m.State = ViewState
				if msg.String() == "esc" {
					m.SearchInput.SetValue("")
				}
				m.Page, m.Cursor = 0, 0
			default:
				m.SearchInput, cmd = m.SearchInput.Update(msg)
				m.Page, m.Cursor = 0, 0
				return m, cmd
			}
			return m, nil
		}

		// --- SCRIPT INPUT MODE ---
		if m.State == ScriptInputState {
			switch msg.String() {
			case "enter":
				val := m.TextInput.Value()
				
				switch m.ScriptInputStep {
				case 0: // Name
					if val == "" { return m, nil } // Name required
					m.ActiveScript.Name = val
					m.ScriptInputStep++
					m.TextInput.SetValue(m.ActiveScript.Command) // Pre-fill if editing
					m.TextInput.Placeholder = " e.g. echo {{msg}}"
				case 1: // Command
					if val == "" { return m, nil } // Command required
					m.ActiveScript.Command = val
					m.ScriptInputStep++
					m.TextInput.SetValue(m.ActiveScript.Description) // Pre-fill if editing
					m.TextInput.Placeholder = " Describe what this does..."
				case 2: // Description
					m.ActiveScript.Description = val
					
					// Save
					if m.IsEditing {
						idx := m.RealScriptIndex()
						if idx >= 0 && idx < len(m.Scripts) {
							m.Scripts[idx] = m.ActiveScript
						}
					} else {
						m.Scripts = append(m.Scripts, m.ActiveScript)
					}
					
					repository.SaveScripts(m.Scripts)
					m.State = ViewState
					m.TextInput.SetValue("")
				}
				return m, nil
			case "esc":
				m.State = ViewState
				m.TextInput.SetValue("")
				return m, nil
			default:
				m.TextInput, cmd = m.TextInput.Update(msg)
				return m, cmd
			}
		}

		// --- INPUT / EDIT MODE (Tasks) ---
		if m.State == InputState || m.State == EditState {
			switch msg.String() {
			case "enter":
				if m.TextInput.Value() != "" {
					if m.State == EditState {
						idx := m.RealIndex()
						if idx >= 0 {
							m.Tasks[idx].Title = m.TextInput.Value()
						}
					} else {
						m.Tasks = append(m.Tasks, model.Task{
							Title:     m.TextInput.Value(),
							CreatedAt: time.Now(),
						})
					}
					m.SortTasks()
					repository.SaveTasks(m.SelectedDate, m.Tasks)
					m.TextInput.SetValue("")
					m.State = ViewState
				}
			case "esc":
				m.State = ViewState
			default:
				m.TextInput, cmd = m.TextInput.Update(msg)
				return m, cmd
			}
			return m, nil
		}

		// --- TAB SPECIFIC LOGIC ---
		if m.ActiveTab == TaskTab {
			return m.updateTaskTab(msg)
		} else {
			return m.updateScriptTab(msg)
		}
	}

	return m, nil
}

func (m Model) updateScriptTab(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.ScriptCursor > 0 {
			m.ScriptCursor--
		} else if m.ScriptPage > 0 {
			m.ScriptPage--
			m.ScriptCursor = m.PageSize() - 1
		}

	case "down", "j":
		if m.ScriptCursor < len(m.PagedScripts())-1 {
			m.ScriptCursor++
		} else if m.ScriptPage < m.ScriptTotalPages()-1 {
			m.ScriptPage++
			m.ScriptCursor = 0
		}

	case "n": // New Script
		m.State = ScriptInputState
		m.ScriptInputStep = 0
		m.IsEditing = false
		m.ActiveScript = model.Script{}
		m.TextInput.SetValue("")
		m.TextInput.Placeholder = " Script Name"
		m.TextInput.Focus()

	case "e": // Edit Script
		if len(m.PagedScripts()) > 0 {
			idx := m.RealScriptIndex()
			if idx >= 0 && idx < len(m.Scripts) {
				m.State = ScriptInputState
				m.ScriptInputStep = 0
				m.IsEditing = true
				m.ActiveScript = m.Scripts[idx]
				m.TextInput.SetValue(m.ActiveScript.Name)
				m.TextInput.Focus()
			}
		}

	case "d": // Delete Script
		if len(m.PagedScripts()) > 0 {
			idx := m.RealScriptIndex()
			if idx >= 0 && idx < len(m.Scripts) {
				m.Scripts = append(m.Scripts[:idx], m.Scripts[idx+1:]...)
				repository.SaveScripts(m.Scripts)
				m.ClampScriptCursor()
			}
		}

	case "enter":
		if len(m.PagedScripts()) == 0 {
			return m, nil
		}
		
		idx := m.RealScriptIndex()
		if idx >= 0 && idx < len(m.Scripts) {
			target := m.Scripts[idx]
			placeholders := script.GetPlaceholders(target.Command)
			
			if len(placeholders) > 0 {
				m.PendingScript = &target
				m.ArgQueue = placeholders
				m.ScriptArgs = make(map[string]string)
				m.State = RunScriptState
				m.TextInput.SetValue("")
				m.TextInput.Focus()
			} else {
				_ = script.Run(target.Command)
			}
		}

	case "q":
		return m, tea.Quit
	}
	m.ClampScriptCursor()
	return m, nil
}

func (m *Model) ClampScriptCursor() {
	paged := m.PagedScripts()
	if len(paged) == 0 {
		m.ScriptCursor = 0
		return
	}
	if m.ScriptCursor < 0 {
		m.ScriptCursor = 0
	}
	if m.ScriptCursor >= len(paged) {
		m.ScriptCursor = len(paged) - 1
	}
}

func (m Model) RealScriptIndex() int {
	ps := m.PageSize()
	return m.ScriptPage*ps + m.ScriptCursor
}

func (m Model) updateTaskTab(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "?":
		m.State = HelpState

	case "g":
		m.State = GotoDateState
		m.DateInput.SetValue("")
		m.DateInput.Focus()

	case "/":
		m.State = SearchState
		m.SearchInput.Focus()
		m.Page, m.Cursor = 0, 0

	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		} else if m.Page > 0 {
			m.Page--
			m.Cursor = m.PageSize() - 1
		}

	case "down", "j":
		if m.Cursor < len(m.PagedTasks())-1 {
			m.Cursor++
		} else if m.Page < m.TotalPages()-1 {
			m.Page++
			m.Cursor = 0
		}

	case "left", "h":
		m.SelectedDate = m.SelectedDate.AddDate(0, 0, -1)
		m.Tasks = repository.LoadTasks(m.SelectedDate)
		m.SortTasks()
		m.Page, m.Cursor = 0, 0

	case "right", "l":
		m.SelectedDate = m.SelectedDate.AddDate(0, 0, 1)
		m.Tasks = repository.LoadTasks(m.SelectedDate)
		m.SortTasks()
		m.Page, m.Cursor = 0, 0

	case "t":
		m.SelectedDate = time.Now()
		m.Tasks = repository.LoadTasks(m.SelectedDate)
		m.SortTasks()
		m.Page, m.Cursor = 0, 0

	case "n":
		m.State = InputState
		m.TextInput.SetValue("")
		m.TextInput.Focus()

	case "e":
		if len(m.PagedTasks()) > 0 {
			m.State = EditState
			m.TextInput.SetValue(m.PagedTasks()[m.Cursor].Title)
			m.TextInput.Focus()
		}

	case " ":
		idx := m.RealIndex()
		if idx >= 0 {
			m.Tasks[idx].Completed = !m.Tasks[idx].Completed
			m.SortTasks()
			repository.SaveTasks(m.SelectedDate, m.Tasks)
		}

	case "d":
		idx := m.RealIndex()
		if idx >= 0 {
			m.Tasks = append(m.Tasks[:idx], m.Tasks[idx+1:]...)
			repository.SaveTasks(m.SelectedDate, m.Tasks)
			m.ClampCursor()
		}
	}
	m.ClampCursor()
	return m, cmd
}

// Helpers

func (m Model) FilteredTasks() []model.Task {
	if m.State != SearchState || m.SearchInput.Value() == "" {
		return m.Tasks
	}

	q := strings.ToLower(m.SearchInput.Value())
	var out []model.Task
	for _, t := range m.Tasks {
		if strings.Contains(strings.ToLower(t.Title), q) {
			out = append(out, t)
		}
	}
	return out
}

func (m Model) PageSize() int {
	ps := m.Height - 10
	if ps < 1 {
		return 1
	}
	return ps
}

func (m Model) TotalPages() int {
	items := len(m.FilteredTasks())
	ps := m.PageSize()
	if items == 0 {
		return 1
	}
	return (items + ps - 1) / ps
}

func (m Model) PagedTasks() []model.Task {
	ps := m.PageSize()
	items := m.FilteredTasks()

	start := m.Page * ps
	end := start + ps

	if start >= len(items) {
		return nil
	}
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func (m *Model) ClampCursor() {
	pageTasks := m.PagedTasks()
	if len(pageTasks) == 0 {
		m.Cursor = 0
		return
	}
	if m.Cursor < 0 {
		m.Cursor = 0
	}
	if m.Cursor >= len(pageTasks) {
		m.Cursor = len(pageTasks) - 1
	}
}

func (m Model) RealIndex() int {
	paged := m.PagedTasks()
	if len(paged) == 0 || m.Cursor >= len(paged) {
		return -1
	}
	target := paged[m.Cursor]
	for i, t := range m.Tasks {
		if t.CreatedAt.Equal(target.CreatedAt) && t.Title == target.Title {
			return i
		}
	}
	return -1
}
