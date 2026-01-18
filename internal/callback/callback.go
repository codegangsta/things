// Package callback provides x-callback-url support for Things URL scheme operations.
// It starts a temporary HTTP server to receive callbacks from Things, enabling
// reliable confirmation of write operations.
package callback

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
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
// It starts a temporary HTTP server, adds x-callback-url parameters to the Things URL,
// opens the URL, and waits for Things to call back with the result.
func ExecuteWithCallback(thingsURL string, timeout time.Duration) (*Result, error) {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	// Start temporary HTTP server on random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Channel to receive the result
	resultCh := make(chan *Result, 1)

	// Create HTTP server with handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		result := &Result{Success: true}

		// Parse x-things-id (single or comma-separated)
		if id := r.URL.Query().Get("x-things-id"); id != "" {
			result.IDs = strings.Split(id, ",")
		}

		// Parse x-things-ids (JSON array)
		if ids := r.URL.Query().Get("x-things-ids"); ids != "" {
			var jsonIDs []string
			if err := json.Unmarshal([]byte(ids), &jsonIDs); err == nil {
				result.IDs = jsonIDs
			}
		}

		w.WriteHeader(http.StatusOK)
		select {
		case resultCh <- result:
		default:
		}
	})

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		result := &Result{
			Success: false,
			Error:   r.URL.Query().Get("errorMessage"),
		}
		if result.Error == "" {
			result.Error = "unknown error from Things"
		}
		w.WriteHeader(http.StatusOK)
		select {
		case resultCh <- result:
		default:
		}
	})

	mux.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
		result := &Result{
			Success: false,
			Error:   "operation cancelled by user",
		}
		w.WriteHeader(http.StatusOK)
		select {
		case resultCh <- result:
		default:
		}
	})

	server := &http.Server{Handler: mux}

	// Start server in goroutine
	go func() {
		_ = server.Serve(listener)
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	// Add callback URLs to the Things URL
	callbackURL, err := addCallbackParams(thingsURL, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to add callback params: %w", err)
	}

	// Open the Things URL
	cmd := exec.Command("open", callbackURL)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to open Things URL: %w", err)
	}

	// Wait for callback or timeout
	select {
	case result := <-resultCh:
		return result, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for Things callback (waited %v)", timeout)
	}
}

// Execute runs a Things URL scheme command without waiting for callbacks (fire-and-forget).
func Execute(thingsURL string) error {
	cmd := exec.Command("open", thingsURL)
	return cmd.Run()
}

// addCallbackParams adds x-success, x-error, and x-cancel parameters to a Things URL.
func addCallbackParams(thingsURL, baseURL string) (string, error) {
	u, err := url.Parse(thingsURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("x-success", baseURL+"/success")
	q.Set("x-error", baseURL+"/error")
	q.Set("x-cancel", baseURL+"/cancel")
	u.RawQuery = q.Encode()

	// Things expects %20 instead of + for spaces
	result := u.String()
	result = strings.ReplaceAll(result, "+", "%20")

	return result, nil
}
