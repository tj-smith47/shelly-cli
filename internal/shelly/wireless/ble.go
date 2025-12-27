// Package wireless provides wireless protocol operations for Shelly devices.
package wireless

import (
	"errors"
	"strings"

	"github.com/tj-smith47/shelly-go/discovery"
)

// IsBLENotSupportedError checks if the error is due to BLE not being supported.
func IsBLENotSupportedError(err error) bool {
	if err == nil {
		return false
	}
	// Check for the BLE not supported error
	var bleErr *discovery.BLEError
	if errors.As(err, &bleErr) {
		// Check if this is the sentinel error or has similar message
		if bleErr == discovery.ErrBLENotSupported {
			return true
		}
		// Check if the error message indicates BLE is not supported
		if strings.Contains(bleErr.Message, "not supported") ||
			strings.Contains(bleErr.Message, "BLE not supported") {
			return true
		}
		// Check if wrapped error is ErrBLENotSupported
		return errors.Is(bleErr.Err, discovery.ErrBLENotSupported)
	}
	return errors.Is(err, discovery.ErrBLENotSupported)
}
