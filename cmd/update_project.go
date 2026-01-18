package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/codegangsta/things/internal/callback"
	"github.com/spf13/cobra"
)

var (
	updateProjectTitle    string
	updateProjectNotes    string
	updateProjectWhen     string
	updateProjectDeadline string
	updateProjectTags     []string
	updateProjectAddTags  []string
	updateProjectArea     string
	updateProjectAppend   bool
	updateProjectComplete bool
	updateProjectCancel   bool
)

var updateProjectCmd = &cobra.Command{
	Use:   "update-project <project-id>",
	Short: "Update an existing project",
	Long: `Update an existing project in Things 3 using the URL scheme.

Requires auth-token to be set. Get your token from:
Things → Settings → General → Enable Things URLs → Manage

Examples:
  things update-project ABC123 --when today
  things update-project ABC123 --notes "Updated description"
  things update-project ABC123 --notes "Additional info" --append
  things update-project ABC123 --add-tags "priority"
  things update-project ABC123 --title "New Project Name"
  things update-project ABC123 --area "Work"
  things update-project ABC123 --complete
  things update-project ABC123 --cancel`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAuthToken()
		if err != nil {
			return fmt.Errorf("auth token required: run 'things auth <token>' first.\nGet your token from Things → Settings → General → Enable Things URLs → Manage")
		}

		projectID := args[0]

		// If appending notes, we need to use the append-notes parameter
		notesParam := "notes"
		if updateProjectAppend && updateProjectNotes != "" {
			notesParam = "append-notes"
		}

		return updateProject(projectID, token, notesParam)
	},
}

func init() {
	rootCmd.AddCommand(updateProjectCmd)

	updateProjectCmd.Flags().StringVar(&updateProjectTitle, "title", "", "New title for the project")
	updateProjectCmd.Flags().StringVar(&updateProjectNotes, "notes", "", "Notes for the project (replaces existing unless --append)")
	updateProjectCmd.Flags().BoolVar(&updateProjectAppend, "append", false, "Append to existing notes instead of replacing")
	updateProjectCmd.Flags().StringVarP(&updateProjectWhen, "when", "w", "", "When to schedule (today, tomorrow, evening, anytime, someday, or YYYY-MM-DD)")
	updateProjectCmd.Flags().StringVarP(&updateProjectDeadline, "deadline", "d", "", "Deadline date (YYYY-MM-DD)")
	updateProjectCmd.Flags().StringSliceVar(&updateProjectTags, "tags", nil, "Replace all tags (comma-separated)")
	updateProjectCmd.Flags().StringSliceVar(&updateProjectAddTags, "add-tags", nil, "Add tags without removing existing (comma-separated)")
	updateProjectCmd.Flags().StringVar(&updateProjectArea, "area", "", "Move to an area by name")
	updateProjectCmd.Flags().BoolVar(&updateProjectComplete, "complete", false, "Mark project as complete")
	updateProjectCmd.Flags().BoolVar(&updateProjectCancel, "cancel", false, "Mark project as canceled")
}

func updateProject(id, token, notesParam string) error {
	params := url.Values{}
	params.Set("id", id)
	params.Set("auth-token", token)

	if updateProjectTitle != "" {
		params.Set("title", updateProjectTitle)
	}
	if updateProjectNotes != "" {
		params.Set(notesParam, updateProjectNotes)
	}
	if updateProjectWhen != "" {
		params.Set("when", updateProjectWhen)
	}
	if updateProjectDeadline != "" {
		params.Set("deadline", updateProjectDeadline)
	}
	if len(updateProjectTags) > 0 {
		params.Set("tags", strings.Join(updateProjectTags, ","))
	}
	if len(updateProjectAddTags) > 0 {
		params.Set("add-tags", strings.Join(updateProjectAddTags, ","))
	}
	if updateProjectArea != "" {
		params.Set("area", updateProjectArea)
	}
	if updateProjectComplete {
		params.Set("completed", "true")
	}
	if updateProjectCancel {
		params.Set("canceled", "true")
	}

	encoded := strings.ReplaceAll(params.Encode(), "+", "%20")
	thingsURL := fmt.Sprintf("things:///update-project?%s", encoded)

	// Fire-and-forget mode
	if noWait {
		if err := callback.Execute(thingsURL); err != nil {
			return fmt.Errorf("failed to update project %s: %w", id, err)
		}
		fmt.Printf("Updated project: %s\n", id)
		return nil
	}

	// Wait for callback confirmation
	result, err := callback.ExecuteWithCallback(thingsURL, callbackTimeout)
	if err != nil {
		return fmt.Errorf("failed to update project %s: %w", id, err)
	}

	if !result.Success {
		return fmt.Errorf("failed to update project %s: %s", id, result.Error)
	}

	fmt.Printf("Updated project: %s\n", id)
	return nil
}
