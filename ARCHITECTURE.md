# Architecture & File Structure

This document outlines the purpose of each file in the `lazyports` codebase.

## Core Application

### [main.go](./main.go)
**Role:** Entry Point
-   Sets up the initial table columns.
-   Initializes the `Bubble Tea` program.
-   Starts the application loop.

### [model.go](./model.go)
**Role:** State & Logic (The Brain)
-   **Structs**: Defines `PortEntry` (data) and `model` (app state).
-   **Update()**: Handles user input (Keypresses: `k`, `r`, `/`, `Enter`).
-   **View()**: Renders the UI (Table, Filter Input, or Details Popup) based on state.

### [utils.go](./utils.go)
**Role:** System Operations (The Engine)
-   **`getPorts()`**: Executes `ss -tulnp`, parses output, and handles sudo/root detection logic.
-   **`killProcess()`**: Sends termination signals to processes.
-   **`getProcessDetails()`**: Fetches detailed process info using `ps`.

### [styles.go](./styles.go)
**Role:** UI Design
-   Contains all `lipgloss` definitions.
-   Defines colors, borders, margins, and text styles.

## Build & Install

### [install.sh](./install.sh)
**Role:** Installation Script
-   Checks for Go installation.
-   Runs `go install`.
-   Moves binary to `/usr/local/bin` for global availability.

### [go.mod](./go.mod)
**Role:** Dependency Management
-   Tracks external libraries:
    -   `bubbletea`: TUI Framework
    -   `lipgloss`: Styling
    -   `bubbles`: UI Components (Table, Text Input)
