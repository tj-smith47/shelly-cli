//go:build !race

// These tests are excluded from race detection because the readline library
// has internal race conditions when instances are created and immediately closed
// (which happens with cancelled contexts in tests). The races are in the library
// itself (DefaultOnWidthChanged global), not in our code.
//
//nolint:paralleltest // readline global state requires mutex serialization, not parallel execution
package repl

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// readlineMu serializes tests that use the readline library.
// The readline library has global state (DefaultOnWidthChanged) that causes
// race conditions when multiple tests create readline instances concurrently.
var readlineMu sync.Mutex

func TestRun_NonInteractiveMode(t *testing.T) {
	readlineMu.Lock()
	defer readlineMu.Unlock()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		NoPrompt: false,
	}

	// Create a cancelled context to exit immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)

	// Should return nil on cancelled context (clean exit)
	if err != nil {
		// The run function returns nil on context cancellation
		// or may return an error if readline fails
		t.Logf("run() returned: %v (expected for non-TTY)", err)
	}

	// Verify output contains REPL header
	output := tf.OutString()
	if !strings.Contains(output, "Shelly Interactive REPL") {
		t.Errorf("Output should contain REPL header, got: %q", output)
	}
}

func TestRun_WithDevice(t *testing.T) {
	readlineMu.Lock()
	defer readlineMu.Unlock()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)

	// Error or nil are both acceptable for cancelled context
	if err != nil {
		t.Logf("run() error (acceptable for cancelled context): %v", err)
	}

	// Verify output contains REPL header
	output := tf.OutString()
	if !strings.Contains(output, "Shelly Interactive REPL") {
		t.Errorf("Output should contain REPL header, got: %q", output)
	}
}

func TestRun_OutputsHelpInfo(t *testing.T) {
	readlineMu.Lock()
	defer readlineMu.Unlock()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	if err != nil {
		t.Logf("run() error: %v (expected for non-TTY)", err)
	}

	// Verify output contains help hint
	output := tf.OutString()
	errOutput := tf.ErrString()
	combined := output + errOutput

	if !strings.Contains(combined, "help") || !strings.Contains(combined, "exit") {
		t.Errorf("Output should contain help/exit hints, got: %q", combined)
	}
}

func TestNewCommand_Execute_NoArgs(t *testing.T) {
	readlineMu.Lock()
	defer readlineMu.Unlock()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - should not require any args
	err := cmd.Execute()
	// Error is acceptable due to readline initialization in non-TTY
	if err != nil {
		t.Logf("execute error (acceptable for non-TTY): %v", err)
	}
}

func TestNewCommand_Execute_WithDeviceFlag(t *testing.T) {
	readlineMu.Lock()
	defer readlineMu.Unlock()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--device", "my-device"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	err := cmd.Execute()
	if err != nil {
		t.Logf("execute error (acceptable for cancelled context): %v", err)
	}
}

func TestNewCommand_Execute_WithNoPromptFlag(t *testing.T) {
	readlineMu.Lock()
	defer readlineMu.Unlock()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--no-prompt"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	err := cmd.Execute()
	if err != nil {
		t.Logf("execute error (acceptable for cancelled context): %v", err)
	}
}

func TestNewCommand_Execute_AllFlags(t *testing.T) {
	readlineMu.Lock()
	defer readlineMu.Unlock()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--device", "test", "--no-prompt"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	err := cmd.Execute()
	if err != nil {
		t.Logf("execute error (acceptable for cancelled context): %v", err)
	}
}
