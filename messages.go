package main

import "time"

// ---------------------------------------------------------------------------
// App modes
// ---------------------------------------------------------------------------

type AppMode int

const (
	ModeNormal      AppMode = iota
	ModeFilter
	ModeConfirmKill
	ModeHelp
)

// ---------------------------------------------------------------------------
// Sort columns
// ---------------------------------------------------------------------------

type SortColumn int

const (
	SortPID    SortColumn = iota // 1
	SortName                     // 2
	SortCPU                      // 3
	SortMem                      // 4
	SortThreads                  // 5
	SortUser                     // 6
)

// ---------------------------------------------------------------------------
// Process row
// ---------------------------------------------------------------------------

type ProcessRow struct {
	PID     int32
	Name    string
	CPU     float64 // percent (0–100*numCPU)
	MemMB   float64
	Threads int32
	User    string
}

// ---------------------------------------------------------------------------
// Tea messages
// ---------------------------------------------------------------------------

type tickMsg time.Time

type sysStatsMsg struct {
	Hostname string
	Uptime   string
	MemUsed  float64 // GB
	MemTotal float64 // GB
	Err      error
}

type processesMsg struct {
	Procs []ProcessRow
	Err   error
}

type killResultMsg struct {
	PID int32
	Err error
}
