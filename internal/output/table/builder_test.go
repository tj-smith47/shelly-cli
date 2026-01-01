package table

import (
	"bytes"
	"strings"
	"testing"
)

func TestBuilderBasic(t *testing.T) {
	t.Parallel()
	tbl := NewBuilder("Name", "Status").
		AddRow("device1", "ON").
		AddRow("device2", "OFF").
		Build()

	if tbl.RowCount() != 2 {
		t.Errorf("expected 2 rows, got %d", tbl.RowCount())
	}
}

func TestBuilderEmpty(t *testing.T) {
	t.Parallel()
	builder := NewBuilder("Name", "Status")

	if !builder.Empty() {
		t.Error("expected builder to be empty")
	}

	builder.AddRow("device1", "ON")

	if builder.Empty() {
		t.Error("expected builder to not be empty after adding row")
	}
}

func TestBuilderRowCount(t *testing.T) {
	t.Parallel()
	builder := NewBuilder("A", "B", "C")

	if builder.RowCount() != 0 {
		t.Errorf("expected 0 rows, got %d", builder.RowCount())
	}

	builder.AddRow("1", "2", "3")
	builder.AddRow("4", "5", "6")

	if builder.RowCount() != 2 {
		t.Errorf("expected 2 rows, got %d", builder.RowCount())
	}
}

func TestBuilderAddRows(t *testing.T) {
	t.Parallel()
	rows := [][]string{
		{"a", "b"},
		{"c", "d"},
		{"e", "f"},
	}

	tbl := NewBuilder("Col1", "Col2").
		AddRows(rows).
		Build()

	if tbl.RowCount() != 3 {
		t.Errorf("expected 3 rows, got %d", tbl.RowCount())
	}
}

func TestBuilderWithStyle(t *testing.T) {
	t.Parallel()
	tbl := NewBuilder("Name", "Value").
		AddRow("key", "val").
		WithPlainStyle().
		Build()

	// Verify plain style was applied
	if tbl.style.ShowBorder {
		t.Error("expected ShowBorder to be false for plain style")
	}
	if !tbl.style.PlainMode {
		t.Error("expected PlainMode to be true for plain style")
	}
}

func TestBuilderWithNoColorStyle(t *testing.T) {
	t.Parallel()
	tbl := NewBuilder("Name", "Value").
		AddRow("key", "val").
		WithNoColorStyle().
		Build()

	// Verify no-color style was applied
	if tbl.style.BorderStyle != BorderASCII {
		t.Errorf("expected BorderASCII for no-color style, got %v", tbl.style.BorderStyle)
	}
}

func TestBuilderHideHeaders(t *testing.T) {
	t.Parallel()
	tbl := NewBuilder("Name", "Value").
		AddRow("key", "val").
		HideHeaders().
		Build()

	if !tbl.hideHeaders {
		t.Error("expected hideHeaders to be true")
	}
}

func TestBuilderChaining(t *testing.T) {
	t.Parallel()
	// Test that all methods return *Builder for chaining
	tbl := NewBuilder("A", "B", "C").
		AddRow("1", "2", "3").
		AddRow("4", "5", "6").
		AddRows([][]string{{"7", "8", "9"}}).
		WithPlainStyle().
		HideHeaders().
		Build()

	if tbl.RowCount() != 3 {
		t.Errorf("expected 3 rows, got %d", tbl.RowCount())
	}
}

func TestBuilderRowPadding(t *testing.T) {
	t.Parallel()
	// Test that rows are padded to match header count
	tbl := NewBuilder("A", "B", "C").
		AddRow("1", "2"). // Missing third column
		Build()

	if tbl.RowCount() != 1 {
		t.Errorf("expected 1 row, got %d", tbl.RowCount())
	}

	// The row should have 3 columns (padded with empty string)
	if len(tbl.rows[0]) != 3 {
		t.Errorf("expected 3 columns, got %d", len(tbl.rows[0]))
	}
}

func TestBuilderPlainOutput(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tbl := NewBuilder("Name", "Status").
		AddRow("device1", "ON").
		AddRow("device2", "OFF").
		WithPlainStyle().
		Build()

	if err := tbl.PrintTo(&buf); err != nil {
		t.Fatalf("PrintTo failed: %v", err)
	}

	output := buf.String()

	// Plain mode should have aligned columns
	if !strings.Contains(output, "NAME") {
		t.Error("expected uppercase header NAME")
	}
	if !strings.Contains(output, "device1") {
		t.Error("expected device1 in output")
	}
	if !strings.Contains(output, "device2") {
		t.Error("expected device2 in output")
	}
}

func TestBuilderWithModeStyle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		plain        bool
		colorEnabled bool
		expectPlain  bool
		expectBorder BorderStyle
	}{
		{
			name:         "default style",
			plain:        false,
			colorEnabled: true,
			expectPlain:  false,
			expectBorder: BorderRounded,
		},
		{
			name:         "plain mode",
			plain:        true,
			colorEnabled: true,
			expectPlain:  true,
			expectBorder: BorderNone,
		},
		{
			name:         "no color mode",
			plain:        false,
			colorEnabled: false,
			expectPlain:  false,
			expectBorder: BorderASCII,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockModeChecker{plain: tt.plain, colorEnabled: tt.colorEnabled}
			tbl := NewBuilder("Name", "Value").
				AddRow("key", "val").
				WithModeStyle(mock).
				Build()

			if tbl.style.PlainMode != tt.expectPlain {
				t.Errorf("expected PlainMode %v, got %v", tt.expectPlain, tbl.style.PlainMode)
			}
			if tbl.style.BorderStyle != tt.expectBorder {
				t.Errorf("expected BorderStyle %v, got %v", tt.expectBorder, tbl.style.BorderStyle)
			}
		})
	}
}

func TestBuilderWithStyleExplicit(t *testing.T) {
	t.Parallel()
	style := DefaultStyle()
	style.Padding = 5

	tbl := NewBuilder("A").
		AddRow("1").
		WithStyle(style).
		Build()

	if tbl.style.Padding != 5 {
		t.Errorf("expected padding 5, got %d", tbl.style.Padding)
	}
}
