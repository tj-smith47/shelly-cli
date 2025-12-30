package jsonf

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
	if !f.Indent {
		t.Error("New() should enable indent by default")
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
		indent  bool
		want    string
		wantErr bool
	}{
		{
			name:   "simple object with indent",
			data:   map[string]string{"key": "value"},
			indent: true,
			want:   "{\n  \"key\": \"value\"\n}\n",
		},
		{
			name:   "simple object without indent",
			data:   map[string]string{"key": "value"},
			indent: false,
			want:   "{\"key\":\"value\"}\n",
		},
		{
			name:   "array",
			data:   []int{1, 2, 3},
			indent: true,
			want:   "[\n  1,\n  2,\n  3\n]\n",
		},
		{
			name:   "string",
			data:   "hello",
			indent: true,
			want:   "\"hello\"\n",
		},
		{
			name:   "number",
			data:   42,
			indent: true,
			want:   "42\n",
		},
		{
			name:   "boolean",
			data:   true,
			indent: true,
			want:   "true\n",
		},
		{
			name:   "null",
			data:   nil,
			indent: true,
			want:   "null\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Formatter{
				Indent:    tt.indent,
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

	// Highlighted output should contain ANSI escape codes
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
	// Channels cannot be marshaled to JSON
	err := f.Format(&buf, make(chan int))
	if err == nil {
		t.Error("Format() should return error for unmarshalable type")
	}
}
