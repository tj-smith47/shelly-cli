// Package tuierrors provides error categorization for TUI components.
package tuierrors

import (
	"github.com/tj-smith47/shelly-cli/internal/errutil"
)

// Category represents the type of error for display purposes.
type Category int

const (
	// CategoryUnknown is an uncategorized error.
	CategoryUnknown Category = iota
	// CategoryNetwork indicates a network connectivity issue.
	CategoryNetwork
	// CategoryTimeout indicates a timeout occurred.
	CategoryTimeout
	// CategoryAuth indicates an authentication failure.
	CategoryAuth
	// CategoryDevice indicates a device-side error.
	CategoryDevice
	// CategoryUnsupported indicates a feature not supported on the device.
	CategoryUnsupported
)

// CategorizedError wraps an error with a category and user-friendly message.
type CategorizedError struct {
	Category Category
	Original error
	Message  string
	Hint     string
}

// Error implements the error interface.
func (e CategorizedError) Error() string {
	return e.Message
}

// Unwrap returns the original error.
func (e CategorizedError) Unwrap() error {
	return e.Original
}

// Categorize analyzes an error and returns a categorized version with user-friendly messaging.
func Categorize(err error) CategorizedError {
	if err == nil {
		return CategorizedError{}
	}

	// Use shared detection, then apply TUI-specific messaging
	cat := errutil.Categorize(err)

	switch cat {
	case errutil.CategoryTimeout:
		return CategorizedError{
			Category: CategoryTimeout,
			Original: err,
			Message:  "Request timed out",
			Hint:     "Device may be unresponsive. Check if device is powered on.",
		}

	case errutil.CategoryDNS, errutil.CategoryNetwork:
		// TUI doesn't distinguish DNS from other network errors
		return CategorizedError{
			Category: CategoryNetwork,
			Original: err,
			Message:  "Network error",
			Hint:     "Check device connectivity and network settings.",
		}

	case errutil.CategoryAuth:
		return CategorizedError{
			Category: CategoryAuth,
			Original: err,
			Message:  "Authentication failed",
			Hint:     "Check device credentials in configuration.",
		}

	case errutil.CategoryUnsupported:
		return CategorizedError{
			Category: CategoryUnsupported,
			Original: err,
			Message:  "Feature not supported",
			Hint:     "This feature may not be available on this device.",
		}

	case errutil.CategoryDevice:
		return CategorizedError{
			Category: CategoryDevice,
			Original: err,
			Message:  "Device error: " + err.Error(),
			Hint:     "The device reported an error. Check device status.",
		}

	default:
		// Unknown error - show original message
		return CategorizedError{
			Category: CategoryUnknown,
			Original: err,
			Message:  "Error: " + err.Error(),
			Hint:     "An unexpected error occurred.",
		}
	}
}

// IsUnsupportedFeature checks if an error indicates a feature not supported on the device.
// This detects Gen1 devices or devices lacking specific capabilities based on common error patterns.
//
// Common indicators:
//   - "404" - HTTP not found
//   - "unknown method" - Method not implemented
//   - "not found" - Resource not found
func IsUnsupportedFeature(err error) bool {
	return errutil.IsUnsupported(err)
}

// UnsupportedMessage returns a standardized message for features not supported on a device.
//
// Example:
//
//	tuierrors.UnsupportedMessage("Scripts") // "Scripts not supported on this device"
func UnsupportedMessage(featureName string) string {
	return featureName + " not supported on this device"
}

// FormatError returns a formatted error string with category-specific messaging.
func FormatError(err error) (message, hint string) {
	cat := Categorize(err)
	return cat.Message, cat.Hint
}

// FormatErrorMsg returns just the categorized error message without the hint.
// Use this when you only need the message and don't want to show additional hints.
func FormatErrorMsg(err error) string {
	cat := Categorize(err)
	return cat.Message
}
