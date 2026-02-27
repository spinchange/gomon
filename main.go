package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	noColor    := flag.Bool("no-color", false, "disable ANSI colour output")
	screenshot := flag.Bool("screenshot", false, "render one frame to stdout and exit")
	scHelp     := flag.Bool("screenshot-help", false, "render the help screen to stdout and exit")
	scWidth    := flag.Int("w", 120, "terminal width for --screenshot")
	scHeight   := flag.Int("h", 35, "terminal height for --screenshot")
	flag.Parse()

	if *noColor {
		os.Setenv("NO_COLOR", "1")
	}

	if *scHelp {
		runScreenshot(*scWidth, *scHeight, ModeHelp)
		return
	}
	if *screenshot {
		runScreenshot(*scWidth, *scHeight, ModeNormal)
		return
	}

	m := NewModel()
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "gomon: %v\n", err)
		os.Exit(1)
	}
}
