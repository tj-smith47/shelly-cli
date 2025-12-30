package deletecmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// testMu ensures tests that modify the global config manager don't run in parallel.
var testMu sync.Mutex

// setupTestManager creates a manager with temp dir and pre-populated templates.
func setupTestManager(t *testing.T, templates map[string]config.DeviceTemplate) *config.Manager {
	t.Helper()
	tmpDir := t.TempDir()
	mgr := config.NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := mgr.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	// Pre-populate templates
	for name, tpl := range templates {
		tpl.Name = name
		if err := mgr.SaveDeviceTemplate(tpl); err != nil {
			t.Fatalf("SaveDeviceTemplate() error: %v", err)
		}
	}
	return mgr
}

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

	// Test Use - ConfigDeleteCommand uses "delete <template>"
	if cmd.Use != "delete <template>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "delete <template>")
	}

	// Test Aliases - ConfigDeleteCommand adds standard aliases
	wantAliases := []string{"rm", "del", "remove"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg valid", []string{"template-name"}, false},
		{"two args", []string{"template1", "template2"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("yes flag not found")
	}
	if yesFlag.Shorthand != "y" {
		t.Errorf("yes shorthand = %q, want y", yesFlag.Shorthand)
	}
	if yesFlag.DefValue != "false" {
		t.Errorf("yes default = %q, want false", yesFlag.DefValue)
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for template completion")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly template delete",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"template",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

//nolint:paralleltest // Modifies global config manager
func TestExecute_TemplateNotFound(t *testing.T) {
	testMu.Lock()
	defer testMu.Unlock()

	mgr := setupTestManager(t, nil)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"nonexistent"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent template")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

//nolint:paralleltest // Modifies global config manager
func TestExecute_DeleteWithYesFlag(t *testing.T) {
	testMu.Lock()
	defer testMu.Unlock()

	templates := map[string]config.DeviceTemplate{
		"my-template": {
			Model:      "SHSW-1",
			Generation: 2,
			Config:     map[string]any{"key": "value"},
		},
	}
	mgr := setupTestManager(t, templates)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"my-template", "--yes"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// Verify template was deleted
	if _, exists := mgr.GetDeviceTemplate("my-template"); exists {
		t.Error("template should have been deleted")
	}

	// Check output contains success message
	output := out.String()
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected output to contain 'deleted', got: %s", output)
	}
}

//nolint:paralleltest // Modifies global config manager
func TestExecute_DeleteTemplateWithModel(t *testing.T) {
	testMu.Lock()
	defer testMu.Unlock()

	templates := map[string]config.DeviceTemplate{
		"test-tpl": {
			Model:      "SHSW-PM",
			Generation: 2,
			Config:     map[string]any{"wifi": "enabled"},
		},
	}
	mgr := setupTestManager(t, templates)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"test-tpl", "-y"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// Verify template was deleted
	if _, exists := mgr.GetDeviceTemplate("test-tpl"); exists {
		t.Error("template should have been deleted")
	}
}

//nolint:paralleltest // Modifies global config manager
func TestExecute_DeleteTemplateWithoutModel(t *testing.T) {
	testMu.Lock()
	defer testMu.Unlock()

	templates := map[string]config.DeviceTemplate{
		"empty-model": {
			Model:      "",
			Generation: 2,
			Config:     map[string]any{},
		},
	}
	mgr := setupTestManager(t, templates)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"empty-model", "--yes"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// Verify template was deleted
	if _, exists := mgr.GetDeviceTemplate("empty-model"); exists {
		t.Error("template should have been deleted")
	}
}

//nolint:paralleltest // Modifies global config manager
func TestExecute_DeleteCancelled(t *testing.T) {
	testMu.Lock()
	defer testMu.Unlock()

	templates := map[string]config.DeviceTemplate{
		"keep-me": {
			Model:      "SHSW-1",
			Generation: 2,
			Config:     map[string]any{},
		},
	}
	mgr := setupTestManager(t, templates)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	// Provide "n" for cancellation
	in := bytes.NewBufferString("n\n")
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"keep-me"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// Verify template was NOT deleted
	if _, exists := mgr.GetDeviceTemplate("keep-me"); !exists {
		t.Error("template should NOT have been deleted when cancelled")
	}

	// Check output contains cancellation message
	output := out.String()
	if !strings.Contains(output, "cancelled") {
		t.Errorf("expected output to contain 'cancelled', got: %s", output)
	}
}

//nolint:paralleltest // Modifies global config manager
func TestExecute_DeleteShortYesFlag(t *testing.T) {
	testMu.Lock()
	defer testMu.Unlock()

	templates := map[string]config.DeviceTemplate{
		"short-flag-tpl": {
			Model:      "SHSW-1",
			Generation: 2,
			Config:     map[string]any{},
		},
	}
	mgr := setupTestManager(t, templates)
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	// Use short flag -y instead of --yes
	cmd.SetArgs([]string{"short-flag-tpl", "-y"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// Verify template was deleted
	if _, exists := mgr.GetDeviceTemplate("short-flag-tpl"); exists {
		t.Error("template should have been deleted with -y flag")
	}

	// Check output contains success message
	output := out.String()
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected output to contain 'deleted', got: %s", output)
	}
}
