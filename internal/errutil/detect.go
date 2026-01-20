// Package errutil provides shared error detection and categorization.
// This package contains low-level detection functions used by both CLI and TUI.
// Rendering and UX-specific handling remain in their respective packages.
package errutil

import (
	"context"
	"errors"
	"net"
	"strings"
)

// Category represents the type of error for categorization purposes.
type Category int

const (
	// CategoryUnknown is an uncategorized error.
	CategoryUnknown Category = iota
	// CategoryTimeout indicates a timeout occurred.
	CategoryTimeout
	// CategoryDNS indicates a DNS lookup failure (subset of network).
	CategoryDNS
	// CategoryNetwork indicates a general network connectivity issue.
	CategoryNetwork
	// CategoryAuth indicates an authentication failure.
	CategoryAuth
	// CategoryDevice indicates a device-side error.
	CategoryDevice
	// CategoryUnsupported indicates a feature not supported on the device.
	CategoryUnsupported
)

// String returns a human-readable category name.
func (c Category) String() string {
	switch c {
	case CategoryTimeout:
		return "timeout"
	case CategoryDNS:
		return "dns"
	case CategoryNetwork:
		return "network"
	case CategoryAuth:
		return "auth"
	case CategoryDevice:
		return "device"
	case CategoryUnsupported:
		return "unsupported"
	default:
		return "unknown"
	}
}

// Categorize returns the category of an error.
// It checks in order of specificity: timeout, DNS, network, auth, unsupported, device.
func Categorize(err error) Category {
	if err == nil {
		return CategoryUnknown
	}

	// Check in order of specificity
	if IsTimeout(err) {
		return CategoryTimeout
	}
	if IsDNS(err) {
		return CategoryDNS
	}
	if IsNetwork(err) {
		return CategoryNetwork
	}
	if IsAuth(err) {
		return CategoryAuth
	}
	if IsUnsupported(err) {
		return CategoryUnsupported
	}
	if IsDevice(err) {
		return CategoryDevice
	}

	return CategoryUnknown
}

// IsTimeout checks if err indicates a timeout.
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") ||
		strings.Contains(errStr, "timed out")
}

// IsDNS checks if err indicates a DNS lookup failure.
// This is a specific subset of network errors related to name resolution.
func IsDNS(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "server misbehaving") ||
		(strings.Contains(errStr, "lookup") && strings.Contains(errStr, "dial tcp"))
}

// IsNetwork checks if err indicates a general network connectivity issue.
// Note: DNS errors are a subset of network errors, so check IsDNS first if you need
// to distinguish between them.
func IsNetwork(err error) bool {
	if err == nil {
		return false
	}

	// Check for structured net.Error type
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no route to host") ||
		strings.Contains(errStr, "network unreachable") ||
		strings.Contains(errStr, "host unreachable") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "dial tcp") ||
		strings.Contains(errStr, "i/o timeout")
}

// IsAuth checks if err indicates an authentication failure.
func IsAuth(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "invalid password") ||
		strings.Contains(errStr, "access denied")
}

// IsDevice checks if err indicates a device-side error.
// This matches errors that mention "device", "shelly", or "component".
func IsDevice(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "device") ||
		strings.Contains(errStr, "shelly") ||
		strings.Contains(errStr, "component")
}

// IsUnsupported checks if err indicates a feature not supported on the device.
// This detects Gen1 devices or devices lacking specific capabilities.
func IsUnsupported(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "404") ||
		strings.Contains(errStr, "unknown method") ||
		strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "not supported")
}
