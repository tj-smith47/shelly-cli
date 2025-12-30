package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/viper"
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

func TestShouldHideHeaders_Default(t *testing.T) {
	t.Parallel()
	// Default should be false (headers visible)
	hide := ShouldHideHeaders()
	if hide {
		t.Error("ShouldHideHeaders() default should be false")
	}
}

func TestNewStyledTable(t *testing.T) {
	t.Parallel()

	t.Run("with plain mode", func(t *testing.T) {
		t.Parallel()
		checker := &mockModeChecker{plain: true, colorEnabled: false}
		table := NewStyledTable(checker, "Name", "Value")
		if table == nil {
			t.Fatal("NewStyledTable returned nil")
		}
	})

	t.Run("with color mode", func(t *testing.T) {
		t.Parallel()
		checker := &mockModeChecker{plain: false, colorEnabled: true}
		table := NewStyledTable(checker, "Col1", "Col2", "Col3")
		if table == nil {
			t.Fatal("NewStyledTable returned nil")
		}
		table.AddRow("a", "b", "c")
		rendered := table.Render()
		if rendered == "" {
			t.Error("expected non-empty rendered output")
		}
	})
}

func TestTable_HideHeaders(t *testing.T) {
	t.Parallel()

	table := NewTable("A", "B")
	table.HideHeaders()
	table.AddRow("1", "2")

	rendered := table.Render()
	// When headers are hidden, output should not contain the header row
	// Headers are typically uppercase (A, B)
	lines := strings.Split(strings.TrimSpace(rendered), "\n")
	// Should have at least the data row
	if len(lines) < 1 {
		t.Error("expected at least one line of output")
	}
}

func TestTable_RenderPlain(t *testing.T) {
	t.Parallel()

	table := NewTable("Name", "Value")
	style := PlainTableStyle()
	table.SetStyle(style)
	table.AddRow("foo", "bar")
	table.AddRow("baz", "qux")

	var buf bytes.Buffer
	if err := table.PrintTo(&buf); err != nil {
		t.Fatalf("PrintTo() error: %v", err)
	}

	output := buf.String()
	// Plain mode uses tab-separated values
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestTableFormatter_BuildTableFromData(t *testing.T) {
	t.Parallel()

	t.Run("slice of structs", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			ID   int
			Name string
		}
		data := []Item{{1, "foo"}, {2, "bar"}}
		var buf bytes.Buffer
		f := NewTableFormatter()
		err := f.Format(&buf, data)
		if err != nil {
			t.Fatalf("Format() error: %v", err)
		}
		output := strings.ToLower(buf.String())
		if !strings.Contains(output, "id") || !strings.Contains(output, "name") {
			t.Error("expected table to have headers")
		}
	})

	t.Run("single struct", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			ID   int
			Name string
		}
		data := Item{1, "foo"}
		var buf bytes.Buffer
		f := NewTableFormatter()
		err := f.Format(&buf, data)
		if err != nil {
			t.Fatalf("Format() error: %v", err)
		}
		output := buf.String()
		if output == "" {
			t.Error("expected non-empty output")
		}
	})

	t.Run("map data", func(t *testing.T) {
		t.Parallel()
		data := map[string]int{"a": 1, "b": 2}
		var buf bytes.Buffer
		f := NewTableFormatter()
		err := f.Format(&buf, data)
		if err != nil {
			t.Fatalf("Format() error: %v", err)
		}
		output := buf.String()
		if output == "" {
			t.Error("expected non-empty output")
		}
	})

	t.Run("bool field", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Active bool
		}
		data := []Item{{true}, {false}}
		var buf bytes.Buffer
		f := NewTableFormatter()
		err := f.Format(&buf, data)
		if err != nil {
			t.Fatalf("Format() error: %v", err)
		}
		output := buf.String()
		if output == "" {
			t.Error("expected non-empty output")
		}
	})

	t.Run("slice field", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Tags []string
		}
		data := []Item{{Tags: []string{"a", "b", "c"}}}
		var buf bytes.Buffer
		f := NewTableFormatter()
		err := f.Format(&buf, data)
		if err != nil {
			t.Fatalf("Format() error: %v", err)
		}
		output := buf.String()
		if output == "" {
			t.Error("expected non-empty output")
		}
	})

	t.Run("nested struct field", func(t *testing.T) {
		t.Parallel()
		type Inner struct {
			Value string
		}
		type Item struct {
			ID    int
			Inner Inner
		}
		data := []Item{{1, Inner{"nested"}}}
		var buf bytes.Buffer
		f := NewTableFormatter()
		err := f.Format(&buf, data)
		if err != nil {
			t.Fatalf("Format() error: %v", err)
		}
		output := buf.String()
		if output == "" {
			t.Error("expected non-empty output")
		}
	})
}

