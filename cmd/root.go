package cmd

import (
	"fmt"
	"os"

	"github.com/codegangsta/things/internal/db"
	"github.com/spf13/cobra"
)

var (
	jsonOutput  bool
	briefOutput bool
	countOnly   bool
	limitOutput int
	database    *db.DB
)

var rootCmd = &cobra.Command{
	Use:   "things",
	Short: "A CLI for interacting with Things 3",
	Long: `things is a command line interface for interacting with Things 3 on macOS.
It reads directly from the Things 3 SQLite database to provide fast access
to your tasks, projects, and areas.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip DB initialization for help commands
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}

		dbPath, err := db.DefaultDBPath()
		if err != nil {
			return fmt.Errorf("failed to get database path: %w", err)
		}

		database = &db.DB{}
		if err := database.Open(dbPath); err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if database != nil {
			database.Close()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&briefOutput, "brief", "b", false, "Brief output (title and tags only)")
	rootCmd.PersistentFlags().BoolVarP(&countOnly, "count", "c", false, "Only show count of results")
	rootCmd.PersistentFlags().IntVarP(&limitOutput, "limit", "l", 0, "Limit number of results (0 = no limit)")
}
