package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <name>",
	Short: "List tasks from a specific list",
	Long: `List tasks from a named list in Things 3.

Available lists:
  inbox, today, upcoming, anytime, someday, logbook, trash

Examples:
  things list inbox
  things list today -b
  things list logbook -l 10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		listName := strings.ToLower(args[0])

		switch listName {
		case "inbox":
			tasks, err := database.GetInbox()
			if err != nil {
				return err
			}
			return outputTasks(cmd.OutOrStdout(), tasks)
		case "today":
			tasks, err := database.GetToday()
			if err != nil {
				return err
			}
			return outputTasks(cmd.OutOrStdout(), tasks)
		case "upcoming":
			tasks, err := database.GetUpcoming()
			if err != nil {
				return err
			}
			return outputTasks(cmd.OutOrStdout(), tasks)
		case "anytime":
			return anytimeCmd.RunE(cmd, args)
		case "someday":
			tasks, err := database.GetSomeday()
			if err != nil {
				return err
			}
			return outputTasks(cmd.OutOrStdout(), tasks)
		case "logbook":
			tasks, err := database.GetLogbook("", "")
			if err != nil {
				return err
			}
			return outputTasks(cmd.OutOrStdout(), tasks)
		case "trash":
			tasks, err := database.GetTrashed()
			if err != nil {
				return err
			}
			return outputTasks(cmd.OutOrStdout(), tasks)
		default:
			return fmt.Errorf("unknown list: %s\nAvailable: inbox, today, upcoming, anytime, someday, logbook, trash", listName)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
