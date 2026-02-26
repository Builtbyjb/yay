package main

import (
	"fmt"

	"github.com/Builtbyjb/yay/pkg/libyay"
)

/*
Todo check list
- [] Fetch and print out all download/installed applications
- [] Create an sqlite3 db on start up if no exist
- [] Create table schema
- [] Figure out tui interface to specify apps
	- [] A user needs to be able to search for a application
	- [] A user needs to be able to select the application and assign a keyboard shortcut
	- [] A user needs to be remove a keyboard shortcut or disable it
	- [] A user needs to be able to map a keyboard hotkey to open dock apps in order
*/

func main() {
	app := libyay.Fetch()
	if app == nil {
		fmt.Println("No applications found.")
		return
	}

	fmt.Println("Applications fetched successfully.")
	fmt.Println("Number of applications found:", len(app))
	fmt.Println("Applications:", app)
}
