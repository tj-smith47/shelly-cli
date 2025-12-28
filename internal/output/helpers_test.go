package output

import (
	"testing"
	"time"
)

func TestFormatConfigValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"nil", nil, "<not set>"},
		{"true", true, "true"},
		{"false", false, "false"},
		{"int as float", float64(42), "42"},
		{"float", float64(3.14), "3.14"},
		{"empty string", "", "<empty>"},
		{"string", "hello", "hello"},
		{"map", map[string]interface{}{"key": "value"}, `{"key":"value"}`},
		{"slice", []interface{}{"a", "b"}, `["a","b"]`},
		{"other type", 123, "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatConfigValue(tt.input)
			if got != tt.want {
				t.Errorf("FormatConfigValue(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatDeviceCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		count int
		want  string
	}{
		{0, "0 (empty)"},
		{1, "1 device"},
		{2, "2 devices"},
		{100, "100 devices"},
	}

	for _, tt := range tests {
		got := FormatDeviceCount(tt.count)
		if got != tt.want {
			t.Errorf("FormatDeviceCount(%d) = %q, want %q", tt.count, got, tt.want)
		}
	}
}

func TestFormatActionCount(t *testing.T) {
	t.Parallel()

	// Just test that it doesn't panic and returns non-empty strings
	got0 := FormatActionCount(0)
	if got0 == "" {
		t.Error("FormatActionCount(0) returned empty string")
	}

	got1 := FormatActionCount(1)
	if got1 == "" {
		t.Error("FormatActionCount(1) returned empty string")
	}

	got5 := FormatActionCount(5)
	if got5 == "" {
		t.Error("FormatActionCount(5) returned empty string")
	}
}

func TestFormatFloat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input float64
		want  string
	}{
		{0.0, "0"},
		{1.0, "1"},
		{3.14159, "3.14159"},
		{100.5, "100.5"},
	}

	for _, tt := range tests {
		got := FormatFloat(tt.input)
		if got != tt.want {
			t.Errorf("FormatFloat(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatFloatPtr(t *testing.T) {
	t.Parallel()

	got := FormatFloatPtr(nil)
	if got != "" {
		t.Errorf("FormatFloatPtr(nil) = %q, want empty", got)
	}

	val := 3.14
	got = FormatFloatPtr(&val)
	if got != "3.14" {
		t.Errorf("FormatFloatPtr(&3.14) = %q, want %q", got, "3.14")
	}
}

func TestFormatSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		size int64
		want string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tt := range tests {
		got := FormatSize(tt.size)
		if got != tt.want {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.size, got, tt.want)
		}
	}
}

func TestFormatJSONValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"nil", nil, "null"},
		{"string", "hello", `"hello"`},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"int as float", float64(42), "42"},
		{"float", float64(3.14), "3.14"},
		{"other", struct{}{}, "{}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatJSONValue(tt.input)
			if got != tt.want {
				t.Errorf("FormatJSONValue(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatDisplayValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"nil", nil, "null"},
		{"short string", "hi", `"hi"`},
		{"long string", "this is a very long string that should be truncated", `"this is a very long string that shoul"...`},
		{"bool", true, "true"},
		{"int as float", float64(42), "42"},
		{"map", map[string]any{"a": 1, "b": 2}, "{2 fields}"},
		{"empty map", map[string]any{}, "{0 fields}"},
		{"array", []any{1, 2, 3}, "[3 items]"},
		{"empty array", []any{}, "[0 items]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatDisplayValue(tt.input)
			if got != tt.want {
				t.Errorf("FormatDisplayValue(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValueType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input any
		want  string
	}{
		{nil, "null"},
		{"string", "string"},
		{true, "boolean"},
		{float64(42), "number"},
		{map[string]any{}, "object"},
		{[]any{}, "array"},
		{123, "unknown"},
	}

	for _, tt := range tests {
		got := ValueType(tt.input)
		if got != tt.want {
			t.Errorf("ValueType(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input time.Duration
		want  string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m"},
		{3 * time.Hour, "3h"},
		{25 * time.Hour, "1d"},
		{48 * time.Hour, "2d"},
	}

	for _, tt := range tests {
		got := FormatDuration(tt.input)
		if got != tt.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatParamsInline(t *testing.T) {
	t.Parallel()

	got := FormatParamsInline(nil)
	if got != "" {
		t.Errorf("FormatParamsInline(nil) = %q, want empty", got)
	}

	got = FormatParamsInline(map[string]any{})
	if got != "" {
		t.Errorf("FormatParamsInline({}) = %q, want empty", got)
	}

	got = FormatParamsInline(map[string]any{"key": "value"})
	if got != "key=value" {
		t.Errorf("FormatParamsInline({key:value}) = %q, want %q", got, "key=value")
	}
}

func TestFormatParamsTable(t *testing.T) {
	t.Parallel()

	got := FormatParamsTable(nil)
	if got != "-" {
		t.Errorf("FormatParamsTable(nil) = %q, want %q", got, "-")
	}

	got = FormatParamsTable(map[string]any{})
	if got != "-" {
		t.Errorf("FormatParamsTable({}) = %q, want %q", got, "-")
	}

	got = FormatParamsTable(map[string]any{"key": "value"})
	if got != "key: value" {
		t.Errorf("FormatParamsTable({key:value}) = %q, want %q", got, "key: value")
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello..."},
		{"hello", 3, "hel"},
		{"hello", 2, "he"},
	}

	for _, tt := range tests {
		got := Truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestPadRight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		width int
		want  string
	}{
		{"hello", 10, "hello     "},
		{"hello", 5, "hello"},
		{"hello", 3, "hello"},
	}

	for _, tt := range tests {
		got := PadRight(tt.input, tt.width)
		if got != tt.want {
			t.Errorf("PadRight(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
		}
	}
}

func TestRenderProgressBar(t *testing.T) {
	t.Parallel()

	// Just test that it doesn't panic and returns expected structure
	got := RenderProgressBar(5, 10)
	if got == "" {
		t.Error("RenderProgressBar returned empty string")
	}

	// Should contain box characters
	if len(got) < 20 {
		t.Error("RenderProgressBar too short")
	}
}

func TestEscapeWiFiQR(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{`back\slash`, `back\\slash`},
		{"semi;colon", `semi\;colon`},
		{"com,ma", `com\,ma`},
		{"col:on", `col\:on`},
		{`all\;,:chars`, `all\\\;\,\:chars`},
	}

	for _, tt := range tests {
		got := EscapeWiFiQR(tt.input)
		if got != tt.want {
			t.Errorf("EscapeWiFiQR(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatWiFiSignalStrength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rssi int
		want string
	}{
		{-40, "excellent"},
		{-50, "excellent"},
		{-55, "good"},
		{-60, "good"},
		{-65, "fair"},
		{-70, "fair"},
		{-80, "weak"},
		{-90, "weak"},
	}

	for _, tt := range tests {
		got := FormatWiFiSignalStrength(tt.rssi)
		if got != tt.want {
			t.Errorf("FormatWiFiSignalStrength(%d) = %q, want %q", tt.rssi, got, tt.want)
		}
	}
}

func TestFormatDeviceGeneration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		gen  int
		want string
	}{
		{1, "Gen1"},
		{2, "Gen2"},
		{3, "Gen3"},
	}

	for _, tt := range tests {
		got := FormatDeviceGeneration(tt.gen)
		if got != tt.want {
			t.Errorf("FormatDeviceGeneration(%d) = %q, want %q", tt.gen, got, tt.want)
		}
	}
}

func TestExtractMapSection(t *testing.T) {
	t.Parallel()

	// Test with nil
	got := ExtractMapSection(nil, "key")
	if got != nil {
		t.Error("ExtractMapSection(nil, ...) should return nil")
	}

	// Test with map
	data := map[string]any{
		"section": map[string]any{
			"field": "value",
		},
	}
	got = ExtractMapSection(data, "section")
	if got == nil {
		t.Fatal("ExtractMapSection should return section")
	}
	if got["field"] != "value" {
		t.Errorf("section[field] = %v, want %q", got["field"], "value")
	}

	// Test with missing section
	got = ExtractMapSection(data, "missing")
	if got != nil {
		t.Error("ExtractMapSection for missing key should return nil")
	}
}

func TestFormatConfigTable(t *testing.T) {
	t.Parallel()

	// Test with non-map
	got := FormatConfigTable("not a map")
	if got != nil {
		t.Error("FormatConfigTable(non-map) should return nil")
	}

	// Test with empty map
	got = FormatConfigTable(map[string]interface{}{})
	if got == nil {
		t.Error("FormatConfigTable({}) should return table")
	}

	// Test with populated map
	got = FormatConfigTable(map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})
	if got == nil {
		t.Fatal("FormatConfigTable should return table")
	}
	if got.RowCount() != 2 {
		t.Errorf("table has %d rows, want 2", got.RowCount())
	}
}
