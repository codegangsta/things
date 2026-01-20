package cmd

import (
	"github.com/codegangsta/things/internal/db"
	"github.com/spf13/cobra"
)

var somedayCmd = &cobra.Command{
	Use:   "someday",
	Short: "List tasks in Someday",
	Long:  `Lists all tasks and projects in the Someday list.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := database.GetSomeday()
		if err != nil {
			return err
		}

		// Combine tasks and projects into a single list
		var tasks []db.Task
		tasks = append(tasks, result.Tasks...)
		tasks = append(tasks, result.Projects...)

		return outputTasks(cmd.OutOrStdout(), tasks)
	},
}

func init() {
	rootCmd.AddCommand(somedayCmd)
}
