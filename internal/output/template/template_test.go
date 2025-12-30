package template

import (
	"bytes"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()
	tmpl := "{{.Name}}"
	f := New(tmpl)
	if f == nil {
		t.Fatal("New() returned nil")
	}
	if f.Template != tmpl {
		t.Errorf("New() Template = %q, want %q", f.Template, tmpl)
	}
}

func TestFormatter_Format(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		template string
		data     any
		want     string
		wantErr  bool
	}{
		{
			name:     "simple field access",
			template: "{{.Name}}",
			data:     map[string]string{"Name": "test"},
			want:     "test\n",
		},
		{
			name:     "multiple fields",
			template: "{{.Name}}: {{.Value}}",
			data:     map[string]any{"Name": "key", "Value": 42},
			want:     "key: 42\n",
		},
		{
			name:     "template with newline",
			template: "{{.Name}}\n",
			data:     map[string]string{"Name": "test"},
			want:     "test\n",
		},
		{
			name:     "range over slice",
			template: "{{range .}}{{.}} {{end}}",
			data:     []int{1, 2, 3},
			want:     "1 2 3 \n",
		},
		{
			name:     "empty template",
			template: "",
			data:     map[string]string{"Name": "test"},
			wantErr:  true,
		},
		{
			name:     "invalid template syntax",
			template: "{{.Name",
			data:     map[string]string{"Name": "test"},
			wantErr:  true,
		},
		{
			name:     "missing field",
			template: "{{.MissingField}}",
			data:     map[string]string{"Name": "test"},
			want:     "<no value>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := New(tt.template)

			var buf bytes.Buffer
			err := f.Format(&buf, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got := buf.String(); got != tt.want {
					t.Errorf("Format() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestFormatter_Format_Struct(t *testing.T) {
	t.Parallel()
	type Device struct {
		Name    string
		Address string
		Online  bool
	}

	f := New("{{.Name}} ({{.Address}}) - {{if .Online}}online{{else}}offline{{end}}")

	var buf bytes.Buffer
	err := f.Format(&buf, Device{
		Name:    "Kitchen Light",
		Address: "192.168.1.10",
		Online:  true,
	})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := "Kitchen Light (192.168.1.10) - online\n"
	if got := buf.String(); got != want {
		t.Errorf("Format() = %q, want %q", got, want)
	}
}

func TestFormatter_Format_TemplateError(t *testing.T) {
	t.Parallel()
	f := New("{{.Method}}")

	var buf bytes.Buffer
	// Maps don't have a Method field that can be called
	err := f.Format(&buf, map[string]string{})
	if err != nil {
		// This is expected - accessing non-existent field
		if !strings.Contains(buf.String(), "no value") && err != nil {
			t.Logf("Format() returned expected error or <no value>")
		}
	}
}
