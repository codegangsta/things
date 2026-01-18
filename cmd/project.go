package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project <id-or-name>",
	Short: "List tasks in a specific project",
	Long: `List all open tasks within a specific project.

You can specify the project by its UUID or by its name.
If using a name, it will be matched case-insensitively.

Examples:
  things project "My Project"
  things project 8A3B4C5D-1234-5678-9ABC-DEF012345678
  things project "home renovation" --json
  things project "Work Tasks" -b`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idOrName := args[0]

		// Try to find the project by UUID or name
		projectUUID, err := database.ResolveProjectUUID(idOrName)
		if err != nil {
			return fmt.Errorf("project not found: %s", idOrName)
		}

		tasks, err := database.GetTasksInProject(projectUUID)
		if err != nil {
			return err
		}
		return outputTasks(cmd.OutOrStdout(), tasks)
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
}
