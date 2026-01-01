package table

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tbl := New("Name", "Value", "Status")

	if len(tbl.headers) != 3 {
		t.Errorf("expected 3 headers, got %d", len(tbl.headers))
	}
	if tbl.headers[0] != "Name" {
		t.Errorf("expected header[0] = 'Name', got %q", tbl.headers[0])
	}
}

func TestTable_AddRow(t *testing.T) {
	t.Parallel()

	tbl := New("A", "B")
	tbl.AddRow("1", "2")
	tbl.AddRow("3", "4")

	if len(tbl.rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(tbl.rows))
	}
	if tbl.rows[0][0] != "1" {
		t.Errorf("expected rows[0][0] = '1', got %q", tbl.rows[0][0])
	}
}

func TestTable_AddRows(t *testing.T) {
	t.Parallel()

	tbl := New("A", "B")
	rows := [][]string{
		{"1", "2"},
		{"3", "4"},
		{"5", "6"},
	}
	tbl.AddRows(rows)

	if tbl.RowCount() != 3 {
		t.Errorf("expected 3 rows, got %d", tbl.RowCount())
	}
}

func TestTable_PrintTo(t *testing.T) {
	t.Parallel()

	tbl := New("Device", "Status")
	tbl.AddRow("Living Room", "ON")
	tbl.AddRow("Bedroom", "OFF")

	var buf bytes.Buffer
	if err := tbl.PrintTo(&buf); err != nil {
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

	tbl := New("A", "B")

	var buf bytes.Buffer
	if err := tbl.PrintTo(&buf); err != nil {
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

	tbl := New("Name")
	style := DefaultStyle()
	style.ShowBorder = true
	tbl = tbl.SetStyle(style)
	tbl.AddRow("Test")

	var buf bytes.Buffer
	if err := tbl.PrintTo(&buf); err != nil {
		t.Fatalf("PrintTo failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestTable_Render(t *testing.T) {
	t.Parallel()

	tbl := New("Col1", "Col2")
	tbl.AddRow("a", "b")

	rendered := tbl.Render()
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

	tbl := New("X")
	tbl.AddRow("Y")

	str := tbl.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}

func TestTable_Empty(t *testing.T) {
	t.Parallel()

	tbl := New("A")
	if !tbl.Empty() {
		t.Error("Empty() should return true for table with no rows")
	}

	tbl.AddRow("1")
	if tbl.Empty() {
		t.Error("Empty() should return false for table with rows")
	}
}

func TestTable_RowCount(t *testing.T) {
	t.Parallel()

	tbl := New("A")
	if tbl.RowCount() != 0 {
		t.Errorf("expected RowCount() = 0, got %d", tbl.RowCount())
	}

	tbl.AddRow("1")
	tbl.AddRow("2")
	if tbl.RowCount() != 2 {
		t.Errorf("expected RowCount() = 2, got %d", tbl.RowCount())
	}
}

func TestPrintTo_Convenience(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := PrintTo(&buf, []string{"A", "B"}, [][]string{{"1", "2"}}); err != nil {
		t.Fatalf("PrintTo failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("PrintTo() produced no output")
	}
}

func TestBorderStyles(t *testing.T) {
	t.Parallel()

	styles := []BorderStyle{
		BorderNone, BorderRounded, BorderSquare,
		BorderDouble, BorderHeavy, BorderASCII,
	}

	for _, bs := range styles {
		tbl := New("Header")
		style := DefaultStyle()
		style.BorderStyle = bs
		tbl.SetStyle(style)
		tbl.AddRow("Value")

		rendered := tbl.Render()
		if rendered == "" {
			t.Errorf("BorderStyle %v produced empty output", bs)
		}
	}
}

func TestSetBorderStyle(t *testing.T) {
	t.Parallel()

	tbl := New("A").SetBorderStyle(BorderDouble)
	if tbl.style.BorderStyle != BorderDouble {
		t.Error("SetBorderStyle did not update style")
	}
}

func TestHideBorders(t *testing.T) {
	t.Parallel()

	tbl := New("A").HideBorders()
	if tbl.style.ShowBorder {
		t.Error("HideBorders should set ShowBorder to false")
	}
	if tbl.style.BorderStyle != BorderNone {
		t.Error("HideBorders should set BorderStyle to BorderNone")
	}
}

func TestTable_HideHeaders(t *testing.T) {
	t.Parallel()

	tbl := New("A", "B")
	tbl.HideHeaders()
	tbl.AddRow("1", "2")

	rendered := tbl.Render()
	// When headers are hidden, output should not contain the header row
	lines := strings.Split(strings.TrimSpace(rendered), "\n")
	if len(lines) < 1 {
		t.Error("expected at least one line of output")
	}
}

func TestTable_RenderPlain(t *testing.T) {
	t.Parallel()

	tbl := New("Name", "Value")
	style := PlainStyle()
	tbl.SetStyle(style)
	tbl.AddRow("foo", "bar")
	tbl.AddRow("baz", "qux")

	var buf bytes.Buffer
	if err := tbl.PrintTo(&buf); err != nil {
		t.Fatalf("PrintTo() error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestTable_Print(t *testing.T) {
	t.Parallel()
	// Actually call Print() to cover it - output goes to stdout
	tbl := New("A")
	tbl.AddRow("value")
	tbl.Print() // This writes to stdout, can't capture but exercises the code

	var buf bytes.Buffer
	err := tbl.PrintTo(&buf)
	if err != nil {
		t.Fatalf("PrintTo() error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestTable_RenderEmptyHeaders(t *testing.T) {
	t.Parallel()

	tbl := New()
	result := tbl.Render()
	if result != "" {
		t.Error("expected empty string for table with no headers")
	}
}

func TestTable_PrepareHeaders_NoUppercase(t *testing.T) {
	t.Parallel()

	tbl := New("Name", "Value")
	style := DefaultStyle()
	style.UppercaseHeaders = false
	tbl.SetStyle(style)

	headers := tbl.prepareHeaders()
	if headers[0] != "Name" {
		t.Errorf("expected 'Name', got %q", headers[0])
	}
}

func TestTable_RenderPlain_HiddenHeaders(t *testing.T) {
	t.Parallel()

	tbl := New("A", "B")
	tbl.SetStyle(PlainStyle())
	tbl.HideHeaders()
	tbl.AddRow("1", "2")

	output := tbl.Render()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line (data only), got %d", len(lines))
	}
}

func TestTable_RenderHiddenBorder(t *testing.T) {
	t.Parallel()

	tbl := New("A", "B")
	style := DefaultStyle()
	style.ShowBorder = false
	tbl.SetStyle(style)
	tbl.AddRow("1", "2")

	output := tbl.Render()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

// Tests using viper cannot be parallel.

//nolint:paralleltest // Uses viper global state via NewStyled->ShouldHideHeaders
func TestNewStyled(t *testing.T) {
	t.Run("with plain mode", func(t *testing.T) { //nolint:paralleltest // parent not parallel
		checker := &mockModeChecker{plain: true, colorEnabled: false}
		tbl := NewStyled(checker, "Name", "Value")
		if tbl == nil {
			t.Fatal("NewStyled returned nil")
		}
	})

	t.Run("with color mode", func(t *testing.T) { //nolint:paralleltest // parent not parallel
		checker := &mockModeChecker{plain: false, colorEnabled: true}
		tbl := NewStyled(checker, "Col1", "Col2", "Col3")
		if tbl == nil {
			t.Fatal("NewStyled returned nil")
		}
		tbl.AddRow("a", "b", "c")
		rendered := tbl.Render()
		if rendered == "" {
			t.Error("expected non-empty rendered output")
		}
	})
}

//nolint:paralleltest // Uses viper global state
func TestNewStyled_WithHiddenHeaders(t *testing.T) {
	// Test with viper "no-headers" set
	original := viper.GetBool("no-headers")
	viper.Set("no-headers", true)
	t.Cleanup(func() {
		viper.Set("no-headers", original)
	})

	checker := &mockModeChecker{plain: false, colorEnabled: true}
	tbl := NewStyled(checker, "A", "B")
	tbl.AddRow("1", "2")

	// With hideHeaders set, output should not have header row
	if !tbl.hideHeaders {
		t.Error("expected hideHeaders to be true")
	}
}
