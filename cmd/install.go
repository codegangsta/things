package cmd

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed bundle/*
var bundleFS embed.FS

var installCmd = &cobra.Command{
	Use:   "install-handler",
	Short: "Install the URL callback handler for Things CLI",
	Long: `Install the ThingsCLICallback.app URL handler.

This app handles the things-cli:// URL scheme callbacks from Things 3,
enabling reliable confirmation of write operations without opening a browser.

The app is installed to ~/Applications and registered with Launch Services.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return installHandler()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func installHandler() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	installDir := filepath.Join(homeDir, "Applications")
	appPath := filepath.Join(installDir, "ThingsCLICallback.app")

	// Create install directory if needed
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Create temp directory for bundle files
	tempDir, err := os.MkdirTemp("", "things-cli-bundle-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract bundle files to temp directory
	infoPlist, err := bundleFS.ReadFile("bundle/Info.plist")
	if err != nil {
		return fmt.Errorf("failed to read Info.plist: %w", err)
	}

	applescript, err := bundleFS.ReadFile("bundle/url-handler.applescript")
	if err != nil {
		return fmt.Errorf("failed to read url-handler.applescript: %w", err)
	}

	// Write files to temp directory
	infoPlistPath := filepath.Join(tempDir, "Info.plist")
	if err := os.WriteFile(infoPlistPath, infoPlist, 0644); err != nil {
		return fmt.Errorf("failed to write Info.plist: %w", err)
	}

	applescriptPath := filepath.Join(tempDir, "url-handler.applescript")
	if err := os.WriteFile(applescriptPath, applescript, 0644); err != nil {
		return fmt.Errorf("failed to write url-handler.applescript: %w", err)
	}

	// Remove existing app if present
	if _, err := os.Stat(appPath); err == nil {
		fmt.Println("Removing existing installation...")
		if err := os.RemoveAll(appPath); err != nil {
			return fmt.Errorf("failed to remove existing app: %w", err)
		}
	}

	// Compile AppleScript to app bundle
	fmt.Println("Compiling AppleScript handler...")
	compileCmd := exec.Command("osacompile", "-o", appPath, applescriptPath)
	if output, err := compileCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to compile AppleScript: %w\n%s", err, output)
	}

	// Copy Info.plist (overwrites default)
	fmt.Println("Configuring app bundle...")
	destInfoPlist := filepath.Join(appPath, "Contents", "Info.plist")
	if err := os.WriteFile(destInfoPlist, infoPlist, 0644); err != nil {
		return fmt.Errorf("failed to write Info.plist to app bundle: %w", err)
	}

	// Register with Launch Services
	fmt.Println("Registering URL handler...")
	lsregister := "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"
	registerCmd := exec.Command(lsregister, "-f", appPath)
	if output, err := registerCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to register URL handler: %w\n%s", err, output)
	}

	fmt.Println("")
	fmt.Println("Installation complete!")
	fmt.Println("The things-cli:// URL scheme is now registered.")
	fmt.Printf("Installed to: %s\n", appPath)

	return nil
}
