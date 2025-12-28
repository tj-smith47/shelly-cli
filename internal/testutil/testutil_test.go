package testutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestTempDir(t *testing.T) {
	t.Parallel()

	dir, cleanup := TempDir(t)
	defer cleanup()

	if dir == "" {
		t.Error("TempDir returned empty string")
	}

	// Verify directory exists
	info, err := os.Stat(dir)
	if err != nil {
		t.Errorf("TempDir directory doesn't exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("TempDir should return a directory")
	}
}

func TestTempDir_Cleanup(t *testing.T) {
	t.Parallel()

	dir, cleanup := TempDir(t)

	// Create a file in the temp directory
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Run cleanup
	cleanup()

	// Verify directory is removed
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("TempDir cleanup should remove directory")
	}
}

func TestTempFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := "hello world"
	path := TempFile(t, dir, "test.txt", content)

	if path == "" {
		t.Error("TempFile returned empty path")
	}

	// Verify file exists with correct content
	data, err := os.ReadFile(path) //nolint:gosec // test file path is safe
	if err != nil {
		t.Errorf("failed to read temp file: %v", err)
	}
	if string(data) != content {
		t.Errorf("TempFile content = %q, want %q", string(data), content)
	}
}

func TestTempFile_MultipleFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	path1 := TempFile(t, dir, "file1.txt", "content1")
	path2 := TempFile(t, dir, "file2.txt", "content2")

	if path1 == path2 {
		t.Error("TempFile should return different paths for different files")
	}

	// Verify both files exist
	data1, err := os.ReadFile(path1) //nolint:gosec // test file path is safe
	if err != nil {
		t.Errorf("failed to read file1: %v", err)
	}
	if string(data1) != "content1" {
		t.Errorf("file1 content = %q, want %q", string(data1), "content1")
	}

	data2, err := os.ReadFile(path2) //nolint:gosec // test file path is safe
	if err != nil {
		t.Errorf("failed to read file2: %v", err)
	}
	if string(data2) != "content2" {
		t.Errorf("file2 content = %q, want %q", string(data2), "content2")
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global os.Stdout/os.Stderr
func TestCaptureOutput(t *testing.T) {
	capture := NewCaptureOutput(t)

	// Write to captured stdout/stderr
	if _, err := os.Stdout.WriteString("captured stdout"); err != nil {
		t.Logf("warning: failed to write to stdout: %v", err)
	}
	if _, err := os.Stderr.WriteString("captured stderr"); err != nil {
		t.Logf("warning: failed to write to stderr: %v", err)
	}

	stdout, stderr := capture.Stop()

	if stdout != "captured stdout" {
		t.Errorf("stdout = %q, want %q", stdout, "captured stdout")
	}
	if stderr != "captured stderr" {
		t.Errorf("stderr = %q, want %q", stderr, "captured stderr")
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global os.Stdout/os.Stderr
func TestCaptureOutput_Empty(t *testing.T) {
	capture := NewCaptureOutput(t)
	stdout, stderr := capture.Stop()

	if stdout != "" {
		t.Errorf("stdout should be empty, got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr should be empty, got %q", stderr)
	}
}

func TestResetViper(t *testing.T) {
	t.Parallel()

	// Just verify it doesn't panic
	ResetViper()
}

func TestSetupTestConfig(t *testing.T) {
	t.Parallel()

	// Just verify it doesn't panic
	SetupTestConfig(t)
}

func TestAssertContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		s         string
		substr    string
		wantError bool
	}{
		{
			name:      "contains substring",
			s:         "hello world",
			substr:    "world",
			wantError: false,
		},
		{
			name:      "contains at start",
			s:         "hello world",
			substr:    "hello",
			wantError: false,
		},
		{
			name:      "contains entire string",
			s:         "hello",
			substr:    "hello",
			wantError: false,
		},
		{
			name:      "does not contain",
			s:         "hello",
			substr:    "world",
			wantError: true,
		},
		{
			name:      "empty substring",
			s:         "hello",
			substr:    "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockT := &testing.T{}
			AssertContains(mockT, tt.s, tt.substr)

			if mockT.Failed() != tt.wantError {
				t.Errorf("AssertContains(%q, %q) failed = %v, want %v",
					tt.s, tt.substr, mockT.Failed(), tt.wantError)
			}
		})
	}
}

func TestAssertNotContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		s         string
		substr    string
		wantError bool
	}{
		{
			name:      "does not contain",
			s:         "hello",
			substr:    "world",
			wantError: false,
		},
		{
			name:      "contains substring",
			s:         "hello world",
			substr:    "world",
			wantError: true,
		},
		{
			name:      "contains entire string",
			s:         "hello",
			substr:    "hello",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockT := &testing.T{}
			AssertNotContains(mockT, tt.s, tt.substr)

			if mockT.Failed() != tt.wantError {
				t.Errorf("AssertNotContains(%q, %q) failed = %v, want %v",
					tt.s, tt.substr, mockT.Failed(), tt.wantError)
			}
		})
	}
}

func TestAssertEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		got       int
		want      int
		wantError bool
	}{
		{
			name:      "equal values",
			got:       42,
			want:      42,
			wantError: false,
		},
		{
			name:      "unequal values",
			got:       42,
			want:      24,
			wantError: true,
		},
		{
			name:      "zero values",
			got:       0,
			want:      0,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockT := &testing.T{}
			AssertEqual(mockT, tt.got, tt.want)

			if mockT.Failed() != tt.wantError {
				t.Errorf("AssertEqual(%v, %v) failed = %v, want %v",
					tt.got, tt.want, mockT.Failed(), tt.wantError)
			}
		})
	}
}

