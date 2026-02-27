package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v3/process"
)

// ---------------------------------------------------------------------------
// Column layout constants
// ---------------------------------------------------------------------------

const (
	colPID    = 7
	colCPU    = 8
	colMem    = 10
	colStatus = 8
	colUser   = 12
	// separator chars between columns: " │ " = 3 chars each, 5 gaps total
	colSeparators  = 5 * 3
	colFixed       = colPID + colCPU + colMem + colStatus + colUser + colSeparators
	colNameMin     = 10
	colNameMax     = 40
	tickInterval   = time.Second
	highCPUThresh  = 50.0
)

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

type Model struct {
	allProcs    []ProcessRow
	visibleProc []ProcessRow

	sysStats sysStatsMsg

	cursor    int
	scrollOff int
	termWidth int
	termHeight int

	sortCol SortColumn
	sortAsc bool

	mode        AppMode
	filterInput textinput.Model
	filterText  string

	killTarget *ProcessRow
	statusMsg  string // ephemeral message in status bar
	err        error
}

// ---------------------------------------------------------------------------
// Init
// ---------------------------------------------------------------------------

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "type to filter"
	ti.CharLimit = 64
	ti.Width = 30

	return Model{
		termWidth:  120,
		termHeight: 30,
		sortCol:    SortCPU,
		sortAsc:    false, // CPU descending by default
		filterInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		fetchSysStats(),
		fetchProcesses(),
	)
}

// ---------------------------------------------------------------------------
// Commands
// ---------------------------------------------------------------------------

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchSysStats() tea.Cmd {
	return func() tea.Msg {
		return CollectSysStats()
	}
}

func fetchProcesses() tea.Cmd {
	return func() tea.Msg {
		return CollectProcesses()
	}
}

func killProcess(pid int32) tea.Cmd {
	return func() tea.Msg {
		p, err := process.NewProcess(pid)
		if err != nil {
			return killResultMsg{PID: pid, Err: err}
		}
		err = p.Kill()
		return killResultMsg{PID: pid, Err: err}
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		return m, nil

	case tickMsg:
		return m, tea.Batch(tickCmd(), fetchSysStats(), fetchProcesses())

	case sysStatsMsg:
		m.sysStats = msg
		if msg.Err != nil {
			m.err = msg.Err
		}
		return m, nil

	case processesMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.allProcs = msg.Procs
		m.applyFilterAndSort()
		m.clampCursor()
		return m, nil

	case killResultMsg:
		m.mode = ModeNormal
		m.killTarget = nil
		if msg.Err != nil {
			m.statusMsg = fmt.Sprintf("kill PID %d failed: %v", msg.PID, msg.Err)
		} else {
			m.statusMsg = fmt.Sprintf("killed PID %d", msg.PID)
		}
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case ModeNormal:
			return m.handleNormalKey(msg)
		case ModeFilter:
			return m.handleFilterKey(msg)
		case ModeConfirmKill:
			return m.handleConfirmKey(msg)
		case ModeHelp:
			return m.handleHelpKey(msg)
		}
	}

	return m, nil
}

// ---------------------------------------------------------------------------
// Key handlers
// ---------------------------------------------------------------------------

func (m Model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.statusMsg = "" // clear ephemeral status

	switch msg.String() {
	case keyQuit, "ctrl+c":
		return m, tea.Quit

	case keyUp, keyVimUp:
		m.moveCursor(-1)

	case keyDown, keyVimDown:
		m.moveCursor(1)

	case keyFilter:
		m.mode = ModeFilter
		m.filterInput.Focus()
		return m, textinput.Blink

	case keyEsc:
		if m.filterText != "" {
			m.filterText = ""
			m.filterInput.SetValue("")
			m.applyFilterAndSort()
			m.clampCursor()
		}

	case keyTab:
		// Cycle sort column forward; each column gets a sensible default direction
		next := SortColumn((int(m.sortCol) + 1) % 6)
		m.sortCol = next
		m.sortAsc = (next == SortPID || next == SortName || next == SortUser)
		m.applyFilterAndSort()
		m.clampCursor()

	case keySortPID:
		m.toggleSort(SortPID)
	case keySortName:
		m.toggleSort(SortName)
	case keySortCPU:
		m.toggleSort(SortCPU)
	case keySortMem:
		m.toggleSort(SortMem)
	case keySortStatus:
		m.toggleSort(SortThreads)
	case keySortUser:
		m.toggleSort(SortUser)

	case keyDel, keyKill:
		if len(m.visibleProc) > 0 {
			target := m.visibleProc[m.cursor]
			m.killTarget = &target
			m.mode = ModeConfirmKill
		}

	case keyHelp:
		m.mode = ModeHelp
	}

	return m, nil
}

