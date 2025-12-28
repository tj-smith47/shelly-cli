// Package utils provides common functionality shared across CLI commands.
package utils

import "testing"

func TestResolveBatchTargets_WithArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		group   string
		all     bool
		args    []string
		wantLen int
		wantErr bool
	}{
		{
			name:    "single device arg",
			args:    []string{"device1"},
			wantLen: 1,
		},
		{
			name:    "multiple device args",
			args:    []string{"device1", "device2", "device3"},
			wantLen: 3,
		},
		{
			name:    "args take precedence over group",
			group:   "mygroup",
			args:    []string{"device1"},
			wantLen: 1,
		},
		{
			name:    "args take precedence over all",
			all:     true,
			args:    []string{"device1", "device2"},
			wantLen: 2,
		},
		{
			name:    "args with group and all",
			group:   "mygroup",
			all:     true,
			args:    []string{"device1"},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			targets, err := ResolveBatchTargets(tt.group, tt.all, tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveBatchTargets() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && len(targets) != tt.wantLen {
				t.Errorf("len(targets) = %d, want %d", len(targets), tt.wantLen)
			}
		})
	}
}

func TestResolveBatchTargets_NoInputError(t *testing.T) {
	t.Parallel()

	// With no args, no group, and all=false, should return error
	_, err := ResolveBatchTargets("", false, nil)
	if err == nil {
		t.Error("ResolveBatchTargets() should error when no input provided")
	}

	// Verify error message is helpful
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestResolveBatchTargets_EmptyArgsSlice(t *testing.T) {
	t.Parallel()

	// Empty slice should be treated same as nil
	_, err := ResolveBatchTargets("", false, []string{})
	if err == nil {
		t.Error("ResolveBatchTargets() should error when args is empty slice")
	}
}

func TestIsJSONObject_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"just opening brace", "{", true},
		{"braces only", "{}", true},
		{"with nested", `{"a":{"b":1}}`, true},
		{"starts with space then brace", " {}", false}, // space first, not brace
		{"array not object", "[{}]", false},
		{"number", "123", false},
		{"boolean", "true", false},
		{"null", "null", false},
		{"string", `"hello"`, false},
		{"unicode object", `{"key":"值"}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsJSONObject(tt.s)
			if got != tt.want {
				t.Errorf("IsJSONObject(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}
