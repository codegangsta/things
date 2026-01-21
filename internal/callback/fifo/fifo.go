// Package fifo provides helper functions for named pipe (FIFO) operations.
package fifo

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
)

// PipeDir is the directory where named pipes are created.
const PipeDir = "/tmp"

// PipePrefix is the prefix for pipe file names.
const PipePrefix = "things-cli-"

// PipeSuffix is the suffix for pipe file names.
const PipeSuffix = ".pipe"

// GenerateID generates a unique identifier for a pipe.
func GenerateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp if crypto/rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// PipePath returns the full path for a pipe with the given ID.
func PipePath(id string) string {
	return fmt.Sprintf("%s/%s%s%s", PipeDir, PipePrefix, id, PipeSuffix)
}

// ReadWithTimeout reads from a named pipe with a timeout.
// The read operation blocks until data is written to the pipe or the timeout expires.
func ReadWithTimeout(pipePath string, timeout time.Duration) ([]byte, error) {
	// Create channels for result and error
	dataCh := make(chan []byte, 1)
	errCh := make(chan error, 1)

	// Start reading in a goroutine (this will block until someone writes)
	go func() {
		// Open the pipe for reading (this blocks until a writer opens it)
		file, err := os.OpenFile(pipePath, os.O_RDONLY, 0)
		if err != nil {
			errCh <- fmt.Errorf("failed to open pipe for reading: %w", err)
			return
		}
		defer file.Close()

		// Read all data from the open file handle
		data, err := io.ReadAll(file)
		if err != nil {
			errCh <- fmt.Errorf("failed to read from pipe: %w", err)
			return
		}
		dataCh <- data
	}()

	// Wait for result or timeout
	select {
	case data := <-dataCh:
		return data, nil
	case err := <-errCh:
		return nil, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for pipe data")
	}
}
