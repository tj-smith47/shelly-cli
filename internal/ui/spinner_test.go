package ui

import (
	"errors"
	"testing"
	"time"

	"github.com/spf13/viper"
)

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestNewSpinner(t *testing.T) {
	resetViper()

	s := NewSpinner("Loading...")

	if s == nil {
		t.Fatal("NewSpinner() returned nil")
	}
	if s.s == nil {
		t.Fatal("NewSpinner().s is nil")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestSpinner_StartStop_NotQuiet(t *testing.T) {
	resetViper()
	viper.Set("quiet", false)

	s := NewSpinner("Loading...")
	s.Start()

	// Give spinner a moment to start
	time.Sleep(10 * time.Millisecond)

	s.Stop()

	// Should not panic and spinner should be stopped
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestSpinner_StartStop_Quiet(t *testing.T) {
	resetViper()
	viper.Set("quiet", true)

	s := NewSpinner("Loading...")
	s.Start() // Should not actually start

	// Give it a moment
	time.Sleep(10 * time.Millisecond)

	s.Stop() // Should not panic
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestSpinner_UpdateMessage(t *testing.T) {
	resetViper()
	viper.Set("quiet", true) // Keep quiet to avoid output

	s := NewSpinner("Initial")
	s.UpdateMessage("Updated message")

	// Just verify it doesn't panic
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestSpinner_StopWithSuccess(t *testing.T) {
	resetViper()
	viper.Set("quiet", true)

	s := NewSpinner("Loading...")
	s.StopWithSuccess("Complete")

	// Verify the final message was set (can't easily capture spinner output)
	if s.s.FinalMSG == "" {
		t.Error("StopWithSuccess should set FinalMSG")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestSpinner_StopWithError(t *testing.T) {
	resetViper()
	viper.Set("quiet", true)

	s := NewSpinner("Loading...")
	s.StopWithError("Failed")

	// Verify the final message was set
	if s.s.FinalMSG == "" {
		t.Error("StopWithError should set FinalMSG")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestWithSpinner_Success(t *testing.T) {
	resetViper()
	viper.Set("quiet", true)

	called := false
	err := WithSpinner("Testing", func() error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("WithSpinner() error = %v, want nil", err)
	}
	if !called {
		t.Error("WithSpinner() did not call the function")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestWithSpinner_Error(t *testing.T) {
	resetViper()
	viper.Set("quiet", true)

	testErr := errors.New("test error")
	err := WithSpinner("Testing", func() error {
		return testErr
	})

	if err == nil {
		t.Error("WithSpinner() should return error")
	}
	if err.Error() != testErr.Error() {
		t.Errorf("WithSpinner() error = %v, want %v", err, testErr)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestSpinner_Options(t *testing.T) {
	resetViper()

	// Test that options can be passed (even if they don't do anything visible)
	customOpt := func(s *Spinner) {
		// Custom option that could modify the spinner
	}

	s := NewSpinner("Loading...", customOpt)

	if s == nil {
		t.Fatal("NewSpinner() with options returned nil")
	}
}
