package cmd

import (
	"encoding/json"
	"fmt"

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

		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(projects)
		}

		for _, project := range projects {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", project.UUID, project.Title)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)
}
