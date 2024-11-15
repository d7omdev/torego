package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/d7omdev/torego/internal/core"
	"github.com/d7omdev/torego/internal/storage"
	"github.com/urfave/cli/v2"
)

func getHomeDir() string {
	HOME := os.Getenv("HOME")
	if HOME == "" {
		fmt.Println("No $HOME environment variable is setup. We need it!")
		return ""
	}
	return HOME
}

var remindCmd = &cli.Command{
	Name:      "remind",
	Aliases:   []string{"r"},
	Usage:     "Set a reminder",
	UsageText: "torego remind <title> [period]",

	Description: `Set a reminder with an optional period.

The title is the text description of the reminder.

The period specifies how often the reminder should trigger.
It can be one of the following:
- "daily"
- "weekly"
- "monthly"
- "annually"
- A custom interval like "2d", "3w", "4m", "5y"

If not provided, the period defaults to "daily".`,

	Action: func(c *cli.Context) error {
		if c.NArg() < 1 {
			fmt.Println("Usage: torego remind <title> [period]")
			return nil
		}
		title := c.Args().Get(0)
		period := ""
		if c.NArg() > 1 {
			period = c.Args().Get(1)
		}
		err := storage.CreateNewReminder(title, period)
		if err != nil {
			fmt.Println("Error creating reminder:", err)
		} else {
			fmt.Println("Reminder set!")
		}
		return nil
	},
}

var forgetCmd = &cli.Command{
	Name:      "forget",
	Aliases:   []string{"f"},
	Usage:     "Forget a reminder",
	UsageText: "torego forget <index>",
	Action: func(c *cli.Context) error {
		if c.NArg() < 1 {
			fmt.Println("Usage: torego forget <index>")
			return nil
		}
		id, err := strconv.Atoi(c.Args().Get(0))
		if err != nil {
			fmt.Println("Invalid index format.")
			return nil
		}
		err = core.DeleteReminder(id)
		if err != nil {
			fmt.Println("Error deleting reminder:", err)
		} else {
			fmt.Println("Reminder forgotten!")
		}
		return nil
	},
}

var listCmd = &cli.Command{
	Name:    "list",
	Aliases: []string{"l"},
	Usage:   "List all reminders with more bloat",
	Action: func(c *cli.Context) error {
		_, err := storage.GetDB()
		if err != nil {
			fmt.Println("Error getting database connection:", err)
			return nil
		}

		reminders, err := core.ListReminders()
		if err != nil {
			fmt.Println("Error listing reminders:", err)
			return nil
		}

		if len(reminders) == 0 {
			fmt.Println("No reminders found.")
			fmt.Println("Use 'torego remind <title> [period] to set a reminder.")
			return nil
		}

		// Color variables
		blue := "\033[1;34m"
		green := "\033[1;32m"
		red := "\033[1;31m"
		reset := "\033[0m"

		// Header
		fmt.Printf("%s╭─────────────────────────────────────────────────────────────────────╮%s\n", blue, reset)
		fmt.Printf("%s| %-4s | %-4s | %-20s | %-18s | %-9s |\n", blue, "#", "ID ", "Title", "ScheduledAt", "Period")
		fmt.Printf("├──────┼──────┼──────────────────────┼────────────────────┼───────────┤\n")

		// Rows
		for i, reminder := range reminders {
			titleLines := wrapText(reminder.Title, 20)
			for j, line := range titleLines {
				if j == 0 {
					fmt.Printf("| %s%-4d%s | %s%-4d%s | %s%-20s%s | %s%-18s%s | %s%-9s%s |\n",
						red, i+1, blue,
						green, reminder.ID, blue,
						green, line, blue,
						green, reminder.ScheduledAt, blue,
						green, reminder.Period, blue,
					)
				} else {
					fmt.Printf("| %s%-4s%s | %s%-4s%s | %s%-20s%s | %s%-18s%s | %s%-9s%s |\n",
						red, "", blue,
						green, "", blue,
						green, line, blue,
						green, "", blue,
						green, "", blue,
					)
				}
			}
		}

		// Footer
		fmt.Printf("%s╰─────────────────────────────────────────────────────────────────────╯%s\n", blue, reset)
		return nil
	},
}

func wrapText(text string, width int) []string {
	words := strings.Fields(text)
	var lines []string
	var currentLine string

	for _, word := range words {
		for len(word) > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = ""
			}
			lines = append(lines, word[:width])
			word = word[width:]
		}

		if len(currentLine)+len(word)+1 > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		} else {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

var notifyCmd = &cli.Command{
	Name:    "notify",
	Aliases: []string{"n"},
	Usage:   "Notify all reminders",
	Action: func(c *cli.Context) error {
		reminders, err := core.GetAllReminders()
		if err != nil {
			fmt.Println("Error notifying reminders:", err)
		} else {
			for _, reminder := range reminders.Items {
				fmt.Println(reminder.Title)
			}
		}
		return nil
	},
}

var dismissCmd = &cli.Command{
	Name:      "dismiss",
	Aliases:   []string{"d"},
	Usage:     "Dismiss a reminder",
	UsageText: "torego dismiss <index>",
	Action: func(c *cli.Context) error {
		if c.NArg() < 1 {
			fmt.Println("Usage: torego dismiss <index>")
			return nil
		}
		id, err := strconv.Atoi(c.Args().Get(0))
		if err != nil {
			fmt.Println("Invalid index format.")
			return nil
		}
		err = core.DeleteReminder(id)
		if err != nil {
			fmt.Println("Error dismissing reminder:", err)
		} else {
			fmt.Println("Reminder dismissed!")
		}
		return nil
	},
}

var checkoutCmd = &cli.Command{
	Name:    "checkout",
	Aliases: []string{"c"},
	Usage:   "Checkout a reminder",
	Action: func(c *cli.Context) error {
		err := core.CheckoutReminder()
		if err != nil {
			fmt.Println("Error checking out reminders:", err)
		} else {
			fmt.Println("Reminders checked out!")
		}
		return nil
	},
}

var serveCmd = &cli.Command{
	Name:    "serve",
	Aliases: []string{"s"},
	Usage:   "Start the Torego server",
	Action: func(c *cli.Context) error {
		// Placeholder for actual server logic
		return nil
	},
}

var initCmd = &cli.Command{
	Name:  "init",
	Usage: "Initialize the Torego database",
	Action: func(c *cli.Context) error {
		HOME := getHomeDir()
		storage.InitDB(HOME)
		return nil
	},
}

var versionCmd = &cli.Command{
	Name:    "version",
	Aliases: []string{"v"},
	Usage:   "Print the version of Torego",
	Action: func(c *cli.Context) error {
		fmt.Println("Torego v0.1.0")
		return nil
	},
}
