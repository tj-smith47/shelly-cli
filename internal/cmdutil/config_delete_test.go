package cmdutil_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestNewConfigDeleteCommand_Structure(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource:   "scene",
		ExistsFunc: func(_ string) (any, bool) { return nil, true },
		DeleteFunc: func(_ string) error { return nil },
	})

	if cmd == nil {
		t.Fatal("NewConfigDeleteCommand returned nil")
	}

	// Check command metadata
	if cmd.Use != "delete <scene>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "delete <scene>")
	}
	if cmd.Short != "Delete a scene" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Delete a scene")
	}

	// Check aliases
	aliases := cmd.Aliases
	expectedAliases := []string{"rm", "del", "remove"}
	if len(aliases) != len(expectedAliases) {
		t.Errorf("Aliases = %v, want %v", aliases, expectedAliases)
	}
	for i, a := range expectedAliases {
		if aliases[i] != a {
			t.Errorf("Aliases[%d] = %q, want %q", i, aliases[i], a)
		}
	}

	// Check --yes flag exists (confirmation commands)
	flag := cmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("--yes flag should exist for commands with confirmation")
	}
}

func TestNewConfigDeleteCommand_SkipConfirmation(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource:         "alias",
		SkipConfirmation: true,
		ExistsFunc:       func(_ string) (any, bool) { return nil, true },
		DeleteFunc:       func(_ string) error { return nil },
	})

	// Check --yes flag does NOT exist (skip confirmation commands)
	flag := cmd.Flags().Lookup("yes")
	if flag != nil {
		t.Error("--yes flag should NOT exist when SkipConfirmation is true")
	}
}

func TestNewConfigDeleteCommand_Execute_Success(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	var deletedName string

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ExistsFunc: func(name string) (any, bool) {
			return map[string]string{"name": name}, true
		},
		DeleteFunc: func(name string) error {
			deletedName = name
			return nil
		},
	})

	cmd.SetArgs([]string{"movie-night", "--yes"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if deletedName != "movie-night" {
		t.Errorf("deletedName = %q, want %q", deletedName, "movie-night")
	}

	// Check success output
	if !strings.Contains(out.String(), "Scene") && !strings.Contains(out.String(), "deleted") {
		t.Errorf("output should contain success message, got: %s", out.String())
	}
}

func TestNewConfigDeleteCommand_Execute_SkipConfirmation(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	var deletedName string

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource:         "alias",
		SkipConfirmation: true,
		ExistsFunc: func(name string) (any, bool) {
			return nil, true
		},
		DeleteFunc: func(name string) error {
			deletedName = name
			return nil
		},
	})

	// No --yes flag needed when SkipConfirmation is true
	cmd.SetArgs([]string{"lights"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if deletedName != "lights" {
		t.Errorf("deletedName = %q, want %q", deletedName, "lights")
	}
}

func TestNewConfigDeleteCommand_Execute_NotFound(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ExistsFunc: func(_ string) (any, bool) {
			return nil, false // Not found
		},
		DeleteFunc: func(_ string) error {
			return nil
		},
	})

	cmd.SetArgs([]string{"nonexistent"})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute should have failed for nonexistent resource")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, should contain 'not found'", err.Error())
	}
}

func TestNewConfigDeleteCommand_Execute_DeleteError(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ExistsFunc: func(_ string) (any, bool) {
			return nil, true
		},
		DeleteFunc: func(_ string) error {
			return errors.New("disk full")
		},
	})

	cmd.SetArgs([]string{"movie-night", "--yes"})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute should have failed")
	}

	if !strings.Contains(err.Error(), "failed to delete") {
		t.Errorf("error = %q, should contain 'failed to delete'", err.Error())
	}
}

func TestNewConfigDeleteCommand_Execute_Cancelled(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	deleteWasCalled := false

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ExistsFunc: func(_ string) (any, bool) {
			return nil, true
		},
		DeleteFunc: func(_ string) error {
			deleteWasCalled = true
			return nil
		},
	})

	// Without --yes, in non-TTY mode, confirmation returns false (default)
	cmd.SetArgs([]string{"movie-night"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if deleteWasCalled {
		t.Error("DeleteFunc should not be called when confirmation is denied")
	}

	// Check cancellation message
	if !strings.Contains(out.String(), "cancelled") {
		t.Errorf("output should contain cancellation message, got: %s", out.String())
	}
}

func TestNewConfigDeleteCommand_Execute_WithInfoFunc(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	type scene struct {
		Name    string
		Actions int
	}

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ExistsFunc: func(name string) (any, bool) {
			return scene{Name: name, Actions: 5}, true
		},
		DeleteFunc: func(_ string) error {
			return nil
		},
		InfoFunc: func(resource any, name string) string {
			if _, ok := resource.(scene); !ok {
				return "invalid type"
			}
			return "Delete scene \"" + name + "\" with 5 action(s)?"
		},
	})

	cmd.SetArgs([]string{"movie-night", "--yes"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestNewConfigDeleteCommand_RequiresArg(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ExistsFunc: func(_ string) (any, bool) {
			return nil, true
		},
		DeleteFunc: func(_ string) error {
			return nil
		},
	})

	// No args - should fail
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute should have failed without name argument")
	}
}

func TestNewConfigDeleteCommand_WithValidArgsFunc(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := factories.NewConfigDeleteCommand(f, factories.ConfigDeleteOpts{
		Resource: "scene",
		ValidArgsFunc: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"scene1", "scene2"}, cobra.ShellCompDirectiveNoFileComp
		},
		ExistsFunc: func(_ string) (any, bool) {
			return nil, true
		},
		DeleteFunc: func(_ string) error {
			return nil
		},
	})

	// Verify ValidArgsFunction is set
	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set when ValidArgsFunc is provided")
	}
}
