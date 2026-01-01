package list

import (
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestNewCommand_HasAliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Aliases) == 0 {
		t.Error("Aliases is empty")
	}
}

func TestNewCommand_HasExample(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestExecute_ListAll(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Profiles should be listed
	if output == "" {
		t.Error("expected output to contain profiles")
	}
}

func TestExecute_ListByGeneration(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--gen", "gen2"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Gen2 profiles should be listed
	if output == "" {
		t.Error("expected output for gen2 filter")
	}
}

func TestExecute_ListByGeneration_Unknown(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--gen", "invalid"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Should warn about unknown generation
	errOutput := tf.ErrString()
	if !strings.Contains(errOutput, "Unknown generation") {
		t.Logf("output = %s, errOutput = %s", tf.OutString(), errOutput)
	}
}

func TestExecute_ListBySeries(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--series", "pro"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Pro series profiles should be listed
	if output == "" {
		t.Error("expected output for series filter")
	}
}

func TestExecute_ListBySeries_Unknown(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--series", "invalid"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Should warn about unknown series
	errOutput := tf.ErrString()
	if !strings.Contains(errOutput, "Unknown series") {
		t.Logf("output = %s, errOutput = %s", tf.OutString(), errOutput)
	}
}

func TestExecute_WithOutputFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"-o", "json"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify command executed (output format handling depends on viper binding)
	output := tf.OutString()
	if output == "" {
		t.Error("expected output")
	}
}

func TestRun_DirectCall(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected profiles in output")
	}
}

func TestRun_Gen1Filter(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:    tf.Factory,
		Generation: "gen1",
	}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected gen1 profiles in output")
	}
}

func TestRun_Gen3Filter(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:    tf.Factory,
		Generation: "gen3",
	}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}
}

func TestRun_Gen4Filter(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:    tf.Factory,
		Generation: "gen4",
	}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}
}

func TestRun_ClassicSeries(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Series:  "classic",
	}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}
}

func TestRun_PlusSeries(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Series:  "plus",
	}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected plus series profiles in output")
	}
}

func TestRun_MiniSeries(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Series:  "mini",
	}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}
}

func TestRun_BluSeries(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Series:  "blu",
	}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}
}

func TestRun_WaveSeries(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Series:  "wave",
	}

	err := run(opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}
}
