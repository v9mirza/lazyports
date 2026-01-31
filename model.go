package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PortEntry represents a single process listening on a port
type PortEntry struct {
	Port     string
	Protocol string
	PID      string
	Process  string
	State    string
	Address  string
}

type model struct {
	table           table.Model
	textInput       textinput.Model
	entries         []PortEntry
	filteredEntries []PortEntry
	err             error
	status          string
	width           int
	height          int
	isFiltering     bool
	showDetails     bool
	detailsContent  string
}

func (m model) Init() tea.Cmd {
	return tea.Batch(loadPorts, textinput.Blink)
}

func loadPorts() tea.Msg {
	entries, err := getPorts()
	if err != nil {
		return err
	}
	return entries
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// If filtering, route messages to textInput
	if m.isFiltering {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter", "esc":
				m.isFiltering = false
				m.table.Focus()
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		m.filterEntries()
		return m, cmd
	}

	// If details view is open
	if m.showDetails {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" || msg.String() == "q" || msg.String() == "enter" {
				m.showDetails = false
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "/":
			m.isFiltering = true
			m.textInput.Focus()
			m.textInput.SetValue("")
			return m, textinput.Blink
		case "r":
			m.status = "Refreshing..."
			return m, loadPorts
		case "enter":
			if len(m.filteredEntries) > 0 {
				selectedIdx := m.table.Cursor()
				if selectedIdx >= 0 && selectedIdx < len(m.filteredEntries) {
					target := m.filteredEntries[selectedIdx]
					details, err := getProcessDetails(target.PID)
					if err != nil {
						m.detailsContent = fmt.Sprintf("Error: %v", err)
					} else {
						// Format the details nicely
						if target.PID == "-" {
							m.detailsContent = details
						} else {
							// ps output is usually: USER STARTED COMMAND (params...)
							// We'll just display it raw but formatted
							m.detailsContent = fmt.Sprintf(
								"Port:      %s/%s\nPID:       %s\nAddress:   %s\nState:     %s\nProcess:   %s\n\n%s",
								target.Port, target.Protocol, target.PID, target.Address, target.State, target.Process, details,
							)
						}
					}
					m.showDetails = true
				}
			}
		case "k":
			if len(m.filteredEntries) > 0 {
				selectedIdx := m.table.Cursor()
				// Safety check
				if selectedIdx >= 0 && selectedIdx < len(m.filteredEntries) {
					target := m.filteredEntries[selectedIdx]
					if target.PID == "-" {
						if os.Geteuid() == 0 {
							m.status = "Cannot kill system process"
						} else {
							m.status = "Run as sudo to kill this process"
						}
						return m, nil
					}
					err := killProcess(target.PID)
					if err != nil {
						m.status = fmt.Sprintf("Error killing %s: %v", target.PID, err)
					} else {
						m.status = fmt.Sprintf("Killed %s (%s)", target.Process, target.PID)
						return m, loadPorts
					}
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Logo (2 lines) + Footer (2 lines) + Borders (2 lines) = ~6 lines of chrome
		// We set proper height constraints to ensure View() string <= m.height
		availableHeight := m.height - 7 // 7 to be safe

		baseStyle = baseStyle.Width(m.width - 2).Height(availableHeight)
		m.table.SetWidth(m.width - 4)

		// Internal table height inside the border
		// baseStyle has borders, so subtract 2 more
		tableHeight := availableHeight - 2
		if tableHeight < 2 {
			tableHeight = 2
		}
		m.table.SetHeight(tableHeight)

	case []PortEntry:
		m.entries = msg
		m.filterEntries() // Initial filter (or reset)
		m.err = nil
		if m.status == "Refreshing..." {
			m.status = "Refreshed"
		}

	case error:
		m.err = msg
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *model) filterEntries() {
	query := strings.ToLower(m.textInput.Value())
	m.filteredEntries = []PortEntry{}

	for _, e := range m.entries {
		if query == "" ||
			strings.Contains(strings.ToLower(e.Process), query) ||
			strings.Contains(e.Port, query) ||
			strings.Contains(e.PID, query) {
			m.filteredEntries = append(m.filteredEntries, e)
		}
	}
	m.updateTable()
}

func (m *model) updateTable() {
	rows := []table.Row{}
	for _, e := range m.filteredEntries {
		stateIcon := "○"
		if strings.Contains(e.State, "LISTEN") {
			stateIcon = "●"
		} else if strings.Contains(e.State, "ESTAB") {
			stateIcon = "↔"
		}

		rows = append(rows, table.Row{
			e.Port,
			e.Protocol,
			stateIcon + " " + e.State,
			e.PID,
			e.Address,
			e.Process,
		})
	}
	m.table.SetRows(rows)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress 'q' to quit", m.err)
	}

	// Details Modal
	if m.showDetails {
		// Calculate center position (simple approximation)
		content := detailsTitleStyle.Render("Connection Details") + "\n" + m.detailsContent
		content += "\n\n" + helpStyle.Render("Press Esc/Enter to close")

		box := detailsStyle.Render(content)

		// Center it manually or just overlay it. For simplicity in Bubble Tea without lipgloss.Place,
		// we'll just render it clear screen.
		// A proper overlay is harder, let's just show the box fullscreen-ish.
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}

	logo := logoStyle.Render("⚡ LazyPorts")
	tableView := baseStyle.Render(m.table.View())

	// Footer
	controls := "↑/↓: Navigate • /: Filter • Enter: Details • k: Kill • r: Refresh • q: Quit"
	if m.isFiltering {
		controls = "Type to search • Esc/Enter: Done"
		// Render Input
		inputView := inputStyle.Render(m.textInput.View())
		// Replace bottom area
		return fmt.Sprintf("%s\n%s\n%s\n%s", logo, tableView, inputView, helpStyle.Render(controls))
	} else {
		if m.status != "" {
			controls = statusStyle.Render(m.status) + " • " + controls
		}
		return fmt.Sprintf("%s\n%s\n%s", logo, tableView, helpStyle.Render(controls))
	}
}
