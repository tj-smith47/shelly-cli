package output

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

//nolint:gocyclo // test function with many sub-tests
func TestCompareConfigs(t *testing.T) {
	t.Parallel()

	t.Run("identical configs", func(t *testing.T) {
		t.Parallel()
		source := map[string]any{"key": "value"}
		target := map[string]any{"key": "value"}
		diffs := CompareConfigs(source, target)
		if len(diffs) != 0 {
			t.Errorf("expected 0 diffs for identical configs, got %d", len(diffs))
		}
	})

	t.Run("value changed", func(t *testing.T) {
		t.Parallel()
		source := map[string]any{"key": "old"}
		target := map[string]any{"key": "new"}
		diffs := CompareConfigs(source, target)
		if len(diffs) != 1 {
			t.Fatalf("expected 1 diff, got %d", len(diffs))
		}
		if diffs[0].DiffType != model.DiffChanged {
			t.Errorf("DiffType = %v, want DiffChanged", diffs[0].DiffType)
		}
		if diffs[0].Path != "key" {
			t.Errorf("Path = %q, want %q", diffs[0].Path, "key")
		}
		if diffs[0].OldValue != "old" {
			t.Errorf("OldValue = %v, want %q", diffs[0].OldValue, "old")
		}
		if diffs[0].NewValue != "new" {
			t.Errorf("NewValue = %v, want %q", diffs[0].NewValue, "new")
		}
	})

	t.Run("key added", func(t *testing.T) {
		t.Parallel()
		source := map[string]any{}
		target := map[string]any{"key": "value"}
		diffs := CompareConfigs(source, target)
		if len(diffs) != 1 {
			t.Fatalf("expected 1 diff, got %d", len(diffs))
		}
		if diffs[0].DiffType != model.DiffAdded {
			t.Errorf("DiffType = %v, want DiffAdded", diffs[0].DiffType)
		}
		if diffs[0].Path != "key" {
			t.Errorf("Path = %q, want %q", diffs[0].Path, "key")
		}
	})

	t.Run("key removed", func(t *testing.T) {
		t.Parallel()
		source := map[string]any{"key": "value"}
		target := map[string]any{}
		diffs := CompareConfigs(source, target)
		if len(diffs) != 1 {
			t.Fatalf("expected 1 diff, got %d", len(diffs))
		}
		if diffs[0].DiffType != model.DiffRemoved {
			t.Errorf("DiffType = %v, want DiffRemoved", diffs[0].DiffType)
		}
	})

	t.Run("nested object changed", func(t *testing.T) {
		t.Parallel()
		source := map[string]any{
			"outer": map[string]any{
				"inner": "old",
			},
		}
		target := map[string]any{
			"outer": map[string]any{
				"inner": "new",
			},
		}
		diffs := CompareConfigs(source, target)
		if len(diffs) != 1 {
			t.Fatalf("expected 1 diff, got %d", len(diffs))
		}
		if diffs[0].Path != "outer.inner" {
			t.Errorf("Path = %q, want %q", diffs[0].Path, "outer.inner")
		}
	})

	t.Run("multiple changes", func(t *testing.T) {
		t.Parallel()
		source := map[string]any{
			"a": "1",
			"b": "2",
			"c": "3",
		}
		target := map[string]any{
			"a": "1",       // same
			"b": "changed", // changed
			"d": "added",   // added (c removed)
		}
		diffs := CompareConfigs(source, target)
		if len(diffs) != 3 {
			t.Fatalf("expected 3 diffs, got %d", len(diffs))
		}
	})

	t.Run("empty configs", func(t *testing.T) {
		t.Parallel()
		source := map[string]any{}
		target := map[string]any{}
		diffs := CompareConfigs(source, target)
		if len(diffs) != 0 {
			t.Errorf("expected 0 diffs for empty configs, got %d", len(diffs))
		}
	})

	t.Run("deep nesting", func(t *testing.T) {
		t.Parallel()
		source := map[string]any{
			"l1": map[string]any{
				"l2": map[string]any{
					"l3": "value",
				},
			},
		}
		target := map[string]any{
			"l1": map[string]any{
				"l2": map[string]any{
					"l3": "changed",
				},
			},
		}
		diffs := CompareConfigs(source, target)
		if len(diffs) != 1 {
			t.Fatalf("expected 1 diff, got %d", len(diffs))
		}
		if diffs[0].Path != "l1.l2.l3" {
			t.Errorf("Path = %q, want %q", diffs[0].Path, "l1.l2.l3")
		}
	})
}
