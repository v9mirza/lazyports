package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	columns := []table.Column{
		{Title: "Port", Width: 8},
		{Title: "Proto", Width: 6},
		{Title: "State", Width: 12},
		{Title: "PID", Width: 8},
		{Title: "Address", Width: 20},
		{Title: "Process", Width: 20}, // Flexible
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10), // Initial height, adjusted on resize
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#6c7086")). // Overlay0
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#cdd6f4")). // Text
		Background(lipgloss.Color("#313244")). // Surface0
		Bold(false)
	t.SetStyles(s)

	ti := textinput.New()
	ti.Placeholder = "Search ports, processes, pids..."
	ti.CharLimit = 156
	ti.Width = 40

	m := model{table: t, textInput: ti}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
