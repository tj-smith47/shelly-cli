// Package kvs provides Key-Value Store operations for Shelly devices.
package kvs

import (
	"context"
	"encoding/json"
	"errors"
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

// ============== Service Method Tests ==============

func TestService_List_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	mockConn := func(_ context.Context, _ string, _ func(*client.Client) error) error {
		return expectedErr
	}

	svc := NewService(mockConn)
	result, err := svc.List(context.Background(), "test-device")

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestService_List_CallbackInvoked(t *testing.T) {
	t.Parallel()

	callbackInvoked := false
	mockConn := func(_ context.Context, identifier string, fn func(*client.Client) error) error {
		if identifier != "test-device" {
			t.Errorf("got identifier=%q, want %q", identifier, "test-device")
		}
		callbackInvoked = true
		return nil
	}

	svc := NewService(mockConn)
	_, _ = svc.List(context.Background(), "test-device")

	if !callbackInvoked {
		t.Error("expected callback to be invoked")
	}
}

func TestService_Get_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	mockConn := func(_ context.Context, _ string, _ func(*client.Client) error) error {
		return expectedErr
	}

	svc := NewService(mockConn)
	result, err := svc.Get(context.Background(), "test-device", "test-key")

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestService_Get_CallbackInvoked(t *testing.T) {
	t.Parallel()

	callbackInvoked := false
	mockConn := func(_ context.Context, identifier string, fn func(*client.Client) error) error {
		if identifier != "test-device" {
			t.Errorf("got identifier=%q, want %q", identifier, "test-device")
		}
		callbackInvoked = true
		return nil
	}

	svc := NewService(mockConn)
	_, _ = svc.Get(context.Background(), "test-device", "test-key")

	if !callbackInvoked {
		t.Error("expected callback to be invoked")
	}
}

func TestService_GetMany_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	mockConn := func(_ context.Context, _ string, _ func(*client.Client) error) error {
		return expectedErr
	}

	svc := NewService(mockConn)
	result, err := svc.GetMany(context.Background(), "test-device", "*")

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestService_GetMany_CallbackInvoked(t *testing.T) {
	t.Parallel()

	callbackInvoked := false
	mockConn := func(_ context.Context, identifier string, fn func(*client.Client) error) error {
		callbackInvoked = true
		return nil
	}

	svc := NewService(mockConn)
	_, _ = svc.GetMany(context.Background(), "test-device", "prefix_*")

	if !callbackInvoked {
		t.Error("expected callback to be invoked")
	}
}

func TestService_GetAll_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	mockConn := func(_ context.Context, _ string, _ func(*client.Client) error) error {
		return expectedErr
	}

	svc := NewService(mockConn)
	result, err := svc.GetAll(context.Background(), "test-device")

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestService_GetAll_CallbackInvoked(t *testing.T) {
	t.Parallel()

	callbackInvoked := false
	mockConn := func(_ context.Context, identifier string, fn func(*client.Client) error) error {
		callbackInvoked = true
		return nil
	}

	svc := NewService(mockConn)
	_, _ = svc.GetAll(context.Background(), "test-device")

	if !callbackInvoked {
		t.Error("expected callback to be invoked")
	}
}

func TestService_Set_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	mockConn := func(_ context.Context, _ string, _ func(*client.Client) error) error {
		return expectedErr
	}

	svc := NewService(mockConn)
	err := svc.Set(context.Background(), "test-device", "key", "value")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestService_Set_CallbackInvoked(t *testing.T) {
	t.Parallel()

	callbackInvoked := false
	mockConn := func(_ context.Context, identifier string, fn func(*client.Client) error) error {
		callbackInvoked = true
		return nil
	}

	svc := NewService(mockConn)
	_ = svc.Set(context.Background(), "test-device", "key", "value")

	if !callbackInvoked {
		t.Error("expected callback to be invoked")
	}
}

func TestService_Delete_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	mockConn := func(_ context.Context, _ string, _ func(*client.Client) error) error {
		return expectedErr
	}

	svc := NewService(mockConn)
	err := svc.Delete(context.Background(), "test-device", "key")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestService_Delete_CallbackInvoked(t *testing.T) {
	t.Parallel()

	callbackInvoked := false
	mockConn := func(_ context.Context, identifier string, fn func(*client.Client) error) error {
		callbackInvoked = true
		return nil
	}

	svc := NewService(mockConn)
	_ = svc.Delete(context.Background(), "test-device", "key")

	if !callbackInvoked {
		t.Error("expected callback to be invoked")
	}
}

func TestService_Export_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	mockConn := func(_ context.Context, _ string, _ func(*client.Client) error) error {
		return expectedErr
	}

	svc := NewService(mockConn)
	result, err := svc.Export(context.Background(), "test-device")

	if result != nil {
		t.Error("expected nil result on error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestService_Export_CallbackInvoked(t *testing.T) {
	t.Parallel()

	callbackInvoked := false
	mockConn := func(_ context.Context, identifier string, fn func(*client.Client) error) error {
		callbackInvoked = true
		return nil
	}

	svc := NewService(mockConn)
	_, _ = svc.Export(context.Background(), "test-device")

	if !callbackInvoked {
		t.Error("expected callback to be invoked")
	}
}

func TestService_Import_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection error")
	mockConn := func(_ context.Context, _ string, _ func(*client.Client) error) error {
		return expectedErr
	}

	svc := NewService(mockConn)
	data := &Export{Items: []Item{{Key: "key1", Value: "value1"}}}
	imported, skipped, err := svc.Import(context.Background(), "test-device", data, true)

	if imported != 0 {
		t.Errorf("got imported=%d, want 0", imported)
	}
	if skipped != 0 {
		t.Errorf("got skipped=%d, want 0", skipped)
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestService_Import_CallbackInvoked(t *testing.T) {
	t.Parallel()

	callbackInvoked := false
	mockConn := func(_ context.Context, identifier string, fn func(*client.Client) error) error {
		callbackInvoked = true
		return nil
	}

	svc := NewService(mockConn)
	data := &Export{Items: []Item{{Key: "key1", Value: "value1"}}}
	_, _, _ = svc.Import(context.Background(), "test-device", data, false)

	if !callbackInvoked {
		t.Error("expected callback to be invoked")
	}
}

// ============== Additional Utility Tests ==============

func TestParseValue_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expectValue any
	}{
		{
			name:        "empty string",
			input:       "",
			expectValue: "",
		},
		{
			name:        "whitespace only",
			input:       "   ",
			expectValue: "   ",
		},
		{
			name:        "negative number",
			input:       "-42",
			expectValue: float64(-42),
		},
		{
			name:        "float number",
			input:       "3.14159",
			expectValue: 3.14159,
		},
		{
			name:        "scientific notation",
			input:       "1e10",
			expectValue: float64(1e10),
		},
		{
			name:        "nested JSON object",
			input:       `{"a": {"b": {"c": 1}}}`,
			expectValue: map[string]any{"a": map[string]any{"b": map[string]any{"c": float64(1)}}},
		},
		{
			name:        "empty JSON object",
			input:       `{}`,
			expectValue: map[string]any{},
		},
		{
			name:        "empty JSON array",
			input:       `[]`,
			expectValue: []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ParseValue(tt.input)

			// For complex types, compare JSON representation
			expectedJSON, _ := json.Marshal(tt.expectValue)
			resultJSON, _ := json.Marshal(result)
			if string(expectedJSON) != string(resultJSON) {
				t.Errorf("ParseValue(%q) = %v, want %v", tt.input, result, tt.expectValue)
			}
		})
	}
}

