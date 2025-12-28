package shelly

import (
	"testing"
)

func TestGeneratePassword_Length(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		length int
		want   int
	}{
		{"default length 16", 16, 16},
		{"short length 8", 8, 8},
		{"long length 32", 32, 32},
		{"minimum enforced 5 -> 8", 5, 8},
		{"minimum enforced 0 -> 8", 0, 8},
		{"minimum enforced 7 -> 8", 7, 8},
		{"boundary length 8", 8, 8},
		{"negative -> 8", -1, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			password, err := GeneratePassword(tt.length)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(password) != tt.want {
				t.Errorf("len(password) = %d, want %d", len(password), tt.want)
			}
		})
	}
}

func TestGeneratePassword_Uniqueness(t *testing.T) {
	t.Parallel()

	// Generate multiple passwords and verify they're different
	passwords := make(map[string]bool)
	for range 10 {
		password, err := GeneratePassword(DefaultPasswordLength)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if passwords[password] {
			t.Error("generated duplicate password")
		}
		passwords[password] = true
	}
}

func TestGeneratePassword_Charset(t *testing.T) {
	t.Parallel()

	password, err := GeneratePassword(100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all characters are from the charset
	charsetMap := make(map[rune]bool)
	for _, c := range PasswordCharset {
		charsetMap[c] = true
	}

	for _, c := range password {
		if !charsetMap[c] {
			t.Errorf("password contains invalid character: %c", c)
		}
	}
}

func TestPasswordConstants(t *testing.T) {
	t.Parallel()

	if DefaultPasswordLength != 16 {
		t.Errorf("DefaultPasswordLength = %d, want 16", DefaultPasswordLength)
	}

	// Verify charset contains expected character types
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, c := range PasswordCharset {
		switch {
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= '0' && c <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if !hasLower {
		t.Error("PasswordCharset missing lowercase letters")
	}
	if !hasUpper {
		t.Error("PasswordCharset missing uppercase letters")
	}
	if !hasDigit {
		t.Error("PasswordCharset missing digits")
	}
	if !hasSpecial {
		t.Error("PasswordCharset missing special characters")
	}
}
