package factories_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestNewCoverCommand_Open(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewCoverCommand(f, factories.CoverOpts{
		Action: factories.CoverActionOpen,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int, _ *int) error {
			return nil
		},
	})

	if cmd == nil {
		t.Fatal("NewCoverCommand returned nil")
	}

	// Check command metadata
	if cmd.Use != "open <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "open <device>")
	}
	if cmd.Short != "Open cover" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Open cover")
	}

	// Check aliases
	aliases := cmd.Aliases
	if len(aliases) != 2 || aliases[0] != "up" || aliases[1] != "raise" {
		t.Errorf("Aliases = %v, want [up raise]", aliases)
	}

	// Check --duration flag exists
	flag := cmd.Flags().Lookup("duration")
	if flag == nil {
		t.Error("--duration flag should exist for open action")
	}

	// Check --id flag exists
	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("--id flag should exist")
	}
}

func TestNewCoverCommand_Close(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewCoverCommand(f, factories.CoverOpts{
		Action: factories.CoverActionClose,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int, _ *int) error {
			return nil
		},
	})

	if cmd == nil {
		t.Fatal("NewCoverCommand returned nil")
	}

	// Check command metadata
	if cmd.Use != "close <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "close <device>")
	}
	if cmd.Short != "Close cover" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Close cover")
	}

	// Check aliases
	aliases := cmd.Aliases
	if len(aliases) != 2 || aliases[0] != "down" || aliases[1] != "lower" {
		t.Errorf("Aliases = %v, want [down lower]", aliases)
	}

	// Check --duration flag exists
	flag := cmd.Flags().Lookup("duration")
	if flag == nil {
		t.Error("--duration flag should exist for close action")
	}
}

func TestNewCoverCommand_Stop(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewCoverCommand(f, factories.CoverOpts{
		Action: factories.CoverActionStop,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int, _ *int) error {
			return nil
		},
	})

	if cmd == nil {
		t.Fatal("NewCoverCommand returned nil")
	}

	// Check command metadata
	if cmd.Use != "stop <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "stop <device>")
	}
	if cmd.Short != "Stop cover" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Stop cover")
	}

	// Check aliases
	aliases := cmd.Aliases
	if len(aliases) != 2 || aliases[0] != "halt" || aliases[1] != "pause" {
		t.Errorf("Aliases = %v, want [halt pause]", aliases)
	}

	// Check --duration flag does NOT exist for stop
	flag := cmd.Flags().Lookup("duration")
	if flag != nil {
		t.Error("--duration flag should NOT exist for stop action")
	}

	// Check --id flag exists
	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("--id flag should exist")
	}
}

func TestNewCoverCommand_Execute_Success(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	var calledDevice string
	var calledID int
	var calledDuration *int

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewCoverCommand(f, factories.CoverOpts{
		Action: factories.CoverActionOpen,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, device string, id int, duration *int) error {
			calledDevice = device
			calledID = id
			calledDuration = duration
			return nil
		},
	})

	cmd.SetArgs([]string{testDevice, "--id", "2", "--duration", "5"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if calledDevice != testDevice {
		t.Errorf("device = %q, want %q", calledDevice, testDevice)
	}
	if calledID != 2 {
		t.Errorf("id = %d, want 2", calledID)
	}
	if calledDuration == nil || *calledDuration != 5 {
		t.Errorf("duration = %v, want 5", calledDuration)
	}
}

func TestNewCoverCommand_Execute_NoDuration(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	var calledDuration *int

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewCoverCommand(f, factories.CoverOpts{
		Action: factories.CoverActionOpen,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int, duration *int) error {
			calledDuration = duration
			return nil
		},
	})

	cmd.SetArgs([]string{testDevice})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Duration should be nil when not specified (or 0)
	if calledDuration != nil {
		t.Errorf("duration = %v, want nil", *calledDuration)
	}
}

func TestNewCoverCommand_Execute_Error(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewCoverCommand(f, factories.CoverOpts{
		Action: factories.CoverActionOpen,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int, _ *int) error {
			return errors.New("device unreachable")
		},
	})

	cmd.SetArgs([]string{testDevice})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute should have failed")
	}

	if !errors.Is(err, errors.Unwrap(err)) && err.Error() != "failed to open cover: device unreachable" {
		t.Errorf("error = %q, want to contain 'failed to open cover'", err.Error())
	}
}

func TestNewCoverCommand_Execute_Stop(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	var calledDevice string
	var calledID int

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewCoverCommand(f, factories.CoverOpts{
		Action: factories.CoverActionStop,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, device string, id int, _ *int) error {
			calledDevice = device
			calledID = id
			return nil
		},
	})

	cmd.SetArgs([]string{testDevice, "--id", "1"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if calledDevice != testDevice {
		t.Errorf("device = %q, want %q", calledDevice, testDevice)
	}
	if calledID != 1 {
		t.Errorf("id = %d, want 1", calledID)
	}
}

func TestNewCoverCommand_RequiresDevice(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewCoverCommand(f, factories.CoverOpts{
		Action: factories.CoverActionOpen,
		ServiceFunc: func(_ context.Context, _ *shelly.Service, _ string, _ int, _ *int) error {
			return nil
		},
	})

	// No args - should fail
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute should have failed without device argument")
	}
}
