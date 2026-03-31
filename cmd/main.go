package main

import (
	"fmt"
	"os"

	"github.com/Builtbyjb/yay/pkg/lib"
	"github.com/Builtbyjb/yay/pkg/tui"
	"github.com/spf13/cobra"
)

const VERSION = "0.1.0"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yay",
	Short: "A light weight application manager",
	// Long:  "A longer description that spans multiple lines and likely contains",
	Run: func(cmd *cobra.Command, args []string) {
		db, settings, err := lib.Fetch()

		if err != nil {
			fmt.Println("Error occurred while fetching applications:", err)
			os.Exit(1)
		}
		defer db.Close()

		if settings == nil {
			fmt.Println("No applications found.")
			os.Exit(0)
		}

		if err := tui.Run(db, settings, VERSION); err != nil {
			fmt.Println("Error running TUI:", err)
			os.Exit(1)
		}
	},
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the version of the application",
	// Long:  "Display the version of the application and exit.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Yay v%s\n", VERSION)
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start background service",
	// Long:  `Start the application with the specified settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		lib.StartBackgroundService()
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run application",
	// Long:  `Start the application with the specified settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running...")

		db, err := lib.GetDatabase()
		if err != nil {
			fmt.Println("Error occurred while fetching database:", err)
			os.Exit(1)
		}
		defer db.Close()

		lib.KeyEventListener(db, nil)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check background service status",
	// Long:  `Stop the application gracefully.`,
	Run: func(cmd *cobra.Command, args []string) {
		lib.CheckBackgroundServiceStatus()
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop background service",
	// Long:  `Stop the application gracefully.`,
	Run: func(cmd *cobra.Command, args []string) {
		lib.StopBackgroundService()
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the application",
	// Long:  `Update the application to the latest version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Updating the application...")
	},
}

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Show help information",
	// Long:  `Show detailed help information about the application and its commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Help information:")
		cmd.Help()
	},
}

func main() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(helpCmd)

	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}