func (m Model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyHelp, keyEsc, keyQuit:
		m.mode = ModeNormal
	}
	return m, nil
}

func (m Model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		m.mode = ModeNormal
		m.filterText = ""
		m.filterInput.SetValue("")
		m.filterInput.Blur()
		m.applyFilterAndSort()
		m.clampCursor()
		return m, nil

	case keyEnter:
		m.filterText = m.filterInput.Value()
		m.mode = ModeNormal
		m.filterInput.Blur()
		m.applyFilterAndSort()
		m.clampCursor()
		return m, nil
	}

	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	m.filterText = m.filterInput.Value()
	m.applyFilterAndSort()
	m.clampCursor()
	return m, cmd
}

func (m Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyConfirmY, keyEnter:
		if m.killTarget != nil {
			pid := m.killTarget.PID
			return m, killProcess(pid)
		}
		m.mode = ModeNormal

	case keyConfirmN, keyEsc:
		m.mode = ModeNormal
		m.killTarget = nil
	}

	return m, nil
}

// ---------------------------------------------------------------------------
// Sort & Filter
// ---------------------------------------------------------------------------

func (m *Model) toggleSort(col SortColumn) {
	if m.sortCol == col {
		m.sortAsc = !m.sortAsc
	} else {
		m.sortCol = col
		// Numeric columns default descending (highest first); text columns ascending
		m.sortAsc = (col == SortPID || col == SortName || col == SortUser)
	}
	m.applyFilterAndSort()
	m.clampCursor()
}

func (m *Model) applyFilterAndSort() {
	filter := strings.ToLower(m.filterText)

	// 1. Filter
	filtered := make([]ProcessRow, 0, len(m.allProcs))
	for _, p := range m.allProcs {
		if filter == "" || strings.Contains(strings.ToLower(p.Name), filter) {
			filtered = append(filtered, p)
		}
	}

	// 2. Sort (stable to avoid jumpiness on equal values)
	sort.SliceStable(filtered, func(i, j int) bool {
		return m.compareRows(filtered[i], filtered[j])
	})

	m.visibleProc = filtered
}

func (m *Model) compareRows(a, b ProcessRow) bool {
	var less bool
	switch m.sortCol {
	case SortPID:
		less = a.PID < b.PID
	case SortName:
		less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
	case SortCPU:
		less = a.CPU < b.CPU
	case SortMem:
		less = a.MemMB < b.MemMB
	case SortThreads:
		less = a.Threads < b.Threads
	case SortUser:
		less = strings.ToLower(a.User) < strings.ToLower(b.User)
	}
	if m.sortAsc {
		return less
	}
	return !less
}

// ---------------------------------------------------------------------------
// Cursor & Scroll
// ---------------------------------------------------------------------------

func (m *Model) moveCursor(delta int) {
	if len(m.visibleProc) == 0 {
		return
	}
	m.cursor += delta
	m.clampCursor()
	m.adjustScroll()
}

func (m *Model) clampCursor() {
	if len(m.visibleProc) == 0 {
		m.cursor = 0
		m.scrollOff = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.visibleProc) {
		m.cursor = len(m.visibleProc) - 1
	}
	m.adjustScroll()
}

func (m *Model) adjustScroll() {
	h := m.tableHeight()
	if m.cursor < m.scrollOff {
		m.scrollOff = m.cursor
	}
	if m.cursor >= m.scrollOff+h {
		m.scrollOff = m.cursor - h + 1
	}
}

// tableHeight returns number of data rows visible.
func (m *Model) tableHeight() int {
	// total: header(1) + colHeader(1) + separator(1) + table rows + filter(1) + sep(1) + status(1) = 6 fixed rows
	// + 2 for top/bottom border rows in the full layout
	reserved := 8
	h := m.termHeight - reserved
	if h < 1 {
		h = 1
	}
	return h
}

