package cmd

import (
	"github.com/spf13/cobra"
)

var upcomingCmd = &cobra.Command{
	Use:   "upcoming",
	Short: "List upcoming scheduled tasks",
	Long:  `Lists all tasks that have a scheduled start date.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := database.GetUpcoming()
		if err != nil {
			return err
		}
		return outputTasks(cmd.OutOrStdout(), tasks)
	},
}

func init() {
	rootCmd.AddCommand(upcomingCmd)
}
