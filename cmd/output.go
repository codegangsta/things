package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/codegangsta/things/internal/db"
	"github.com/mattn/go-runewidth"
)

// isEmoji returns true if the rune is an emoji or emoji-related character
func isEmoji(r rune) bool {
	// Variation selectors
	if r == 0xFE0E || r == 0xFE0F {
		return true
	}
	// Common emoji ranges
	if r >= 0x1F300 && r <= 0x1F9FF { // Misc Symbols, Emoticons, etc.
		return true
	}
	if r >= 0x2600 && r <= 0x26FF { // Misc Symbols (includes ☕)
		return true
	}
	if r >= 0x2700 && r <= 0x27BF { // Dingbats
		return true
	}
	if r >= 0x1F600 && r <= 0x1F64F { // Emoticons
		return true
	}
	if r >= 0x1F680 && r <= 0x1F6FF { // Transport symbols
		return true
	}
	if r >= 0x200D && r <= 0x200D { // Zero-width joiner
		return true
	}
	return false
}

// stripEmojis removes all emoji characters from a string
func stripEmojis(s string) string {
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if !isEmoji(r) {
			result = append(result, r)
		}
	}
	// Trim leading space if emoji was at start
	return strings.TrimLeft(string(result), " ")
}

// stringWidth calculates display width
func stringWidth(s string) int {
	width := 0
	for _, r := range s {
		width += runewidth.RuneWidth(r)
	}
	return width
}

// truncate shortens a string to maxWidth display columns, adding "..." if truncated
func truncate(s string, maxWidth int) string {
	width := stringWidth(s)
	if width <= maxWidth {
		return s
	}

	// Manually truncate respecting display width
	result := []rune{}
	currentWidth := 0
	targetWidth := maxWidth - 3 // leave room for "..."
	if targetWidth < 0 {
		targetWidth = maxWidth
	}

	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if currentWidth+rw > targetWidth {
			break
		}
		result = append(result, r)
		currentWidth += rw
	}

	if maxWidth > 3 {
		return string(result) + "..."
	}
	return string(result)
}

// padRight pads a string to the given display width
func padRight(s string, width int) string {
	w := stringWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// decodeStartDate decodes Things 3 startDate format: (year << 16) + (dayOfYear+32)*128
func decodeStartDate(encoded int64) time.Time {
	year := int(encoded >> 16)
	dayOfYear := int((encoded&0xFFFF)/128) - 32
	// Create date from year and day of year
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, dayOfYear-1)
}

// formatWhen returns a human-readable "when" value for a task
func formatWhen(task db.Task) string {
	// Check if there's a specific start date
	if task.StartDate.Valid {
		date := decodeStartDate(task.StartDate.Int64)
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		tomorrow := today.AddDate(0, 0, 1)

		if date.Year() == today.Year() && date.YearDay() == today.YearDay() {
			return "today"
		}
		if date.Year() == tomorrow.Year() && date.YearDay() == tomorrow.YearDay() {
			return "tomorrow"
		}
		return date.Format("Jan 2")
	}

	// Fall back to start type
	switch task.Start {
	case db.StartTypeToday:
		return "anytime"
	case db.StartTypeSomeday:
		return "someday"
	default:
		return "-"
	}
}

// Column widths for table output
const (
	colWidthID    = 22
	colWidthTitle = 45
	colWidthType  = 7
	colWidthList  = 20
	colWidthTags  = 15
)

