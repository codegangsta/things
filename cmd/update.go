package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/codegangsta/things/internal/callback"
	"github.com/spf13/cobra"
)

var (
	updateTitle            string
	updateNotes            string
	updateWhen             string
	updateDeadline         string
	updateTags             []string
	updateAddTags          []string
	updateList             string
	updateHeading          string
	appendNotes            bool
	updateAppendChecklist  []string
	updatePrependChecklist []string
	updateComplete         bool
	updateCancel           bool
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
  things update ABC123 --title "New title" --when tomorrow
  things update ABC123 --list "Work"
  things update ABC123 --list "Project Name" --heading "Phase 1"
  things update ABC123 --append-checklist "New item"
  things update ABC123 --prepend-checklist "First item"
  things update ABC123 --complete
  things update ABC123 --cancel`,
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
	updateCmd.Flags().StringVar(&updateList, "list", "", "Move to a project or area by name")
	updateCmd.Flags().StringVar(&updateHeading, "heading", "", "Move under a heading within a project")
	updateCmd.Flags().StringArrayVar(&updateAppendChecklist, "append-checklist", nil, "Add checklist items to the end (can be specified multiple times)")
	updateCmd.Flags().StringArrayVar(&updatePrependChecklist, "prepend-checklist", nil, "Add checklist items to the beginning (can be specified multiple times)")
	updateCmd.Flags().BoolVar(&updateComplete, "complete", false, "Mark task as complete")
	updateCmd.Flags().BoolVar(&updateCancel, "cancel", false, "Mark task as canceled")
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
	if updateList != "" {
		params.Set("list", updateList)
	}
	if updateHeading != "" {
		params.Set("heading", updateHeading)
	}
	if len(updateAppendChecklist) > 0 {
		params.Set("append-checklist-items", strings.Join(updateAppendChecklist, "\n"))
	}
	if len(updatePrependChecklist) > 0 {
		params.Set("prepend-checklist-items", strings.Join(updatePrependChecklist, "\n"))
	}
	if updateComplete {
		params.Set("completed", "true")
	}
	if updateCancel {
		params.Set("canceled", "true")
	}

	// Replace + with %20 since Things URL scheme expects %20 for spaces
	encoded := strings.ReplaceAll(params.Encode(), "+", "%20")
	thingsURL := fmt.Sprintf("things:///update?%s", encoded)

	// Fire-and-forget mode
	if noWait {
		if err := callback.Execute(thingsURL); err != nil {
			return fmt.Errorf("failed to update task %s: %w", id, err)
		}
		fmt.Printf("Updated: %s\n", id)
		return nil
	}

	// Wait for callback confirmation
	result, err := callback.ExecuteWithCallback(thingsURL, callbackTimeout)
	if err != nil {
		return fmt.Errorf("failed to update task %s: %w", id, err)
	}

	if !result.Success {
		return fmt.Errorf("failed to update task %s: %s", id, result.Error)
	}

	fmt.Printf("Updated: %s\n", id)
	return nil
}
