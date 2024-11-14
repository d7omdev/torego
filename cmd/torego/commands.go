package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/d7omdev/torego/internal/core"
	"github.com/d7omdev/torego/internal/storage"
	"github.com/spf13/cobra"
)

func getHomeDir() string {
	HOME := os.Getenv("HOME")
	if HOME == "" {
		fmt.Println("No $HOME environment variable is setup. We need it!")
		return ""
	}
	return HOME
}

var remindCmd = &cobra.Command{
	Use:     "remind <title> [period]",
	Short:   "Set a reminder",
	Aliases: []string{"r"},
	Long: `Set a reminder with a title and an optional period.

The period can be one of the following:
- "daily"
- "weekly"
- "monthly"
- "annually"
- A custom interval like "2d", "3w", "4m", "5y"

If not provided, the period defaults to "daily".`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Usage: torego remind <title> [period]")
			return
		}
		title := args[0]
		period := ""
		if len(args) > 1 {
			period = args[1]
		}
		err := storage.CreateNewReminder(title, period)
		if err != nil {
			fmt.Println("Error creating reminder:", err)
		} else {
			fmt.Println("Reminder set!")
		}
	},
}

var forgetCmd = &cobra.Command{
	Use:     "forget",
	Short:   "Forget a reminder",
	Aliases: []string{"f"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Usage: torego forget <id>")
			return
		}
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid ID format.")
			return
		}
		err = core.DeleteReminder(id)
		if err != nil {
			fmt.Println("Error deleting reminder:", err)
		} else {
			fmt.Println("Reminder forgotten!")
		}
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all reminders with more bloat",
	Aliases: []string{"l"},
	Run: func(cmd *cobra.Command, args []string) {
		HOME := getHomeDir()
		_, err := storage.GetDB(HOME)
		if err != nil {
			fmt.Println("Error getting database connection:", err)
			return
		}

		reminders, err := core.ListReminders()
		if err != nil {
			fmt.Println("Error listing reminders:", err)
			return
		}
		// Sort reminders by ID
		sort.Slice(reminders, func(i, j int) bool {
			return reminders[i].ID < reminders[j].ID
		})

		if len(reminders) == 0 {
			fmt.Println("No reminders found.")
			fmt.Println("Use 'torego remind <title> [period] to set a reminder.")
			return
		}

		// Is this less bloat or more bloat?
		// at least im not using extra packages
		// right?

		// Color variables
		blue := "\033[1;34m"
		green := "\033[1;32m"
		reset := "\033[0m"

		// Header
		fmt.Printf("%s╭───────────────────────────────────────────────────────────────╮%s\n", blue, reset)
		fmt.Printf("%s| %-4s | %-21s | %-18s | %-9s |\n", blue, "ID ", "Title", "ScheduledAt", "Period")
		fmt.Printf("├──────┼───────────────────────┼────────────────────┼───────────┤\n")

		// Rows
		for _, reminder := range reminders {
			fmt.Printf("| %s%-4d%s | %s%-21s%s | %s%-18s%s | %s%-9s%s |\n",
				green, reminder.ID, blue,
				green, reminder.Title, blue,
				green, reminder.ScheduledAt, blue,
				green, reminder.Period, blue,
			)
		}

		// Footer
		fmt.Printf("%s╰───────────────────────────────────────────────────────────────╯%s\n", blue, reset)
	},
}

var notifyCmd = &cobra.Command{
	Use:     "notify",
	Short:   "Notify all reminders",
	Aliases: []string{"n"},
	Run: func(cmd *cobra.Command, args []string) {
		reminders, err := core.GetAllReminders()
		if err != nil {
			fmt.Println("Error notifying reminders:", err)
		} else {
			for _, reminder := range reminders.Items {
				fmt.Println(reminder.Title)
			}
		}
	},
}

var dismissCmd = &cobra.Command{
	Use:     "dismiss",
	Short:   "Dismiss a reminder",
	Aliases: []string{"d"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Usage: torego dismiss <id>")
			return
		}
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid ID format.")
			return
		}
		err = core.DeleteReminder(id)
		if err != nil {
			fmt.Println("Error dismissing reminder:", err)
		} else {
			fmt.Println("Reminder dismissed!")
		}
	},
}

var checkoutCmd = &cobra.Command{
	Use:     "checkout",
	Short:   "Checkout a reminder",
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {
		err := core.CheckoutReminder()
		if err != nil {
			fmt.Println("Error checking out reminders:", err)
		} else {
			fmt.Println("Reminders checked out!")
		}
	},
}

var serveCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Start the Torego server",
	Aliases: []string{"s"},
	Run:     func(cmd *cobra.Command, args []string) {},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the Torego database",
	Run: func(cmd *cobra.Command, args []string) {
		HOME := getHomeDir()

		storage.InitDB(HOME)
	},
}
