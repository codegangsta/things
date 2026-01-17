package cmd

import (
	"encoding/json"
	"fmt"

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

		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(areas)
		}

		for _, area := range areas {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", area.UUID, area.Title)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(areasCmd)
}
