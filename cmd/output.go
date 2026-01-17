package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/codegangsta/things/internal/db"
)

// outputTasks handles formatting and outputting a list of tasks
// respecting the global flags: jsonOutput, briefOutput, countOnly, limitOutput
func outputTasks(w io.Writer, tasks []db.Task) error {
	// Apply limit if specified
	if limitOutput > 0 && len(tasks) > limitOutput {
		tasks = tasks[:limitOutput]
	}

	if jsonOutput {
		return json.NewEncoder(w).Encode(tasks)
	}

	if countOnly {
		fmt.Fprintln(w, len(tasks))
		return nil
	}

	for _, task := range tasks {
		if briefOutput {
			// Get tags for this task
			tags, _ := database.GetTaskTags(task.UUID)
			tagStrs := make([]string, len(tags))
			for i, t := range tags {
				tagStrs[i] = t.Title
			}
			if len(tagStrs) > 0 {
				fmt.Fprintf(w, "%s\t%s (%s)\n", task.UUID, task.Title, strings.Join(tagStrs, ", "))
			} else {
				fmt.Fprintf(w, "%s\t%s\n", task.UUID, task.Title)
			}
		} else {
			fmt.Fprintf(w, "%s\t%s\n", task.UUID, task.Title)
		}
	}
	return nil
}
