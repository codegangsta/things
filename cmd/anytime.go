package cmd

import (
	"github.com/codegangsta/things/internal/db"
	"github.com/spf13/cobra"
)

var anytimeCmd = &cobra.Command{
	Use:   "anytime",
	Short: "List tasks in Anytime",
	Long:  `Lists all tasks in the Anytime list (not scheduled for today, not in someday).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := database.GetAnytime()
		if err != nil {
			return err
		}

		// Flatten tasks and projects into a single list
		var tasks []db.Task
		tasks = append(tasks, result.Tasks...)
		for _, p := range result.Projects {
			tasks = append(tasks, p.Project)
			tasks = append(tasks, p.Tasks...)
		}

		return outputTasks(cmd.OutOrStdout(), tasks)
	},
}

func init() {
	rootCmd.AddCommand(anytimeCmd)
}
