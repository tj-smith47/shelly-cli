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
