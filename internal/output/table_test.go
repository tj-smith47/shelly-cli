package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewTable(t *testing.T) {
	t.Parallel()

	table := NewTable("Name", "Value", "Status")

	if len(table.headers) != 3 {
		t.Errorf("expected 3 headers, got %d", len(table.headers))
	}
	if table.headers[0] != "Name" {
		t.Errorf("expected header[0] = 'Name', got %q", table.headers[0])
	}
}

func TestTable_AddRow(t *testing.T) {
	t.Parallel()

	table := NewTable("A", "B")
	table.AddRow("1", "2")
	table.AddRow("3", "4")

	if len(table.rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(table.rows))
	}
	if table.rows[0][0] != "1" {
		t.Errorf("expected rows[0][0] = '1', got %q", table.rows[0][0])
	}
}

func TestTable_PrintTo(t *testing.T) {
	t.Parallel()

	table := NewTable("Device", "Status")
	table.AddRow("Living Room", "ON")
	table.AddRow("Bedroom", "OFF")

	var buf bytes.Buffer
	if err := table.PrintTo(&buf); err != nil {
		t.Fatalf("PrintTo failed: %v", err)
	}

	output := buf.String()

	// Check headers present
	if !strings.Contains(output, "Device") {
		t.Error("expected output to contain 'Device'")
	}
	if !strings.Contains(output, "Status") {
		t.Error("expected output to contain 'Status'")
	}

	// Check data present
	if !strings.Contains(output, "Living Room") {
		t.Error("expected output to contain 'Living Room'")
	}
	if !strings.Contains(output, "ON") {
		t.Error("expected output to contain 'ON'")
	}
}

func TestTable_EmptyTable(t *testing.T) {
	t.Parallel()

	table := NewTable("A", "B")

	var buf bytes.Buffer
	if err := table.PrintTo(&buf); err != nil {
		t.Fatalf("PrintTo failed: %v", err)
	}

	output := buf.String()
	// Should still have headers
	if !strings.Contains(output, "A") {
		t.Error("expected output to contain header 'A'")
	}
}

func TestTable_SetStyle(t *testing.T) {
	t.Parallel()

	table := NewTable("Name")
	style := DefaultTableStyle()
	style.ShowBorder = true
	table = table.SetStyle(style)
	table.AddRow("Test")

	var buf bytes.Buffer
	if err := table.PrintTo(&buf); err != nil {
		t.Fatalf("PrintTo failed: %v", err)
	}

	// Should produce output without errors
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestTable_Render(t *testing.T) {
	t.Parallel()

	table := NewTable("Col1", "Col2")
	table.AddRow("a", "b")

	rendered := table.Render()
	if rendered == "" {
		t.Error("Render() returned empty string")
	}
	if !strings.Contains(rendered, "Col1") {
		t.Error("Render() should contain header")
	}
}

func TestTable_String(t *testing.T) {
	t.Parallel()

	table := NewTable("X")
	table.AddRow("Y")

	str := table.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}

func TestTable_Empty(t *testing.T) {
	t.Parallel()

	table := NewTable("A")
	if !table.Empty() {
		t.Error("Empty() should return true for table with no rows")
	}

	table.AddRow("1")
	if table.Empty() {
		t.Error("Empty() should return false for table with rows")
	}
}

func TestTable_RowCount(t *testing.T) {
	t.Parallel()

	table := NewTable("A")
	if table.RowCount() != 0 {
		t.Errorf("expected RowCount() = 0, got %d", table.RowCount())
	}

	table.AddRow("1")
	table.AddRow("2")
	if table.RowCount() != 2 {
		t.Errorf("expected RowCount() = 2, got %d", table.RowCount())
	}
}

func TestKeyValueTable(t *testing.T) {
	t.Parallel()

	pairs := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	table := KeyValueTable(pairs)
	if table.RowCount() != 2 {
		t.Errorf("expected 2 rows, got %d", table.RowCount())
	}
}

func TestPrintTableTo(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := PrintTableTo(&buf, []string{"A", "B"}, [][]string{{"1", "2"}}); err != nil {
		t.Fatalf("PrintTableTo failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("PrintTableTo() produced no output")
	}
}
