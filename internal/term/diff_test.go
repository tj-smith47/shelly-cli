package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplayConfigDiffs_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayConfigDiffs(ios, []model.ConfigDiff{}, false)

	if out.String() != "" {
		t.Error("expected no output for empty diffs")
	}
}

func TestDisplayConfigDiffs_Added(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diffs := []model.ConfigDiff{
		{Path: "sys.device.name", DiffType: model.DiffAdded, NewValue: "Kitchen Light"},
	}
	DisplayConfigDiffs(ios, diffs, false)

	output := out.String()
	if !strings.Contains(output, "+") {
		t.Error("expected + prefix for added")
	}
	if !strings.Contains(output, "sys.device.name") {
		t.Error("expected path")
	}
}

func TestDisplayConfigDiffs_Removed(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diffs := []model.ConfigDiff{
		{Path: "wifi.sta.ssid", DiffType: model.DiffRemoved, OldValue: "OldNetwork"},
	}
	DisplayConfigDiffs(ios, diffs, false)

	output := out.String()
	if !strings.Contains(output, "-") {
		t.Error("expected - prefix for removed")
	}
}

func TestDisplayConfigDiffs_Changed(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diffs := []model.ConfigDiff{
		{Path: "switch:0.name", DiffType: model.DiffChanged, OldValue: "Old", NewValue: "New"},
	}
	DisplayConfigDiffs(ios, diffs, false)

	output := out.String()
	if !strings.Contains(output, "~") {
		t.Error("expected ~ prefix for changed")
	}
}

func TestDisplayConfigDiffs_Verbose(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diffs := []model.ConfigDiff{
		{Path: "sys.device.name", DiffType: model.DiffAdded, NewValue: "Kitchen"},
		{Path: "wifi.sta.ssid", DiffType: model.DiffRemoved, OldValue: "Old"},
		{Path: "switch:0.name", DiffType: model.DiffChanged, OldValue: "A", NewValue: "B"},
	}
	DisplayConfigDiffs(ios, diffs, true)

	output := out.String()
	if !strings.Contains(output, "will be added from backup") {
		t.Error("expected verbose message for added")
	}
	if !strings.Contains(output, "exists on device") {
		t.Error("expected verbose message for removed")
	}
	if !strings.Contains(output, "will be updated") {
		t.Error("expected verbose message for changed")
	}
}

func TestDisplayScriptDiffs_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayScriptDiffs(ios, []model.ScriptDiff{}, false)

	if out.String() != "" {
		t.Error("expected no output for empty diffs")
	}
}

func TestDisplayScriptDiffs_WithDiffs(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diffs := []model.ScriptDiff{
		{Name: "automation.js", DiffType: model.DiffAdded},
		{Name: "old_script.js", DiffType: model.DiffRemoved},
		{Name: "updated.js", DiffType: model.DiffChanged},
	}
	DisplayScriptDiffs(ios, diffs, false)

	output := out.String()
	if !strings.Contains(output, "Script") {
		t.Error("expected Script header")
	}
	if !strings.Contains(output, "automation.js") {
		t.Error("expected script name")
	}
}

func TestDisplayScheduleDiffs_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayScheduleDiffs(ios, []model.ScheduleDiff{}, false)

	if out.String() != "" {
		t.Error("expected no output for empty diffs")
	}
}

func TestDisplayScheduleDiffs_WithDiffs(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diffs := []model.ScheduleDiff{
		{Timespec: "0 7 * * *", DiffType: model.DiffAdded},
		{Timespec: "0 22 * * *", DiffType: model.DiffRemoved},
	}
	DisplayScheduleDiffs(ios, diffs, false)

	output := out.String()
	if !strings.Contains(output, "Schedule") {
		t.Error("expected Schedule header")
	}
	if !strings.Contains(output, "0 7 * * *") {
		t.Error("expected timespec")
	}
}

func TestDisplayWebhookDiffs_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayWebhookDiffs(ios, []model.WebhookDiff{}, false)

	if out.String() != "" {
		t.Error("expected no output for empty diffs")
	}
}

func TestDisplayWebhookDiffs_WithName(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diffs := []model.WebhookDiff{
		{Name: "my_webhook", Event: "switch.on", DiffType: model.DiffAdded},
	}
	DisplayWebhookDiffs(ios, diffs, false)

	output := out.String()
	if !strings.Contains(output, "my_webhook") {
		t.Error("expected webhook name")
	}
}

func TestDisplayWebhookDiffs_WithEvent(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diffs := []model.WebhookDiff{
		{Name: "", Event: "switch.off", DiffType: model.DiffAdded},
	}
	DisplayWebhookDiffs(ios, diffs, false)

	output := out.String()
	if !strings.Contains(output, "switch.off") {
		t.Error("expected event when name is empty")
	}
}

func TestDisplayConfigDiffsSummary_AllTypes(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diffs := []model.ConfigDiff{
		{Path: "added.path", DiffType: model.DiffAdded, NewValue: "new"},
		{Path: "removed.path", DiffType: model.DiffRemoved, OldValue: "old"},
		{Path: "changed.path", DiffType: model.DiffChanged, OldValue: "before", NewValue: "after"},
	}
	DisplayConfigDiffsSummary(ios, diffs)

	output := out.String()
	if !strings.Contains(output, "1 added") {
		t.Error("expected added count")
	}
	if !strings.Contains(output, "1 removed") {
		t.Error("expected removed count")
	}
	if !strings.Contains(output, "1 changed") {
		t.Error("expected changed count")
	}
}

func TestDisplayConfigDiffsSummary_OnlyAdded(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	diffs := []model.ConfigDiff{
		{Path: "path1", DiffType: model.DiffAdded, NewValue: "v1"},
		{Path: "path2", DiffType: model.DiffAdded, NewValue: "v2"},
	}
	DisplayConfigDiffsSummary(ios, diffs)

	output := out.String()
	if !strings.Contains(output, "2 added") {
		t.Error("expected 2 added")
	}
	if !strings.Contains(output, "0 removed") {
		t.Error("expected 0 removed")
	}
}

func TestDisplayConfigMapDiff_NewKeys(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	current := map[string]any{
		"existing": "value",
	}
	incoming := map[string]any{
		"existing": "value",
		"new_key":  "new_value",
	}
	DisplayConfigMapDiff(ios, current, incoming)

	output := out.String()
	if !strings.Contains(output, "new_key") {
		t.Error("expected new key")
	}
	if !strings.Contains(output, "(new)") {
		t.Error("expected (new) marker")
	}
}

func TestDisplayConfigMapDiff_ChangedKeys(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	current := map[string]any{
		"key": "old_value",
	}
	incoming := map[string]any{
		"key": "new_value",
	}
	DisplayConfigMapDiff(ios, current, incoming)

	output := out.String()
	if !strings.Contains(output, "old_value") {
		t.Error("expected old value")
	}
	if !strings.Contains(output, "new_value") {
		t.Error("expected new value")
	}
	if !strings.Contains(output, "->") {
		t.Error("expected arrow for change")
	}
}

func TestDisplayConfigMapDiff_NoChanges(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	current := map[string]any{
		"key": "value",
	}
	incoming := map[string]any{
		"key": "value",
	}
	DisplayConfigMapDiff(ios, current, incoming)

	output := out.String()
	if output != "" {
		t.Error("expected no output when configs are identical")
	}
}
