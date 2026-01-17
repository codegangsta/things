package cmd

import (
	"github.com/spf13/cobra"
)

var inboxCmd = &cobra.Command{
	Use:   "inbox",
	Short: "List tasks in the Inbox",
	Long:  `Lists all tasks in the Inbox (tasks with no project, no area, and not scheduled).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := database.GetInbox()
		if err != nil {
			return err
		}
		return outputTasks(cmd.OutOrStdout(), tasks)
	},
}

func init() {
	rootCmd.AddCommand(inboxCmd)
}
