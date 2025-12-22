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

	// Check headers present (uppercase by default)
	if !strings.Contains(output, "DEVICE") {
		t.Error("expected output to contain 'DEVICE'")
	}
	if !strings.Contains(output, "STATUS") {
		t.Error("expected output to contain 'STATUS'")
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
	// Headers are uppercase by default
	if !strings.Contains(rendered, "COL1") {
		t.Error("Render() should contain uppercase header")
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

func TestTableBorderStyles(t *testing.T) {
	t.Parallel()

	styles := []TableBorderStyle{
		BorderNone, BorderRounded, BorderSquare,
		BorderDouble, BorderHeavy, BorderASCII,
	}

	for _, bs := range styles {
		table := NewTable("Header")
		style := DefaultTableStyle()
		style.BorderStyle = bs
		table.SetStyle(style)
		table.AddRow("Value")

		rendered := table.Render()
		if rendered == "" {
			t.Errorf("BorderStyle %v produced empty output", bs)
		}
	}
}

func TestPlainTableStyle(t *testing.T) {
	t.Parallel()

	style := PlainTableStyle()
	// Plain mode uses no borders for tab-separated output
	if style.BorderStyle != BorderNone {
		t.Error("PlainTableStyle should use no borders")
	}
	if style.ShowBorder {
		t.Error("PlainTableStyle should have ShowBorder disabled")
	}
	if !style.PlainMode {
		t.Error("PlainTableStyle should have PlainMode enabled")
	}
}

func TestNoColorTableStyle(t *testing.T) {
	t.Parallel()

	style := NoColorTableStyle()
	// No-color mode uses ASCII borders
	if style.BorderStyle != BorderASCII {
		t.Error("NoColorTableStyle should use ASCII borders")
	}
	if !style.ShowBorder {
		t.Error("NoColorTableStyle should have ShowBorder enabled")
	}
	if style.PlainMode {
		t.Error("NoColorTableStyle should not have PlainMode enabled")
	}
}

func TestSetBorderStyle(t *testing.T) {
	t.Parallel()

	table := NewTable("A").SetBorderStyle(BorderDouble)
	if table.style.BorderStyle != BorderDouble {
		t.Error("SetBorderStyle did not update style")
	}
}

func TestHideBorders(t *testing.T) {
	t.Parallel()

	table := NewTable("A").HideBorders()
	if table.style.ShowBorder {
		t.Error("HideBorders should set ShowBorder to false")
	}
	if table.style.BorderStyle != BorderNone {
		t.Error("HideBorders should set BorderStyle to BorderNone")
	}
}

// mockModeChecker implements ModeChecker for testing.
type mockModeChecker struct {
	plain        bool
	colorEnabled bool
}

func (m *mockModeChecker) IsPlainMode() bool {
	return m.plain
}

func (m *mockModeChecker) ColorEnabled() bool {
	return m.colorEnabled
}

func TestGetTableStyle(t *testing.T) {
	t.Parallel()

	t.Run("nil returns default", func(t *testing.T) {
		t.Parallel()
		style := GetTableStyle(nil)
		if style.BorderStyle != BorderRounded {
			t.Error("nil checker should return default style with rounded borders")
		}
	})

	t.Run("plain mode returns plain style with no borders", func(t *testing.T) {
		t.Parallel()
		checker := &mockModeChecker{plain: true, colorEnabled: true}
		style := GetTableStyle(checker)
		if style.BorderStyle != BorderNone {
			t.Error("plain mode should return plain style with no borders")
		}
		if !style.PlainMode {
			t.Error("plain mode should have PlainMode=true for tab-separated output")
		}
	})

	t.Run("no-color mode returns ASCII borders", func(t *testing.T) {
		t.Parallel()
		checker := &mockModeChecker{plain: false, colorEnabled: false}
		style := GetTableStyle(checker)
		if style.BorderStyle != BorderASCII {
			t.Error("no-color mode should return ASCII borders")
		}
	})

	t.Run("color enabled returns default style", func(t *testing.T) {
		t.Parallel()
		checker := &mockModeChecker{plain: false, colorEnabled: true}
		style := GetTableStyle(checker)
		if style.BorderStyle != BorderRounded {
			t.Error("color mode should return default style with rounded borders")
		}
	})
}
