package yaml

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/output/syntax"
)

func TestNew(t *testing.T) {
	t.Parallel()
	f := New()
	if f == nil {
		t.Fatal("New() returned nil")
	}
}

//nolint:paralleltest // Tests modify shared syntax.IsTTY state
func TestFormatter_Format(t *testing.T) {
	// Disable highlighting for predictable output
	oldIsTTY := syntax.IsTTY
	syntax.IsTTY = func() bool { return false }
	defer func() { syntax.IsTTY = oldIsTTY }()

	tests := []struct {
		name    string
		data    any
		want    string
		wantErr bool
	}{
		{
			name: "simple object",
			data: map[string]string{"key": "value"},
			want: "key: value\n",
		},
		{
			name: "nested object",
			data: map[string]any{
				"outer": map[string]string{
					"inner": "value",
				},
			},
			want: "outer:\n  inner: value\n",
		},
		{
			name: "array",
			data: []int{1, 2, 3},
			want: "- 1\n- 2\n- 3\n",
		},
		{
			name: "string",
			data: "hello",
			want: "hello\n",
		},
		{
			name: "number",
			data: 42,
			want: "42\n",
		},
		{
			name: "boolean",
			data: true,
			want: "true\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Formatter{
				Highlight: false,
			}

			var buf bytes.Buffer
			err := f.Format(&buf, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got := buf.String(); got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

//nolint:paralleltest // Tests modify shared syntax.IsTTY state
func TestFormatter_Format_Highlight(t *testing.T) {
	// Enable highlighting
	oldIsTTY := syntax.IsTTY
	syntax.IsTTY = func() bool { return true }
	defer func() { syntax.IsTTY = oldIsTTY }()

	f := New()
	f.Highlight = true

	var buf bytes.Buffer
	err := f.Format(&buf, map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Highlighted output should contain the key
	output := buf.String()
	if !strings.Contains(output, "key") {
		t.Error("Format() output should contain 'key'")
	}
}

func TestFormatter_Format_Error(t *testing.T) {
	t.Parallel()
	f := New()
	f.Highlight = false

	var buf bytes.Buffer
	// Functions cannot be marshaled to YAML - yaml.v3 panics on this
	defer func() {
		if r := recover(); r == nil {
			t.Error("Format() should panic for unmarshalable type (function)")
		}
	}()
	//nolint:errcheck // Intentionally ignoring error - testing panic behavior
	f.Format(&buf, func() {})
}
