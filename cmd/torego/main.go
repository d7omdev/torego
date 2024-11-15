package main

import (
	"fmt"
	"os"

	"github.com/d7omdev/torego/internal/core"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "torego",
		Usage: "Torego - a lightweight reminder and notification tool",
		Commands: []*cli.Command{
			initCmd, versionCmd, checkoutCmd, dismissCmd, serveCmd, notifyCmd, remindCmd, forgetCmd, listCmd,
		},
		Action: func(c *cli.Context) error {
			core.ShowActiveNotifications()
			return nil
		},
	}

	// Run the application
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
