package jq

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApply_FieldSelection(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := map[string]any{"name": "kitchen", "status": "online"}

	err := Apply(&buf, data, ".name")
	require.NoError(t, err)
	assert.Equal(t, "\"kitchen\"\n", buf.String())
}

func TestApply_ArrayFiltering(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := []map[string]any{
		{"name": "kitchen", "online": true},
		{"name": "bedroom", "online": false},
		{"name": "garage", "online": true},
	}

	err := Apply(&buf, data, ".[] | select(.online) | .name")
	require.NoError(t, err)

	lines := strings.TrimSpace(buf.String())
	assert.Equal(t, "\"kitchen\"\n\"garage\"", lines)
}

func TestApply_NestedField(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := map[string]any{
		"wifi": map[string]any{
			"rssi": -42,
			"ssid": "HomeNet",
		},
	}

	err := Apply(&buf, data, ".wifi.rssi")
	require.NoError(t, err)
	assert.Equal(t, "-42\n", buf.String())
}

func TestApply_ArrayNames(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := []map[string]any{
		{"name": "a"},
		{"name": "b"},
		{"name": "c"},
	}

	err := Apply(&buf, data, ".[].name")
	require.NoError(t, err)
	assert.Equal(t, "\"a\"\n\"b\"\n\"c\"\n", buf.String())
}

func TestApply_InvalidExpression(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := Apply(&buf, map[string]any{}, ".foo[")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid jq expression")
}

func TestApply_NullResult(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := map[string]any{"name": "kitchen"}

	err := Apply(&buf, data, ".missing")
	require.NoError(t, err)
	assert.Equal(t, "null\n", buf.String())
}

func TestApply_PipeChain(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{
			map[string]any{"x": 1.0},
			map[string]any{"x": 2.0},
			map[string]any{"x": 3.0},
		},
	}

	err := Apply(&buf, data, ".items | map(.x) | add")
	require.NoError(t, err)
	assert.Equal(t, "6\n", buf.String())
}

func TestApply_MultipleFiltersPiped(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{
			map[string]any{"name": "a", "x": 1.0},
			map[string]any{"name": "b", "x": 2.0},
		},
	}

	// Simulate what happens when multiple --jq flags are joined with " | "
	err := Apply(&buf, data, ".items | map(.name)")
	require.NoError(t, err)
	assert.Equal(t, "[\n  \"a\",\n  \"b\"\n]\n", buf.String())
}

func TestApply_StructInput(t *testing.T) {
	t.Parallel()
	type device struct {
		Name   string `json:"name"`
		Online bool   `json:"online"`
	}
	var buf bytes.Buffer
	data := device{Name: "lamp", Online: true}

	err := Apply(&buf, data, ".name")
	require.NoError(t, err)
	assert.Equal(t, "\"lamp\"\n", buf.String())
}

func TestPrintFields_FlatMap(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := map[string]any{"name": "kitchen", "online": true, "id": 1}

	err := PrintFields(&buf, data)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Equal(t, []string{"id", "name", "online"}, lines)
}

func TestPrintFields_NestedMap(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := map[string]any{
		"name": "kitchen",
		"wifi": map[string]any{
			"rssi": -42,
			"ssid": "HomeNet",
		},
	}

	err := PrintFields(&buf, data)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Contains(t, lines, "name")
	assert.Contains(t, lines, "wifi")
	assert.Contains(t, lines, "wifi.rssi")
	assert.Contains(t, lines, "wifi.ssid")
}

func TestPrintFields_Array(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := []map[string]any{
		{"name": "kitchen", "status": "online"},
		{"name": "bedroom", "status": "offline"},
	}

	err := PrintFields(&buf, data)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "(array element fields)\n")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// First line is the header, rest are field names
	assert.Equal(t, "(array element fields)", lines[0])
	assert.Contains(t, lines, "name")
	assert.Contains(t, lines, "status")
}

func TestPrintFields_ArrayWithNestedObjects(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{
			map[string]any{"id": 1, "label": "a"},
		},
	}

	err := PrintFields(&buf, data)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Contains(t, lines, "items")
	assert.Contains(t, lines, "items[].id")
	assert.Contains(t, lines, "items[].label")
}

func TestPrintFields_Scalar(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := "just a string"

	err := PrintFields(&buf, data)
	require.NoError(t, err)
	assert.Equal(t, "(scalar value, no fields available)\n", buf.String())
}

func TestPrintFields_Struct(t *testing.T) {
	t.Parallel()
	type device struct {
		Name   string `json:"name"`
		Online bool   `json:"online"`
	}
	var buf bytes.Buffer
	data := device{Name: "lamp", Online: true}

	err := PrintFields(&buf, data)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Equal(t, []string{"name", "online"}, lines)
}

func TestPrintFields_SliceOfStructs(t *testing.T) {
	t.Parallel()
	type device struct {
		Name   string `json:"name"`
		Online bool   `json:"online"`
	}
	var buf bytes.Buffer
	data := []device{
		{Name: "lamp", Online: true},
		{Name: "plug", Online: false},
	}

	err := PrintFields(&buf, data)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "(array element fields)\n")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, "(array element fields)", lines[0])
	assert.Contains(t, lines, "name")
	assert.Contains(t, lines, "online")
}
