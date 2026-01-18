package cmd

import (
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
		return outputTags(cmd.OutOrStdout(), tags)
	},
}

func init() {
	rootCmd.AddCommand(tagsCmd)
}
