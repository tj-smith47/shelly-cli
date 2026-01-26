package diagram

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"

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

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "diagram" {
		t.Errorf("Use = %q, want 'diagram'", cmd.Use)
	}

	if len(cmd.Aliases) == 0 {
		t.Fatal("Aliases should not be empty")
	}

	aliasMap := make(map[string]bool)
	for _, alias := range cmd.Aliases {
		aliasMap[alias] = true
	}
	for _, expected := range []string{"wiring", "diag"} {
		if !aliasMap[expected] {
			t.Errorf("expected alias %q not found", expected)
		}
	}

	if cmd.Short == "" {
		t.Error("Short is empty")
	}

	if cmd.Long == "" {
		t.Error("Long is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	modelFlag := cmd.Flags().Lookup("model")
	if modelFlag == nil {
		t.Fatal("--model flag not found")
	}
	if modelFlag.Shorthand != "m" {
		t.Errorf("model shorthand = %q, want 'm'", modelFlag.Shorthand)
	}

	genFlag := cmd.Flags().Lookup("generation")
	if genFlag == nil {
		t.Fatal("--generation flag not found")
	}
	if genFlag.Shorthand != "g" {
		t.Errorf("generation shorthand = %q, want 'g'", genFlag.Shorthand)
	}

	styleFlag := cmd.Flags().Lookup("style")
	if styleFlag == nil {
		t.Fatal("--style flag not found")
	}
	if styleFlag.Shorthand != "s" {
		t.Errorf("style shorthand = %q, want 's'", styleFlag.Shorthand)
	}
	if styleFlag.DefValue != "schematic" {
		t.Errorf("style default = %q, want 'schematic'", styleFlag.DefValue)
	}
}

func TestNewCommand_ModelRequired(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	modelFlag := cmd.Flags().Lookup("model")
	if modelFlag == nil {
		t.Fatal("--model flag not found")
	}

	// Cobra annotations for required flags
	ann := modelFlag.Annotations
	if ann == nil {
		t.Fatal("model flag should have annotations (required)")
	}
	required, ok := ann["cobra_annotation_bash_completion_one_required_flag"]
	if !ok || len(required) == 0 {
		t.Error("model flag should be marked as required")
	}
}

func TestRun_SchematicDefault(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Model:   "plus-1",
		Style:   "schematic",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Shelly Plus 1") {
		t.Errorf("output should contain device name, got:\n%s", output)
	}
	if !strings.Contains(output, "Gen2") {
		t.Errorf("output should contain generation, got:\n%s", output)
	}
	if !strings.Contains(output, "Specs:") {
		t.Errorf("output should contain specs footer, got:\n%s", output)
	}
}

func TestRun_CompactStyle(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Model:   "pro-4pm",
		Style:   "compact",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Shelly Pro 4PM") {
		t.Errorf("output should contain device name, got:\n%s", output)
	}
	if !strings.Contains(output, "O1") {
		t.Errorf("output should contain terminal labels, got:\n%s", output)
	}
}

func TestRun_DetailedStyle(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Model:   "dimmer-2",
		Style:   "detailed",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Shelly Dimmer 2") {
		t.Errorf("output should contain device name, got:\n%s", output)
	}
	if !strings.Contains(output, "POWER SUPPLY") {
		t.Errorf("detailed output should contain section headers, got:\n%s", output)
	}
}

func TestRun_WithGeneration(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:    tf.Factory,
		Model:      "1",
		Generation: "1",
		Style:      "schematic",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Shelly 1") {
		t.Errorf("output should contain 'Shelly 1', got:\n%s", output)
	}
	if !strings.Contains(output, "Gen1") {
		t.Errorf("output should show Gen1, got:\n%s", output)
	}
}

func TestRun_Gen4Model(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Model:   "1-gen4",
		Style:   "schematic",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Shelly 1 Gen4") {
		t.Errorf("output should contain 'Shelly 1 Gen4', got:\n%s", output)
	}
	if !strings.Contains(output, "Gen4") {
		t.Errorf("output should show Gen4, got:\n%s", output)
	}
}

func TestRun_GenFilterWord(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:    tf.Factory,
		Model:      "1",
		Generation: "gen4",
		Style:      "schematic",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Shelly 1 Gen4") {
		t.Errorf("output should contain 'Shelly 1 Gen4', got:\n%s", output)
	}
}

func TestRun_InvalidStyle(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Model:   "plus-1",
		Style:   "invalid-style",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for invalid style")
	}
	if !strings.Contains(err.Error(), "invalid style") {
		t.Errorf("error should mention invalid style, got: %v", err)
	}
}

func TestRun_InvalidGeneration(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:    tf.Factory,
		Model:      "plus-1",
		Generation: "gen99",
		Style:      "schematic",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for invalid generation")
	}
	if !strings.Contains(err.Error(), "invalid generation") {
		t.Errorf("error should mention invalid generation, got: %v", err)
	}
}

func TestRun_MultiGenDefaultsToLatest(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Model:   "1",
		Style:   "schematic",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Gen4") {
		t.Errorf("multi-gen slug should default to latest (Gen4), got:\n%s", output)
	}
	if !strings.Contains(output, "Shelly 1 Gen4") {
		t.Errorf("output should contain 'Shelly 1 Gen4', got:\n%s", output)
	}
	if !strings.Contains(output, "Also available") {
		t.Errorf("multi-gen output should show generation note, got:\n%s", output)
	}
	if !strings.Contains(output, "-g") {
		t.Errorf("generation note should mention -g flag, got:\n%s", output)
	}
}

func TestRun_SingleGenNoNote(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Model:   "plus-1",
		Style:   "schematic",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if strings.Contains(output, "Also available") {
		t.Errorf("single-gen slug should not show generation note, got:\n%s", output)
	}
}

