package flags

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newFormatTree builds a root with the global -o flag and one child that has
// its own --format flag, then parses args the way cobra would before a run.
func newFormatTree(t *testing.T, args ...string) (child *cobra.Command, opts *OutputFlags) {
	t.Helper()
	root := &cobra.Command{Use: "shelly"}
	root.PersistentFlags().StringP("output", "o", "table", "Output format")
	opts = &OutputFlags{}
	child = &cobra.Command{Use: "list", Run: func(*cobra.Command, []string) {}}
	AddOutputFlagsCustom(child, opts, "text", "text", "json")
	root.AddCommand(child)
	if err := child.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags(%v): %v", args, err)
	}
	return child, opts
}

func TestApplyGlobalOutputFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr string
	}{
		{"no flag keeps the default", nil, "text", ""},
		{"global -o json is honoured", []string{"-o", "json"}, "json", ""},
		{"local --format wins over -o", []string{"-o", "json", "--format", "text"}, "text", ""},
		{"global table means the readable default", []string{"-o", "table"}, "text", ""},
		{"unsupported global format is an error", []string{"-o", "yaml"}, "", `cannot print "yaml"`},
		{"unsupported local format is an error", []string{"--format", "xml"}, "", `cannot print "xml"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd, opts := newFormatTree(t, tt.args...)
			err := ApplyGlobalOutputFormat(cmd)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want one containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opts.Format != tt.want {
				t.Errorf("Format = %q, want %q", opts.Format, tt.want)
			}
		})
	}
}

func TestApplyGlobalOutputFormat_IgnoresCommandsWithoutTheHelperFlag(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "shelly"}
	root.PersistentFlags().StringP("output", "o", "table", "Output format")
	child := &cobra.Command{Use: "export", Run: func(*cobra.Command, []string) {}}
	var format string
	child.Flags().StringVar(&format, "format", "csv", "File format")
	root.AddCommand(child)
	if err := child.ParseFlags([]string{"-o", "json"}); err != nil {
		t.Fatal(err)
	}
	if err := ApplyGlobalOutputFormat(child); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if format != "csv" {
		t.Errorf("format = %q, want it untouched", format)
	}
}

func TestApplyGlobalOutputFormat_RejectsUnknownFormat(t *testing.T) {
	t.Parallel()

	for format, wantErr := range map[string]bool{"csv": true, "jsno": true, "json": false, "yml": false, "table": false} {
		root := &cobra.Command{Use: "shelly"}
		root.PersistentFlags().StringP("output", "o", "table", "Output format")
		child := &cobra.Command{Use: "status", Run: func(*cobra.Command, []string) {}}
		root.AddCommand(child)
		if err := child.ParseFlags([]string{"-o", format}); err != nil {
			t.Fatalf("ParseFlags: %v", err)
		}
		err := ApplyGlobalOutputFormat(child)
		if (err != nil) != wantErr {
			t.Errorf("-o %s: err = %v, wantErr %v", format, err, wantErr)
		}
	}
}