// taskTypeString returns a human-readable string for the task type
func taskTypeString(t db.TaskType) string {
	switch t {
	case db.TaskTypeProject:
		return "project"
	default:
		return "task"
	}
}

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

	// Print header
	if !briefOutput {
		fmt.Fprintf(w, "%s  %s  %s  %s  %s  %s\n",
			padRight("ID", colWidthID),
			padRight("TITLE", colWidthTitle),
			padRight("TYPE", colWidthType),
			padRight("LIST", colWidthList),
			padRight("TAGS", colWidthTags),
			"WHEN")
	}

	for _, task := range tasks {
		// Get tags for this task
		tags, _ := database.GetTaskTags(task.UUID)
		tagStrs := make([]string, len(tags))
		for i, t := range tags {
			tagStrs[i] = t.Title
		}
		tagsStr := strings.Join(tagStrs, ", ")

		// Get list name (project or area), strip emojis
		listName := ""
		if task.Project.Valid {
			if proj, err := database.GetTask(task.Project.String); err == nil {
				listName = stripEmojis(proj.Title)
			}
		} else if task.Area.Valid {
			if area, err := database.GetAreaByUUID(task.Area.String); err == nil {
				listName = stripEmojis(area.Title)
			}
		}

		// Strip emojis from title for clean terminal output
		title := stripEmojis(task.Title)

		if briefOutput {
			// Brief: just ID, title, and tags inline
			if len(tagStrs) > 0 {
				fmt.Fprintf(w, "%s  %s (%s)\n", task.UUID, title, tagsStr)
			} else {
				fmt.Fprintf(w, "%s  %s\n", task.UUID, title)
			}
		} else {
			// Full table format with proper padding
			titleDisplay := truncate(title, colWidthTitle)
			typeDisplay := taskTypeString(task.Type)
			listDisplay := truncate(listName, colWidthList)
			tagsDisplay := truncate(tagsStr, colWidthTags)
			when := formatWhen(task)
			fmt.Fprintf(w, "%s  %s  %s  %s  %s  %s\n",
				padRight(task.UUID, colWidthID),
				padRight(titleDisplay, colWidthTitle),
				padRight(typeDisplay, colWidthType),
				padRight(listDisplay, colWidthList),
				padRight(tagsDisplay, colWidthTags),
				when)
		}
	}

	return nil
}

// outputProjects handles formatting and outputting a list of projects
func outputProjects(w io.Writer, projects []db.Task) error {
	if limitOutput > 0 && len(projects) > limitOutput {
		projects = projects[:limitOutput]
	}

	if jsonOutput {
		return json.NewEncoder(w).Encode(projects)
	}

	if countOnly {
		fmt.Fprintln(w, len(projects))
		return nil
	}

	// Print header
	if !briefOutput {
		fmt.Fprintf(w, "%s  %s  %s\n",
			padRight("ID", colWidthID),
			padRight("TITLE", colWidthTitle),
			"AREA")
	}

	for _, project := range projects {
		title := stripEmojis(project.Title)

		// Get area name
		areaName := ""
		if project.Area.Valid {
			if area, err := database.GetAreaByUUID(project.Area.String); err == nil {
				areaName = stripEmojis(area.Title)
			}
		}

		if briefOutput {
			fmt.Fprintf(w, "%s  %s\n", project.UUID, title)
		} else {
			titleDisplay := truncate(title, colWidthTitle)
			fmt.Fprintf(w, "%s  %s  %s\n",
				padRight(project.UUID, colWidthID),
				padRight(titleDisplay, colWidthTitle),
				areaName)
		}
	}

	return nil
}

// outputAreas handles formatting and outputting a list of areas
func outputAreas(w io.Writer, areas []db.Area) error {
	if limitOutput > 0 && len(areas) > limitOutput {
		areas = areas[:limitOutput]
	}

	if jsonOutput {
		return json.NewEncoder(w).Encode(areas)
	}

	if countOnly {
		fmt.Fprintln(w, len(areas))
		return nil
	}

	// Print header
	if !briefOutput {
		fmt.Fprintf(w, "%s  %s\n",
			padRight("ID", colWidthID),
			"TITLE")
	}

	for _, area := range areas {
		title := stripEmojis(area.Title)

		if briefOutput {
			fmt.Fprintf(w, "%s  %s\n", area.UUID, title)
		} else {
			fmt.Fprintf(w, "%s  %s\n",
				padRight(area.UUID, colWidthID),
				title)
		}
	}

	return nil
}

