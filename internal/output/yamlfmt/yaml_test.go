package yamlfmt

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/output/synfmt"
)

func TestNew(t *testing.T) {
	t.Parallel()
	f := New()
	if f == nil {
		t.Fatal("New() returned nil")
	}
}

//nolint:paralleltest // Tests modify shared synfmt.IsTTY state
func TestFormatter_Format(t *testing.T) {
	// Disable highlighting for predictable output
	oldIsTTY := synfmt.IsTTY
	synfmt.IsTTY = func() bool { return false }
	defer func() { synfmt.IsTTY = oldIsTTY }()

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

//nolint:paralleltest // Tests modify shared synfmt.IsTTY state
func TestFormatter_Format_Highlight(t *testing.T) {
	// Enable highlighting
	oldIsTTY := synfmt.IsTTY
	synfmt.IsTTY = func() bool { return true }
	defer func() { synfmt.IsTTY = oldIsTTY }()

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
	if err := f.Format(&buf, func() {}); err == nil {
		t.Errorf("Format(func) = nil error, output %q; want an error", buf.String())
	}
}

func TestFormatter_Format_UsesJSONKeys(t *testing.T) {
	t.Parallel()

	type row struct {
		DeviceName string          `json:"device_name"`
		Skipped    string          `json:"-"`
		Empty      string          `json:"empty,omitempty"`
		Version    string          `json:"version"`
		Raw        json.RawMessage `json:"raw"`
		Notes      string          `json:"notes"`
	}
	var buf bytes.Buffer
	err := (&Formatter{}).Format(&buf, []row{{DeviceName: "kitchen", Skipped: "x", Version: "1.0", Raw: json.RawMessage(`{"on":true}`), Notes: "a\nb"}})
	if err != nil {
		t.Fatal(err)
	}
	want := "- device_name: kitchen\n  version: \"1.0\"\n  raw:\n    \"on\": true\n  notes: |-\n    a\n    b\n"
	if buf.String() != want {
		t.Errorf("Format() =\n%s\nwant\n%s", buf.String(), want)
	}
}
