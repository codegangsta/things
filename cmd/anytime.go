package cmd

import (
	"github.com/spf13/cobra"
)

var anytimeCmd = &cobra.Command{
	Use:   "anytime",
	Short: "List tasks in Anytime",
	Long:  `Lists all tasks in the Anytime list (not scheduled for today, not in someday).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := database.GetAnytime()
		if err != nil {
			return err
		}
		return outputTasks(cmd.OutOrStdout(), tasks)
	},
}

func init() {
	rootCmd.AddCommand(anytimeCmd)
}
