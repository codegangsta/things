package cmd

import (
	"github.com/spf13/cobra"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "List tasks scheduled for today",
	Long:  `Lists all tasks that are scheduled for Today in Things 3.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := database.GetToday()
		if err != nil {
			return err
		}
		return outputTasks(cmd.OutOrStdout(), tasks)
	},
}

func init() {
	rootCmd.AddCommand(todayCmd)
}
