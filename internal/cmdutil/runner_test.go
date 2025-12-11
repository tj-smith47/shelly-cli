package cmdutil_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func testIOStreams() (ios *iostreams.IOStreams, out, errOut *bytes.Buffer) {
	in := &bytes.Buffer{}
	out = &bytes.Buffer{}
	errOut = &bytes.Buffer{}
	ios = iostreams.Test(in, out, errOut)
	return ios, out, errOut
}

func TestRunWithSpinner(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()

		err := cmdutil.RunWithSpinner(context.Background(), ios, "Working...", func(_ context.Context) error {
			return nil
		})

		if err != nil {
			t.Errorf("RunWithSpinner() error = %v, want nil", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()
		expectedErr := errors.New("test error")

		err := cmdutil.RunWithSpinner(context.Background(), ios, "Working...", func(_ context.Context) error {
			return expectedErr
		})

		if !errors.Is(err, expectedErr) {
			t.Errorf("RunWithSpinner() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := cmdutil.RunWithSpinner(ctx, ios, "Working...", func(ctx context.Context) error {
			return ctx.Err()
		})

		if !errors.Is(err, context.Canceled) {
			t.Errorf("RunWithSpinner() error = %v, want context.Canceled", err)
		}
	})
}

func TestRunWithSpinnerResult(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()

		result, err := cmdutil.RunWithSpinnerResult(context.Background(), ios, "Working...", func(_ context.Context) (string, error) {
			return "test result", nil
		})

		if err != nil {
			t.Errorf("RunWithSpinnerResult() error = %v, want nil", err)
		}
		if result != "test result" {
			t.Errorf("RunWithSpinnerResult() result = %q, want %q", result, "test result")
		}
	})

	t.Run("error with partial result", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()
		expectedErr := errors.New("test error")

		result, err := cmdutil.RunWithSpinnerResult(context.Background(), ios, "Working...", func(_ context.Context) (string, error) {
			return "partial", expectedErr
		})

		if !errors.Is(err, expectedErr) {
			t.Errorf("RunWithSpinnerResult() error = %v, want %v", err, expectedErr)
		}
		if result != "partial" {
			t.Errorf("RunWithSpinnerResult() result = %q, want %q", result, "partial")
		}
	})

	t.Run("with struct result", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()

		type Result struct {
			Value int
			Name  string
		}

		result, err := cmdutil.RunWithSpinnerResult(context.Background(), ios, "Working...", func(_ context.Context) (Result, error) {
			return Result{Value: 42, Name: "test"}, nil
		})

		if err != nil {
			t.Errorf("RunWithSpinnerResult() error = %v, want nil", err)
		}
		if result.Value != 42 || result.Name != "test" {
			t.Errorf("RunWithSpinnerResult() result = %+v, want {Value: 42, Name: test}", result)
		}
	})
}

func TestRunBatch(t *testing.T) {
	t.Parallel()

	t.Run("empty targets", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()
		svc := shelly.NewService()

		err := cmdutil.RunBatch(context.Background(), ios, svc, []string{}, 5, func(_ context.Context, _ *shelly.Service, _ string) error {
			return nil
		})

		if err != nil {
			t.Errorf("RunBatch() with empty targets error = %v, want nil", err)
		}
	})

	t.Run("single target success", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()
		svc := shelly.NewService()
		called := false

		err := cmdutil.RunBatch(context.Background(), ios, svc, []string{"device1"}, 5, func(_ context.Context, _ *shelly.Service, device string) error {
			called = true
			if device != "device1" {
				t.Errorf("action called with device = %q, want %q", device, "device1")
			}
			return nil
		})

		if err != nil {
			t.Errorf("RunBatch() error = %v, want nil", err)
		}
		if !called {
			t.Error("action was not called")
		}
	})

	t.Run("multiple targets", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()
		svc := shelly.NewService()
		targets := []string{"device1", "device2", "device3"}
		callCount := 0

		err := cmdutil.RunBatch(context.Background(), ios, svc, targets, 5, func(_ context.Context, _ *shelly.Service, _ string) error {
			callCount++
			return nil
		})

		if err != nil {
			t.Errorf("RunBatch() error = %v, want nil", err)
		}
		if callCount != 3 {
			t.Errorf("action called %d times, want 3", callCount)
		}
	})

	t.Run("individual errors don't stop batch", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()
		svc := shelly.NewService()
		targets := []string{"device1", "device2", "device3"}

		err := cmdutil.RunBatch(context.Background(), ios, svc, targets, 5, func(_ context.Context, _ *shelly.Service, device string) error {
			if device == "device2" {
				return errors.New("device2 failed")
			}
			return nil
		})

		// Batch should not return error even if individual operations fail
		if err != nil {
			t.Errorf("RunBatch() error = %v, want nil (errors logged but don't stop batch)", err)
		}
	})
}

func TestRunBatchComponent(t *testing.T) {
	t.Parallel()

	t.Run("passes component ID", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()
		svc := shelly.NewService()
		receivedID := -1

		err := cmdutil.RunBatchComponent(context.Background(), ios, svc, []string{"device1"}, 42, 5, func(_ context.Context, _ *shelly.Service, _ string, id int) error {
			receivedID = id
			return nil
		})

		if err != nil {
			t.Errorf("RunBatchComponent() error = %v, want nil", err)
		}
		if receivedID != 42 {
			t.Errorf("action received componentID = %d, want 42", receivedID)
		}
	})
}