// nameColWidth computes the dynamic Name column width.
func (m *Model) nameColWidth() int {
	w := m.termWidth - colFixed - 2 // 2 for left margin
	if w < colNameMin {
		w = colNameMin
	}
	if w > colNameMax {
		w = colNameMax
	}
	return w
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (m Model) View() string {
	if m.mode == ModeHelp {
		return m.renderHelpScreen()
	}

	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n")
	b.WriteString(m.renderSeparator())
	b.WriteString("\n")
	b.WriteString(m.renderColHeader())
	b.WriteString("\n")
	b.WriteString(m.renderSeparator())
	b.WriteString("\n")
	b.WriteString(m.renderTable())
	b.WriteString(m.renderSeparator())
	b.WriteString("\n")
	b.WriteString(m.renderFilterBar())
	b.WriteString("\n")
	b.WriteString(m.renderSeparator())
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	out := b.String()

	if m.mode == ModeConfirmKill && m.killTarget != nil {
		out = m.renderKillOverlay(out)
	}

	return out
}

// ---------------------------------------------------------------------------
// Render helpers
// ---------------------------------------------------------------------------

func (m *Model) renderHeader() string {
	mem := fmt.Sprintf("%.1f / %.1f GB", m.sysStats.MemUsed, m.sysStats.MemTotal)
	line := fmt.Sprintf(
		"%s   host: %s   uptime: %s   RAM: %s",
		styleHeaderLabel.Render("gomon"),
		styleHeaderValue.Render(m.sysStats.Hostname),
		styleHeaderValue.Render(m.sysStats.Uptime),
		styleHeaderValue.Render(mem),
	)
	return styleHeader.Width(m.termWidth).Render(line)
}

func (m *Model) renderSeparator() string {
	return styleBorder.Render(strings.Repeat("─", m.termWidth))
}

func (m *Model) renderColHeader() string {
	nameW := m.nameColWidth()
	cols := []struct {
		col   SortColumn
		label string
		width int
		right bool
	}{
		{SortPID, "PID", colPID, true},
		{SortName, "NAME", nameW, false},
		{SortCPU, "CPU%", colCPU, true},
		{SortMem, "MEM(MB)", colMem, true},
		{SortThreads, "THRD", colStatus, true},
		{SortUser, "USER", colUser, false},
	}

	var parts []string
	for _, c := range cols {
		indicator := " "
		if m.sortCol == c.col {
			if m.sortAsc {
				indicator = "▲"
			} else {
				indicator = "▼"
			}
		}
		label := c.label + indicator
		var cell string
		if c.right {
			cell = padLeft(label, c.width)
		} else {
			cell = padRight(label, c.width)
		}

		if m.sortCol == c.col {
			parts = append(parts, styleColHeaderSelected.Render(cell))
		} else {
			parts = append(parts, styleColHeader.Render(cell))
		}
	}

	return " " + strings.Join(parts, styleBorder.Render(" │ "))
}

func (m *Model) renderTable() string {
	var b strings.Builder
	h := m.tableHeight()
	nameW := m.nameColWidth()

	for i := 0; i < h; i++ {
		idx := m.scrollOff + i
		if idx >= len(m.visibleProc) {
			b.WriteString("\n")
			continue
		}
		row := m.visibleProc[idx]
		selected := idx == m.cursor

		// cursor glyph
		cursor := " "
		if selected {
			cursor = styleCursor.Render("▶")
		}

		// cells
		pid := padLeft(fmt.Sprintf("%d", row.PID), colPID)
		name := truncate(row.Name, nameW)
		name = padRight(name, nameW)
		cpu := padLeft(fmt.Sprintf("%.2f", row.CPU), colCPU)
		memStr := padLeft(fmt.Sprintf("%.1f", row.MemMB), colMem)
		threads := padLeft(fmt.Sprintf("%d", row.Threads), colStatus)
		user := padRight(row.User, colUser)

		line := cursor + pid + styleBorder.Render(" │ ") +
			name + styleBorder.Render(" │ ") +
			cpu + styleBorder.Render(" │ ") +
			memStr + styleBorder.Render(" │ ") +
			threads + styleBorder.Render(" │ ") +
			user

		highCPU := row.CPU >= highCPUThresh

		if selected {
			if highCPU {
				b.WriteString(styleRowHighCPUSelected.Width(m.termWidth).Render(line))
			} else {
				b.WriteString(styleRowSelected.Width(m.termWidth).Render(line))
			}
		} else {
			if highCPU {
				b.WriteString(styleRowHighCPU.Render(line))
			} else {
				b.WriteString(styleRowNormal.Render(line))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m *Model) renderFilterBar() string {
	label := styleFilterLabel.Render("  Filter: ")
	hint := styleFilterHint.Render("   Esc clear · Enter confirm")

	var inputView string
	if m.mode == ModeFilter {
		inputView = "[" + m.filterInput.View() + "]"
	} else if m.filterText != "" {
		inputView = "[" + m.filterText + "]"
	} else {
		inputView = "[          ]"
	}

	return label + inputView + hint
}

func (m *Model) renderStatusBar() string {
	if m.statusMsg != "" {
		return styleStatusError.Render("  " + m.statusMsg)
	}
	if m.err != nil {
		return styleStatusError.Render("  Error: " + m.err.Error())
	}
	return styleStatusBar.Render("  " + helpText)
}

func (m *Model) renderKillOverlay(base string) string {
	if m.killTarget == nil {
		return base
	}

	windowsNote := ""
	if runtime.GOOS == "windows" {
		windowsNote = "\n  " + styleOverlayHint.Render("(Windows: force-terminate, no SIGTERM)")
	}

	content := styleOverlayTitle.Render("Kill Process?") + "\n\n" +
		fmt.Sprintf("  Kill PID %d (%s)\n", m.killTarget.PID, m.killTarget.Name) +
		fmt.Sprintf("  owned by: %s", m.killTarget.User) +
		windowsNote + "\n\n" +
		"  " + styleOverlayHint.Render("Press y to confirm, n or Esc to cancel")

	box := styleOverlayBorder.Render(content)

	// Centre the overlay on screen using lipgloss.Place
	return lipgloss.Place(
		m.termWidth, m.termHeight,
		lipgloss.Center, lipgloss.Center,
		box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}

// ---------------------------------------------------------------------------
// Help screen
// ---------------------------------------------------------------------------

func (m *Model) renderHelpScreen() string {
	var b strings.Builder
	indent := "  "

	b.WriteString(m.renderHeader())
	b.WriteString("\n")
	b.WriteString(m.renderSeparator())
	b.WriteString("\n")
	b.WriteString(styleHelpTitle.Render(indent + "gomon — keyboard reference"))
	b.WriteString("\n")

	type row struct{ key, desc string }
	section := func(title string, rows []row) {
		b.WriteString("\n")
		b.WriteString(styleHelpSection.Render(indent + title))
		b.WriteString("\n")
		for _, r := range rows {
			k := styleHelpKey.Render(padRight(r.key, 14))
			d := styleHelpDesc.Render(r.desc)
			b.WriteString(indent + "  " + k + d + "\n")
		}
	}

	section("Navigation", []row{
		{"j / ↓", "Move cursor down"},
		{"k / ↑", "Move cursor up"},
		{"q", "Quit gomon"},
		{"Ctrl+C", "Force quit"},
	})

	section("Filter", []row{
		{"/", "Enter filter mode — type to search by process name"},
		{"Esc", "Clear filter and return to normal mode"},
		{"Enter", "Confirm filter and return to normal mode"},
	})

	section("Sorting", []row{
		{"Tab", "Cycle sort column forward"},
		{"1", "Sort by PID (ascending)"},
		{"2", "Sort by Name (A→Z)"},
		{"3", "Sort by CPU%   (default — highest first)"},
		{"4", "Sort by Memory in MB (highest first)"},
		{"5", "Sort by Thread count (highest first)"},
		{"6", "Sort by User (A→Z)"},
	})

	section("Process Actions", []row{
		{"Del / K", "Kill selected process — shows confirmation dialog"},
		{"y / Enter", "Confirm kill"},
		{"n / Esc", "Cancel kill"},
	})

	section("Columns", []row{
		{"PID", "Process ID assigned by the operating system"},
		{"NAME", "Executable name (truncated with … if longer than column)"},
		{"CPU%", "CPU usage across all cores — can exceed 100% on multi-core"},
		{"", "  First tick always shows 0% — real values appear after ~1s"},
		{"MEM(MB)", "Resident set size: physical RAM in use, in megabytes"},
		{"THRD", "Number of OS threads owned by the process"},
		{"USER", "Account that owns the process (N/A if access is denied)"},
	})

	// Pad remaining lines so the status bar sits at the bottom
	built := b.String()
	lines := strings.Count(built, "\n")
	for i := lines; i < m.termHeight-2; i++ {
		b.WriteString("\n")
	}

	b.WriteString(m.renderSeparator())
	b.WriteString("\n")
	b.WriteString(styleStatusBar.Render(indent + "Press  ?  Esc  or  q  to close help"))

	return b.String()
}

// ---------------------------------------------------------------------------
// String utilities
// ---------------------------------------------------------------------------

func padRight(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

func padLeft(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return strings.Repeat(" ", width-n) + s
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}

