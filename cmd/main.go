package main

import (
	"fmt"
	"os"

	"github.com/Builtbyjb/yay/pkg/lib"
	"github.com/Builtbyjb/yay/pkg/tui"
)

const VERSION = "0.1.0"

func main() {
	settings, err := lib.Fetch()
	if err != nil {
		fmt.Println("Error occurred while fetching applications:", err)
		os.Exit(1)
	}

	if settings == nil {
		fmt.Println("No applications found.")
		os.Exit(0)
	}

	changes, err := tui.Run(settings, VERSION)
	if err != nil {
		fmt.Println("Error running TUI:", err)
		os.Exit(1)
	}

	tui.PrintChanges(changes)
}