func TestRun_ExplicitGenNoNote(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:    tf.Factory,
		Model:      "1",
		Generation: "1",
		Style:      "schematic",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if strings.Contains(output, "Also available") {
		t.Errorf("explicit generation should not show generation note, got:\n%s", output)
	}
}

func TestRun_ModelNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Model:   "nonexistent-device",
		Style:   "schematic",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for nonexistent model")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestRun_ModelNotFoundWithGeneration(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:    tf.Factory,
		Model:      "plus-1",
		Generation: "1",
		Style:      "schematic",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for plus-1 in gen1")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestRun_AllTopologies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		model string
		gen   string
	}{
		{"SingleRelay", "plus-1", ""},
		{"DualRelay", "plus-2pm", ""},
		{"QuadRelay", "pro-4pm", ""},
		{"Dimmer", "dimmer-2", ""},
		{"InputOnly", "i3", ""},
		{"Plug", "plug-s", "1"},
		{"EnergyMonitor", "em", ""},
		{"RGBW", "rgbw2", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			opts := &Options{
				Factory:    tf.Factory,
				Model:      tt.model,
				Generation: tt.gen,
				Style:      "schematic",
			}

			err := run(context.Background(), opts)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tt.name, err)
			}

			output := tf.OutString()
			if output == "" {
				t.Errorf("%s: expected non-empty output", tt.name)
			}
			if !strings.Contains(output, "Specs:") {
				t.Errorf("%s: output should contain specs", tt.name)
			}
		})
	}
}

func TestRun_AllStyles(t *testing.T) {
	t.Parallel()

	for _, style := range []string{"schematic", "compact", "detailed"} {
		t.Run(style, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			opts := &Options{
				Factory: tf.Factory,
				Model:   "plus-1",
				Style:   style,
			}

			err := run(context.Background(), opts)
			if err != nil {
				t.Fatalf("unexpected error for style %s: %v", style, err)
			}

			output := tf.OutString()
			if !strings.Contains(output, "Shelly Plus 1") {
				t.Errorf("style %s: output should contain device name", style)
			}
		})
	}
}

func TestRun_CaseInsensitiveModel(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Model:   "PLUS-1",
		Style:   "schematic",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Shelly Plus 1") {
		t.Errorf("case-insensitive lookup should work, got:\n%s", output)
	}
}

func TestRun_AltSlug(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Model:   "shelly25",
		Style:   "compact",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Shelly 2.5") {
		t.Errorf("alt slug should resolve, got:\n%s", output)
	}
}

func TestRun_PlugTopology(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:    tf.Factory,
		Model:      "plug-s",
		Generation: "1",
		Style:      "schematic",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Plug") {
		t.Errorf("plug output should mention 'Plug', got:\n%s", output)
	}
	// Plug has no wiring terminals
	if !strings.Contains(output, "Relay") {
		t.Errorf("plug output should mention relay, got:\n%s", output)
	}
}

func TestRun_EnergyMonitor3Phase(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:    tf.Factory,
		Model:      "3em",
		Generation: "1",
		Style:      "detailed",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "CT3") {
		t.Errorf("3EM should show CT3, got:\n%s", output)
	}
	if !strings.Contains(output, "3-channel") {
		t.Errorf("3EM should mention 3-channel, got:\n%s", output)
	}
}

func TestRun_InputOnly4Inputs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Model:   "plus-i4",
		Style:   "detailed",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "SW4") {
		t.Errorf("i4 should show SW4, got:\n%s", output)
	}
	if !strings.Contains(output, "4 digital inputs") {
		t.Errorf("i4 should mention 4 digital inputs, got:\n%s", output)
	}
}

func TestRun_RGBWTopology(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Model:   "rgbw2",
		Style:   "detailed",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	for _, channel := range []string{"V+", "GND", "Red", "Green", "Blue", "White"} {
		if !strings.Contains(output, channel) {
			t.Errorf("RGBW output should contain %q, got:\n%s", channel, output)
		}
	}
}

func TestNewCommand_ExecuteViaRunE(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--model", "plus-1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Shelly Plus 1") {
		t.Errorf("output should contain device name, got:\n%s", output)
	}
}

func TestNewCommand_CompletionModel(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cmd := NewCommand(f)

	fn, ok := cmd.GetFlagCompletionFunc("model")
	if !ok {
		t.Fatal("model flag completion function not registered")
	}
	results, dir := fn(cmd, nil, "")
	if len(results) == 0 {
		t.Error("model completion should return slugs")
	}
	if dir != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %d, want ShellCompDirectiveNoFileComp", dir)
	}
}

func TestNewCommand_CompletionGeneration(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cmd := NewCommand(f)

	fn, ok := cmd.GetFlagCompletionFunc("generation")
	if !ok {
		t.Fatal("generation flag completion function not registered")
	}
	results, dir := fn(cmd, nil, "")
	if len(results) == 0 {
		t.Error("generation completion should return values")
	}
	if dir != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %d, want ShellCompDirectiveNoFileComp", dir)
	}
}

func TestNewCommand_CompletionStyle(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cmd := NewCommand(f)

	fn, ok := cmd.GetFlagCompletionFunc("style")
	if !ok {
		t.Fatal("style flag completion function not registered")
	}
	results, dir := fn(cmd, nil, "")
	if len(results) == 0 {
		t.Error("style completion should return values")
	}
	if dir != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %d, want ShellCompDirectiveNoFileComp", dir)
	}
}

func TestRun_OutputContainsNewlines(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Model:   "plus-1",
		Style:   "schematic",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := tf.OutString()
	lines := strings.Split(output, "\n")
	if len(lines) < 5 {
		t.Errorf("output should have multiple lines, got %d", len(lines))
	}
}
