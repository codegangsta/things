package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/things/internal/callback"
	"github.com/spf13/cobra"
)

var completeCmd = &cobra.Command{
	Use:   "complete [task-id]",
	Short: "Mark a task as complete",
	Long: `Mark a task as complete in Things 3 using the URL scheme.

You can find task IDs using the other commands with --json flag.

Requires auth-token to be set. Get your token from:
Things → Settings → General → Enable Things URLs → Manage

Set it via: things auth <token>

Examples:
  things complete ABC123
  things complete ABC123 DEF456  # Complete multiple tasks`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAuthToken()
		if err != nil {
			return fmt.Errorf("auth token required: run 'things auth <token>' first.\nGet your token from Things → Settings → General → Enable Things URLs → Manage")
		}

		for _, id := range args {
			if err := completeTask(id, token); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)
}

func getAuthTokenPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "things", "auth-token"), nil
}

func getAuthToken() (string, error) {
	path, err := getAuthTokenPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func completeTask(id, token string) error {
	// Things URL scheme for completing a task (requires auth-token)
	thingsURL := fmt.Sprintf("things:///update?id=%s&auth-token=%s&completed=true", id, token)

	// Fire-and-forget mode
	if noWait {
		if err := callback.Execute(thingsURL); err != nil {
			return fmt.Errorf("failed to complete task %s: %w", id, err)
		}
		fmt.Printf("Completed: %s\n", id)
		return nil
	}

	// Wait for callback confirmation
	result, err := callback.ExecuteWithCallback(thingsURL, callbackTimeout)
	if err != nil {
		return fmt.Errorf("failed to complete task %s: %w", id, err)
	}

	if !result.Success {
		return fmt.Errorf("failed to complete task %s: %s", id, result.Error)
	}

	fmt.Printf("Completed: %s\n", id)
	return nil
}
