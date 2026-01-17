package cmd

import (
	"github.com/spf13/cobra"
)

var (
	logbookStart string
	logbookEnd   string
)

var logbookCmd = &cobra.Command{
	Use:   "logbook",
	Short: "List completed tasks in the Logbook",
	Long: `Lists completed tasks from the Logbook, optionally filtered by date range.

Examples:
  things logbook
  things logbook --start 2024-01-01
  things logbook --start 2024-01-01 --end 2024-01-31
  things logbook -l 10`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := database.GetLogbook(logbookStart, logbookEnd)
		if err != nil {
			return err
		}
		return outputTasks(cmd.OutOrStdout(), tasks)
	},
}

func init() {
	rootCmd.AddCommand(logbookCmd)
	logbookCmd.Flags().StringVar(&logbookStart, "start", "", "Start date (YYYY-MM-DD)")
	logbookCmd.Flags().StringVar(&logbookEnd, "end", "", "End date (YYYY-MM-DD)")
}
