package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
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
	table   table.Model
	entries []PortEntry
	err     error
	status  string
	width   int
	height  int
}

func (m model) Init() tea.Cmd {
	return loadPorts
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.status = "Refreshing..."
			return m, loadPorts
		case "k":
			if len(m.entries) > 0 {
				selectedIdx := m.table.Cursor()
				// Safety check
				if selectedIdx >= 0 && selectedIdx < len(m.entries) {
					// The table rows correspond 1:1 to m.entries
					target := m.entries[selectedIdx]
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
		case "enter":
			// Optional: maybe show details? For now do nothing
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
		rows := []table.Row{}
		for _, e := range m.entries {
			stateIcon := "○"
			if strings.Contains(e.State, "LISTEN") {
				stateIcon = "●" // Green-ish usually, but we rely on text for now
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

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress 'q' to quit", m.err)
	}

	logo := logoStyle.Render("⚡ LazyPorts")
	tableView := baseStyle.Render(m.table.View())

	// Footer
	controls := "↑/↓: Navigate • k: Kill • r: Refresh • q: Quit"
	if m.status != "" {
		controls = statusStyle.Render(m.status) + " • " + controls
	}

	// Use explicit newlines instead of JoinVertical to ensure rendering
	return fmt.Sprintf("%s\n%s\n%s", logo, tableView, helpStyle.Render(controls))
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

	m := model{table: t}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
