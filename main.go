package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#6c7086")) // Overlay0

	// Status/Footer styles
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f5c2e7"))              // Pink
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).MarginTop(1) // Overlay0
	inputStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#cba6f7")).Padding(0, 1).Width(60)

	// Details View Styles
	detailsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#cba6f7")).
			Padding(1, 2).
			Width(60)
	detailsTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#cba6f7")).
				Bold(true).
				MarginBottom(1)
	detailsLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6c7086")).
				Width(12)

	// Logo Style
	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cba6f7")). // Mauve
			Bold(true).
			MarginBottom(1)

	// Common ports mapping for guessing services when running without sudo
	commonPorts = map[string]string{
		"21":    "FTP",
		"22":    "SSH",
		"23":    "Telnet",
		"25":    "SMTP",
		"53":    "DNS",
		"80":    "HTTP",
		"110":   "POP3",
		"143":   "IMAP",
		"443":   "HTTPS",
		"3306":  "MySQL",
		"5432":  "PostgreSQL",
		"6379":  "Redis",
		"8080":  "HTTP-Alt",
		"27017": "MongoDB",
	}
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

func getPorts() ([]PortEntry, error) {
	cmd := exec.Command("ss", "-tulnp")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run ss: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	var entries []PortEntry

	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 || fields[0] == "Netid" {
			continue
		}

		proto := fields[0]
		state := fields[1]
		localAddr := fields[4]
		address := localAddr
		port := ""
		lastColon := strings.LastIndex(localAddr, ":")
		if lastColon != -1 {
			port = localAddr[lastColon+1:]
			address = localAddr[:lastColon]
		}
		if address == "*" || address == "0.0.0.0" || address == "[::]" {
			address = "All Interfaces"
		}

		pid := ""
		process := ""
		for _, f := range fields {
			if strings.Contains(f, "users:((") {
				content := strings.TrimPrefix(f, "users:((")
				content = strings.TrimSuffix(content, "))")
				content = strings.TrimSuffix(content, ")")
				parts := strings.Split(content, ",")
				for _, p := range parts {
					if strings.HasPrefix(p, "\"") {
						process = strings.Trim(p, "\"")
					}
					if strings.HasPrefix(p, "pid=") {
						pid = strings.TrimPrefix(p, "pid=")
					}
				}
			}
		}

		if pid == "" {
			pid = "-"
			if service, ok := commonPorts[port]; ok {
				process = fmt.Sprintf("%s (requires sudo)", service)
			} else {
				process = "(requires sudo)"
			}
		}

		entries = append(entries, PortEntry{
			Port:     port,
			Protocol: proto,
			PID:      pid,
			Process:  process,
			State:    state,
			Address:  address,
		})
	}

	// Sort entries by Port (numerically) and then Protocol
	sort.Slice(entries, func(i, j int) bool {
		p1, err1 := strconv.Atoi(entries[i].Port)
		p2, err2 := strconv.Atoi(entries[j].Port)

		if err1 == nil && err2 == nil {
			if p1 == p2 {
				return entries[i].Protocol < entries[j].Protocol
			}
			return p1 < p2
		}
		// Fallback for non-numeric ports (rare)
		return entries[i].Port < entries[j].Port
	})

	return entries, nil
}

func killProcess(pid string) error {
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return err
	}
	proc, err := os.FindProcess(pidInt)
	if err != nil {
		return err
	}
	return proc.Kill()
}

func getProcessDetails(pid string) (string, error) {
	if pid == "-" {
		return "Process details require sudo privileges.", nil
	}

	// Run ps to get details: User, Start Time, Command
	cmd := exec.Command("ps", "-p", pid, "-o", "user,lstart,cmd", "--no-headers")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get details: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
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
						m.status = "Run as sudo to kill this process"
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
