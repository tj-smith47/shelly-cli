package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestShort(t *testing.T) {
	t.Parallel()

	// Default should be "dev"
	v := Short()
	if v == "" {
		t.Error("Short() returned empty string")
	}
}

func TestLong(t *testing.T) {
	t.Parallel()

	long := Long()

	// Should contain version
	if !strings.Contains(long, "shelly") {
		t.Error("Long() should contain 'shelly'")
	}

	// Should contain Go version
	if !strings.Contains(long, "go:") {
		t.Error("Long() should contain 'go:'")
	}

	// Should contain OS/arch
	if !strings.Contains(long, runtime.GOOS) {
		t.Errorf("Long() should contain OS %q", runtime.GOOS)
	}
	if !strings.Contains(long, runtime.GOARCH) {
		t.Errorf("Long() should contain arch %q", runtime.GOARCH)
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	info := Get()

	if info.Version == "" {
		t.Error("Get().Version is empty")
	}
	if info.GoVersion == "" {
		t.Error("Get().GoVersion is empty")
	}
	if info.OS == "" {
		t.Error("Get().OS is empty")
	}
	if info.Arch == "" {
		t.Error("Get().Arch is empty")
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	s := String()
	if s != Short() {
		t.Errorf("String() = %q, expected Short() = %q", s, Short())
	}
}

func TestIsDevelopment(t *testing.T) {
	t.Parallel()

	isDev := IsDevelopment()
	// During tests, Version is usually "dev" or empty
	if Version == "" || Version == DevVersion {
		if !isDev {
			t.Error("IsDevelopment() should return true for dev builds")
		}
	} else {
		if isDev {
			t.Error("IsDevelopment() should return false for release builds")
		}
	}
}

func TestDevVersionConstant(t *testing.T) {
	t.Parallel()

	if DevVersion != "dev" {
		t.Errorf("DevVersion = %q, want 'dev'", DevVersion)
	}
}

func TestInfoFields(t *testing.T) {
	t.Parallel()

	info := Get()

	// Verify all fields are populated
	if info.Version == "" {
		t.Error("Version should not be empty")
	}
	if info.Commit == "" {
		t.Error("Commit should not be empty")
	}
	if info.Date == "" {
		t.Error("Date should not be empty")
	}
	if info.BuiltBy == "" {
		t.Error("BuiltBy should not be empty")
	}
	if info.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
	if info.OS == "" {
		t.Error("OS should not be empty")
	}
	if info.Arch == "" {
		t.Error("Arch should not be empty")
	}
}
