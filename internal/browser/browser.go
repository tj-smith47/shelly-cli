// Package browser provides utilities for opening URLs in the default web browser.
package browser

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

// Browser is an interface for opening URLs in a web browser.
type Browser interface {
	// Browse opens a URL in the default web browser.
	Browse(ctx context.Context, url string) error
}

// browserImpl is the default implementation of Browser.
type browserImpl struct{}

// New creates a new Browser instance.
func New() Browser {
	return &browserImpl{}
}

// Browse opens a URL in the default web browser.
// It detects the operating system and uses the appropriate command.
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
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// OpenDeviceUI opens a Shelly device's web interface by IP address.
// It constructs the HTTP URL and opens it in the browser.
func OpenDeviceUI(ctx context.Context, browser Browser, deviceIP string) error {
	url := fmt.Sprintf("http://%s", deviceIP)
	return browser.Browse(ctx, url)
}
