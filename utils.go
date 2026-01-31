package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
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
			isRoot := os.Geteuid() == 0
			suffix := "(requires sudo)"
			if isRoot {
				suffix = "(system)"
			}

			if service, ok := commonPorts[port]; ok {
				process = fmt.Sprintf("%s %s", service, suffix)
			} else {
				process = suffix
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

func getProcessDetails(pid string) (string, error) {
	if pid == "-" {
		if os.Geteuid() == 0 {
			return "System process (no detailed information available).", nil
		}
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
