package cmd

import (
	"fmt"

	"github.com/codegangsta/things/internal/callback"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [task-id...]",
	Short: "Move a task to the trash",
	Long: `Move one or more tasks to the Trash in Things 3.

This uses the URL scheme to set canceled=true, which moves the task to Trash.
Tasks in Trash can be restored from within Things 3.

You can find task IDs using the other commands with --json flag.

Requires auth-token to be set. Get your token from:
Things → Settings → General → Enable Things URLs → Manage

Set it via: things auth <token>

Examples:
  things delete ABC123
  things delete ABC123 DEF456  # Delete multiple tasks`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAuthToken()
		if err != nil {
			return fmt.Errorf("auth token required: run 'things auth <token>' first.\nGet your token from Things → Settings → General → Enable Things URLs → Manage")
		}

		for _, id := range args {
			if err := deleteTask(id, token); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func deleteTask(id, token string) error {
	// Things URL scheme: setting canceled=true moves to trash
	thingsURL := fmt.Sprintf("things:///update?id=%s&auth-token=%s&canceled=true", id, token)

	// Fire-and-forget mode
	if noWait {
		if err := callback.Execute(thingsURL); err != nil {
			return fmt.Errorf("failed to delete task %s: %w", id, err)
		}
		fmt.Printf("Deleted: %s\n", id)
		return nil
	}

	// Wait for callback confirmation
	result, err := callback.ExecuteWithCallback(thingsURL, callbackTimeout)
	if err != nil {
		return fmt.Errorf("failed to delete task %s: %w", id, err)
	}

	if !result.Success {
		return fmt.Errorf("failed to delete task %s: %s", id, result.Error)
	}

	fmt.Printf("Deleted: %s\n", id)
	return nil
}
