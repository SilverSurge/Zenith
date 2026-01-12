package ui

import "github.com/charmbracelet/lipgloss"

var (
	AccentColor = lipgloss.Color("#94B4C1")
	GrayColor   = lipgloss.Color("#E4E4E4")

	RedColor    = lipgloss.Color("#F39EB6")

	HeaderStyle = lipgloss.NewStyle().
			Background(AccentColor).
			Foreground(lipgloss.Color("#213448")).
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
	FooterTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EAE0CF"))
)