func TestTable_Print(t *testing.T) {
	t.Parallel()
	// Actually call Print() to cover it - output goes to stdout
	table := NewTable("A")
	table.AddRow("value")
	table.Print() // This writes to stdout, can't capture but exercises the code

	var buf bytes.Buffer
	err := table.PrintTo(&buf)
	if err != nil {
		t.Fatalf("PrintTo() error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestBuildTableFromData_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nil data", func(t *testing.T) {
		t.Parallel()
		table := buildTableFromData(nil)
		if table != nil {
			t.Error("expected nil for nil data")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		t.Parallel()
		var nilPtr *struct{ Name string }
		table := buildTableFromData(nilPtr)
		if table != nil {
			t.Error("expected nil for nil pointer")
		}
	})

	t.Run("pointer to slice of structs", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			ID int
		}
		items := []Item{{1}, {2}}
		table := buildTableFromData(&items)
		if table == nil {
			t.Error("expected table for pointer to slice")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			ID int
		}
		items := []Item{}
		table := buildTableFromData(items)
		if table != nil {
			t.Error("expected nil for empty slice")
		}
	})

	t.Run("slice of non-struct", func(t *testing.T) {
		t.Parallel()
		items := []int{1, 2, 3}
		table := buildTableFromData(items)
		if table != nil {
			t.Error("expected nil for slice of primitives")
		}
	})

	t.Run("array of structs", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Name string
		}
		items := [2]Item{{"a"}, {"b"}}
		table := buildTableFromData(items)
		if table == nil {
			t.Error("expected table for array of structs")
		}
	})

	t.Run("struct with unexported fields", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Name    string
			private string
		}
		items := []Item{{"pub", "priv"}}
		table := buildTableFromData(items)
		if table == nil {
			t.Fatal("expected table for struct with unexported fields")
		}
		// Should only have one column (Name)
		if len(table.headers) != 1 {
			t.Errorf("expected 1 header (only exported), got %d", len(table.headers))
		}
	})

	t.Run("struct with json dash tag", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Name   string `json:"name"`
			Hidden string `json:"-"`
		}
		items := []Item{{"visible", "hidden"}}
		table := buildTableFromData(items)
		if table == nil {
			t.Fatal("expected table for struct with json tags")
		}
		// Should only have one column (name)
		if len(table.headers) != 1 {
			t.Errorf("expected 1 header (excluding json:-), got %d", len(table.headers))
		}
	})

	t.Run("struct with json tag name", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			MyField string `json:"custom_name,omitempty"`
		}
		items := []Item{{"value"}}
		table := buildTableFromData(items)
		if table == nil {
			t.Fatal("expected table for struct with json tag name")
		}
		if table.headers[0] != "custom_name" {
			t.Errorf("expected header 'custom_name', got %q", table.headers[0])
		}
	})

	t.Run("slice of pointer structs", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			ID int
		}
		items := []*Item{{1}, {2}, {3}}
		table := buildTableFromData(items)
		if table == nil {
			t.Fatal("expected table for slice of pointer structs")
		}
		if table.RowCount() != 3 {
			t.Errorf("expected 3 rows, got %d", table.RowCount())
		}
	})
}

