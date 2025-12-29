package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

func TestDisplayMigrationPreview_NoDifferences(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diff := &model.BackupDiff{}
	DisplayMigrationPreview(ios, "source-device", "device", "target-device", diff)

	output := out.String()
	if !strings.Contains(output, "Migration Preview") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "dry run") {
		t.Error("expected dry run notice")
	}
	if !strings.Contains(output, "source-device") {
		t.Error("expected source")
	}
	if !strings.Contains(output, "target-device") {
		t.Error("expected target")
	}
	if !strings.Contains(output, "No differences found") {
		t.Error("expected no differences message")
	}
}

func TestDisplayMigrationPreview_WithDifferences(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diff := &model.BackupDiff{
		ConfigDiffs: []model.ConfigDiff{
			{Path: "sys.device.name", DiffType: model.DiffAdded, NewValue: "NewName"},
		},
	}
	DisplayMigrationPreview(ios, "/path/to/backup.json", "file", "target", diff)

	output := out.String()
	if !strings.Contains(output, "/path/to/backup.json") {
		t.Error("expected source path")
	}
	if !strings.Contains(output, "file") {
		t.Error("expected source type")
	}
}

func TestDisplayMigrationDiff_AllTypes(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diff := &model.BackupDiff{
		ConfigDiffs: []model.ConfigDiff{
			{Path: "config.path", DiffType: model.DiffAdded},
		},
		ScriptDiffs: []model.ScriptDiff{
			{Name: "script.js", DiffType: model.DiffChanged},
		},
		ScheduleDiffs: []model.ScheduleDiff{
			{Timespec: "0 8 * * *", DiffType: model.DiffRemoved},
		},
		WebhookDiffs: []model.WebhookDiff{
			{Name: "webhook", DiffType: model.DiffAdded},
		},
	}
	DisplayMigrationDiff(ios, diff)

	output := out.String()
	if !strings.Contains(output, "config.path") {
		t.Error("expected config diff")
	}
	if !strings.Contains(output, "script.js") {
		t.Error("expected script diff")
	}
	if !strings.Contains(output, "0 8 * * *") {
		t.Error("expected schedule diff")
	}
	if !strings.Contains(output, "webhook") {
		t.Error("expected webhook diff")
	}
}

func TestDisplayMigrationResult_Full(t *testing.T) {
	t.Parallel()

	ios, out, errOut := testIOStreams()
	result := &backup.RestoreResult{
		ConfigRestored:    true,
		ScriptsRestored:   3,
		SchedulesRestored: 2,
		WebhooksRestored:  1,
		Warnings:          []string{"Script had modified permissions"},
	}
	DisplayMigrationResult(ios, result)

	output := out.String()
	if !strings.Contains(output, "Config:") {
		t.Error("expected config status")
	}
	if !strings.Contains(output, "migrated") {
		t.Error("expected migrated status")
	}
	if !strings.Contains(output, "Scripts:   3") {
		t.Error("expected scripts count")
	}
	if !strings.Contains(output, "Schedules: 2") {
		t.Error("expected schedules count")
	}
	if !strings.Contains(output, "Webhooks:  1") {
		t.Error("expected webhooks count")
	}
	// Warning header goes to stderr
	if !strings.Contains(errOut.String(), "Warnings") {
		t.Error("expected warnings header")
	}
	// Warning details go to stdout via Printf
	if !strings.Contains(output, "modified permissions") {
		t.Error("expected warning message")
	}
}

func TestDisplayMigrationResult_NoWarnings(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	result := &backup.RestoreResult{
		ConfigRestored:  true,
		ScriptsRestored: 1,
	}
	DisplayMigrationResult(ios, result)

	output := out.String()
	if strings.Contains(output, "Warnings") {
		t.Error("should not show warnings when empty")
	}
}

func TestDisplayMigrationResult_ScriptsOnly(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	result := &backup.RestoreResult{
		ScriptsRestored: 5,
	}
	DisplayMigrationResult(ios, result)

	output := out.String()
	if strings.Contains(output, "Config:") {
		t.Error("should not show config when not restored")
	}
	if !strings.Contains(output, "Scripts:   5") {
		t.Error("expected scripts count")
	}
}