func TestAssertEqual_StringType(t *testing.T) {
	t.Parallel()

	mockT := &testing.T{}
	AssertEqual(mockT, "hello", "hello")
	if mockT.Failed() {
		t.Error("AssertEqual should pass for equal strings")
	}

	mockT2 := &testing.T{}
	AssertEqual(mockT2, "hello", "world")
	if !mockT2.Failed() {
		t.Error("AssertEqual should fail for unequal strings")
	}
}

func TestAssertNotEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		got       int
		notWant   int
		wantError bool
	}{
		{
			name:      "unequal values",
			got:       42,
			notWant:   24,
			wantError: false,
		},
		{
			name:      "equal values",
			got:       42,
			notWant:   42,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockT := &testing.T{}
			AssertNotEqual(mockT, tt.got, tt.notWant)

			if mockT.Failed() != tt.wantError {
				t.Errorf("AssertNotEqual(%v, %v) failed = %v, want %v",
					tt.got, tt.notWant, mockT.Failed(), tt.wantError)
			}
		})
	}
}

func TestAssertNil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		err       error
		wantError bool
	}{
		{
			name:      "nil error",
			err:       nil,
			wantError: false,
		},
		{
			name:      "non-nil error",
			err:       errors.New("test error"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockT := &testing.T{}
			AssertNil(mockT, tt.err)

			if mockT.Failed() != tt.wantError {
				t.Errorf("AssertNil(%v) failed = %v, want %v",
					tt.err, mockT.Failed(), tt.wantError)
			}
		})
	}
}

func TestAssertError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		err       error
		wantError bool
	}{
		{
			name:      "non-nil error",
			err:       errors.New("test error"),
			wantError: false,
		},
		{
			name:      "nil error",
			err:       nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockT := &testing.T{}
			AssertError(mockT, tt.err)

			if mockT.Failed() != tt.wantError {
				t.Errorf("AssertError(%v) failed = %v, want %v",
					tt.err, mockT.Failed(), tt.wantError)
			}
		})
	}
}

func TestAssertErrorContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		err       error
		substr    string
		wantError bool
	}{
		{
			name:      "error contains substring",
			err:       errors.New("connection failed: timeout"),
			substr:    "timeout",
			wantError: false,
		},
		{
			name:      "error does not contain substring",
			err:       errors.New("connection failed"),
			substr:    "timeout",
			wantError: true,
		},
		{
			name:      "nil error",
			err:       nil,
			substr:    "any",
			wantError: true,
		},
		{
			name:      "exact match",
			err:       errors.New("timeout"),
			substr:    "timeout",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockT := &testing.T{}
			AssertErrorContains(mockT, tt.err, tt.substr)

			if mockT.Failed() != tt.wantError {
				t.Errorf("AssertErrorContains(%v, %q) failed = %v, want %v",
					tt.err, tt.substr, mockT.Failed(), tt.wantError)
			}
		})
	}
}

func TestAssertTrue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		condition bool
		wantError bool
	}{
		{
			name:      "true condition",
			condition: true,
			wantError: false,
		},
		{
			name:      "false condition",
			condition: false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockT := &testing.T{}
			AssertTrue(mockT, tt.condition, "test message")

			if mockT.Failed() != tt.wantError {
				t.Errorf("AssertTrue(%v) failed = %v, want %v",
					tt.condition, mockT.Failed(), tt.wantError)
			}
		})
	}
}

func TestAssertFalse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		condition bool
		wantError bool
	}{
		{
			name:      "false condition",
			condition: false,
			wantError: false,
		},
		{
			name:      "true condition",
			condition: true,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockT := &testing.T{}
			AssertFalse(mockT, tt.condition, "test message")

			if mockT.Failed() != tt.wantError {
				t.Errorf("AssertFalse(%v) failed = %v, want %v",
					tt.condition, mockT.Failed(), tt.wantError)
			}
		})
	}
}

func TestPtr(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		intPtr := Ptr(42)
		if intPtr == nil {
			t.Fatal("Ptr returned nil")
		}
		if *intPtr != 42 {
			t.Errorf("*Ptr(42) = %d, want 42", *intPtr)
		}
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		strPtr := Ptr("hello")
		if strPtr == nil {
			t.Fatal("Ptr returned nil for string")
		}
		if *strPtr != "hello" {
			t.Errorf("*Ptr(hello) = %q, want hello", *strPtr)
		}
	})

	t.Run("bool", func(t *testing.T) {
		t.Parallel()

		boolPtr := Ptr(true)
		if boolPtr == nil {
			t.Fatal("Ptr returned nil for bool")
		}
		if !*boolPtr {
			t.Error("*Ptr(true) should be true")
		}
	})

	t.Run("float", func(t *testing.T) {
		t.Parallel()

		floatPtr := Ptr(3.14)
		if floatPtr == nil {
			t.Fatal("Ptr returned nil for float")
		}
		if *floatPtr != 3.14 {
			t.Errorf("*Ptr(3.14) = %f, want 3.14", *floatPtr)
		}
	})

	t.Run("struct", func(t *testing.T) {
		t.Parallel()

		type testStruct struct {
			Value int
		}
		structPtr := Ptr(testStruct{Value: 100})
		if structPtr == nil {
			t.Fatal("Ptr returned nil for struct")
		}
		if structPtr.Value != 100 {
			t.Errorf("Ptr(struct).Value = %d, want 100", structPtr.Value)
		}
	})
}

func TestErrMock(t *testing.T) {
	t.Parallel()

	if ErrMock == nil {
		t.Error("ErrMock should not be nil")
	}
	if ErrMock.Error() != "mock error" {
		t.Errorf("ErrMock.Error() = %q, want 'mock error'", ErrMock.Error())
	}
}
