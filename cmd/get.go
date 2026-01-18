package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codegangsta/things/internal/db"
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
	Short: "Get detailed information about a task, project, area, or tag",
	Long:  `Shows detailed information about a specific item by UUID. Supports tasks, projects, areas, and tags.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		uuid := args[0]

		// Try to get as task/project first
		task, taskErr := database.GetTask(uuid)
		if taskErr == nil {
			return outputTaskDetail(cmd, task)
		}

		// Try to get as area
		area, areaErr := database.GetAreaByUUID(uuid)
		if areaErr == nil {
			return outputAreaDetail(cmd, area)
		}

		// Try to get as tag (by UUID - less common but possible)
		// For tags, we'll return the original task error since tags are usually accessed by title
		return taskErr
	},
}

func outputTaskDetail(cmd *cobra.Command, task *db.Task) error {
	uuid := task.UUID

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

	// Show tasks for projects, or checklist items for tasks
	if task.Type == 1 { // Project
		tasks, err := database.GetAllTasksInProject(uuid)
		if err != nil {
			return fmt.Errorf("failed to get project tasks: %w", err)
		}
		if len(tasks) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\nTasks:\n")
			return outputTasksWithStatus(cmd.OutOrStdout(), tasks)
		}
	} else if task.Type == 0 { // Task
		items, err := database.GetChecklistItems(uuid)
		if err != nil {
			return fmt.Errorf("failed to get checklist items: %w", err)
		}
		if len(items) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\nChecklist:\n")
			return outputChecklistItems(cmd.OutOrStdout(), items)
		}
	}

	return nil
}

// AreaDetail holds detailed area information for output
type AreaDetail struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

func outputAreaDetail(cmd *cobra.Command, area *db.Area) error {
	detail := AreaDetail{
		UUID:  area.UUID,
		Title: area.Title,
		Type:  "area",
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(detail)
	}

	// Pretty print
	fmt.Fprintf(cmd.OutOrStdout(), "UUID:       %s\n", detail.UUID)
	fmt.Fprintf(cmd.OutOrStdout(), "Title:      %s\n", detail.Title)
	fmt.Fprintf(cmd.OutOrStdout(), "Type:       %s\n", detail.Type)

	// Show projects in this area
	projects, err := database.GetProjectsInArea(area.UUID)
	if err != nil {
		return fmt.Errorf("failed to get area projects: %w", err)
	}
	if len(projects) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "\nProjects:\n")
		if err := outputProjects(cmd.OutOrStdout(), projects); err != nil {
			return err
		}
	}

	// Show tasks directly in this area (not in a project)
	tasks, err := database.GetTasksInArea(area.UUID)
	if err != nil {
		return fmt.Errorf("failed to get area tasks: %w", err)
	}
	if len(tasks) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "\nTasks:\n")
		return outputTasks(cmd.OutOrStdout(), tasks)
	}

	return nil
}

// formatTimestamp converts a Unix timestamp to a readable format
func formatTimestamp(timestamp float64) string {
	t := time.Unix(int64(timestamp), 0)
	return t.Format("2006-01-02 15:04:05")
}

// formatThingsDate converts a Things date integer to a readable format.
// Things 3 encodes dates as: (year << 16) + (day_of_year + 32) * 128
func formatThingsDate(date int64) string {
	year := date >> 16
	dayOfYear := ((date & 0xFFFF) / 128) - 32
	// Convert day of year to month/day
	t := time.Date(int(year), 1, 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, int(dayOfYear)-1)
	return t.Format("2006-01-02")
}

func init() {
	rootCmd.AddCommand(getCmd)
}
