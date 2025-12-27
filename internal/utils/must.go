// Package utils provides common functionality shared across CLI commands.
package utils

// Must panics if err is not nil.
// Use for errors that indicate programming bugs, not runtime errors.
// Common use: binding flags to viper in init().
func Must(err error) {
	if err != nil {
		panic(err)
	}
}
