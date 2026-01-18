package cmd

import (
	"encoding/json"
	"fmt"

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

		w := cmd.OutOrStdout()

		if jsonOutput {
			return json.NewEncoder(w).Encode(result)
		}

		if countOnly {
			total := len(result.Tasks)
			for _, p := range result.Projects {
				total += len(p.Tasks)
			}
			fmt.Fprintln(w, total)
			return nil
		}

		// Output standalone tasks
		for _, task := range result.Tasks {
			fmt.Fprintf(w, "%s\t%s\n", task.UUID, task.Title)
		}

		// Output projects with their tasks
		for _, p := range result.Projects {
			fmt.Fprintf(w, "\n## %s\n", p.Project.Title)
			for _, task := range p.Tasks {
				fmt.Fprintf(w, "%s\t  %s\n", task.UUID, task.Title)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(anytimeCmd)
}
