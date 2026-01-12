package main

import (
	"fmt"
	"os"
	"zenith/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if _, err := tea.NewProgram(ui.InitialModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
