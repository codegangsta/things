package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/codegangsta/things/internal/callback"
	"github.com/spf13/cobra"
)

// JSONTask represents a task in JSON format for bulk import
type JSONTask struct {
	Title       string   `json:"title"`
	Notes       string   `json:"notes,omitempty"`
	When        string   `json:"when,omitempty"`
	Deadline    string   `json:"deadline,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	List        string   `json:"list,omitempty"`
	Heading     string   `json:"heading,omitempty"`
	ChecklistItems []string `json:"checklist_items,omitempty"`
}

var addJSONCmd = &cobra.Command{
	Use:   "add-json",
	Short: "Add multiple tasks from JSON input",
	Long: `Add multiple tasks from JSON input (stdin or file).

JSON format (array of tasks):
[
  {
    "title": "Task 1",
    "notes": "Optional notes",
    "when": "today",
    "tags": ["@phone", "5m"],
    "list": "Project Name"
  },
  {
    "title": "Task 2",
    "deadline": "2024-12-31"
  }
]

Examples:
  echo '[{"title": "Test task", "when": "today"}]' | things add-json
  things add-json < tasks.json
  cat tasks.json | things add-json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		var tasks []JSONTask
		if err := json.Unmarshal(data, &tasks); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks to add")
			return nil
		}

		var allIDs []string
		for i, task := range tasks {
			if task.Title == "" {
				return fmt.Errorf("task %d: title is required", i+1)
			}

			ids, err := addJSONTask(task)
			if err != nil {
				return fmt.Errorf("task %d (%s): %w", i+1, task.Title, err)
			}
			allIDs = append(allIDs, ids...)
			if len(ids) > 0 {
				fmt.Printf("Added: %s (ID: %s)\n", task.Title, ids[0])
			} else {
				fmt.Printf("Added: %s\n", task.Title)
			}
		}

		fmt.Printf("\nAdded %d tasks\n", len(tasks))
		if len(allIDs) > 0 {
			fmt.Printf("IDs: %s\n", strings.Join(allIDs, ", "))
		}
		return nil
	},
}

func addJSONTask(task JSONTask) ([]string, error) {
	params := url.Values{}
	params.Set("title", task.Title)

	if task.Notes != "" {
		params.Set("notes", task.Notes)
	}
	if task.When != "" {
		params.Set("when", task.When)
	}
	if task.Deadline != "" {
		params.Set("deadline", task.Deadline)
	}
	if len(task.Tags) > 0 {
		params.Set("tags", strings.Join(task.Tags, ","))
	}
	if task.List != "" {
		params.Set("list", task.List)
	}
	if task.Heading != "" {
		params.Set("heading", task.Heading)
	}
	if len(task.ChecklistItems) > 0 {
		params.Set("checklist-items", strings.Join(task.ChecklistItems, "\n"))
	}

	encoded := strings.ReplaceAll(params.Encode(), "+", "%20")
	thingsURL := fmt.Sprintf("things:///add?%s", encoded)

	// Fire-and-forget mode
	if noWait {
		if err := callback.Execute(thingsURL); err != nil {
			return nil, err
		}
		return nil, nil
	}

	// Wait for callback confirmation
	result, err := callback.ExecuteWithCallback(thingsURL, callbackTimeout)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("%s", result.Error)
	}

	return result.IDs, nil
}

func init() {
	rootCmd.AddCommand(addJSONCmd)
}
