# gomon

A cross-platform terminal UI system monitor. View and manage running processes with real-time CPU, memory, and thread stats — right from your terminal.

![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-blue)
![Language](https://img.shields.io/badge/language-Go-00ADD8)

## Features

- **Real-time process table** — updates every second with PID, name, CPU%, memory (MB), thread count, and user
- **Multi-column sorting** — sort by any column via keyboard shortcuts or Tab to cycle
- **Process filtering** — press `/` and type to filter by process name
- **Kill processes** — press `Del` or `K` to terminate the selected process (with confirmation)
- **System header** — shows hostname, uptime, and RAM usage at a glance
- **Vim-style navigation** — `j`/`k` or arrow keys
- **Cross-platform** — Windows, Linux, macOS

## Installation

**Prerequisites:** [Go 1.21+](https://go.dev/dl/)

```bash
git clone https://github.com/spinchange/gomon
cd gomon
go build -o gomon .
./gomon
```

**Windows:**
```bat
git clone https://github.com/spinchange/gomon
cd gomon
build.bat
gomon.exe
```

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Tab` | Cycle sort column |
| `1`–`6` | Sort by PID / Name / CPU / Mem / Threads / User |
| `/` | Enter filter mode |
| `Esc` | Clear filter / cancel |
| `Del` / `K` | Kill selected process |
| `y` / `Enter` | Confirm kill |
| `?` | Toggle help screen |
| `q` / `Ctrl+C` | Quit |

## Options

```
-no-color          Disable ANSI color output (also respects NO_COLOR env var)
-screenshot        Render one frame to stdout and exit
-screenshot-help   Render the help screen to stdout and exit
-w <cols>          Terminal width for screenshot mode (default 120)
-h <rows>          Terminal height for screenshot mode (default 35)
```

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) — TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — terminal styling
- [gopsutil](https://github.com/shirou/gopsutil) — cross-platform system/process info
