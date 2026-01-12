package ui

import (
	"sort"
	"time"
	"zenith/internal/model"
	"zenith/internal/repository"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type sessionState int

const (
	ViewState sessionState = iota
	InputState
	EditState
	SearchState
	GotoDateState
	HelpState
	ScriptInputState // For adding/editing scripts
	RunScriptState   // For answering placeholders
)

type Tab int

const (
	TaskTab Tab = iota
	ScriptTab
)

type Model struct {
	// Tabs
	ActiveTab Tab

	// Tasks
	Tasks        []model.Task
	Cursor       int
	Page         int
	SelectedDate time.Time

	// Scripts
	Scripts       []model.Script
	ScriptCursor  int
	ScriptPage    int
	PendingScript *model.Script // Script currently being run
	ArgQueue      []string      // Placeholders waiting for input
	ScriptArgs    map[string]string

	// Script Editing/Creation
	ScriptInputStep int          // 0: Name, 1: Command, 2: Description
	ActiveScript    model.Script // Temporary holder for script being edited/created
	IsEditing       bool

	// Inputs
	TextInput   textinput.Model
	SearchInput textinput.Model
	DateInput   textinput.Model

	State  sessionState
	Width  int
	Height int
}

func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = " Description..."

	si := textinput.New()
	si.Placeholder = " Search..."

	di := textinput.New()
	di.Placeholder = " YYYY-MM-DD"
	di.CharLimit = 10

	now := time.Now()
	m := Model{
		ActiveTab:    TaskTab,
		Tasks:        repository.LoadTasks(now),
		Scripts:      repository.LoadScripts(),
		SelectedDate: now,
		TextInput:    ti,
		SearchInput:  si,
		DateInput:    di,
		State:        ViewState,
		ScriptArgs:   make(map[string]string),
	}
	m.SortTasks()
	return m
}

func (m *Model) SortTasks() {
	sort.SliceStable(m.Tasks, func(i, j int) bool {
		if m.Tasks[i].Completed != m.Tasks[j].Completed {
			return !m.Tasks[i].Completed
		}
		return m.Tasks[i].CreatedAt.Before(m.Tasks[j].CreatedAt)
	})
}

func (m Model) Init() tea.Cmd { return nil }
