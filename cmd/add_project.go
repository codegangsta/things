package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/codegangsta/things/internal/callback"
	"github.com/spf13/cobra"
)

var (
	addProjectNotes    string
	addProjectWhen     string
	addProjectDeadline string
	addProjectTags     []string
	addProjectArea     string
	addProjectTodos    []string
)

var addProjectCmd = &cobra.Command{
	Use:   "add-project [title]",
	Short: "Add a new project to Things 3",
	Long: `Add a new project to Things 3 using the URL scheme.

Examples:
  things add-project "Home Renovation"
  things add-project "Q1 Goals" --when today
  things add-project "Book Launch" --deadline 2024-12-31
  things add-project "Work Project" --area "Work" --tags "important"
  things add-project "Shopping List" --todos "Groceries" --todos "Hardware store"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		return addProject(title)
	},
}

func init() {
	rootCmd.AddCommand(addProjectCmd)

	addProjectCmd.Flags().StringVarP(&addProjectNotes, "notes", "n", "", "Notes for the project")
	addProjectCmd.Flags().StringVarP(&addProjectWhen, "when", "w", "", "When to schedule (today, tomorrow, evening, anytime, someday, or YYYY-MM-DD)")
	addProjectCmd.Flags().StringVarP(&addProjectDeadline, "deadline", "d", "", "Deadline date (YYYY-MM-DD)")
	addProjectCmd.Flags().StringSliceVarP(&addProjectTags, "tags", "t", nil, "Tags to apply (comma-separated)")
	addProjectCmd.Flags().StringVar(&addProjectArea, "area", "", "Area name to add project to")
	addProjectCmd.Flags().StringArrayVar(&addProjectTodos, "todos", nil, "To-dos to create inside the project (can be specified multiple times)")
}

func addProject(title string) error {
	params := url.Values{}
	params.Set("title", title)

	if addProjectNotes != "" {
		params.Set("notes", addProjectNotes)
	}
	if addProjectWhen != "" {
		params.Set("when", addProjectWhen)
	}
	if addProjectDeadline != "" {
		params.Set("deadline", addProjectDeadline)
	}
	if len(addProjectTags) > 0 {
		params.Set("tags", strings.Join(addProjectTags, ","))
	}
	if addProjectArea != "" {
		params.Set("area", addProjectArea)
	}
	if len(addProjectTodos) > 0 {
		params.Set("to-dos", strings.Join(addProjectTodos, "\n"))
	}

	encoded := strings.ReplaceAll(params.Encode(), "+", "%20")
	thingsURL := fmt.Sprintf("things:///add-project?%s", encoded)

	// Fire-and-forget mode
	if noWait {
		if err := callback.Execute(thingsURL); err != nil {
			return fmt.Errorf("failed to add project: %w", err)
		}
		fmt.Printf("Added project: %s\n", title)
		return nil
	}

	// Wait for callback confirmation
	result, err := callback.ExecuteWithCallback(thingsURL, callbackTimeout)
	if err != nil {
		return fmt.Errorf("failed to add project: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("failed to add project: %s", result.Error)
	}

	if len(result.IDs) > 0 {
		fmt.Printf("Added project: %s (ID: %s)\n", title, result.IDs[0])
	} else {
		fmt.Printf("Added project: %s\n", title)
	}
	return nil
}
