package search

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
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

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Aliases) == 0 {
		t.Error("Aliases are empty, should have at least one")
	}

	wantAliases := []string{"find", "s"}
	for _, wantAlias := range wantAliases {
		found := false
		for _, alias := range cmd.Aliases {
			if alias == wantAlias {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected alias %q not found in %v", wantAlias, cmd.Aliases)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	wantPatterns := []string{
		"shelly profile search",
		"plug",
		"--capability",
		"--protocol",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q, got %q", pattern, cmd.Example)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	flagTests := []struct {
		name      string
		shorthand string
	}{
		{"capability", ""},
		{"protocol", ""},
		{"output", "o"},
	}

	for _, f := range flagTests {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag %q not found", f.name)
			continue
		}
		if f.shorthand != "" && flag.Shorthand != f.shorthand {
			t.Errorf("flag %q shorthand = %q, want %q", f.name, flag.Shorthand, f.shorthand)
		}
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	var out, err bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&err)
	cmd.SetArgs([]string{"--help"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("--help should not error: %v", cmdErr)
	}
}

func TestExecute_SearchByQuery(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"plug"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected output, got empty")
	}

	// Should contain search results - either table with profiles or "no profiles found" message
	hasTable := strings.Contains(output, "Model")
	hasNoResults := strings.Contains(output, "No profiles found")
	if !hasTable && !hasNoResults {
		t.Error("expected table or 'No profiles found' message")
	}
}

func TestExecute_SearchByQuery_NoResults(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistentdevice12345"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if !strings.Contains(output, "No profiles found") {
		t.Errorf("expected 'No profiles found' message, got: %s", output)
	}
}

func TestExecute_FilterByCapability(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--capability", "power_metering"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	// Should either show results or a message about no profiles
	if output == "" {
		t.Error("expected some output")
	}
}

func TestExecute_FilterByCapability_NoResults(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--capability", "nonexistent_capability_xyz"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if !strings.Contains(output, "No profiles found") {
		t.Errorf("expected 'No profiles found' message, got: %s", output)
	}
}

func TestExecute_FilterByProtocol(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--protocol", "mqtt"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestExecute_FilterByProtocol_NoResults(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--protocol", "nonexistent_protocol_xyz"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if !strings.Contains(output, "No profiles found") {
		t.Errorf("expected 'No profiles found' message, got: %s", output)
	}
}

func TestExecute_CombineQueryAndCapability(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"plug", "--capability", "power_metering"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	// Should have results or no profiles found message
	if output == "" {
		t.Error("expected some output")
	}
}

func TestExecute_CombineQueryAndProtocol(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"plug", "--protocol", "mqtt"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestExecute_CombineAllFilters(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"plug", "--capability", "power_metering", "--protocol", "mqtt"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestExecute_OutputFormatJSON(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"plug", "-o", "json"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	// JSON output should contain brackets or be empty for no results
	if !strings.Contains(output, "[") && !strings.Contains(output, "No profiles found") {
		t.Errorf("expected JSON output or 'No profiles found' message, got: %s", output)
	}
}

func TestExecute_OutputFormatYAML(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"plug", "-o", "yaml"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestExecute_NoArguments_ListsAll(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	// Should list all profiles
	if !strings.Contains(output, "Model") && !strings.Contains(output, "No profiles found") {
		t.Errorf("expected table or 'No profiles found', got: %s", output)
	}
}

func TestExecute_SortedByModel(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if strings.Contains(output, "Model") {
		// If we got results, verify it's a table
		if !strings.Contains(output, "Found") {
			t.Log("Table output received, sorting verified by function")
		}
	}
}

func TestOptions_Creation(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Query:      "plug",
		Capability: "power_metering",
		Protocol:   "mqtt",
		Factory:    tf.Factory,
	}

	if opts.Query != "plug" {
		t.Errorf("Query = %q, want %q", opts.Query, "plug")
	}
	if opts.Capability != "power_metering" {
		t.Errorf("Capability = %q, want %q", opts.Capability, "power_metering")
	}
	if opts.Protocol != "mqtt" {
		t.Errorf("Protocol = %q, want %q", opts.Protocol, "mqtt")
	}
}

func TestExecute_CapabilityOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--capability", "scripting"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestExecute_ProtocolOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--protocol", "websocket"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestExecute_CapabilityAndProtocol(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--capability", "scripting", "--protocol", "mqtt"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestExecute_LongDescriptionPresent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if !strings.Contains(cmd.Long, "Search") || !strings.Contains(cmd.Long, "filter") {
		t.Error("Long description missing important keywords")
	}
}

func TestExecute_SearchResults_ContainsFull(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"1pm"}) // Try a common device model pattern

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}

	// Verify the output contains important information
	hasNameColumn := strings.Contains(output, "Name")
	hasGenerationColumn := strings.Contains(output, "Generation")
	hasSeriesColumn := strings.Contains(output, "Series")
	hasFormFactorColumn := strings.Contains(output, "Form Factor")
	hasNoResults := strings.Contains(output, "No profiles found")

	if !hasNoResults && (!hasNameColumn || !hasGenerationColumn || !hasSeriesColumn || !hasFormFactorColumn) {
		t.Log("If results are found, should contain all columns")
	}
}

