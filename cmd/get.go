package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// TaskDetail holds detailed task information for output
type TaskDetail struct {
	UUID      string   `json:"uuid"`
	Title     string   `json:"title"`
	Notes     string   `json:"notes,omitempty"`
	Type      string   `json:"type"`
	Status    string   `json:"status"`
	Start     string   `json:"start"`
	StartDate string   `json:"start_date,omitempty"`
	Deadline  string   `json:"deadline,omitempty"`
	Project   string   `json:"project,omitempty"`
	Area      string   `json:"area,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	CreatedAt string   `json:"created_at"`
}

var getCmd = &cobra.Command{
	Use:   "get <uuid>",
	Short: "Get detailed information about a task",
	Long:  `Shows detailed information about a specific task, including title, notes, tags, project, area, and dates.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		uuid := args[0]

		task, err := database.GetTask(uuid)
		if err != nil {
			return err
		}

		// Get tags for this task
		tags, err := database.GetTaskTags(uuid)
		if err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}

		// Build detail struct
		detail := TaskDetail{
			UUID:  task.UUID,
			Title: task.Title,
		}
		if task.CreationDate.Valid {
			detail.CreatedAt = formatTimestamp(task.CreationDate.Float64)
		}

		if task.Notes.Valid {
			detail.Notes = task.Notes.String
		}

		// Type
		switch task.Type {
		case 0:
			detail.Type = "task"
		case 1:
			detail.Type = "project"
		case 2:
			detail.Type = "heading"
		default:
			detail.Type = fmt.Sprintf("unknown(%d)", task.Type)
		}

		// Status
		switch task.Status {
		case 0:
			detail.Status = "open"
		case 2:
			detail.Status = "canceled"
		case 3:
			detail.Status = "completed"
		default:
			detail.Status = fmt.Sprintf("unknown(%d)", task.Status)
		}

		// Start
		switch task.Start {
		case 0:
			detail.Start = "not_started"
		case 1:
			detail.Start = "today"
		case 2:
			detail.Start = "someday"
		default:
			detail.Start = fmt.Sprintf("unknown(%d)", task.Start)
		}

		if task.StartDate.Valid {
			detail.StartDate = formatThingsDate(task.StartDate.Int64)
		}

		if task.Deadline.Valid {
			detail.Deadline = formatThingsDate(task.Deadline.Int64)
		}

		// Get project name if task has a project
		if task.Project.Valid {
			project, err := database.GetTask(task.Project.String)
			if err == nil {
				detail.Project = project.Title
			} else {
				detail.Project = task.Project.String
			}
		}

		// Get area name if task has an area
		if task.Area.Valid {
			area, err := database.GetAreaByUUID(task.Area.String)
			if err == nil {
				detail.Area = area.Title
			} else {
				detail.Area = task.Area.String
			}
		}

		// Tags
		for _, tag := range tags {
			detail.Tags = append(detail.Tags, tag.Title)
		}

		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(detail)
		}

		// Pretty print
		fmt.Fprintf(cmd.OutOrStdout(), "UUID:       %s\n", detail.UUID)
		fmt.Fprintf(cmd.OutOrStdout(), "Title:      %s\n", detail.Title)
		fmt.Fprintf(cmd.OutOrStdout(), "Type:       %s\n", detail.Type)
		fmt.Fprintf(cmd.OutOrStdout(), "Status:     %s\n", detail.Status)
		fmt.Fprintf(cmd.OutOrStdout(), "Start:      %s\n", detail.Start)

		if detail.StartDate != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Start Date: %s\n", detail.StartDate)
		}
		if detail.Deadline != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Deadline:   %s\n", detail.Deadline)
		}
		if detail.Project != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Project:    %s\n", detail.Project)
		}
		if detail.Area != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Area:       %s\n", detail.Area)
		}
		if len(detail.Tags) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Tags:       %s\n", strings.Join(detail.Tags, ", "))
		}
		if detail.Notes != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Notes:\n%s\n", detail.Notes)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created:    %s\n", detail.CreatedAt)

		return nil
	},
}

// formatTimestamp converts a Core Data timestamp (seconds since 2001-01-01) to a readable format
func formatTimestamp(timestamp float64) string {
	// Core Data reference date is 2001-01-01 00:00:00 UTC
	referenceDate := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	t := referenceDate.Add(time.Duration(timestamp) * time.Second)
	return t.Format("2006-01-02 15:04:05")
}

// formatThingsDate converts a Things date integer (YYYYMMDD) to a readable format
func formatThingsDate(date int64) string {
	year := date / 10000
	month := (date % 10000) / 100
	day := date % 100
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func init() {
	rootCmd.AddCommand(getCmd)
}
