// Package download provides HTTP download utilities for the CLI.
package download

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Result holds the result of downloading a file.
type Result struct {
	LocalPath string
	Cleanup   func()
}

// FromURL downloads a file from a URL to a temp file.
// Returns the path to the downloaded file and a cleanup function.
func FromURL(ctx context.Context, downloadURL string) (*Result, error) {
	fs := config.Fs()

	// Create temp file
	tmpFile, err := afero.TempFile(fs, "", "shelly-download-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	cleanup := func() {
		// Best effort cleanup - ignore errors since we can't report them
		_ = fs.Remove(tmpPath)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "shelly-cli")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // response body close in defer

	if resp.StatusCode != http.StatusOK {
		cleanup()
		return nil, fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if cerr := tmpFile.Close(); cerr != nil && err == nil {
		err = cerr
	}

	if err != nil {
		cleanup()
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Make executable - extensions must be executable binaries
	if err := fs.Chmod(tmpPath, 0o755); err != nil { //nolint:gosec // G302: extensions require executable permissions
		cleanup()
		return nil, fmt.Errorf("failed to make executable: %w", err)
	}

	return &Result{
		LocalPath: tmpPath,
		Cleanup:   cleanup,
	}, nil
}
