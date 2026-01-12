# Zenith Project Context

## Project Overview
**Zenith** is a terminal-based user interface (TUI) daily task manager built in Go. It allows users to manage their daily to-dos with a clean, keyboard-driven interface.

### Key Technologies
*   **Language:** Go (1.24.5)
*   **TUI Framework:** [Bubbletea](https://github.com/charmbracelet/bubbletea)
*   **Styling:** [Lipgloss](https://github.com/charmbracelet/lipgloss)
*   **Components:** [Bubbles](https://github.com/charmbracelet/bubbles)

### Architecture
The application follows The Elm Architecture (Model-View-Update) provided by the Bubbletea framework:
*   **Model:** Holds the state (list of tasks, current date, UI state, input fields).
*   **Update:** Handles messages (keypresses, window resizes) and updates the model.
*   **View:** Renders the UI based on the current state using Lipgloss for styling.

**Persistence:**
Tasks are persisted as JSON files in a local directory (currently hardcoded to `D:\.zenith`). Each day has its own file named `tasks_YYYY-MM-DD.json`.

## Building and Running

### Prerequisites
*   Go 1.24 or later.

### Commands
*   **Run:** `go run main.go`
*   **Build:** `go build -o zenith.exe`
*   **Test:** (No tests currently implemented)

## Key Features & Controls
*   **Navigation:**
    *   `h` / `l`: Previous / Next Day
    *   `t`: Jump to Today
    *   `g`: Go to specific date (YYYY-MM-DD)
*   **Task Management:**
    *   `n`: New Task
    *   `e`: Edit Task
    *   `d`: Delete Task
    *   `Space`: Toggle Complete
*   **Search:**
    *   `/`: Search Tasks
*   **General:**
    *   `?`: Help
    *   `q`: Quit

## Development Conventions
*   **Code Structure:** Currently, all logic resides in `main.go`. Future refactoring might separate the model, update, and view logic into packages.
*   **Styling:** UI styles (colors, borders) are defined using Lipgloss at the package level.
*   **Error Handling:** Basic error handling for file I/O.