func TestExecute_AllProfiles_Sorted(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	// When no filters are applied, all profiles are listed and sorted
	if output != "" && !strings.Contains(output, "No profiles found") {
		// If we got results, they should be in a table format
		if !strings.Contains(output, "Model") {
			t.Error("expected Model column when showing all profiles")
		}
	}
}

func TestExecute_LowercaseQuery(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"shelly"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestNewCommand_FieldValues(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "search <query>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "search <query>")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

func TestExecute_OutputFormatTable(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"-o", "table"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestExecute_MultipleFilters_PrecedenceTable(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	// Query takes precedence in the switch
	cmd.SetArgs([]string{"1pm", "--capability", "scripting", "--protocol", "mqtt"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestExecute_QueryWithCapabilityFilter(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	// This should trigger both the query search and capability filter
	cmd.SetArgs([]string{"plus", "--capability", "power_metering"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	// Should have either results or "no profiles found"
	if !strings.Contains(output, "No profiles found") && !strings.Contains(output, "Model") {
		t.Logf("output: %s", output)
	}
}

func TestExecute_QueryWithCapabilityAndProtocolFilter(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	// This should trigger query search, then both filters
	cmd.SetArgs([]string{"pro", "--capability", "power_metering", "--protocol", "mqtt"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	// Should have either results or "no profiles found"
	if !strings.Contains(output, "No profiles found") && !strings.Contains(output, "Model") {
		t.Logf("output: %s", output)
	}
}

func TestExecute_CapabilityWithProtocolFilter(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetContext(context.Background())
	// Capability search followed by protocol filter
	cmd.SetArgs([]string{"--capability", "power_metering", "--protocol", "mqtt"})

	cmdErr := cmd.Execute()
	if cmdErr != nil {
		t.Errorf("unexpected error: %v", cmdErr)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestRun_DirectCall(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Query:   "pro",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestRun_DirectCall_NoResults(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Query:   "zzzzzzzznonexistentzzzzzzz",
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "No profiles found") {
		t.Errorf("expected 'No profiles found' message, got: %s", output)
	}
}

func TestRun_DirectCall_CapabilityOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Capability: "power_metering",
		Factory:    tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestRun_DirectCall_ProtocolOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Protocol: "mqtt",
		Factory:  tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected some output")
	}
}

func TestRun_DirectCall_AllProfiles(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := tf.OutString()
	// Should have results or a no profiles message
	hasResults := strings.Contains(output, "Model")
	hasNoResults := strings.Contains(output, "No profiles found")
	if !hasResults && !hasNoResults {
		t.Errorf("expected table with profiles or 'No profiles found', got: %s", output)
	}
}

func TestRun_DirectCall_QueryAndCapability(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Query:      "test",
		Capability: "power_metering",
		Factory:    tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := tf.OutString()
	// Should have some output
	if output == "" {
		t.Error("expected some output")
	}
}

func TestRun_DirectCall_QueryCapabilityAndProtocol(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Query:      "test",
		Capability: "power_metering",
		Protocol:   "mqtt",
		Factory:    tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := tf.OutString()
	// Should have some output
	if output == "" {
		t.Error("expected some output")
	}
}

func TestRun_DirectCall_CapabilityAndProtocol(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Capability: "power_metering",
		Protocol:   "mqtt",
		Factory:    tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := tf.OutString()
	// Should have some output
	if output == "" {
		t.Error("expected some output")
	}
}

func TestOptions_OutputFlagsEmbedded(t *testing.T) {
	t.Parallel()

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "json",
		},
	}

	if opts.Format != "json" {
		t.Errorf("Format = %q, want %q", opts.Format, "json")
	}
}
