package cmd

import (
	"github.com/spf13/cobra"
)

var trashCmd = &cobra.Command{
	Use:   "trash",
	Short: "List tasks in the Trash",
	Long:  `Lists all trashed tasks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := database.GetTrashed()
		if err != nil {
			return err
		}
		return outputTasks(cmd.OutOrStdout(), tasks)
	},
}

func init() {
	rootCmd.AddCommand(trashCmd)
}
