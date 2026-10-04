package firmware

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestCheckEntries_JSONShape(t *testing.T) {
	t.Parallel()
	entries := CheckEntries([]CheckResult{
		{Name: "kitchen", Info: &Info{Current: "1.0.0", Available: "1.1.0", HasUpdate: true}},
		{Name: "attic", Err: errors.New("timeout")},
	})
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	// examples/scripts/bulk-update.sh reads these four keys from every row.
	for _, key := range []string{"name", "update_available", "current_version", "new_version"} {
		if _, ok := rows[0][key]; !ok {
			t.Errorf("row is missing %q: %s", key, data)
		}
	}
	if rows[0]["update_available"] != true || rows[0]["new_version"] != "1.1.0" {
		t.Errorf("row = %v, want update_available=true new_version=1.1.0", rows[0])
	}
	if rows[1]["error"] != "timeout" || rows[1]["name"] != "attic" {
		t.Errorf("failed row = %v, want name=attic error=timeout", rows[1])
	}
}