func TestFormatFieldValue_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("map field", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Data map[string]int
		}
		items := []Item{{Data: map[string]int{"a": 1, "b": 2}}}
		var buf bytes.Buffer
		f := NewTableFormatter()
		if err := f.Format(&buf, items); err != nil {
			t.Fatalf("Format error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "2 keys") {
			t.Error("expected output to contain '[2 keys]' for map")
		}
	})

	t.Run("empty map field", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Data map[string]int
		}
		items := []Item{{Data: map[string]int{}}}
		var buf bytes.Buffer
		f := NewTableFormatter()
		if err := f.Format(&buf, items); err != nil {
			t.Fatalf("Format error: %v", err)
		}
		output := buf.String()
		// Empty collection shows placeholder
		if !strings.Contains(output, LabelPlaceholder) {
			t.Error("expected output to contain placeholder for empty map")
		}
	})

	t.Run("pointer field with nil", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Value *int
		}
		items := []Item{{Value: nil}}
		var buf bytes.Buffer
		f := NewTableFormatter()
		if err := f.Format(&buf, items); err != nil {
			t.Fatalf("Format error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, LabelPlaceholder) {
			t.Error("expected output to contain placeholder for nil pointer")
		}
	})

	t.Run("pointer field with value", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Value *int
		}
		val := 42
		items := []Item{{Value: &val}}
		var buf bytes.Buffer
		f := NewTableFormatter()
		if err := f.Format(&buf, items); err != nil {
			t.Fatalf("Format error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "42") {
			t.Error("expected output to contain pointer value")
		}
	})

	t.Run("uint field", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			ID uint64
		}
		items := []Item{{ID: 12345}}
		var buf bytes.Buffer
		f := NewTableFormatter()
		if err := f.Format(&buf, items); err != nil {
			t.Fatalf("Format error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "12345") {
			t.Error("expected output to contain uint value")
		}
	})

	t.Run("float field", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Value float64
		}
		items := []Item{{Value: 3.14159}}
		var buf bytes.Buffer
		f := NewTableFormatter()
		if err := f.Format(&buf, items); err != nil {
			t.Fatalf("Format error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "3.14") {
			t.Error("expected output to contain float value")
		}
	})

	t.Run("empty string field", func(t *testing.T) {
		t.Parallel()
		type Item struct {
			Name string
		}
		items := []Item{{Name: ""}}
		var buf bytes.Buffer
		f := NewTableFormatter()
		if err := f.Format(&buf, items); err != nil {
			t.Fatalf("Format error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, LabelPlaceholder) {
			t.Error("expected output to contain placeholder for empty string")
		}
	})

	t.Run("stringer struct", func(t *testing.T) {
		t.Parallel()
		type MyStringer struct{}
		type Item struct {
			S MyStringer
		}
		items := []Item{{}}
		var buf bytes.Buffer
		f := NewTableFormatter()
		err := f.Format(&buf, items)
		if err != nil {
			t.Fatalf("Format error: %v", err)
		}
		// Should not panic - struct doesn't implement Stringer so falls back to %+v
	})
}

func TestTable_AddRows(t *testing.T) {
	t.Parallel()

	table := NewTable("A", "B")
	rows := [][]string{
		{"1", "2"},
		{"3", "4"},
		{"5", "6"},
	}
	table.AddRows(rows)

	if table.RowCount() != 3 {
		t.Errorf("expected 3 rows, got %d", table.RowCount())
	}
}

func TestTable_RenderEmptyHeaders(t *testing.T) {
	t.Parallel()

	table := NewTable()
	result := table.Render()
	if result != "" {
		t.Error("expected empty string for table with no headers")
	}
}

func TestTable_PrepareHeaders_NoUppercase(t *testing.T) {
	t.Parallel()

	table := NewTable("Name", "Value")
	style := DefaultTableStyle()
	style.UppercaseHeaders = false
	table.SetStyle(style)

	headers := table.prepareHeaders()
	if headers[0] != "Name" {
		t.Errorf("expected 'Name', got %q", headers[0])
	}
}

func TestTable_RenderPlain_HiddenHeaders(t *testing.T) {
	t.Parallel()

	table := NewTable("A", "B")
	table.SetStyle(PlainTableStyle())
	table.HideHeaders()
	table.AddRow("1", "2")

	output := table.Render()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line (data only), got %d", len(lines))
	}
}