// outputTags handles formatting and outputting a list of tags
func outputTags(w io.Writer, tags []db.Tag) error {
	if limitOutput > 0 && len(tags) > limitOutput {
		tags = tags[:limitOutput]
	}

	if jsonOutput {
		return json.NewEncoder(w).Encode(tags)
	}

	if countOnly {
		fmt.Fprintln(w, len(tags))
		return nil
	}

	// Print header
	if !briefOutput {
		fmt.Fprintf(w, "%s  %s\n",
			padRight("ID", colWidthID),
			"TITLE")
	}

	for _, tag := range tags {
		if briefOutput {
			fmt.Fprintf(w, "%s  %s\n", tag.UUID, tag.Title)
		} else {
			fmt.Fprintf(w, "%s  %s\n",
				padRight(tag.UUID, colWidthID),
				tag.Title)
		}
	}

	return nil
}

// formatStatus returns a human-readable status string
func formatStatus(status db.TaskStatus) string {
	switch status {
	case db.TaskStatusOpen:
		return "todo"
	case db.TaskStatusCompleted:
		return "done"
	case db.TaskStatusCanceled:
		return "canceled"
	default:
		return "unknown"
	}
}

// outputTasksWithStatus outputs tasks with STATUS column instead of LIST column
// Used for showing tasks within a project
func outputTasksWithStatus(w io.Writer, tasks []db.Task) error {
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

	// Print header
	if !briefOutput {
		fmt.Fprintf(w, "%s  %s  %s  %s  %s\n",
			padRight("ID", colWidthID),
			padRight("TITLE", colWidthTitle),
			padRight("TAGS", colWidthTags),
			padRight("STATUS", 10),
			"WHEN")
	}

	for _, task := range tasks {
		// Get tags for this task
		tags, _ := database.GetTaskTags(task.UUID)
		tagStrs := make([]string, len(tags))
		for i, t := range tags {
			tagStrs[i] = t.Title
		}
		tagsStr := strings.Join(tagStrs, ", ")

		// Strip emojis from title for clean terminal output
		title := stripEmojis(task.Title)
		status := formatStatus(task.Status)

		if briefOutput {
			if len(tagStrs) > 0 {
				fmt.Fprintf(w, "%s  %s (%s) [%s]\n", task.UUID, title, tagsStr, status)
			} else {
				fmt.Fprintf(w, "%s  %s [%s]\n", task.UUID, title, status)
			}
		} else {
			titleDisplay := truncate(title, colWidthTitle)
			tagsDisplay := truncate(tagsStr, colWidthTags)
			when := formatWhen(task)
			fmt.Fprintf(w, "%s  %s  %s  %s  %s\n",
				padRight(task.UUID, colWidthID),
				padRight(titleDisplay, colWidthTitle),
				padRight(tagsDisplay, colWidthTags),
				padRight(status, 10),
				when)
		}
	}

	return nil
}

// outputChecklistItems outputs checklist items as a table
func outputChecklistItems(w io.Writer, items []db.ChecklistItem) error {
	if limitOutput > 0 && len(items) > limitOutput {
		items = items[:limitOutput]
	}

	if jsonOutput {
		return json.NewEncoder(w).Encode(items)
	}

	if countOnly {
		fmt.Fprintln(w, len(items))
		return nil
	}

	// Print header
	if !briefOutput {
		fmt.Fprintf(w, "%s  %s  %s\n",
			padRight("#", 3),
			padRight("STATUS", 10),
			"TITLE")
	}

	for i, item := range items {
		status := formatStatus(item.Status)
		index := fmt.Sprintf("%d", i+1) // 1-based index for user-friendliness
		if briefOutput {
			fmt.Fprintf(w, "%s. [%s] %s\n", index, status, item.Title)
		} else {
			fmt.Fprintf(w, "%s  %s  %s\n",
				padRight(index, 3),
				padRight(status, 10),
				item.Title)
		}
	}

	return nil
}
