package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth [token]",
	Short: "Set the Things URL scheme auth token",
	Long: `Set the auth token required for write operations (complete, update, etc).

Get your token from:
Things → Settings → General → Enable Things URLs → Manage

The token is stored in ~/.config/things/auth-token

Examples:
  things auth abc123def456`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return saveAuthToken(args[0])
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}

func saveAuthToken(token string) error {
	path, err := getAuthTokenPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write token with restricted permissions
	if err := os.WriteFile(path, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println("Auth token saved successfully")
	return nil
}
