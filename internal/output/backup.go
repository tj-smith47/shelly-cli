// Package output provides output formatting utilities for the CLI.
package output

import (
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// FormatBackupsTable builds a table of backup file information.
func FormatBackupsTable(backups []model.BackupFileInfo) *Table {
	table := NewTable("FILENAME", "DEVICE", "MODEL", "CREATED", "ENCRYPTED", "SIZE")
	for _, b := range backups {
		encrypted := ""
		if b.Encrypted {
			encrypted = "yes"
		}
		table.AddRow(
			b.Filename,
			b.DeviceID,
			b.DeviceModel,
			b.CreatedAt.Format("2006-01-02 15:04:05"),
			encrypted,
			FormatSize(b.Size),
		)
	}
	return table
}