func TestListResult_EmptyKeys(t *testing.T) {
	t.Parallel()

	result := ListResult{
		Keys: []string{},
		Rev:  0,
	}

	if len(result.Keys) != 0 {
		t.Errorf("got %d keys, want 0", len(result.Keys))
	}
	if result.Rev != 0 {
		t.Errorf("got Rev=%d, want 0", result.Rev)
	}
}

func TestExport_EmptyItems(t *testing.T) {
	t.Parallel()

	export := Export{
		Items:   []Item{},
		Version: 1,
		Rev:     0,
	}

	if len(export.Items) != 0 {
		t.Errorf("got %d items, want 0", len(export.Items))
	}
}

func TestItem_ComplexValue(t *testing.T) {
	t.Parallel()

	complexValue := map[string]any{
		"nested": map[string]any{
			"array": []any{1, 2, 3},
			"bool":  true,
		},
	}

	item := Item{
		Key:   "complex",
		Value: complexValue,
		Etag:  "etag",
	}

	// Verify we can marshal and unmarshal
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
}

func TestParseImportFile_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty.json")

	// Write empty JSON object
	if err := os.WriteFile(filePath, []byte("{}"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseImportFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Items) != 0 {
		t.Errorf("got %d items, want 0", len(result.Items))
	}
}

func TestParseImportFile_YAMLWithAllTypes(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "alltypes.yaml")

	yamlContent := `
items:
  - key: string_key
    value: hello
  - key: number_key
    value: 42
  - key: bool_key
    value: true
  - key: float_key
    value: 3.14
  - key: null_key
    value: null
  - key: array_key
    value: [1, 2, 3]
  - key: object_key
    value:
      nested: value
version: 1
rev: 100
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
	if len(result.Items) != 7 {
		t.Errorf("got %d items, want 7", len(result.Items))
	}
	if result.Rev != 100 {
		t.Errorf("got Rev=%d, want 100", result.Rev)
	}
}

func TestNewService_NilConnection(t *testing.T) {
	t.Parallel()

	svc := NewService(nil)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	// withConnection will be nil, which should cause panic if used
	// This tests that the service can be created but shouldn't be used without a connection
}

func TestGetResult_ZeroValue(t *testing.T) {
	t.Parallel()

	var result GetResult

	if result.Value != nil {
		t.Error("expected nil Value")
	}
	if result.Etag != "" {
		t.Errorf("got Etag=%q, want empty", result.Etag)
	}
}

func TestItem_EtagOmitEmpty(t *testing.T) {
	t.Parallel()

	item := Item{
		Key:   "key",
		Value: "value",
		// Etag is empty
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Check that etag is omitted in JSON
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, exists := parsed["etag"]; exists {
		t.Error("expected etag to be omitted when empty")
	}
}