func TestRunBatchWithResults(t *testing.T) {
	t.Parallel()

	t.Run("empty targets", func(t *testing.T) {
		t.Parallel()
		svc := shelly.NewService()

		results := cmdutil.RunBatchWithResults(context.Background(), svc, []string{}, 5, func(_ context.Context, _ *shelly.Service, _ string) error {
			return nil
		})

		if results != nil {
			t.Errorf("RunBatchWithResults() with empty targets = %v, want nil", results)
		}
	})

	t.Run("all success", func(t *testing.T) {
		t.Parallel()
		svc := shelly.NewService()
		targets := []string{"device1", "device2"}

		results := cmdutil.RunBatchWithResults(context.Background(), svc, targets, 5, func(_ context.Context, _ *shelly.Service, _ string) error {
			return nil
		})

		if len(results) != 2 {
			t.Fatalf("RunBatchWithResults() returned %d results, want 2", len(results))
		}

		for i, r := range results {
			if !r.Success {
				t.Errorf("results[%d].Success = false, want true", i)
			}
			if r.Error != nil {
				t.Errorf("results[%d].Error = %v, want nil", i, r.Error)
			}
		}
	})

	t.Run("mixed results", func(t *testing.T) {
		t.Parallel()
		svc := shelly.NewService()
		targets := []string{"device1", "device2", "device3"}
		testErr := errors.New("device2 failed")

		results := cmdutil.RunBatchWithResults(context.Background(), svc, targets, 5, func(_ context.Context, _ *shelly.Service, device string) error {
			if device == "device2" {
				return testErr
			}
			return nil
		})

		if len(results) != 3 {
			t.Fatalf("RunBatchWithResults() returned %d results, want 3", len(results))
		}

		// Check device1 and device3 succeeded
		if !results[0].Success || !results[2].Success {
			t.Error("device1 and device3 should succeed")
		}

		// Check device2 failed
		if results[1].Success {
			t.Error("device2 should fail")
		}
		if !errors.Is(results[1].Error, testErr) {
			t.Errorf("device2 error = %v, want %v", results[1].Error, testErr)
		}
	})

	t.Run("preserves order", func(t *testing.T) {
		t.Parallel()
		svc := shelly.NewService()
		targets := []string{"a", "b", "c", "d", "e"}

		results := cmdutil.RunBatchWithResults(context.Background(), svc, targets, 5, func(_ context.Context, _ *shelly.Service, _ string) error {
			return nil
		})

		for i, target := range targets {
			if results[i].Device != target {
				t.Errorf("results[%d].Device = %q, want %q", i, results[i].Device, target)
			}
		}
	})
}

func TestBatchResult(t *testing.T) {
	t.Parallel()

	t.Run("success result", func(t *testing.T) {
		t.Parallel()
		result := cmdutil.BatchResult{
			Device:  "device1",
			Success: true,
			Message: "success",
		}

		if result.Device != "device1" {
			t.Errorf("Device = %q, want %q", result.Device, "device1")
		}
		if !result.Success {
			t.Error("Success = false, want true")
		}
	})

	t.Run("error result", func(t *testing.T) {
		t.Parallel()
		testErr := errors.New("test error")
		result := cmdutil.BatchResult{
			Device:  "device1",
			Success: false,
			Message: "test error",
			Error:   testErr,
		}

		if result.Success {
			t.Error("Success = true, want false")
		}
		if !errors.Is(result.Error, testErr) {
			t.Errorf("Error = %v, want %v", result.Error, testErr)
		}
	})
}

func TestRunSimple(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		svc := shelly.NewService()

		err := cmdutil.RunSimple(context.Background(), ios, svc, "device1", 0, "Working...", "Done!", func(_ context.Context, _ *shelly.Service, _ string, _ int) error {
			return nil
		})

		if err != nil {
			t.Errorf("RunSimple() error = %v, want nil", err)
		}

		// Check success message was printed
		output := out.String()
		if output == "" {
			t.Error("success message should be printed")
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()
		svc := shelly.NewService()
		expectedErr := errors.New("test error")

		err := cmdutil.RunSimple(context.Background(), ios, svc, "device1", 0, "Working...", "Done!", func(_ context.Context, _ *shelly.Service, _ string, _ int) error {
			return expectedErr
		})

		if !errors.Is(err, expectedErr) {
			t.Errorf("RunSimple() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("passes device and component ID", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()
		svc := shelly.NewService()
		receivedDevice := ""
		receivedID := -1

		err := cmdutil.RunSimple(context.Background(), ios, svc, "mydevice", 5, "Working...", "Done!", func(_ context.Context, _ *shelly.Service, device string, id int) error {
			receivedDevice = device
			receivedID = id
			return nil
		})

		if err != nil {
			t.Errorf("RunSimple() error = %v, want nil", err)
		}
		if receivedDevice != "mydevice" {
			t.Errorf("device = %q, want %q", receivedDevice, "mydevice")
		}
		if receivedID != 5 {
			t.Errorf("componentID = %d, want 5", receivedID)
		}
	})
}

func TestPrintBatchSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		results     []cmdutil.BatchResult
		checkStdout bool // if false, check stderr
	}{
		{
			name: "all success",
			results: []cmdutil.BatchResult{
				{Device: "d1", Success: true},
				{Device: "d2", Success: true},
			},
			checkStdout: true,
		},
		{
			name: "all failed",
			results: []cmdutil.BatchResult{
				{Device: "d1", Success: false},
				{Device: "d2", Success: false},
			},
			checkStdout: false, // Error output goes to stderr
		},
		{
			name: "mixed results",
			results: []cmdutil.BatchResult{
				{Device: "d1", Success: true},
				{Device: "d2", Success: false},
			},
			checkStdout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ios, out, errOut := testIOStreams()

			cmdutil.PrintBatchSummary(ios, tt.results)

			var output string
			if tt.checkStdout {
				output = out.String()
			} else {
				output = errOut.String()
			}
			if output == "" {
				t.Error("PrintBatchSummary should print output")
			}
		})
	}
}
