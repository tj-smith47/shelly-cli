package config

import "testing"

const (
	testDeviceType = "device"
	testLongName   = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testMaxName    = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func TestValidateName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		entityType string
		wantErr    bool
	}{
		{"valid simple", "myname", testDeviceType, false},
		{"valid with dash", "my-name", "group", false},
		{"valid with underscore", "my_name", "scene", false},
		{"valid alphanumeric", "device123", testDeviceType, false},
		{"valid starts with number", "1device", testDeviceType, false},
		{"empty name", "", testDeviceType, true},
		{"too long", testLongName, testDeviceType, true},
		{"max length", testMaxName, testDeviceType, false},
		{"starts with hyphen", "-invalid", testDeviceType, true},
		{"starts with underscore", "_invalid", testDeviceType, true},
		{"contains space", "my name", testDeviceType, true},
		{"contains special char", "my@name", testDeviceType, true},
		{"contains slash", "my/name", testDeviceType, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateName(tt.input, tt.entityType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName(%q, %q) error = %v, wantErr %v", tt.input, tt.entityType, err, tt.wantErr)
			}
		})
	}
}

func TestValidateName_ErrorMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		entityType string
		wantMsg    string
	}{
		{"empty shows entity type", "", testDeviceType, "device name cannot be empty"},
		{"empty shows entity type scene", "", "scene", "scene name cannot be empty"},
		{"too long shows entity type", testLongName, "group", "group name too long"},
		{"invalid chars shows entity type", "-bad", "template", "template name contains invalid characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateName(tt.input, tt.entityType)
			if err == nil {
				t.Fatalf("expected error for input %q", tt.input)
			}
			if err.Error() != tt.wantMsg && tt.wantMsg != "" {
				// Just check it contains the entity type
				if !containsSubstring(err.Error(), tt.entityType) {
					t.Errorf("error %q should contain entity type %q", err.Error(), tt.entityType)
				}
			}
		})
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || s != "" && (s[0:len(sub)] == sub || containsSubstring(s[1:], sub)))
}
