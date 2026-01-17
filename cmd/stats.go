package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show task statistics",
	Long:  `Shows aggregated counts for all lists in Things 3.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stats, err := database.GetStats()
		if err != nil {
			return err
		}

		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(stats)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Inbox:     %d\n", stats.Inbox)
		fmt.Fprintf(cmd.OutOrStdout(), "Today:     %d\n", stats.Today)
		fmt.Fprintf(cmd.OutOrStdout(), "Upcoming:  %d\n", stats.Upcoming)
		fmt.Fprintf(cmd.OutOrStdout(), "Anytime:   %d\n", stats.Anytime)
		fmt.Fprintf(cmd.OutOrStdout(), "Someday:   %d\n", stats.Someday)
		fmt.Fprintf(cmd.OutOrStdout(), "Completed: %d\n", stats.Completed)
		fmt.Fprintf(cmd.OutOrStdout(), "---\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Projects:  %d\n", stats.Projects)
		fmt.Fprintf(cmd.OutOrStdout(), "Areas:     %d\n", stats.Areas)
		fmt.Fprintf(cmd.OutOrStdout(), "Tags:      %d\n", stats.Tags)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
