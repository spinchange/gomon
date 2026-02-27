package main

import (
	"fmt"
	"time"
)

// runScreenshot renders one frame of the TUI to stdout (no bubbletea needed).
// Collects real process data (two ticks so CPU% is populated), prints View().
func runScreenshot(width, height int, mode AppMode) {
	m := NewModel()
	m.termWidth = width
	m.termHeight = height

	// Seed sysStats
	stats := CollectSysStats()
	m.sysStats = stats

	// Tick 1 — seeds gopsutil CPU baseline (values will be 0)
	CollectProcesses()

	// Wait one second so tick 2 produces real CPU deltas
	fmt.Println("Collecting process data (1s)...")
	time.Sleep(1100 * time.Millisecond)

	// Tick 2 — real CPU%
	result := CollectProcesses()
	m.allProcs = result.Procs
	m.applyFilterAndSort()
	m.clampCursor()
	m.mode = mode

	fmt.Print("\033[2J\033[H") // clear screen
	fmt.Print(m.View())
}
