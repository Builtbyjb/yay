package main

import (
	"fmt"

	"github.com/Builtbyjb/yay/pkg/libyay"
)

const VERSION = "0.1.0"

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
