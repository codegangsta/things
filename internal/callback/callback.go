// Package callback provides x-callback-url support for Things URL scheme operations.
// It uses named pipes (FIFOs) and an AppleScript URL handler for IPC, enabling
// reliable confirmation of write operations without HTTP servers or browsers.
package callback

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/codegangsta/things/internal/callback/fifo"
)

// Result represents the outcome of a Things URL scheme operation.
type Result struct {
	Success bool     // Whether the operation succeeded
	IDs     []string // IDs returned by Things (task IDs, project IDs, etc.)
	Error   string   // Error message if operation failed
}

// DefaultTimeout is the default time to wait for a callback response.
const DefaultTimeout = 5 * time.Second

// ExecuteWithCallback executes a Things URL scheme command and waits for a callback.
// It creates a named pipe (FIFO), adds x-callback-url parameters pointing to the
// things-cli:// URL scheme, opens the URL, and waits for the AppleScript handler
// to write the callback result to the pipe.
func ExecuteWithCallback(thingsURL string, timeout time.Duration) (*Result, error) {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	// Generate unique pipe ID and path
	pipeID := fifo.GenerateID()
	pipePath := fifo.PipePath(pipeID)

	// Create named pipe (FIFO)
	if err := syscall.Mkfifo(pipePath, 0600); err != nil {
		return nil, fmt.Errorf("failed to create pipe: %w", err)
	}
	defer os.Remove(pipePath)

	// Add callback params with pipe ID
	callbackURL, err := addCallbackParams(thingsURL, pipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to add callback params: %w", err)
	}

	// Open Things URL
	cmd := exec.Command("open", callbackURL)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to open Things URL: %w", err)
	}

	// Read from pipe with timeout (blocks until AppleScript writes)
	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		data, err := fifo.ReadWithTimeout(pipePath, timeout)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- string(data)
	}()

	select {
	case urlStr := <-resultCh:
		return parseCallbackURL(urlStr)
	case err := <-errCh:
		return nil, err
	case <-time.After(timeout + 100*time.Millisecond):
		return nil, fmt.Errorf("timeout waiting for Things callback (waited %v)", timeout)
	}
}

// Execute runs a Things URL scheme command without waiting for callbacks (fire-and-forget).
func Execute(thingsURL string) error {
	cmd := exec.Command("open", thingsURL)
	return cmd.Run()
}

// addCallbackParams adds x-success, x-error, and x-cancel parameters to a Things URL.
func addCallbackParams(thingsURL string, pipeID string) (string, error) {
	u, err := url.Parse(thingsURL)
	if err != nil {
		return "", err
	}

	// Build callback URLs using things-cli:// scheme
	successURL := fmt.Sprintf("things-cli://success?pipe=%s", pipeID)
	errorURL := fmt.Sprintf("things-cli://error?pipe=%s", pipeID)
	cancelURL := fmt.Sprintf("things-cli://cancel?pipe=%s", pipeID)

	q := u.Query()
	q.Set("x-success", successURL)
	q.Set("x-error", errorURL)
	q.Set("x-cancel", cancelURL)
	u.RawQuery = q.Encode()

	// Things expects %20 instead of + for spaces
	result := u.String()
	result = strings.ReplaceAll(result, "+", "%20")

	return result, nil
}

// parseCallbackURL parses a callback URL from the AppleScript handler.
// URL format: things-cli://success?pipe=UUID&x-things-id=...
// or: things-cli://error?pipe=UUID&errorMessage=...
func parseCallbackURL(urlStr string) (*Result, error) {
	urlStr = strings.TrimSpace(urlStr)

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse callback URL: %w", err)
	}

	result := &Result{}

	// Determine success/error/cancel from host (path in URL scheme)
	switch u.Host {
	case "success":
		result.Success = true

		// Parse x-things-id (single or comma-separated)
		if id := u.Query().Get("x-things-id"); id != "" {
			result.IDs = strings.Split(id, ",")
		}

		// Parse x-things-ids (JSON array)
		if ids := u.Query().Get("x-things-ids"); ids != "" {
			var jsonIDs []string
			if err := json.Unmarshal([]byte(ids), &jsonIDs); err == nil {
				result.IDs = jsonIDs
			}
		}

	case "error":
		result.Success = false
		result.Error = u.Query().Get("errorMessage")
		if result.Error == "" {
			result.Error = "unknown error from Things"
		}

	case "cancel":
		result.Success = false
		result.Error = "operation cancelled by user"

	default:
		return nil, fmt.Errorf("unknown callback type: %s", u.Host)
	}

	return result, nil
}
