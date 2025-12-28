// Package kvs provides Key-Value Store operations for Shelly devices.
package kvs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

func TestItem_Fields(t *testing.T) {
	t.Parallel()

	item := Item{
		Key:   "test_key",
		Value: "test_value",
		Etag:  "abc123",
	}

	if item.Key != "test_key" {
		t.Errorf("got Key=%q, want %q", item.Key, "test_key")
	}
	if item.Value != "test_value" {
		t.Errorf("got Value=%v, want %q", item.Value, "test_value")
	}
	if item.Etag != "abc123" {
		t.Errorf("got Etag=%q, want %q", item.Etag, "abc123")
	}
}

func TestListResult_Fields(t *testing.T) {
	t.Parallel()

	result := ListResult{
		Keys: []string{"key1", "key2", "key3"},
		Rev:  5,
	}

	if len(result.Keys) != 3 {
		t.Errorf("got %d keys, want 3", len(result.Keys))
	}
	if result.Keys[0] != "key1" {
		t.Errorf("got Keys[0]=%q, want %q", result.Keys[0], "key1")
	}
	if result.Rev != 5 {
		t.Errorf("got Rev=%d, want 5", result.Rev)
	}
}

func TestGetResult_Fields(t *testing.T) {
	t.Parallel()

	result := GetResult{
		Value: map[string]any{"nested": "value"},
		Etag:  "etag123",
	}

	if result.Value == nil {
		t.Error("expected non-nil Value")
	}
	if result.Etag != "etag123" {
		t.Errorf("got Etag=%q, want %q", result.Etag, "etag123")
	}
}

func TestExport_Fields(t *testing.T) {
	t.Parallel()

	export := Export{
		Items: []Item{
			{Key: "key1", Value: "value1"},
			{Key: "key2", Value: 123},
		},
		Version: 1,
		Rev:     10,
	}

	if len(export.Items) != 2 {
		t.Errorf("got %d items, want 2", len(export.Items))
	}
	if export.Version != 1 {
		t.Errorf("got Version=%d, want 1", export.Version)
	}
	if export.Rev != 10 {
		t.Errorf("got Rev=%d, want 10", export.Rev)
	}
}

func TestParseValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantType string
	}{
		{
			name:     "string value",
			input:    "hello",
			wantType: "string",
		},
		{
			name:     "number value",
			input:    "123",
			wantType: "float64", // JSON numbers are float64
		},
		{
			name:     "boolean true",
			input:    "true",
			wantType: "bool",
		},
		{
			name:     "boolean false",
			input:    "false",
			wantType: "bool",
		},
		{
			name:     "null value",
			input:    "null",
			wantType: "nil",
		},
		{
			name:     "JSON object",
			input:    `{"key": "value"}`,
			wantType: "map[string]interface {}",
		},
		{
			name:     "JSON array",
			input:    `[1, 2, 3]`,
			wantType: "[]interface {}",
		},
		{
			name:     "plain string with spaces",
			input:    "hello world",
			wantType: "string",
		},
		{
			name:     "quoted string",
			input:    `"hello"`,
			wantType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ParseValue(tt.input)

			var gotType string
			if result == nil {
				gotType = "nil"
			} else {
				gotType = getTypeName(result)
			}

			if gotType != tt.wantType {
				t.Errorf("ParseValue(%q) type = %s, want %s", tt.input, gotType, tt.wantType)
			}
		})
	}
}

func getTypeName(v any) string {
	if v == nil {
		return "nil"
	}
	switch v.(type) {
	case string:
		return "string"
	case float64:
		return "float64"
	case bool:
		return "bool"
	case map[string]any:
		return "map[string]interface {}"
	case []any:
		return "[]interface {}"
	default:
		return "unknown"
	}
}

func TestParseImportFile_JSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "import.json")

	exportData := Export{
		Items: []Item{
			{Key: "key1", Value: "value1"},
			{Key: "key2", Value: 123},
		},
		Version: 1,
		Rev:     5,
	}

	data, err := json.Marshal(exportData)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseImportFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Items) != 2 {
		t.Errorf("got %d items, want 2", len(result.Items))
	}
	if result.Version != 1 {
		t.Errorf("got Version=%d, want 1", result.Version)
	}
}

func TestParseImportFile_YAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "import.yaml")

	yamlContent := `
items:
  - key: key1
    value: value1
  - key: key2
    value: 123
version: 1
rev: 5
`

	if err := os.WriteFile(filePath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseImportFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Items) != 2 {
		t.Errorf("got %d items, want 2", len(result.Items))
	}
}

func TestParseImportFile_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := ParseImportFile("/nonexistent/file.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestParseImportFile_InvalidContent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(filePath, []byte("not valid json or yaml {{{{"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := ParseImportFile(filePath)
	if err == nil {
		t.Error("expected error for invalid content")
	}
}

func TestNewService(t *testing.T) {
	t.Parallel()

	mockConn := func(_ context.Context, _ string, _ func(*client.Client) error) error {
		return nil
	}

	svc := NewService(mockConn)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.withConnection == nil {
		t.Error("expected withConnection to be set")
	}
}

func TestItem_JSONMarshaling(t *testing.T) {
	t.Parallel()

	item := Item{
		Key:   "test_key",
		Value: map[string]any{"nested": "value"},
		Etag:  "abc123",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed Item
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.Key != item.Key {
		t.Errorf("got Key=%q, want %q", parsed.Key, item.Key)
	}
	if parsed.Etag != item.Etag {
		t.Errorf("got Etag=%q, want %q", parsed.Etag, item.Etag)
	}
}

func TestExport_JSONMarshaling(t *testing.T) {
	t.Parallel()

	export := Export{
		Items: []Item{
			{Key: "key1", Value: "value1", Etag: "etag1"},
			{Key: "key2", Value: 123, Etag: "etag2"},
		},
		Version: 1,
		Rev:     10,
	}

	data, err := json.Marshal(export)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed Export
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(parsed.Items) != len(export.Items) {
		t.Errorf("got %d items, want %d", len(parsed.Items), len(export.Items))
	}
	if parsed.Version != export.Version {
		t.Errorf("got Version=%d, want %d", parsed.Version, export.Version)
	}
	if parsed.Rev != export.Rev {
		t.Errorf("got Rev=%d, want %d", parsed.Rev, export.Rev)
	}
}
