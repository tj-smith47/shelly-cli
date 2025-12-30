package connection

import (
	"errors"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// isConnectionError returns true if the error indicates a connection failure
// that could be resolved by trying a different IP address.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, model.ErrConnectionFailed) {
		return true
	}
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no route to host") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "network is unreachable") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "dial tcp")
}
