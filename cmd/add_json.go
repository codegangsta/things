package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"

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

		for i, task := range tasks {
			if task.Title == "" {
				return fmt.Errorf("task %d: title is required", i+1)
			}

			if err := addJSONTask(task); err != nil {
				return fmt.Errorf("task %d (%s): %w", i+1, task.Title, err)
			}
			fmt.Printf("Added: %s\n", task.Title)
		}

		fmt.Printf("\nAdded %d tasks\n", len(tasks))
		return nil
	},
}

func addJSONTask(task JSONTask) error {
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

	cmd := exec.Command("open", thingsURL)
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(addJSONCmd)
}
