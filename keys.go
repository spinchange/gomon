package main

// Keybinding string constants used in Update and the help footer.
const (
	keyQuit      = "q"
	keyUp        = "up"
	keyDown      = "down"
	keyVimUp     = "k"
	keyVimDown   = "j"
	keyFilter    = "/"
	keyEsc       = "esc"
	keyTab       = "tab"
	keyDel       = "delete"
	keyKill      = "K"
	keyConfirmY  = "y"
	keyConfirmN  = "n"
	keyEnter     = "enter"
	keySortPID   = "1"
	keySortName  = "2"
	keySortCPU   = "3"
	keySortMem   = "4"
	keySortStatus = "5"
	keySortUser  = "6"
	keyHelp      = "?"
)

// helpText is rendered in the footer status bar during Normal mode.
const helpText = "q quit  / filter  Tab sort  Del/K kill  j↓  k↑  1=PID 2=Name 3=CPU 4=Mem 5=Thrd 6=User  ? help"
