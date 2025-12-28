// Package browser provides utilities for opening URLs in the default web browser.
package browser

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/atotto/clipboard"
)

// Browser is an interface for opening URLs in a web browser.
type Browser interface {
	// Browse opens a URL in the default web browser.
	Browse(ctx context.Context, url string) error
	// OpenDeviceUI opens a Shelly device's web interface by IP address.
	OpenDeviceUI(ctx context.Context, deviceIP string) error
	// CopyToClipboard copies a URL to the system clipboard.
	CopyToClipboard(url string) error
}

// ClipboardFallbackError indicates the URL was copied to clipboard
// because the browser could not be opened.
type ClipboardFallbackError struct {
	URL string
}

// Error implements the error interface.
func (e *ClipboardFallbackError) Error() string {
	return fmt.Sprintf("URL copied to clipboard: %s", e.URL)
}

// browserImpl is the default implementation of Browser.
type browserImpl struct{}

// New creates a new Browser instance.
func New() Browser {
	return &browserImpl{}
}

// Browse opens a URL in the default web browser.
// It detects the operating system and uses the appropriate command.
// If the browser cannot be opened, it copies the URL to the clipboard
// and returns a ClipboardFallbackError.
func (b *browserImpl) Browse(ctx context.Context, url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", url)
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", url)
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", url)
	default:
		// Unsupported OS - try clipboard fallback
		if err := clipboard.WriteAll(url); err == nil {
			return &ClipboardFallbackError{URL: url}
		}
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err := cmd.Start(); err != nil {
		// Browser failed - try clipboard fallback
		if clipErr := clipboard.WriteAll(url); clipErr == nil {
			return &ClipboardFallbackError{URL: url}
		}
		return fmt.Errorf("failed to open browser and clipboard unavailable: %w", err)
	}
	return nil
}

// OpenDeviceUI opens a Shelly device's web interface by IP address.
func (b *browserImpl) OpenDeviceUI(ctx context.Context, deviceIP string) error {
	url := fmt.Sprintf("http://%s", deviceIP)
	return b.Browse(ctx, url)
}

// CopyToClipboard copies a URL to the system clipboard.
func (b *browserImpl) CopyToClipboard(url string) error {
	return clipboard.WriteAll(url)
}