// customStringerStruct is a struct that implements fmt.Stringer for testing.
type customStringerStruct struct {
	value string
}

func (s customStringerStruct) String() string {
	return s.value + "-stringified"
}

func TestFormatStruct_WithStringer(t *testing.T) {
	t.Parallel()

	type Item struct {
		Inner customStringerStruct
	}
	items := []Item{{Inner: customStringerStruct{"test"}}}
	var buf bytes.Buffer
	f := NewTableFormatter()
	if err := f.Format(&buf, items); err != nil {
		t.Fatalf("Format error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "test-stringified") {
		t.Errorf("expected stringer output, got %q", output)
	}
}

func TestBuildTableFromData_SliceOfPointers(t *testing.T) {
	t.Parallel()

	type Item struct {
		ID int
	}
	items := []*Item{{1}, {2}}
	table := buildTableFromData(items)
	if table == nil {
		t.Fatal("expected table for slice of pointers")
	}
	if table.RowCount() != 2 {
		t.Errorf("expected 2 rows, got %d", table.RowCount())
	}
}

func TestTable_RenderHiddenBorder(t *testing.T) {
	t.Parallel()

	table := NewTable("A", "B")
	style := DefaultTableStyle()
	style.ShowBorder = false
	table.SetStyle(style)
	table.AddRow("1", "2")

	output := table.Render()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestNewStyledTable_WithHiddenHeaders(t *testing.T) {
	t.Parallel()

	// Test with viper "no-headers" set
	original := viper.GetBool("no-headers")
	viper.Set("no-headers", true)
	t.Cleanup(func() {
		viper.Set("no-headers", original)
	})

	checker := &mockModeChecker{plain: false, colorEnabled: true}
	table := NewStyledTable(checker, "A", "B")
	table.AddRow("1", "2")

	// With hideHeaders set, output should not have header row
	if !table.hideHeaders {
		t.Error("expected hideHeaders to be true")
	}
}

func TestFormatDefault_InterfaceField(t *testing.T) {
	t.Parallel()

	// Test with an interface field that will hit formatDefault
	type Item struct {
		Value interface{}
	}
	// Use a complex number which isn't handled by other formatters
	items := []Item{{Value: complex(1, 2)}}
	var buf bytes.Buffer
	f := NewTableFormatter()
	if err := f.Format(&buf, items); err != nil {
		t.Fatalf("Format error: %v", err)
	}
	output := buf.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestBuildTableFromData_StructWithAllTypes(t *testing.T) {
	t.Parallel()

	type AllTypes struct {
		IntField    int
		Int8Field   int8
		Int16Field  int16
		Int32Field  int32
		Int64Field  int64
		UintField   uint
		Uint8Field  uint8
		Uint16Field uint16
		Uint32Field uint32
		Uint64Field uint64
		Float32F    float32
		Float64F    float64
		BoolField   bool
		StrField    string
	}
	items := []AllTypes{{
		IntField:    1,
		Int8Field:   2,
		Int16Field:  3,
		Int32Field:  4,
		Int64Field:  5,
		UintField:   6,
		Uint8Field:  7,
		Uint16Field: 8,
		Uint32Field: 9,
		Uint64Field: 10,
		Float32F:    1.5,
		Float64F:    2.5,
		BoolField:   true,
		StrField:    "test",
	}}
	table := buildTableFromData(items)
	if table == nil {
		t.Fatal("expected table for struct with all types")
	}
}

func TestFormatFieldValue_EmptySlice(t *testing.T) {
	t.Parallel()

	type Item struct {
		Tags []string
	}
	items := []Item{{Tags: []string{}}}
	var buf bytes.Buffer
	f := NewTableFormatter()
	if err := f.Format(&buf, items); err != nil {
		t.Fatalf("Format error: %v", err)
	}
	output := buf.String()
	// Empty slice shows placeholder
	if !strings.Contains(output, LabelPlaceholder) {
		t.Error("expected placeholder for empty slice")
	}
}
