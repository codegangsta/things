package cmd

import (
	"github.com/spf13/cobra"
)

var areasCmd = &cobra.Command{
	Use:   "areas",
	Short: "List all areas",
	Long:  `Lists all areas in Things 3.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		areas, err := database.GetAreas()
		if err != nil {
			return err
		}
		return outputAreas(cmd.OutOrStdout(), areas)
	},
}

func init() {
	rootCmd.AddCommand(areasCmd)
}
