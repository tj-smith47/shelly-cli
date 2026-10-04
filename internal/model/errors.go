// Package model defines core domain types for the Shelly CLI.
package model

import "errors"

// Domain errors.
var (
	// ErrDeviceNotFound indicates the device was not found in the registry.
	ErrDeviceNotFound = errors.New("device not found")

	// ErrDeviceExists indicates a device with the same name already exists.
	ErrDeviceExists = errors.New("device already exists")

	// ErrConnectionFailed indicates failure to connect to the device.
	ErrConnectionFailed = errors.New("failed to connect to device")

	// ErrComponentNotFound indicates the requested component was not found.
	ErrComponentNotFound = errors.New("component not found")

	// ErrInvalidDeviceName indicates the device name is invalid.
	ErrInvalidDeviceName = errors.New("invalid device name")

	// ErrAuthRequired indicates the device requires authentication.
	ErrAuthRequired = errors.New("device requires authentication")

	// ErrTimeout indicates the operation timed out.
	ErrTimeout = errors.New("operation timed out")

	// ErrNoDevices indicates no devices were found.
	ErrNoDevices = errors.New("no devices found")

	// ErrNoComponents indicates no matching components were found.
	ErrNoComponents = errors.New("no matching components found")
)

// ErrCredentialsRejected indicates a device refused the credentials it was
// given. Match it with errors.Is; the error itself is a
// *CredentialsRejectedError naming the user.
var ErrCredentialsRejected = errors.New("device rejected the credentials")

// CredentialsRejectedError reports that a device answered an authenticated
// request with "unauthorized" although credentials were sent.
type CredentialsRejectedError struct {
	// User is the user name the rejected password was sent for.
	User string
}

// Error implements the error interface.
func (e *CredentialsRejectedError) Error() string {
	return "device rejected the password for user " + e.User
}

// Is reports a match for ErrCredentialsRejected.
func (e *CredentialsRejectedError) Is(target error) bool {
	return target == ErrCredentialsRejected
}
