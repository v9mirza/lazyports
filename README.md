# lazyports

> **Visual port manager for Linux**

`lazyports` is a simple Terminal UI (TUI) tool to visualize which local processes are using which network ports, allowing you to easily kill them.

Built with **Go**, **Bubble Tea**, and **Lipgloss**.

## Features

- 🔍 ** visualize** open ports (TCP/UDP) and their owning processes.
- 🎯 **Kill** processes directly from the list.
- 🔄 **Auto-refresh** after killing a process.
- ⌨️ **Keyboard-driven** interface.

## Requirements

- Linux
- `ss` command (usually available by default in `iproute2` package)
- Go 1.21+ (to build/run)

## Installation

```bash
# Clone the repo
git clone https://github.com/v9mirza/lazyports.git
cd lazyports

# Install dependencies
go mod tidy
```

## Usage

Run the tool:

```bash
go run main.go
# OR build it
go build -o lazyports
./lazyports
```

## Controls

| Key | Action |
| --- | --- |
| `↑` / `↓` | Navigate the list |
| `k` | Kill the selected process (SIGTERM) |
| `r` | Refresh the list manually |
| `q` | Quit |

## License

MIT
# lazyports
