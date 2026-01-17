package cmd

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	addNotes    string
	addWhen     string
	addDeadline string
	addTags     []string
	addList     string
	addHeading  string
)

var addCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new task to Things 3",
	Long: `Add a new task to Things 3 using the URL scheme.

Examples:
  things add "Buy groceries"
  things add "Call mom" --when today
  things add "Project deadline" --deadline 2024-12-31
  things add "Work task" --list "Work" --tags "urgent,important"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		return addTask(title)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&addNotes, "notes", "n", "", "Notes for the task")
	addCmd.Flags().StringVarP(&addWhen, "when", "w", "", "When to schedule (today, tomorrow, evening, anytime, someday, or YYYY-MM-DD)")
	addCmd.Flags().StringVarP(&addDeadline, "deadline", "d", "", "Deadline date (YYYY-MM-DD)")
	addCmd.Flags().StringSliceVarP(&addTags, "tags", "t", nil, "Tags to apply (comma-separated)")
	addCmd.Flags().StringVar(&addList, "list", "", "Project or area name to add to")
	addCmd.Flags().StringVar(&addHeading, "heading", "", "Heading within a project")
}

func addTask(title string) error {
	// Build the Things URL scheme
	params := url.Values{}
	params.Set("title", title)

	if addNotes != "" {
		params.Set("notes", addNotes)
	}
	if addWhen != "" {
		params.Set("when", addWhen)
	}
	if addDeadline != "" {
		params.Set("deadline", addDeadline)
	}
	if len(addTags) > 0 {
		params.Set("tags", strings.Join(addTags, ","))
	}
	if addList != "" {
		params.Set("list", addList)
	}
	if addHeading != "" {
		params.Set("heading", addHeading)
	}

	// Replace + with %20 since Things URL scheme expects %20 for spaces
	encoded := strings.ReplaceAll(params.Encode(), "+", "%20")
	thingsURL := fmt.Sprintf("things:///add?%s", encoded)

	// Use macOS open command to trigger the URL scheme
	cmd := exec.Command("open", thingsURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}

	fmt.Printf("Added: %s\n", title)
	return nil
}
