package cmdutil_test

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestParseKeyValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		want    []cmdutil.KeyValue
		wantErr bool
	}{
		{
			name: "inline equals",
			args: []string{"telemetry=true"},
			want: []cmdutil.KeyValue{{Key: "telemetry", Value: "true"}},
		},
		{
			name: "inline colon",
			args: []string{"telemetry:true"},
			want: []cmdutil.KeyValue{{Key: "telemetry", Value: "true"}},
		},
		{
			name: "space separated",
			args: []string{"telemetry", "true"},
			want: []cmdutil.KeyValue{{Key: "telemetry", Value: "true"}},
		},
		{
			name: "multiple inline equals",
			args: []string{"defaults.timeout=30s", "defaults.output=json"},
			want: []cmdutil.KeyValue{
				{Key: "defaults.timeout", Value: "30s"},
				{Key: "defaults.output", Value: "json"},
			},
		},
		{
			name: "multiple space pairs",
			args: []string{"a", "1", "b", "2"},
			want: []cmdutil.KeyValue{{Key: "a", Value: "1"}, {Key: "b", Value: "2"}},
		},
		{
			name: "mixed inline and space",
			args: []string{"a=1", "b", "2"},
			want: []cmdutil.KeyValue{{Key: "a", Value: "1"}, {Key: "b", Value: "2"}},
		},
		{
			name: "two-arg value may contain separator",
			args: []string{"editor", "vim:tag"},
			want: []cmdutil.KeyValue{{Key: "editor", Value: "vim:tag"}},
		},
		{
			name: "inline value keeps later separators",
			args: []string{"k=a=b"},
			want: []cmdutil.KeyValue{{Key: "k", Value: "a=b"}},
		},
		{
			name: "earliest separator wins",
			args: []string{"k:a=b"},
			want: []cmdutil.KeyValue{{Key: "k", Value: "a=b"}},
		},
		{
			name:    "bare key missing value",
			args:    []string{"telemetry"},
			wantErr: true,
		},
		{
			name:    "bare key followed by inline pair in multi-set is ambiguous",
			args:    []string{"verbose", "telemetry=true", "x=1"},
			wantErr: true,
		},
		{
			name:    "empty args",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := cmdutil.ParseKeyValues(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseKeyValues(%q) err = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("ParseKeyValues(%q) = %v, want %v", tt.args, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("pair %d = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
