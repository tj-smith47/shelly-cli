// Package tuierrors provides error categorization for TUI components.
package tuierrors

import (
	"context"
	"errors"
	"net"
	"strings"
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

	errStr := strings.ToLower(err.Error())

	if cat := checkTimeout(err, errStr); cat != nil {
		return *cat
	}

	if cat := checkNetwork(err, errStr); cat != nil {
		return *cat
	}

	if cat := checkAuth(err, errStr); cat != nil {
		return *cat
	}

	if cat := checkDevice(err, errStr); cat != nil {
		return *cat
	}

	// Unknown error - show original message
	return CategorizedError{
		Category: CategoryUnknown,
		Original: err,
		Message:  "Error: " + err.Error(),
		Hint:     "An unexpected error occurred.",
	}
}

func checkTimeout(err error, errStr string) *CategorizedError {
	if errors.Is(err, context.DeadlineExceeded) ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") ||
		strings.Contains(errStr, "timed out") {
		return &CategorizedError{
			Category: CategoryTimeout,
			Original: err,
			Message:  "Request timed out",
			Hint:     "Device may be unresponsive. Check if device is powered on.",
		}
	}
	return nil
}

func checkNetwork(err error, errStr string) *CategorizedError {
	var netErr net.Error
	if errors.As(err, &netErr) ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no route to host") ||
		strings.Contains(errStr, "network unreachable") ||
		strings.Contains(errStr, "host unreachable") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "dial tcp") ||
		strings.Contains(errStr, "i/o timeout") {
		return &CategorizedError{
			Category: CategoryNetwork,
			Original: err,
			Message:  "Network error",
			Hint:     "Check device connectivity and network settings.",
		}
	}
	return nil
}

func checkAuth(err error, errStr string) *CategorizedError {
	if strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "invalid password") ||
		strings.Contains(errStr, "access denied") {
		return &CategorizedError{
			Category: CategoryAuth,
			Original: err,
			Message:  "Authentication failed",
			Hint:     "Check device credentials in configuration.",
		}
	}
	return nil
}

func checkDevice(err error, errStr string) *CategorizedError {
	if strings.Contains(errStr, "device") ||
		strings.Contains(errStr, "shelly") ||
		strings.Contains(errStr, "component") {
		return &CategorizedError{
			Category: CategoryDevice,
			Original: err,
			Message:  "Device error: " + err.Error(),
			Hint:     "The device reported an error. Check device status.",
		}
	}
	return nil
}

// FormatError returns a formatted error string with category-specific messaging.
func FormatError(err error) (message, hint string) {
	cat := Categorize(err)
	return cat.Message, cat.Hint
}
