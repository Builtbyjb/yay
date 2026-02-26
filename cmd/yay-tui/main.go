package main

import (
	"fmt"
	"os"

	"github.com/Builtbyjb/yay/pkg/libyay"
)

const VERSION = "0.1.0"

func main() {
	settings, err := libyay.Fetch()
	if err != nil {
		fmt.Println("Error occurred while fetching applications:", err)
		os.Exit(1)
	}

	if settings == nil {
		fmt.Println("No applications found.")
		os.Exit(0)
	}

	changes, err := RunTUI(settings, VERSION)
	if err != nil {
		fmt.Println("Error running TUI:", err)
		os.Exit(1)
	}

	PrintChanges(changes)
}
