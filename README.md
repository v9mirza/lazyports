# LazyPorts

> **Visual port manager for Linux**

![License](https://img.shields.io/badge/license-MIT-blue.svg) ![Go](https://img.shields.io/badge/Go-1.24+-00ADD8.svg)

`lazyports` is a terminal UI tool to visualize and manage network ports. It provides an interactive table to inspect listening processes and kill them easily.

![LazyPorts Demo](assets/image.png)

Built with Bubble Tea and Lipgloss.

## Prerequisites

-   **Go 1.24** or higher (required for installation)

## Installation

Install with a single command:

```bash
curl -sL https://raw.githubusercontent.com/v9mirza/lazyports/main/install.sh | bash
```

*Note: The script may request `sudo` permission to install the binary to `/usr/local/bin` for system-wide access.*

## Uninstallation

To remove `lazyports` from your system, simply delete the binary:

```bash
# Remove system-wide installation
sudo rm /usr/local/bin/lazyports

# Remove user-local installation
rm ~/go/bin/lazyports
```

## Usage

Run the tool from your terminal:

```bash
lazyports
```

### Controls

| Key | Action |
| :--- | :--- |
| `↑` / `↓` | Navigate list |
| `/` | **Filter / Search** ports |
| `Enter` | **View Details** (User, Path, Time) |
| `k` | Kill selected process |
| `r` | Refresh list |
| `q` | Quit |

## Features

-   **Interactive Table**: Clean visualization of open ports (TCP/UDP).
-   **Smart Filtering**: Type `/` to instantly filter by port, PID, or process name.
-   **Detailed Inspection**: Press `Enter` to see full command, user, and start time.
-   **Process Management**: Terminate blocking processes instantly.
-   **Smart Detection**: Identifies system services (like SSH/DNS) and their ports even without sudo.
-   **Auto-Sorting**: Ports are automatically sorted numerically for easy scanning.
-   **Visual States**: Distinct indicators for `LISTEN` and `ESTAB` connections.
-   **Zero Config**: Works out of the box with automatic shell path configuration.

## License

MIT
