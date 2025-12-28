package model

import "testing"

func TestDiffConstants(t *testing.T) {
	t.Parallel()

	if DiffAdded != "added" {
		t.Errorf("DiffAdded = %q, want %q", DiffAdded, "added")
	}
	if DiffRemoved != "removed" {
		t.Errorf("DiffRemoved = %q, want %q", DiffRemoved, "removed")
	}
	if DiffChanged != "changed" {
		t.Errorf("DiffChanged = %q, want %q", DiffChanged, "changed")
	}
}

func TestBackupDiff_HasDifferences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		diff BackupDiff
		want bool
	}{
		{
			name: "empty diff",
			diff: BackupDiff{},
			want: false,
		},
		{
			name: "config diffs only",
			diff: BackupDiff{
				ConfigDiffs: []ConfigDiff{{Path: "sys.name", DiffType: DiffChanged}},
			},
			want: true,
		},
		{
			name: "script diffs only",
			diff: BackupDiff{
				ScriptDiffs: []ScriptDiff{{Name: "test.js", DiffType: DiffAdded}},
			},
			want: true,
		},
		{
			name: "schedule diffs only",
			diff: BackupDiff{
				ScheduleDiffs: []ScheduleDiff{{Timespec: "0 0 * * *", DiffType: DiffRemoved}},
			},
			want: true,
		},
		{
			name: "webhook diffs only",
			diff: BackupDiff{
				WebhookDiffs: []WebhookDiff{{Event: "switch.on", DiffType: DiffChanged}},
			},
			want: true,
		},
		{
			name: "warnings only (not a difference)",
			diff: BackupDiff{
				Warnings: []string{"could not compare scripts"},
			},
			want: false,
		},
		{
			name: "multiple diff types",
			diff: BackupDiff{
				ConfigDiffs:   []ConfigDiff{{Path: "name"}},
				ScriptDiffs:   []ScriptDiff{{Name: "script.js"}},
				ScheduleDiffs: []ScheduleDiff{{Timespec: "* * * * *"}},
				WebhookDiffs:  []WebhookDiff{{Event: "input.on"}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.diff.HasDifferences()
			if got != tt.want {
				t.Errorf("HasDifferences() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigDiff(t *testing.T) {
	t.Parallel()

	diff := ConfigDiff{
		Path:     "wifi.sta.ssid",
		DiffType: DiffChanged,
		OldValue: "OldNetwork",
		NewValue: "NewNetwork",
	}

	if diff.Path != "wifi.sta.ssid" {
		t.Errorf("Path = %q, want %q", diff.Path, "wifi.sta.ssid")
	}
	if diff.DiffType != DiffChanged {
		t.Errorf("DiffType = %q, want %q", diff.DiffType, DiffChanged)
	}
	if diff.OldValue != "OldNetwork" {
		t.Errorf("OldValue = %v, want %q", diff.OldValue, "OldNetwork")
	}
	if diff.NewValue != "NewNetwork" {
		t.Errorf("NewValue = %v, want %q", diff.NewValue, "NewNetwork")
	}
}

func TestScriptDiff(t *testing.T) {
	t.Parallel()

	diff := ScriptDiff{
		Name:     "automation.js",
		DiffType: DiffAdded,
		Details:  "New automation script",
	}

	if diff.Name != "automation.js" {
		t.Errorf("Name = %q, want %q", diff.Name, "automation.js")
	}
	if diff.DiffType != DiffAdded {
		t.Errorf("DiffType = %q, want %q", diff.DiffType, DiffAdded)
	}
}

func TestScheduleDiff(t *testing.T) {
	t.Parallel()

	diff := ScheduleDiff{
		Timespec: "0 8 * * 1-5",
		DiffType: DiffRemoved,
		Details:  "Weekday morning schedule",
	}

	if diff.Timespec != "0 8 * * 1-5" {
		t.Errorf("Timespec = %q, want %q", diff.Timespec, "0 8 * * 1-5")
	}
	if diff.DiffType != DiffRemoved {
		t.Errorf("DiffType = %q, want %q", diff.DiffType, DiffRemoved)
	}
}

func TestWebhookDiff(t *testing.T) {
	t.Parallel()

	diff := WebhookDiff{
		Event:    "button.single_push",
		Name:     "NotifyHomeAssistant",
		DiffType: DiffChanged,
		Details:  "URL changed",
	}

	if diff.Event != "button.single_push" {
		t.Errorf("Event = %q, want %q", diff.Event, "button.single_push")
	}
	if diff.Name != "NotifyHomeAssistant" {
		t.Errorf("Name = %q, want %q", diff.Name, "NotifyHomeAssistant")
	}
	if diff.DiffType != DiffChanged {
		t.Errorf("DiffType = %q, want %q", diff.DiffType, DiffChanged)
	}
}
