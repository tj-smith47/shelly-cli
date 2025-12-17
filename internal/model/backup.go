// Package model provides domain types for the CLI.
package model

import "time"

// BackupFileInfo holds information about a backup file.
type BackupFileInfo struct {
	Filename    string    `json:"filename"`
	DeviceID    string    `json:"device_id"`
	DeviceModel string    `json:"device_model"`
	FWVersion   string    `json:"fw_version"`
	CreatedAt   time.Time `json:"created_at"`
	Encrypted   bool      `json:"encrypted"`
	Size        int64     `json:"size"`
}
