package cmd

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	updateTitle    string
	updateNotes    string
	updateWhen     string
	updateDeadline string
	updateTags     []string
	updateAddTags  []string
	appendNotes    bool
)

var updateCmd = &cobra.Command{
	Use:   "update <task-id>",
	Short: "Update an existing task",
	Long: `Update an existing task in Things 3 using the URL scheme.

Requires auth-token to be set. Get your token from:
Things → Settings → General → Enable Things URLs → Manage

Examples:
  things update ABC123 --when today
  things update ABC123 --notes "Updated context"
  things update ABC123 --notes "Additional info" --append
  things update ABC123 --add-tags "@phone,5m"
  things update ABC123 --title "New title" --when tomorrow`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAuthToken()
		if err != nil {
			return fmt.Errorf("auth token required: run 'things auth <token>' first.\nGet your token from Things → Settings → General → Enable Things URLs → Manage")
		}

		taskID := args[0]

		// If appending notes, we need to get the existing notes first
		if appendNotes && updateNotes != "" {
			task, err := database.GetTask(taskID)
			if err != nil {
				return fmt.Errorf("failed to get task: %w", err)
			}
			if task.Notes.Valid && task.Notes.String != "" {
				updateNotes = task.Notes.String + "\n\n---\n" + updateNotes
			}
		}

		return updateTask(taskID, token)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New title for the task")
	updateCmd.Flags().StringVar(&updateNotes, "notes", "", "Notes for the task (replaces existing unless --append)")
	updateCmd.Flags().BoolVar(&appendNotes, "append", false, "Append to existing notes instead of replacing")
	updateCmd.Flags().StringVarP(&updateWhen, "when", "w", "", "When to schedule (today, tomorrow, evening, anytime, someday, or YYYY-MM-DD)")
	updateCmd.Flags().StringVarP(&updateDeadline, "deadline", "d", "", "Deadline date (YYYY-MM-DD)")
	updateCmd.Flags().StringSliceVar(&updateTags, "tags", nil, "Replace all tags (comma-separated)")
	updateCmd.Flags().StringSliceVar(&updateAddTags, "add-tags", nil, "Add tags without removing existing (comma-separated)")
}

func updateTask(id, token string) error {
	params := url.Values{}
	params.Set("id", id)
	params.Set("auth-token", token)

	if updateTitle != "" {
		params.Set("title", updateTitle)
	}
	if updateNotes != "" {
		params.Set("notes", updateNotes)
	}
	if updateWhen != "" {
		params.Set("when", updateWhen)
	}
	if updateDeadline != "" {
		params.Set("deadline", updateDeadline)
	}
	if len(updateTags) > 0 {
		params.Set("tags", strings.Join(updateTags, ","))
	}
	if len(updateAddTags) > 0 {
		params.Set("add-tags", strings.Join(updateAddTags, ","))
	}

	// Replace + with %20 since Things URL scheme expects %20 for spaces
	encoded := strings.ReplaceAll(params.Encode(), "+", "%20")
	thingsURL := fmt.Sprintf("things:///update?%s", encoded)

	cmd := exec.Command("open", thingsURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update task %s: %w", id, err)
	}

	fmt.Printf("Updated: %s\n", id)
	return nil
}
