package cmd

import (
	"github.com/spf13/cobra"
)

var somedayCmd = &cobra.Command{
	Use:   "someday",
	Short: "List tasks in Someday",
	Long:  `Lists all tasks in the Someday list.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := database.GetSomeday()
		if err != nil {
			return err
		}
		return outputTasks(cmd.OutOrStdout(), tasks)
	},
}

func init() {
	rootCmd.AddCommand(somedayCmd)
}
