package iostreams_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestNewSpinner(t *testing.T) {
	t.Parallel()

	s := iostreams.NewSpinner("Loading...")
	if s == nil {
		t.Fatal("NewSpinner() returned nil")
	}
}

func TestNewSpinnerWithWriter(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	s := iostreams.NewSpinnerWithWriter("Loading...", buf)
	if s == nil {
		t.Fatal("NewSpinnerWithWriter() returned nil")
	}
}

func TestSpinner_Options(t *testing.T) {
	t.Parallel()

	// Test all option functions don't panic
	buf := &bytes.Buffer{}
	s := iostreams.NewSpinnerWithWriter("Test",
		buf,
		iostreams.WithSuffix("suffix"),
		iostreams.WithPrefix("prefix"),
		iostreams.WithFinalMessage("final"),
		iostreams.WithCharSet([]string{"|", "/", "-", "\\"}),
		iostreams.WithColor("green"),
	)

	if s == nil {
		t.Fatal("NewSpinnerWithWriter with options returned nil")
	}
}

func TestSpinner_StartStop(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	s := iostreams.NewSpinnerWithWriter("Loading...", buf)

	// Should not panic
	s.Start()
	s.Stop()
}

func TestSpinner_StopWithSuccess(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	s := iostreams.NewSpinnerWithWriter("Loading...", buf)

	s.Start()
	s.StopWithSuccess("Done!")

	// Note: We can't easily verify output due to spinner's async nature
	// Just verify it doesn't panic
}

func TestSpinner_StopWithError(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	s := iostreams.NewSpinnerWithWriter("Loading...", buf)

	s.Start()
	s.StopWithError("Failed!")
}

func TestSpinner_StopWithWarning(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	s := iostreams.NewSpinnerWithWriter("Loading...", buf)

	s.Start()
	s.StopWithWarning("Warning!")
}

func TestSpinner_UpdateMessage(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	s := iostreams.NewSpinnerWithWriter("Loading...", buf)

	s.Start()
	s.UpdateMessage("Still loading...")
	s.Stop()
}

func TestSpinner_Active(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	s := iostreams.NewSpinnerWithWriter("Loading...", buf)

	if s.Active() {
		t.Error("Active() should be false before Start()")
	}

	s.Start()
	// Note: Active() state depends on spinner's goroutine timing
	s.Stop()
}

func TestSpinner_Reverse(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	s := iostreams.NewSpinnerWithWriter("Loading...", buf)

	// Should not panic
	s.Reverse()
}

// Test spinner presets

func TestDiscoverySpinner(t *testing.T) {
	t.Parallel()

	s := iostreams.DiscoverySpinner()
	if s == nil {
		t.Fatal("DiscoverySpinner() returned nil")
	}
}

func TestFirmwareSpinner(t *testing.T) {
	t.Parallel()

	s := iostreams.FirmwareSpinner("Updating")
	if s == nil {
		t.Fatal("FirmwareSpinner() returned nil")
	}
}

func TestBackupSpinner(t *testing.T) {
	t.Parallel()

	s := iostreams.BackupSpinner("Creating")
	if s == nil {
		t.Fatal("BackupSpinner() returned nil")
	}
}

func TestConnectingSpinner(t *testing.T) {
	t.Parallel()

	s := iostreams.ConnectingSpinner("device.local")
	if s == nil {
		t.Fatal("ConnectingSpinner() returned nil")
	}
}

func TestLoadingSpinner(t *testing.T) {
	t.Parallel()

	s := iostreams.LoadingSpinner("Please wait...")
	if s == nil {
		t.Fatal("LoadingSpinner() returned nil")
	}
}

func TestWithSpinner_Success(t *testing.T) {
	t.Parallel()

	err := iostreams.WithSpinner("Working...", func() error {
		return nil
	})

	if err != nil {
		t.Errorf("WithSpinner() with successful fn should not return error, got %v", err)
	}
}

func TestWithSpinner_Error(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("test error")
	err := iostreams.WithSpinner("Working...", func() error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("WithSpinner() should return fn error, got %v", err)
	}
}

func TestWithSpinnerResult_Success(t *testing.T) {
	t.Parallel()

	result, err := iostreams.WithSpinnerResult("Working...", func() (string, error) {
		return "result value", nil
	})

	if err != nil {
		t.Errorf("WithSpinnerResult() should not return error, got %v", err)
	}
	if result != "result value" {
		t.Errorf("WithSpinnerResult() result = %q, want %q", result, "result value")
	}
}

func TestWithSpinnerResult_Error(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("test error")
	result, err := iostreams.WithSpinnerResult("Working...", func() (string, error) {
		return "partial", expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("WithSpinnerResult() should return fn error, got %v", err)
	}
	if result != "partial" {
		t.Errorf("WithSpinnerResult() should return partial result, got %q", result)
	}
}
