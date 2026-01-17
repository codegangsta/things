package cmd

import (
	"github.com/spf13/cobra"
)

var taggedCmd = &cobra.Command{
	Use:   "tagged <tag>",
	Short: "List tasks with a specific tag",
	Long: `List all open tasks that have a specific tag.

Examples:
  things tagged @claude
  things tagged @phone
  things tagged 25m
  things tagged @claude -b
  things tagged @waiting -l 10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tagName := args[0]

		tasks, err := database.GetTasksByTag(tagName)
		if err != nil {
			return err
		}
		return outputTasks(cmd.OutOrStdout(), tasks)
	},
}

func init() {
	rootCmd.AddCommand(taggedCmd)
}
