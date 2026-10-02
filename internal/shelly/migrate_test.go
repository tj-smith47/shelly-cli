package shelly

import (
	"testing"

	backuppkg "github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
)

func TestCompareConfigs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config1 map[string]any
		config2 map[string]any
		wantLen int
	}{
		{
			name:    "identical configs",
			config1: map[string]any{"key": "value"},
			config2: map[string]any{"key": "value"},
			wantLen: 0,
		},
		{
			name:    "added key",
			config1: map[string]any{},
			config2: map[string]any{"key": "value"},
			wantLen: 1,
		},
		{
			name:    "removed key",
			config1: map[string]any{"key": "value"},
			config2: map[string]any{},
			wantLen: 1,
		},
		{
			name:    "changed value",
			config1: map[string]any{"key": "value1"},
			config2: map[string]any{"key": "value2"},
			wantLen: 1,
		},
		{
			name: "nested changes",
			config1: map[string]any{
				"sys": map[string]any{
					"device": map[string]any{
						"name": "Old Name",
					},
				},
			},
			config2: map[string]any{
				"sys": map[string]any{
					"device": map[string]any{
						"name": "New Name",
					},
				},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			diffs := backuppkg.CompareConfigs(tt.config1, tt.config2)
			testutil.AssertEqual(t, len(diffs), tt.wantLen)
		})
	}
}
