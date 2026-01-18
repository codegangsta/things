package cmd

import (
	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List all active projects",
	Long:  `Lists all active projects in Things 3.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projects, err := database.GetProjects()
		if err != nil {
			return err
		}
		return outputProjects(cmd.OutOrStdout(), projects)
	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)
}
