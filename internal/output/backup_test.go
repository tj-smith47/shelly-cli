package output

import (
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestFormatBackupsTable(t *testing.T) {
	t.Parallel()

	t.Run("empty backups", func(t *testing.T) {
		t.Parallel()
		table := FormatBackupsTable(nil)
		if !table.Empty() {
			t.Error("expected empty table for nil backups")
		}
	})

	t.Run("with backups", func(t *testing.T) {
		t.Parallel()
		backups := []model.BackupFileInfo{
			{
				Filename:    "backup1.json",
				DeviceID:    "device-001",
				DeviceModel: "SHSW-1",
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Encrypted:   false,
				Size:        1024,
			},
			{
				Filename:    "backup2.json",
				DeviceID:    "device-002",
				DeviceModel: "SHSW-2",
				CreatedAt:   time.Date(2024, 1, 16, 11, 30, 0, 0, time.UTC),
				Encrypted:   true,
				Size:        2048,
			},
		}

		table := FormatBackupsTable(backups)
		if table.Empty() {
			t.Error("expected non-empty table")
		}
		if table.RowCount() != 2 {
			t.Errorf("RowCount() = %d, want 2", table.RowCount())
		}
	})
}
