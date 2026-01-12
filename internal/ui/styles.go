package ui

import "github.com/charmbracelet/lipgloss"

var (
	AccentColor = lipgloss.Color("99")
	GrayColor   = lipgloss.Color("245")
	RedColor    = lipgloss.Color("196")

	HeaderStyle = lipgloss.NewStyle().
			Background(AccentColor).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			Bold(true)

	DateStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)

	CursorCol = lipgloss.NewStyle().Width(3)
	CheckCol  = lipgloss.NewStyle().Width(3)

	HelpKeyStyle   = lipgloss.NewStyle().Foreground(AccentColor).Bold(true)
	HelpValueStyle = lipgloss.NewStyle().Foreground(GrayColor)
	GrayTextStyle  = lipgloss.NewStyle().Foreground(GrayColor)
)
