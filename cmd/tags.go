package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all tags",
	Long:  `Lists all tags in Things 3.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tags, err := database.GetTags()
		if err != nil {
			return err
		}

		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(tags)
		}

		for _, tag := range tags {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", tag.UUID, tag.Title)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagsCmd)
}
