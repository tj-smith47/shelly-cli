// Package model provides domain types for the shelly-cli.
package model

// AuditResult holds the results of a device security audit.
type AuditResult struct {
	Device     string
	Address    string
	Issues     []string
	Warnings   []string
	InfoItems  []string
	Reachable  bool
	AuthStatus *AuthAudit
	CloudAudit *CloudAudit
	FWAudit    *FirmwareAudit
}

// AuthAudit holds authentication audit results.
type AuthAudit struct {
	AuthEnabled bool
}

// CloudAudit holds cloud audit results.
type CloudAudit struct {
	Connected bool
}

// FirmwareAudit holds firmware audit results.
type FirmwareAudit struct {
	Current   string
	Available string
	HasUpdate bool
}
