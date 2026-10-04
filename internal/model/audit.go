// Package model provides domain types for the shelly-cli.
package model

// AuditResult holds the results of a device security audit.
type AuditResult struct {
	Device     string         `json:"device" yaml:"device"`
	Address    string         `json:"address" yaml:"address"`
	Issues     []string       `json:"issues" yaml:"issues"`
	Warnings   []string       `json:"warnings" yaml:"warnings"`
	InfoItems  []string       `json:"info" yaml:"info"`
	Reachable  bool           `json:"reachable" yaml:"reachable"`
	AuthStatus *AuthAudit     `json:"auth,omitempty" yaml:"auth,omitempty"`
	CloudAudit *CloudAudit    `json:"cloud,omitempty" yaml:"cloud,omitempty"`
	FWAudit    *FirmwareAudit `json:"firmware,omitempty" yaml:"firmware,omitempty"`
}

// AuthAudit holds authentication audit results.
type AuthAudit struct {
	AuthEnabled bool `json:"enabled" yaml:"enabled"`
}

// CloudAudit holds cloud audit results.
type CloudAudit struct {
	Connected bool `json:"connected" yaml:"connected"`
}

// FirmwareAudit holds firmware audit results.
type FirmwareAudit struct {
	Current   string `json:"current" yaml:"current"`
	Available string `json:"available,omitempty" yaml:"available,omitempty"`
	HasUpdate bool   `json:"has_update" yaml:"has_update"`
}

// AuditChecks selects which checks a security audit runs.
type AuditChecks struct {
	// Auth reports whether the device requires a password.
	Auth bool
	// Cloud reports whether the device is connected to Shelly Cloud, and flags
	// a cloud connection on a device without a password.
	Cloud bool
	// Firmware reports whether a firmware update is available.
	Firmware bool
}

// AllAuditChecks returns a selection with every check enabled.
func AllAuditChecks() AuditChecks {
	return AuditChecks{Auth: true, Cloud: true, Firmware: true}
}

// Any reports whether at least one check is selected.
func (c AuditChecks) Any() bool {
	return c.Auth || c.Cloud || c.Firmware
}
