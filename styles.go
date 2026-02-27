package main

import "github.com/charmbracelet/lipgloss"

var (
	// -------------------------------------------------------------------------
	// Colours
	// -------------------------------------------------------------------------
	colorAccent   = lipgloss.Color("39")  // bright cyan
	colorSelected = lipgloss.Color("63")  // medium purple
	colorHighCPU  = lipgloss.Color("196") // bright red
	colorMuted    = lipgloss.Color("241") // dark grey
	colorGreen    = lipgloss.Color("82")  // bright green
	colorWhite    = lipgloss.Color("255")
	colorBg       = lipgloss.Color("235") // header background

	// -------------------------------------------------------------------------
	// Header panel
	// -------------------------------------------------------------------------
	styleHeader = lipgloss.NewStyle().
			Background(colorBg).
			Foreground(colorWhite).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	styleHeaderLabel = lipgloss.NewStyle().
				Foreground(colorAccent).
				Background(colorBg).
				Bold(true)

	styleHeaderValue = lipgloss.NewStyle().
				Foreground(colorWhite).
				Background(colorBg)

	// -------------------------------------------------------------------------
	// Table header row
	// -------------------------------------------------------------------------
	styleColHeader = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	styleColHeaderSelected = lipgloss.NewStyle().
				Foreground(colorBg).
				Background(colorAccent).
				Bold(true)

	// -------------------------------------------------------------------------
	// Table rows
	// -------------------------------------------------------------------------
	styleRowNormal = lipgloss.NewStyle().
			Foreground(colorWhite)

	styleRowSelected = lipgloss.NewStyle().
				Foreground(colorWhite).
				Background(colorSelected).
				Bold(true)

	styleRowHighCPU = lipgloss.NewStyle().
			Foreground(colorHighCPU)

	styleRowHighCPUSelected = lipgloss.NewStyle().
				Foreground(colorHighCPU).
				Background(colorSelected).
				Bold(true)

	// Cursor indicator (▶ / space)
	styleCursor = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	// -------------------------------------------------------------------------
	// Filter bar
	// -------------------------------------------------------------------------
	styleFilterLabel = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	styleFilterHint = lipgloss.NewStyle().
			Foreground(colorMuted)

	// -------------------------------------------------------------------------
	// Kill overlay border + text
	// -------------------------------------------------------------------------
	styleOverlayBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorHighCPU).
				Padding(1, 3)

	styleOverlayTitle = lipgloss.NewStyle().
				Foreground(colorHighCPU).
				Bold(true)

	styleOverlayHint = lipgloss.NewStyle().
				Foreground(colorMuted)

	// -------------------------------------------------------------------------
	// Status bar (bottom)
	// -------------------------------------------------------------------------
	styleStatusBar = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleStatusError = lipgloss.NewStyle().
				Foreground(colorHighCPU)

	// -------------------------------------------------------------------------
	// Help screen
	// -------------------------------------------------------------------------
	styleHelpTitle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	styleHelpSection = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	styleHelpKey = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	styleHelpDesc = lipgloss.NewStyle().
			Foreground(colorWhite)

	// -------------------------------------------------------------------------
	// Border / separator
	// -------------------------------------------------------------------------
	styleBorder = lipgloss.NewStyle().
			Foreground(colorMuted)
)
