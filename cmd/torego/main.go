package main

import (
	"fmt"
	"os"

	"github.com/d7omdev/torego/internal/core"
	"github.com/d7omdev/torego/internal/storage"
	"github.com/spf13/cobra"
)

func main() {
	HOME := getHomeDir()

	// Root command setup
	rootCmd := &cobra.Command{
		Use:   "torego",
		Short: "Torego - a lightweight reminder and notification tool",
		Long:  "Torego is a lightweight reminder and notification tool\n\nRunning 'torego' without any arguments will show active notifications.",
		Run: func(cmd *cobra.Command, args []string) {
			if !storage.IsDBInitialized(HOME) {
				fmt.Fprintln(os.Stderr, "\033[31m[FATAL] Database is not initialized. Please run 'torego init' to initialize the database.\033[0m")
				return
			}
			db, err := storage.GetDB(HOME)
			if err != nil {
				fmt.Println("Error getting database connection:", err)
				return
			}
			defer db.Close()

			// Show active notifications
			err = core.ShowActiveNotifications()
			if err != nil {
				fmt.Println("Error showing active notifications:", err)
			}
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Torego",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Torego v0.1")
		},
	}

	for _, cmd := range []*cobra.Command{initCmd, versionCmd, checkoutCmd, dismissCmd, serveCmd, notifyCmd, remindCmd, forgetCmd, listCmd} {
		cmd.DisableFlagsInUseLine = true
	}

	// Add commands to root
	rootCmd.AddCommand(initCmd, versionCmd, checkoutCmd, dismissCmd, serveCmd, notifyCmd, remindCmd, forgetCmd, listCmd)

	// Execute root command or show notifications if no args
	if len(os.Args) == 1 {
		rootCmd.Run(rootCmd, os.Args[1:])
	} else {
		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
