// Package model provides domain types for the shelly-cli.
package model

// BatchRPCResult holds the result of a batch RPC operation for a single device.
// Used for JSON/YAML output of batch command results.
type BatchRPCResult struct {
	Device   string `json:"device" yaml:"device"`
	Response any    `json:"response,omitempty" yaml:"response,omitempty"`
	Error    string `json:"error,omitempty" yaml:"error,omitempty"`
}
