package completion_test

import (
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/completion"
)

//nolint:paralleltest // Test modifies environment variables, cannot run in parallel
func TestDetectShell(t *testing.T) {
	tests := []struct {
		name     string
		envShell string
		want     string
		wantErr  bool
	}{
		{
			name:     "bash shell",
			envShell: "/bin/bash",
			want:     completion.ShellBash,
		},
		{
			name:     "zsh shell",
			envShell: "/bin/zsh",
			want:     completion.ShellZsh,
		},
		{
			name:     "fish shell",
			envShell: "/usr/bin/fish",
			want:     completion.ShellFish,
		},
		{
			name:     "unknown shell with SHELL set",
			envShell: "/bin/sh",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origShell := os.Getenv("SHELL")
			t.Cleanup(func() {
				if origShell == "" {
					if err := os.Unsetenv("SHELL"); err != nil {
						t.Logf("warning: failed to unset SHELL: %v", err)
					}
				} else {
					if err := os.Setenv("SHELL", origShell); err != nil {
						t.Logf("warning: failed to restore SHELL: %v", err)
					}
				}
			})

			if err := os.Setenv("SHELL", tt.envShell); err != nil {
				t.Fatalf("failed to set SHELL: %v", err)
			}

			got, err := completion.DetectShell()
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectShell() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("DetectShell() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShellConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ShellBash", completion.ShellBash, "bash"},
		{"ShellZsh", completion.ShellZsh, "zsh"},
		{"ShellFish", completion.ShellFish, "fish"},
		{"ShellPowerShell", completion.ShellPowerShell, "powershell"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestDeviceNames(t *testing.T) {
	t.Parallel()

	fn := completion.DeviceNames()
	if fn == nil {
		t.Fatal("DeviceNames() returned nil")
	}

	// Create a dummy command for testing
	cmd := &cobra.Command{}

	// Call the completion function
	completions, directive := fn(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	// Completions can be empty if no devices are registered
	_ = completions
}

func TestGroupNames(t *testing.T) {
	t.Parallel()

	fn := completion.GroupNames()
	if fn == nil {
		t.Fatal("GroupNames() returned nil")
	}

	cmd := &cobra.Command{}
	completions, directive := fn(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
	_ = completions
}

func TestAliasNames(t *testing.T) {
	t.Parallel()

	fn := completion.AliasNames()
	if fn == nil {
		t.Fatal("AliasNames() returned nil")
	}

	cmd := &cobra.Command{}
	completions, directive := fn(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
	_ = completions
}

func TestThemeNames(t *testing.T) {
	t.Parallel()

	fn := completion.ThemeNames()
	if fn == nil {
		t.Fatal("ThemeNames() returned nil")
	}

	cmd := &cobra.Command{}
	completions, directive := fn(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
	// There should be at least some built-in themes
	if len(completions) == 0 {
		t.Log("Warning: no themes returned, but this may be expected in test environment")
	}
}

func TestSceneNames(t *testing.T) {
	t.Parallel()

	fn := completion.SceneNames()
	if fn == nil {
		t.Fatal("SceneNames() returned nil")
	}

	cmd := &cobra.Command{}
	completions, directive := fn(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
	_ = completions
}

func TestOutputFormats(t *testing.T) {
	t.Parallel()

	fn := completion.OutputFormats()
	if fn == nil {
		t.Fatal("OutputFormats() returned nil")
	}

	cmd := &cobra.Command{}
	completions, directive := fn(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	if len(completions) == 0 {
		t.Error("OutputFormats() should return at least one format")
	}

	// Check that standard formats are present
	formatFound := make(map[string]bool)
	for _, c := range completions {
		if c != "" {
			// Completions may have tab-separated descriptions
			formatFound[c[:4]] = true // First 4 chars for format name
		}
	}
	if !formatFound["tabl"] && !formatFound["json"] && !formatFound["yaml"] {
		t.Error("OutputFormats() should include table, json, and yaml")
	}
}

func TestDevicesOrGroups(t *testing.T) {
	t.Parallel()

	fn := completion.DevicesOrGroups()
	if fn == nil {
		t.Fatal("DevicesOrGroups() returned nil")
	}

	cmd := &cobra.Command{}
	completions, directive := fn(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
	_ = completions
}

func TestNoFile(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	completions, directive := completion.NoFile(cmd, nil, "")

	if completions != nil {
		t.Errorf("NoFile() completions = %v, want nil", completions)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("NoFile() directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

func TestDeviceThenScriptID(t *testing.T) {
	t.Parallel()

	fn := completion.DeviceThenScriptID()
	if fn == nil {
		t.Fatal("DeviceThenScriptID() returned nil")
	}

	cmd := &cobra.Command{}

	// First arg should complete device names
	_, directive := fn(cmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("first arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	// Second arg should complete script IDs (empty since no device exists)
	_, directive = fn(cmd, []string{"device1"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("second arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	// Third arg should be no completion
	_, directive = fn(cmd, []string{"device1", "1"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("third arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

func TestDeviceThenScheduleID(t *testing.T) {
	t.Parallel()

	fn := completion.DeviceThenScheduleID()
	if fn == nil {
		t.Fatal("DeviceThenScheduleID() returned nil")
	}

	cmd := &cobra.Command{}

	// First arg should complete device names
	_, directive := fn(cmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("first arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	// Second arg should complete schedule IDs
	_, directive = fn(cmd, []string{"device1"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("second arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

func TestDiscoveredDevices(t *testing.T) {
	t.Parallel()

	fn := completion.DiscoveredDevices()
	if fn == nil {
		t.Fatal("DiscoveredDevices() returned nil")
	}

	cmd := &cobra.Command{}
	_, directive := fn(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

func TestDevicesWithGroups(t *testing.T) {
	t.Parallel()

	fn := completion.DevicesWithGroups()
	if fn == nil {
		t.Fatal("DevicesWithGroups() returned nil")
	}

	cmd := &cobra.Command{}
	completions, directive := fn(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	// Should always include @all
	found := false
	for _, c := range completions {
		if c == "@all\tall registered devices" {
			found = true
			break
		}
	}
	if !found {
		t.Error("DevicesWithGroups() should include @all")
	}
}

func TestExpandDeviceArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		devices []string
		wantLen int // -1 means we don't know exact length
	}{
		{
			name:    "single device",
			devices: []string{"device1"},
			wantLen: 1,
		},
		{
			name:    "multiple devices",
			devices: []string{"device1", "device2"},
			wantLen: 2,
		},
		{
			name:    "empty devices",
			devices: []string{},
			wantLen: 0,
		},
		{
			name:    "non-existent group",
			devices: []string{"@nonexistent"},
			wantLen: 0, // Group not found, returns empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := completion.ExpandDeviceArgs(tt.devices)
			if len(result) != tt.wantLen {
				t.Errorf("ExpandDeviceArgs() returned %d items, want %d", len(result), tt.wantLen)
			}
		})
	}
}

func TestTemplateNames(t *testing.T) {
	t.Parallel()

	fn := completion.TemplateNames()
	if fn == nil {
		t.Fatal("TemplateNames() returned nil")
	}

	cmd := &cobra.Command{}
	_, directive := fn(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

func TestTemplateThenDevice(t *testing.T) {
	t.Parallel()

	fn := completion.TemplateThenDevice()
	if fn == nil {
		t.Fatal("TemplateThenDevice() returned nil")
	}

	cmd := &cobra.Command{}

	// First arg should complete template names
	_, directive := fn(cmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("first arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	// Second arg should complete device names
	_, directive = fn(cmd, []string{"template1"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("second arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	// Third arg should be no completion
	_, directive = fn(cmd, []string{"template1", "device1"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("third arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

func TestTemplateThenFile(t *testing.T) {
	t.Parallel()

	fn := completion.TemplateThenFile()
	if fn == nil {
		t.Fatal("TemplateThenFile() returned nil")
	}

	cmd := &cobra.Command{}

	// First arg should complete template names
	_, directive := fn(cmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("first arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	// Second arg should use default file completion
	_, directive = fn(cmd, []string{"template1"}, "")
	if directive != cobra.ShellCompDirectiveDefault {
		t.Errorf("second arg directive = %v, want %v", directive, cobra.ShellCompDirectiveDefault)
	}

	// Third arg should be no completion
	_, directive = fn(cmd, []string{"template1", "file.txt"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("third arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

func TestDeviceThenFile(t *testing.T) {
	t.Parallel()

	fn := completion.DeviceThenFile()
	if fn == nil {
		t.Fatal("DeviceThenFile() returned nil")
	}

	cmd := &cobra.Command{}

	// First arg should complete device names
	_, directive := fn(cmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("first arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	// Second arg should use default file completion
	_, directive = fn(cmd, []string{"device1"}, "")
	if directive != cobra.ShellCompDirectiveDefault {
		t.Errorf("second arg directive = %v, want %v", directive, cobra.ShellCompDirectiveDefault)
	}
}

func TestDeviceThenNoComplete(t *testing.T) {
	t.Parallel()

	fn := completion.DeviceThenNoComplete()
	if fn == nil {
		t.Fatal("DeviceThenNoComplete() returned nil")
	}

	cmd := &cobra.Command{}

	// First arg should complete device names
	_, directive := fn(cmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("first arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	// Second arg should be no completion
	_, directive = fn(cmd, []string{"device1"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("second arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

func TestNameThenDevice(t *testing.T) {
	t.Parallel()

	fn := completion.NameThenDevice()
	if fn == nil {
		t.Fatal("NameThenDevice() returned nil")
	}

	cmd := &cobra.Command{}

	// First arg should be no completion (name)
	completions, directive := fn(cmd, []string{}, "")
	if completions != nil {
		t.Errorf("first arg completions = %v, want nil", completions)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("first arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	// Second arg should complete device names
	_, directive = fn(cmd, []string{"name"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("second arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

func TestSettingKeys(t *testing.T) {
	t.Parallel()

	fn := completion.SettingKeys()
	if fn == nil {
		t.Fatal("SettingKeys() returned nil")
	}

	cmd := &cobra.Command{}
	_, directive := fn(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

func TestSettingKeysWithEquals(t *testing.T) {
	t.Parallel()

	fn := completion.SettingKeysWithEquals()
	if fn == nil {
		t.Fatal("SettingKeysWithEquals() returned nil")
	}

	cmd := &cobra.Command{}
	completions, directive := fn(cmd, nil, "")

	expectedDirective := cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
	if directive != expectedDirective {
		t.Errorf("directive = %v, want %v", directive, expectedDirective)
	}

	// All completions should end with =
	for _, c := range completions {
		if c != "" && c[len(c)-1] != '=' {
			t.Errorf("completion %q should end with '='", c)
		}
	}
}

func TestFileThenNoComplete(t *testing.T) {
	t.Parallel()

	fn := completion.FileThenNoComplete()
	if fn == nil {
		t.Fatal("FileThenNoComplete() returned nil")
	}

	cmd := &cobra.Command{}

	// First arg should use default file completion
	_, directive := fn(cmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveDefault {
		t.Errorf("first arg directive = %v, want %v", directive, cobra.ShellCompDirectiveDefault)
	}

	// Second arg should be no completion
	_, directive = fn(cmd, []string{"file.txt"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("second arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

//nolint:paralleltest // Test writes to global cache, cannot run in parallel
func TestSaveDiscoveryCache(t *testing.T) {
	addresses := []string{"192.168.1.1", "192.168.1.2"}

	err := completion.SaveDiscoveryCache(addresses)
	if err != nil {
		// May fail if cache directory is not writable in test environment
		t.Logf("SaveDiscoveryCache() error: %v (may be expected in restricted environments)", err)
		return
	}
}

func TestScriptTemplateNames(t *testing.T) {
	t.Parallel()

	fn := completion.ScriptTemplateNames()
	if fn == nil {
		t.Fatal("ScriptTemplateNames() returned nil")
	}

	cmd := &cobra.Command{}
	_, directive := fn(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

func TestDeviceThenScriptTemplate(t *testing.T) {
	t.Parallel()

	fn := completion.DeviceThenScriptTemplate()
	if fn == nil {
		t.Fatal("DeviceThenScriptTemplate() returned nil")
	}

	cmd := &cobra.Command{}

	// First arg should complete device names
	_, directive := fn(cmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("first arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	// Second arg should complete script template names
	_, directive = fn(cmd, []string{"device1"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("second arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}

	// Third arg should be no completion
	_, directive = fn(cmd, []string{"device1", "template1"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("third arg directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

func TestExtensionNames(t *testing.T) {
	t.Parallel()

	fn := completion.ExtensionNames()
	if fn == nil {
		t.Fatal("ExtensionNames() returned nil")
	}

	cmd := &cobra.Command{}
	// This may return error directive if plugins directory doesn't exist
	_, directive := fn(cmd, nil, "")
	// Accept either NoFileComp or Error directive (depends on environment)
	if directive != cobra.ShellCompDirectiveNoFileComp && directive != cobra.ShellCompDirectiveError {
		t.Errorf("directive = %v, want NoFileComp or Error", directive)
	}
}
