// Package model provides domain types for the CLI.
package model

// Diff type constants for diff operations.
const (
	DiffAdded   = "added"
	DiffRemoved = "removed"
	DiffChanged = "changed"
)

// ConfigDiff represents a configuration difference between two sources.
type ConfigDiff struct {
	Path     string `json:"path"`                // Config path (dot notation)
	DiffType string `json:"diff_type"`           // added, removed, changed
	OldValue any    `json:"old_value,omitempty"` // Previous/source value
	NewValue any    `json:"new_value,omitempty"` // New/target value
}

// ScriptDiff represents a script difference.
type ScriptDiff struct {
	Name     string `json:"name"`
	DiffType string `json:"diff_type"`
	Details  string `json:"details,omitempty"`
}

// ScheduleDiff represents a schedule difference.
type ScheduleDiff struct {
	Timespec string `json:"timespec"`
	DiffType string `json:"diff_type"`
	Details  string `json:"details,omitempty"`
}

// WebhookDiff represents a webhook difference.
type WebhookDiff struct {
	Event    string `json:"event"`
	Name     string `json:"name,omitempty"`
	DiffType string `json:"diff_type"`
	Details  string `json:"details,omitempty"`
}

// BackupDiff contains differences between a backup and current device state.
type BackupDiff struct {
	ConfigDiffs   []ConfigDiff   `json:"config_diffs,omitempty"`
	ScriptDiffs   []ScriptDiff   `json:"script_diffs,omitempty"`
	ScheduleDiffs []ScheduleDiff `json:"schedule_diffs,omitempty"`
	WebhookDiffs  []WebhookDiff  `json:"webhook_diffs,omitempty"`
}

// HasDifferences returns true if there are any differences.
func (d *BackupDiff) HasDifferences() bool {
	return len(d.ConfigDiffs) > 0 || len(d.ScriptDiffs) > 0 ||
		len(d.ScheduleDiffs) > 0 || len(d.WebhookDiffs) > 0
}
