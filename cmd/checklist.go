package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/codegangsta/things/internal/callback"
	"github.com/codegangsta/things/internal/db"
	"github.com/spf13/cobra"
)

var checklistCmd = &cobra.Command{
	Use:   "checklist",
	Short: "Manage checklist items within a task",
	Long: `Manage checklist items within a task.

Subcommands:
  add         Add a new checklist item to a task
  complete    Mark a checklist item as complete
  uncomplete  Mark a checklist item as incomplete

Examples:
  things checklist add ABC123 "Buy milk"
  things checklist complete ABC123 1
  things checklist uncomplete ABC123 2`,
}

var checklistAddCmd = &cobra.Command{
	Use:   "add <task-id> <item-text>",
	Short: "Add a checklist item to a task",
	Long: `Add a new checklist item to an existing task.

The item is appended to the end of the checklist.

Examples:
  things checklist add ABC123 "Buy milk"
  things checklist add ABC123 "Call the doctor"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAuthToken()
		if err != nil {
			return fmt.Errorf("auth token required: run 'things auth <token>' first")
		}

		taskID := args[0]
		itemText := args[1]

		return addChecklistItem(taskID, itemText, token)
	},
}

var checklistCompleteCmd = &cobra.Command{
	Use:   "complete <task-id> <item-index>",
	Short: "Mark a checklist item as complete",
	Long: `Mark a checklist item as complete by its index (1-based).

Use 'things get <task-id>' to see the checklist with indices.

Examples:
  things checklist complete ABC123 1
  things checklist complete ABC123 3`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAuthToken()
		if err != nil {
			return fmt.Errorf("auth token required: run 'things auth <token>' first")
		}

		taskID := args[0]
		index, err := strconv.Atoi(args[1])
		if err != nil || index < 1 {
			return fmt.Errorf("invalid index: must be a positive integer")
		}

		return setChecklistItemStatus(taskID, index, true, token)
	},
}

var checklistUncompleteCmd = &cobra.Command{
	Use:   "uncomplete <task-id> <item-index>",
	Short: "Mark a checklist item as incomplete",
	Long: `Mark a checklist item as incomplete by its index (1-based).

Use 'things get <task-id>' to see the checklist with indices.

Examples:
  things checklist uncomplete ABC123 1
  things checklist uncomplete ABC123 3`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAuthToken()
		if err != nil {
			return fmt.Errorf("auth token required: run 'things auth <token>' first")
		}

		taskID := args[0]
		index, err := strconv.Atoi(args[1])
		if err != nil || index < 1 {
			return fmt.Errorf("invalid index: must be a positive integer")
		}

		return setChecklistItemStatus(taskID, index, false, token)
	},
}

func init() {
	rootCmd.AddCommand(checklistCmd)
	checklistCmd.AddCommand(checklistAddCmd)
	checklistCmd.AddCommand(checklistCompleteCmd)
	checklistCmd.AddCommand(checklistUncompleteCmd)
}

func addChecklistItem(taskID, itemText, token string) error {
	params := url.Values{}
	params.Set("id", taskID)
	params.Set("auth-token", token)
	params.Set("append-checklist-items", itemText)

	encoded := strings.ReplaceAll(params.Encode(), "+", "%20")
	thingsURL := fmt.Sprintf("things:///update?%s", encoded)

	// Fire-and-forget mode
	if noWait {
		if err := callback.Execute(thingsURL); err != nil {
			return fmt.Errorf("failed to add checklist item: %w", err)
		}
		fmt.Printf("Added checklist item to %s: %s\n", taskID, itemText)
		return nil
	}

	// Wait for callback confirmation
	result, err := callback.ExecuteWithCallback(thingsURL, callbackTimeout)
	if err != nil {
		return fmt.Errorf("failed to add checklist item: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("failed to add checklist item: %s", result.Error)
	}

	fmt.Printf("Added checklist item to %s: %s\n", taskID, itemText)
	return nil
}

// checklistItemJSON represents the JSON structure for a checklist item
type checklistItemJSON struct {
	Type       string                    `json:"type"`
	Attributes checklistItemAttributeJSON `json:"attributes"`
}

type checklistItemAttributeJSON struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	Canceled  bool   `json:"canceled"`
}

func setChecklistItemStatus(taskID string, index int, completed bool, token string) error {
	// Get existing checklist items
	items, err := database.GetChecklistItems(taskID)
	if err != nil {
		return fmt.Errorf("failed to get checklist items: %w", err)
	}

	if len(items) == 0 {
		return fmt.Errorf("task %s has no checklist items", taskID)
	}

	if index < 1 || index > len(items) {
		return fmt.Errorf("invalid index %d: task has %d checklist items (use 1-%d)", index, len(items), len(items))
	}

	// Build JSON payload with all items, modifying the target one
	jsonItems := make([]checklistItemJSON, len(items))
	for i, item := range items {
		isCompleted := item.Status == db.TaskStatusCompleted
		isCanceled := item.Status == db.TaskStatusCanceled

		// Modify the target item
		if i == index-1 { // Convert 1-based to 0-based
			isCompleted = completed
			isCanceled = false // Uncomplete also clears canceled
		}

		jsonItems[i] = checklistItemJSON{
			Type: "checklist-item",
			Attributes: checklistItemAttributeJSON{
				Title:     item.Title,
				Completed: isCompleted,
				Canceled:  isCanceled,
			},
		}
	}

	// Build the update payload
	updatePayload := []map[string]interface{}{
		{
			"type": "to-do",
			"id":   taskID,
			"attributes": map[string]interface{}{
				"checklist-items": jsonItems,
			},
		},
	}

	jsonData, err := json.Marshal(updatePayload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Use the JSON URL scheme
	params := url.Values{}
	params.Set("auth-token", token)
	params.Set("data", string(jsonData))

	encoded := strings.ReplaceAll(params.Encode(), "+", "%20")
	thingsURL := fmt.Sprintf("things:///json?%s", encoded)

	action := "Completed"
	if !completed {
		action = "Uncompleted"
	}

	// Fire-and-forget mode
	if noWait {
		if err := callback.Execute(thingsURL); err != nil {
			return fmt.Errorf("failed to update checklist item: %w", err)
		}
		fmt.Printf("%s checklist item %d in %s: %s\n", action, index, taskID, items[index-1].Title)
		return nil
	}

	// Wait for callback confirmation
	result, err := callback.ExecuteWithCallback(thingsURL, callbackTimeout)
	if err != nil {
		return fmt.Errorf("failed to update checklist item: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("failed to update checklist item: %s", result.Error)
	}

	fmt.Printf("%s checklist item %d in %s: %s\n", action, index, taskID, items[index-1].Title)
	return nil
}
