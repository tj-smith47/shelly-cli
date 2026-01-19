// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// EnhanceDeviceError improves error messages for device-related failures.
// It detects DNS lookup failures and suggests similar device names from the config.
func EnhanceDeviceError(err error, device string) error {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	// Check for DNS lookup failures
	if isDNSError(errStr) {
		return enhanceWithSuggestions(device)
	}

	return err
}

// isDNSError checks if the error message indicates a DNS lookup failure.
func isDNSError(errStr string) bool {
	return strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "server misbehaving") ||
		(strings.Contains(errStr, "lookup") && strings.Contains(errStr, "dial tcp"))
}

// enhanceWithSuggestions adds device name suggestions to the error message.
func enhanceWithSuggestions(device string) error {
	similar := findSimilarDevices(device, 3)
	if len(similar) == 0 {
		return fmt.Errorf("device %q not found (DNS lookup failed)\n"+
			"Check the device name for typos, or use 'shelly discover' to find devices", device)
	}

	return fmt.Errorf("device %q not found (DNS lookup failed)\n"+
		"Did you mean: %s?", device, strings.Join(similar, ", "))
}

// findSimilarDevices finds device names that are similar to the given name.
// Uses simple substring matching and Levenshtein-like heuristics.
func findSimilarDevices(name string, maxResults int) []string {
	devices := config.ListDevices()
	if len(devices) == 0 {
		return nil
	}

	nameLower := strings.ToLower(name)
	var matches []string

	for key, dev := range devices {
		// Check device key
		keyLower := strings.ToLower(key)
		if isSimilar(nameLower, keyLower) {
			matches = append(matches, key)
			continue
		}

		// Check device display name
		displayLower := strings.ToLower(dev.Name)
		if displayLower != keyLower && isSimilar(nameLower, displayLower) {
			matches = append(matches, dev.Name)
		}
	}

	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}
	return matches
}

// isSimilar checks if two strings are similar enough to suggest.
// Uses substring matching and edit distance heuristics.
func isSimilar(a, b string) bool {
	// Exact substring match
	if strings.Contains(a, b) || strings.Contains(b, a) {
		return true
	}

	// Check if strings share a common prefix of at least 3 characters
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	if minLen >= 3 {
		prefixLen := 0
		for i := range minLen {
			if a[i] == b[i] {
				prefixLen++
			} else {
				break
			}
		}
		if prefixLen >= 3 {
			return true
		}
	}

	// Simple edit distance check - if lengths differ by at most 2 and
	// most characters match, consider them similar
	lenDiff := len(a) - len(b)
	if lenDiff < 0 {
		lenDiff = -lenDiff
	}
	if lenDiff <= 2 && len(a) >= 3 {
		matches := 0
		for i := range a {
			if i < len(b) && a[i] == b[i] {
				matches++
			}
		}
		// If more than 70% of characters match, consider similar
		if float64(matches)/float64(len(a)) > 0.7 {
			return true
		}
	}

	return false
}
