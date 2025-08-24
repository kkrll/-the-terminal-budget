package main

import (
	"fmt"
	"os"

	"github.com/kkrll/the-terminal-budget/tui"
)

func main() {
	runTUI()
}

func runTUI() {
	if err := tui.RunTUI(); err != nil {
		fmt.Println("Error running TUI:", err)
		os.Exit(1)
	}
}
