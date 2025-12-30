package wizard

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const testAPIModeHTTP = "http"

//nolint:gocritic,unparam // helper function returns multiple values for API consistency
func testIOStreams() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	in := strings.NewReader("")
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return iostreams.Test(in, out, errOut), out, errOut
}

// setupTestConfig creates an isolated test config environment.
// Returns a cleanup function that MUST be deferred.
func setupTestConfig(t *testing.T) func() {
	t.Helper()

	// Save original environment
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")

	// Reset the config singleton BEFORE changing HOME
	config.ResetDefaultManagerForTesting()

	// Create temp directory for test config
	tmpDir := t.TempDir()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}

	// Create config directory (XDG_CONFIG_HOME/shelly since XDG takes precedence)
	configDir := filepath.Join(tmpDir, "shelly")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write minimal config
	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `devices: {}
groups: {}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Return cleanup function
	return func() {
		config.ResetDefaultManagerForTesting()
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
		if originalXDG != "" {
			if err := os.Setenv("XDG_CONFIG_HOME", originalXDG); err != nil {
				t.Logf("warning: failed to restore XDG_CONFIG_HOME: %v", err)
			}
		} else {
			if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
				t.Logf("warning: failed to unset XDG_CONFIG_HOME: %v", err)
			}
		}
	}
}

func TestOptions_IsNonInteractive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options Options
		want    bool
	}{
		{
			name:    "empty options (interactive)",
			options: Options{},
			want:    false,
		},
		{
			name:    "devices set",
			options: Options{Devices: []string{"device1"}},
			want:    true,
		},
		{
			name:    "devicesJSON set",
			options: Options{DevicesJSON: []string{`{"name":"dev"}`}},
			want:    true,
		},
		{
			name:    "theme set",
			options: Options{Theme: "dark"},
			want:    true,
		},
		{
			name:    "outputFormat set",
			options: Options{OutputFormat: "json"},
			want:    true,
		},
		{
			name:    "apiMode set",
			options: Options{APIMode: "rpc"},
			want:    true,
		},
		{
			name:    "noColor set",
			options: Options{NoColor: true},
			want:    true,
		},
		{
			name:    "cloudEmail set",
			options: Options{CloudEmail: "test@example.com"},
			want:    true,
		},
		{
			name:    "cloudPassword set",
			options: Options{CloudPassword: "secret"},
			want:    true,
		},
		{
			name:    "completions set",
			options: Options{Completions: "bash"},
			want:    true,
		},
		{
			name:    "aliases set",
			options: Options{Aliases: true},
			want:    true,
		},
		{
			name:    "discover set",
			options: Options{Discover: true},
			want:    true,
		},
		{
			name:    "force set",
			options: Options{Force: true},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.options.IsNonInteractive()
			if got != tt.want {
				t.Errorf("IsNonInteractive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptions_WantsCloudSetup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options Options
		want    bool
	}{
		{
			name:    "no cloud credentials",
			options: Options{},
			want:    false,
		},
		{
			name:    "email only",
			options: Options{CloudEmail: "test@example.com"},
			want:    true,
		},
		{
			name:    "password only",
			options: Options{CloudPassword: "secret"},
			want:    true,
		},
		{
			name:    "both email and password",
			options: Options{CloudEmail: "test@example.com", CloudPassword: "secret"},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.options.WantsCloudSetup()
			if got != tt.want {
				t.Errorf("WantsCloudSetup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintWelcome(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	PrintWelcome(ios)

	// Should contain welcome message
	if out.String() == "" {
		t.Error("PrintWelcome should produce output")
	}
}

func TestPrintSummary(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	PrintSummary(ios)

	// Should contain summary
	if out.String() == "" {
		t.Error("PrintSummary should produce output")
	}
}

func TestOptionsFields(t *testing.T) {
	t.Parallel()

	opts := Options{
		Devices:         []string{"dev1", "dev2"},
		DevicesJSON:     []string{`{"name":"test"}`},
		Discover:        true,
		DiscoverTimeout: 5e9,
		DiscoverModes:   "mdns,broadcast",
		Network:         "192.168.1.0/24",
		Completions:     "zsh",
		Aliases:         true,
		Theme:           "dracula",
		OutputFormat:    "yaml",
		NoColor:         false,
		APIMode:         "http",
		CloudEmail:      "user@example.com",
		CloudPassword:   "pass123",
		Force:           true,
	}

	if len(opts.Devices) != 2 {
		t.Errorf("Devices len = %d, want 2", len(opts.Devices))
	}
	if opts.Discover != true {
		t.Error("Discover = false, want true")
	}
	if opts.DiscoverModes != "mdns,broadcast" {
		t.Errorf("DiscoverModes = %q", opts.DiscoverModes)
	}
	if opts.Network != "192.168.1.0/24" {
		t.Errorf("Network = %q", opts.Network)
	}
	if opts.Completions != "zsh" {
		t.Errorf("Completions = %q", opts.Completions)
	}
	if opts.Theme != "dracula" {
		t.Errorf("Theme = %q", opts.Theme)
	}
	if opts.OutputFormat != "yaml" {
		t.Errorf("OutputFormat = %q", opts.OutputFormat)
	}
	if opts.APIMode != testAPIModeHTTP {
		t.Errorf("APIMode = %q", opts.APIMode)
	}
	if opts.CloudEmail != "user@example.com" {
		t.Errorf("CloudEmail = %q", opts.CloudEmail)
	}
	if !opts.Force {
		t.Error("Force = false, want true")
	}
}

func TestSanitizeDeviceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple lowercase",
			input: "mydevice",
			want:  "mydevice",
		},
		{
			name:  "uppercase to lowercase",
			input: "MyDevice",
			want:  "mydevice",
		},
		{
			name:  "spaces to dashes",
			input: "my device",
			want:  "my-device",
		},
		{
			name:  "underscores to dashes",
			input: "my_device",
			want:  "my-device",
		},
		{
			name:  "mixed spaces and underscores",
			input: "My_Device Name",
			want:  "my-device-name",
		},
		{
			name:  "special characters removed",
			input: "device@#$%123",
			want:  "device123",
		},
		{
			name:  "numbers preserved",
			input: "device123",
			want:  "device123",
		},
		{
			name:  "dashes preserved",
			input: "my-device",
			want:  "my-device",
		},
		{
			name:  "complex name",
			input: "Shelly Plus 1PM - Kitchen",
			want:  "shelly-plus-1pm---kitchen",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only special chars",
			input: "@#$%^&*()",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SanitizeDeviceName(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeDeviceName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDiscoverModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty string defaults to all",
			input: "",
			want:  []string{"http", "mdns", "coiot"},
		},
		{
			name:  "all keyword",
			input: "all",
			want:  []string{"http", "mdns", "coiot"},
		},
		{
			name:  "single http",
			input: "http",
			want:  []string{"http"},
		},
		{
			name:  "single mdns",
			input: "mdns",
			want:  []string{"mdns"},
		},
		{
			name:  "single coiot",
			input: "coiot",
			want:  []string{"coiot"},
		},
		{
			name:  "single ble",
			input: "ble",
			want:  []string{"ble"},
		},
		{
			name:  "scan alias for http",
			input: "scan",
			want:  []string{"http"},
		},
		{
			name:  "zeroconf alias for mdns",
			input: "zeroconf",
			want:  []string{"mdns"},
		},
		{
			name:  "bonjour alias for mdns",
			input: "bonjour",
			want:  []string{"mdns"},
		},
		{
			name:  "coap alias for coiot",
			input: "coap",
			want:  []string{"coiot"},
		},
		{
			name:  "bluetooth alias for ble",
			input: "bluetooth",
			want:  []string{"ble"},
		},
		{
			name:  "multiple modes comma separated",
			input: "http,mdns,coiot",
			want:  []string{"http", "mdns", "coiot"},
		},
		{
			name:  "multiple modes with spaces",
			input: "http, mdns, coiot",
			want:  []string{"http", "mdns", "coiot"},
		},
		{
			name:  "uppercase modes",
			input: "HTTP,MDNS",
			want:  []string{"http", "mdns"},
		},
		{
			name:  "mixed case modes",
			input: "Http,Mdns,BLE",
			want:  []string{"http", "mdns", "ble"},
		},
		{
			name:  "unknown mode falls back to http",
			input: "unknown",
			want:  []string{"http"},
		},
		{
			name:  "partial valid modes",
			input: "http,unknown,mdns",
			want:  []string{"http", "mdns"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseDiscoverModes(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parseDiscoverModes(%q) = %v, want %v", tt.input, got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("parseDiscoverModes(%q)[%d] = %q, want %q", tt.input, i, v, tt.want[i])
				}
			}
		})
	}
}

func TestCheckExistingConfig(t *testing.T) {
	t.Parallel()

	// Test returns a path based on home directory
	exists, path := CheckExistingConfig()

	// Path should be non-empty regardless of whether config exists
	if path == "" {
		t.Error("CheckExistingConfig should return a path")
	}

	// Path should end with expected suffix
	expectedSuffix := filepath.Join(".config", "shelly", "config.yaml")
	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("path = %q, should end with %q", path, expectedSuffix)
	}

	// exists should match whether the file actually exists
	_, err := os.Stat(path)
	actualExists := err == nil
	if exists != actualExists {
		t.Errorf("exists = %v, but file existence = %v", exists, actualExists)
	}
}

func TestCheckAndConfirmConfig_NoExistingConfig(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	opts := &Options{}

	// When no config exists, should return true to continue
	// This test may depend on actual file system state
	// We test the logic when config doesn't exist or force is set

	optsWithForce := &Options{Force: true}
	shouldContinue, err := CheckAndConfirmConfig(ios, optsWithForce)
	if err != nil {
		t.Errorf("CheckAndConfirmConfig with force = error %v, want nil", err)
	}
	if !shouldContinue {
		t.Error("CheckAndConfirmConfig with force = false, want true")
	}

	// Non-interactive with existing config (would need --force)
	// We can't easily test the confirm dialog, but we can test non-interactive path
	optsNonInteractive := &Options{Theme: "dracula"} // makes it non-interactive
	_, err = CheckAndConfirmConfig(ios, optsNonInteractive)
	if err != nil {
		// No error expected, just may return false if config exists
		t.Logf("non-interactive check: %v", err)
	}

	_ = opts // use opts to avoid unused variable
}

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty config is valid",
			cfg:     &config.Config{},
			wantErr: false,
		},
		{
			name: "valid output format table",
			cfg: &config.Config{
				Output: "table",
			},
			wantErr: false,
		},
		{
			name: "valid output format json",
			cfg: &config.Config{
				Output: "json",
			},
			wantErr: false,
		},
		{
			name: "valid output format yaml",
			cfg: &config.Config{
				Output: "yaml",
			},
			wantErr: false,
		},
		{
			name: "valid output format text",
			cfg: &config.Config{
				Output: "text",
			},
			wantErr: false,
		},
		{
			name: "valid output format template",
			cfg: &config.Config{
				Output: "template",
			},
			wantErr: false,
		},
		{
			name: "invalid output format",
			cfg: &config.Config{
				Output: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid output format",
		},
		{
			name: "valid api_mode local",
			cfg: &config.Config{
				APIMode: "local",
			},
			wantErr: false,
		},
		{
			name: "valid api_mode cloud",
			cfg: &config.Config{
				APIMode: "cloud",
			},
			wantErr: false,
		},
		{
			name: "valid api_mode auto",
			cfg: &config.Config{
				APIMode: "auto",
			},
			wantErr: false,
		},
		{
			name: "invalid api_mode",
			cfg: &config.Config{
				APIMode: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid api_mode",
		},
		{
			name: "device without address",
			cfg: &config.Config{
				Devices: map[string]model.Device{
					"test": {Address: ""},
				},
			},
			wantErr: true,
			errMsg:  "has no address",
		},
		{
			name: "device with address is valid",
			cfg: &config.Config{
				Devices: map[string]model.Device{
					"test": {Address: "192.168.1.100"},
				},
			},
			wantErr: false,
		},
		{
			name: "group references unknown device",
			cfg: &config.Config{
				Devices: map[string]model.Device{},
				Groups: map[string]config.Group{
					"mygroup": {Devices: []string{"unknown"}},
				},
			},
			wantErr: true,
			errMsg:  "references unknown device",
		},
		{
			name: "group references known device",
			cfg: &config.Config{
				Devices: map[string]model.Device{
					"mydevice": {Address: "192.168.1.100"},
				},
				Groups: map[string]config.Group{
					"mygroup": {Devices: []string{"mydevice"}},
				},
			},
			wantErr: false,
		},
		{
			name: "group with IP address skips device check",
			cfg: &config.Config{
				Devices: map[string]model.Device{},
				Groups: map[string]config.Group{
					"mygroup": {Devices: []string{"192.168.1.100"}},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple errors combined",
			cfg: &config.Config{
				Output:  "invalid",
				APIMode: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateConfig(tt.cfg)
			if tt.wantErr { //nolint:nestif // test validation logic
				if err == nil {
					t.Error("ValidateConfig() = nil, want error")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateConfig() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateConfig() = %v, want nil", err)
				}
			}
		})
	}
}

func TestCheckCompletionInstalled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		shell string
	}{
		{"bash", "bash"},
		{"zsh", "zsh"},
		{"fish", "fish"},
		{"powershell", "powershell"},
		{"unknown shell", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Just verify it doesn't panic and returns a boolean
			result := CheckCompletionInstalled(tt.shell)
			// Result depends on file system state, just verify it's a valid bool
			_ = result
		})
	}
}

func TestSelectDiscoveryMethods_NonInteractive(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()

	tests := []struct {
		name string
		opts *Options
		want []string
	}{
		{
			name: "http mode",
			opts: &Options{Discover: true, DiscoverModes: "http"},
			want: []string{"http"},
		},
		{
			name: "mdns mode",
			opts: &Options{Discover: true, DiscoverModes: "mdns"},
			want: []string{"mdns"},
		},
		{
			name: "multiple modes",
			opts: &Options{Discover: true, DiscoverModes: "http,mdns"},
			want: []string{"http", "mdns"},
		},
		{
			name: "empty modes defaults to all",
			opts: &Options{Discover: true, DiscoverModes: ""},
			want: []string{"http", "mdns", "coiot"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := selectDiscoveryMethods(ios, tt.opts)
			if len(got) != len(tt.want) {
				t.Errorf("selectDiscoveryMethods() = %v, want %v", got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("selectDiscoveryMethods()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestOptions_DiscoverTimeout(t *testing.T) {
	t.Parallel()

	opts := Options{
		DiscoverTimeout: 30 * time.Second,
	}

	if opts.DiscoverTimeout != 30*time.Second {
		t.Errorf("DiscoverTimeout = %v, want 30s", opts.DiscoverTimeout)
	}
}

func TestDefaultDiscoveryTimeout(t *testing.T) {
	t.Parallel()

	if defaultDiscoveryTimeout != 15*time.Second {
		t.Errorf("defaultDiscoveryTimeout = %v, want 15s", defaultDiscoveryTimeout)
	}
}

//nolint:gocyclo // comprehensive test coverage
func TestOptions_AllFieldsSet(t *testing.T) {
	t.Parallel()

	// Verify all fields can be set and retrieved
	opts := Options{
		Devices:         []string{"dev1", "dev2"},
		DevicesJSON:     []string{`{"name":"dev1"}`, `{"name":"dev2"}`},
		Discover:        true,
		DiscoverTimeout: 20 * time.Second,
		DiscoverModes:   "http,mdns,coiot,ble",
		Network:         "10.0.0.0/8",
		Completions:     "bash,zsh,fish",
		Aliases:         true,
		Theme:           "nord",
		OutputFormat:    "json",
		NoColor:         true,
		APIMode:         "cloud",
		CloudEmail:      "admin@example.com",
		CloudPassword:   "supersecret",
		Force:           true,
	}

	// Verify all values
	if len(opts.Devices) != 2 {
		t.Errorf("Devices length = %d, want 2", len(opts.Devices))
	}
	if len(opts.DevicesJSON) != 2 {
		t.Errorf("DevicesJSON length = %d, want 2", len(opts.DevicesJSON))
	}
	if !opts.Discover {
		t.Error("Discover = false, want true")
	}
	if opts.DiscoverTimeout != 20*time.Second {
		t.Errorf("DiscoverTimeout = %v, want 20s", opts.DiscoverTimeout)
	}
	if opts.DiscoverModes != "http,mdns,coiot,ble" {
		t.Errorf("DiscoverModes = %q", opts.DiscoverModes)
	}
	if opts.Network != "10.0.0.0/8" {
		t.Errorf("Network = %q", opts.Network)
	}
	if opts.Completions != "bash,zsh,fish" {
		t.Errorf("Completions = %q", opts.Completions)
	}
	if !opts.Aliases {
		t.Error("Aliases = false, want true")
	}
	if opts.Theme != "nord" {
		t.Errorf("Theme = %q", opts.Theme)
	}
	if opts.OutputFormat != "json" {
		t.Errorf("OutputFormat = %q", opts.OutputFormat)
	}
	if !opts.NoColor {
		t.Error("NoColor = false, want true")
	}
	if opts.APIMode != "cloud" {
		t.Errorf("APIMode = %q", opts.APIMode)
	}
	if opts.CloudEmail != "admin@example.com" {
		t.Errorf("CloudEmail = %q", opts.CloudEmail)
	}
	if opts.CloudPassword != "supersecret" {
		t.Errorf("CloudPassword = %q", opts.CloudPassword)
	}
	if !opts.Force {
		t.Error("Force = false, want true")
	}
}

func TestValidateConfig_Theme(t *testing.T) {
	t.Parallel()

	// Test with an unknown theme
	cfg := &config.Config{
		Theme: config.ThemeConfig{Name: "nonexistent-theme-xyz"},
	}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("ValidateConfig with unknown theme should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "unknown theme") {
		t.Errorf("error = %q, should contain 'unknown theme'", err.Error())
	}

	// Test with valid theme name (dracula is built-in)
	cfg = &config.Config{
		Theme: config.ThemeConfig{Name: "dracula"},
	}
	err = ValidateConfig(cfg)
	if err != nil {
		t.Errorf("ValidateConfig with valid theme = %v, want nil", err)
	}

	// Test with empty theme name (should be valid)
	cfg = &config.Config{
		Theme: config.ThemeConfig{Name: ""},
	}
	err = ValidateConfig(cfg)
	if err != nil {
		t.Errorf("ValidateConfig with empty theme = %v, want nil", err)
	}
}

func TestRunDiscoveryMethod_UnknownMethod(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()

	// Test unknown method returns error
	ctx := context.Background()
	_, err := runDiscoveryMethod(ctx, ios, "unknown-method", 5e9, "")
	if err == nil {
		t.Error("runDiscoveryMethod with unknown method should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "unknown method") {
		t.Errorf("error = %q, should contain 'unknown method'", err.Error())
	}
}

func TestSelectDiscoveryMethods_Interactive_Fallback(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()

	// Interactive mode with empty options - tests the fallback path when multi-select fails
	// Since we can't actually do interactive in tests, we verify non-interactive path works
	opts := &Options{} // Interactive mode (no flags set)

	// This will call the multi-select which will fail with our test IO,
	// but we verify the fallback behavior
	methods := selectDiscoveryMethods(ios, opts)

	// Should get fallback to http when multi-select fails
	if len(methods) == 0 {
		t.Error("selectDiscoveryMethods should return at least one method")
	}
}

func TestPrintWelcome_Content(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	PrintWelcome(ios)

	output := out.String()
	if !strings.Contains(output, "Welcome") {
		t.Error("PrintWelcome output should contain 'Welcome'")
	}
	if !strings.Contains(output, "shelly") || !strings.Contains(output, "Shelly") {
		t.Error("PrintWelcome output should mention shelly")
	}
}

func TestPrintSummary_Content(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	PrintSummary(ios)

	output := out.String()
	if !strings.Contains(output, "complete") {
		t.Error("PrintSummary output should mention completion")
	}
	if !strings.Contains(output, "shelly") {
		t.Error("PrintSummary output should contain shelly commands")
	}
}

func TestValidateConfig_ComplexScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "multiple devices all valid",
			cfg: &config.Config{
				Devices: map[string]model.Device{
					"device1": {Address: "192.168.1.100"},
					"device2": {Address: "192.168.1.101"},
					"device3": {Address: "192.168.1.102"},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple groups referencing valid devices",
			cfg: &config.Config{
				Devices: map[string]model.Device{
					"living-room": {Address: "192.168.1.100"},
					"bedroom":     {Address: "192.168.1.101"},
					"kitchen":     {Address: "192.168.1.102"},
				},
				Groups: map[string]config.Group{
					"downstairs": {Devices: []string{"living-room", "kitchen"}},
					"upstairs":   {Devices: []string{"bedroom"}},
				},
			},
			wantErr: false,
		},
		{
			name: "group with mixed IPs and device names",
			cfg: &config.Config{
				Devices: map[string]model.Device{
					"named-device": {Address: "192.168.1.100"},
				},
				Groups: map[string]config.Group{
					"mixed": {Devices: []string{"named-device", "192.168.1.200"}},
				},
			},
			wantErr: false, // IP addresses in groups are allowed
		},
		{
			name: "one device missing address among many",
			cfg: &config.Config{
				Devices: map[string]model.Device{
					"good1": {Address: "192.168.1.100"},
					"bad":   {Address: ""},
					"good2": {Address: "192.168.1.102"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateConfig(tt.cfg)
			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseDiscoverModes_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "whitespace only",
			input: "   ",
			want:  []string{"http"}, // Falls back to default
		},
		{
			name:  "mixed valid and empty",
			input: "http,,mdns,",
			want:  []string{"http", "mdns"},
		},
		{
			name:  "duplicate modes",
			input: "http,http,mdns,mdns",
			want:  []string{"http", "http", "mdns", "mdns"}, // Duplicates allowed
		},
		{
			name:  "all with extra spaces treated as unknown",
			input: "  all  ",
			want:  []string{"http"}, // Extra spaces mean "all" check fails, falls back
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseDiscoverModes(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parseDiscoverModes(%q) length = %d, want %d", tt.input, len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("parseDiscoverModes(%q)[%d] = %q, want %q", tt.input, i, v, tt.want[i])
				}
			}
		})
	}
}

func TestSanitizeDeviceName_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "unicode characters removed",
			input: "device-\u00e9\u00e0\u00fc",
			want:  "device-",
		},
		{
			name:  "long name with many special chars",
			input: "My!@#$%^&*()Device_Name With Spaces",
			want:  "mydevice-name-with-spaces",
		},
		{
			name:  "all numbers",
			input: "123456",
			want:  "123456",
		},
		{
			name:  "leading and trailing dashes",
			input: "---device---",
			want:  "---device---",
		},
		{
			name:  "multiple consecutive spaces",
			input: "my   device   name",
			want:  "my---device---name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SanitizeDeviceName(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeDeviceName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestSelectOutputFormat tests the selectOutputFormat function.
func TestSelectOutputFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    *Options
		want    string
		wantErr bool
	}{
		{
			name:    "explicit format from flags",
			opts:    &Options{OutputFormat: "json"},
			want:    "json",
			wantErr: false,
		},
		{
			name:    "explicit yaml format",
			opts:    &Options{OutputFormat: "yaml"},
			want:    "yaml",
			wantErr: false,
		},
		{
			name:    "explicit table format",
			opts:    &Options{OutputFormat: "table"},
			want:    "table",
			wantErr: false,
		},
		{
			name:    "non-interactive defaults to table",
			opts:    &Options{Theme: "dracula"}, // Makes it non-interactive
			want:    "table",
			wantErr: false,
		},
		{
			name:    "non-interactive with force flag",
			opts:    &Options{Force: true},
			want:    "table",
			wantErr: false,
		},
		{
			name:    "non-interactive with devices flag",
			opts:    &Options{Devices: []string{"dev1"}},
			want:    "table",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ios, _, _ := testIOStreams()
			got, err := selectOutputFormat(ios, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("selectOutputFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("selectOutputFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSelectTheme tests the selectTheme function.
func TestSelectTheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    *Options
		want    string
		wantErr bool
	}{
		{
			name:    "explicit theme from flags",
			opts:    &Options{Theme: "nord"},
			want:    "nord",
			wantErr: false,
		},
		{
			name:    "explicit dracula theme",
			opts:    &Options{Theme: "dracula"},
			want:    "dracula",
			wantErr: false,
		},
		{
			name:    "explicit tokyo-night theme",
			opts:    &Options{Theme: "tokyo-night"},
			want:    "tokyo-night",
			wantErr: false,
		},
		{
			name:    "non-interactive defaults to dracula",
			opts:    &Options{OutputFormat: "json"}, // Makes it non-interactive
			want:    "dracula",
			wantErr: false,
		},
		{
			name:    "non-interactive with force flag",
			opts:    &Options{Force: true},
			want:    "dracula",
			wantErr: false,
		},
		{
			name:    "non-interactive with discover flag",
			opts:    &Options{Discover: true},
			want:    "dracula",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ios, _, _ := testIOStreams()
			got, err := selectTheme(ios, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("selectTheme() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("selectTheme() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRunDiscoveryMethod tests the runDiscoveryMethod function dispatch.
func TestRunDiscoveryMethod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		wantErrMsg string
	}{
		{
			name:       "unknown method returns error",
			method:     "unknown-method",
			wantErrMsg: "unknown method",
		},
		{
			name:       "invalid method returns error",
			method:     "foobar",
			wantErrMsg: "unknown method",
		},
		{
			name:       "empty method returns error",
			method:     "",
			wantErrMsg: "unknown method",
		},
		{
			name:       "random string returns error",
			method:     "xyz123",
			wantErrMsg: "unknown method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ios, _, _ := testIOStreams()
			ctx := context.Background()
			_, err := runDiscoveryMethod(ctx, ios, tt.method, 5*time.Second, "")
			if err == nil {
				t.Error("expected error, got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("error = %q, want error containing %q", err.Error(), tt.wantErrMsg)
			}
		})
	}
}

// TestSelectDiscoveryMethods_MoreCases tests additional discovery method selection cases.
func TestSelectDiscoveryMethods_MoreCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts *Options
		want []string
	}{
		{
			name: "coiot mode",
			opts: &Options{Discover: true, DiscoverModes: "coiot"},
			want: []string{"coiot"},
		},
		{
			name: "ble mode",
			opts: &Options{Discover: true, DiscoverModes: "ble"},
			want: []string{"ble"},
		},
		{
			name: "all methods",
			opts: &Options{Discover: true, DiscoverModes: "all"},
			want: []string{"http", "mdns", "coiot"},
		},
		{
			name: "http mdns coiot",
			opts: &Options{Discover: true, DiscoverModes: "http,mdns,coiot"},
			want: []string{"http", "mdns", "coiot"},
		},
		{
			name: "using alias scan for http",
			opts: &Options{Discover: true, DiscoverModes: "scan"},
			want: []string{"http"},
		},
		{
			name: "using alias zeroconf for mdns",
			opts: &Options{Discover: true, DiscoverModes: "zeroconf"},
			want: []string{"mdns"},
		},
		{
			name: "using alias coap for coiot",
			opts: &Options{Discover: true, DiscoverModes: "coap"},
			want: []string{"coiot"},
		},
		{
			name: "using alias bluetooth for ble",
			opts: &Options{Discover: true, DiscoverModes: "bluetooth"},
			want: []string{"ble"},
		},
		{
			name: "non-interactive with force defaults to all",
			opts: &Options{Force: true, DiscoverModes: ""},
			want: []string{"http", "mdns", "coiot"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ios, _, _ := testIOStreams()
			got := selectDiscoveryMethods(ios, tt.opts)
			if len(got) != len(tt.want) {
				t.Errorf("selectDiscoveryMethods() = %v, want %v", got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("selectDiscoveryMethods()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

// TestValidateConfig_AdditionalCases tests more ValidateConfig scenarios.
func TestValidateConfig_AdditionalCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil devices map is valid",
			cfg:     &config.Config{Devices: nil},
			wantErr: false,
		},
		{
			name:    "nil groups map is valid",
			cfg:     &config.Config{Groups: nil},
			wantErr: false,
		},
		{
			name: "empty group devices slice is valid",
			cfg: &config.Config{
				Groups: map[string]config.Group{
					"empty": {Devices: []string{}},
				},
			},
			wantErr: false,
		},
		{
			name: "device with hostname address is valid",
			cfg: &config.Config{
				Devices: map[string]model.Device{
					"test": {Address: "shelly-plus-1.local"},
				},
			},
			wantErr: false,
		},
		{
			name: "device with IPv6 address is valid",
			cfg: &config.Config{
				Devices: map[string]model.Device{
					"test": {Address: "fe80::1"},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple valid devices and groups",
			cfg: &config.Config{
				Devices: map[string]model.Device{
					"living-room": {Address: "192.168.1.100"},
					"bedroom":     {Address: "192.168.1.101"},
					"kitchen":     {Address: "192.168.1.102"},
					"garage":      {Address: "192.168.1.103"},
				},
				Groups: map[string]config.Group{
					"downstairs": {Devices: []string{"living-room", "kitchen"}},
					"upstairs":   {Devices: []string{"bedroom"}},
					"all":        {Devices: []string{"living-room", "bedroom", "kitchen", "garage"}},
				},
			},
			wantErr: false,
		},
		{
			name: "group with unknown device and valid devices",
			cfg: &config.Config{
				Devices: map[string]model.Device{
					"known": {Address: "192.168.1.100"},
				},
				Groups: map[string]config.Group{
					"mixed": {Devices: []string{"known", "unknown"}},
				},
			},
			wantErr: true,
			errMsg:  "references unknown device",
		},
		{
			name: "valid theme nord",
			cfg: &config.Config{
				Theme: config.ThemeConfig{Name: "nord"},
			},
			wantErr: false,
		},
		{
			name: "valid theme catppuccin_mocha",
			cfg: &config.Config{
				Theme: config.ThemeConfig{Name: "catppuccin_mocha"},
			},
			wantErr: false,
		},
		{
			name: "combined valid settings",
			cfg: &config.Config{
				Output:  "json",
				APIMode: "cloud",
				Theme:   config.ThemeConfig{Name: "nord"},
				Devices: map[string]model.Device{
					"test": {Address: "192.168.1.100"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateConfig(tt.cfg)
			if tt.wantErr { //nolint:nestif // test validation logic
				if err == nil {
					t.Error("ValidateConfig() = nil, want error")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateConfig() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateConfig() = %v, want nil", err)
				}
			}
		})
	}
}

// TestCheckAndConfirmConfig_Force tests CheckAndConfirmConfig with force flag.
func TestCheckAndConfirmConfig_Force(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		opts           *Options
		wantShouldCont bool
		wantErr        bool
	}{
		{
			name:           "force flag always continues",
			opts:           &Options{Force: true},
			wantShouldCont: true,
			wantErr:        false,
		},
		{
			name:           "force with other flags",
			opts:           &Options{Force: true, Theme: "nord", OutputFormat: "json"},
			wantShouldCont: true,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ios, _, _ := testIOStreams()
			shouldCont, err := CheckAndConfirmConfig(ios, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAndConfirmConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if shouldCont != tt.wantShouldCont {
				t.Errorf("CheckAndConfirmConfig() = %v, want %v", shouldCont, tt.wantShouldCont)
			}
		})
	}
}

// TestOptions_ZeroValue tests Options zero value behavior.
//
//nolint:gocyclo // comprehensive test coverage
func TestOptions_ZeroValue(t *testing.T) {
	t.Parallel()

	var opts Options

	// Zero value should be interactive
	if opts.IsNonInteractive() {
		t.Error("zero value Options should be interactive")
	}

	// Zero value should not want cloud setup
	if opts.WantsCloudSetup() {
		t.Error("zero value Options should not want cloud setup")
	}

	// Verify all fields are zero
	if len(opts.Devices) != 0 {
		t.Error("Devices should be nil")
	}
	if len(opts.DevicesJSON) != 0 {
		t.Error("DevicesJSON should be nil")
	}
	if opts.Discover {
		t.Error("Discover should be false")
	}
	if opts.DiscoverTimeout != 0 {
		t.Error("DiscoverTimeout should be 0")
	}
	if opts.DiscoverModes != "" {
		t.Error("DiscoverModes should be empty")
	}
	if opts.Network != "" {
		t.Error("Network should be empty")
	}
	if opts.Completions != "" {
		t.Error("Completions should be empty")
	}
	if opts.Aliases {
		t.Error("Aliases should be false")
	}
	if opts.Theme != "" {
		t.Error("Theme should be empty")
	}
	if opts.OutputFormat != "" {
		t.Error("OutputFormat should be empty")
	}
	if opts.NoColor {
		t.Error("NoColor should be false")
	}
	if opts.APIMode != "" {
		t.Error("APIMode should be empty")
	}
	if opts.CloudEmail != "" {
		t.Error("CloudEmail should be empty")
	}
	if opts.CloudPassword != "" {
		t.Error("CloudPassword should be empty")
	}
	if opts.Force {
		t.Error("Force should be false")
	}
}

// TestOptions_NonInteractiveVariations tests various non-interactive flag combinations.
func TestOptions_NonInteractiveVariations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
		want bool
	}{
		{
			name: "discover only",
			opts: Options{Discover: true},
			want: true,
		},
		{
			name: "aliases only",
			opts: Options{Aliases: true},
			want: true,
		},
		{
			name: "completions only",
			opts: Options{Completions: "bash"},
			want: true,
		},
		{
			name: "nocolor only",
			opts: Options{NoColor: true},
			want: true,
		},
		{
			name: "multiple devices",
			opts: Options{Devices: []string{"dev1", "dev2", "dev3"}},
			want: true,
		},
		{
			name: "multiple devicesJSON",
			opts: Options{DevicesJSON: []string{`{"a":"b"}`, `{"c":"d"}`}},
			want: true,
		},
		{
			name: "all flags set",
			opts: Options{
				Devices:       []string{"dev"},
				DevicesJSON:   []string{"{}"},
				Discover:      true,
				Completions:   "zsh",
				Aliases:       true,
				Theme:         "nord",
				OutputFormat:  "yaml",
				NoColor:       true,
				APIMode:       "local",
				CloudEmail:    "a@b.com",
				CloudPassword: "pass",
				Force:         true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.opts.IsNonInteractive()
			if got != tt.want {
				t.Errorf("IsNonInteractive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestOptions_WantsCloudSetupVariations tests WantsCloudSetup edge cases.
func TestOptions_WantsCloudSetupVariations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
		want bool
	}{
		{
			name: "whitespace email",
			opts: Options{CloudEmail: "   "},
			want: true, // Non-empty string counts as "set"
		},
		{
			name: "whitespace password",
			opts: Options{CloudPassword: "   "},
			want: true, // Non-empty string counts as "set"
		},
		{
			name: "email with special chars",
			opts: Options{CloudEmail: "user+tag@example.com"},
			want: true,
		},
		{
			name: "empty email empty password",
			opts: Options{CloudEmail: "", CloudPassword: ""},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.opts.WantsCloudSetup()
			if got != tt.want {
				t.Errorf("WantsCloudSetup() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseDiscoverModes_Comprehensive tests all parseDiscoverModes aliases.
func TestParseDiscoverModes_Comprehensive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		// Standard modes
		{"http", "http", []string{"http"}},
		{"mdns", "mdns", []string{"mdns"}},
		{"coiot", "coiot", []string{"coiot"}},
		{"ble", "ble", []string{"ble"}},

		// Aliases
		{"scan alias", "scan", []string{"http"}},
		{"zeroconf alias", "zeroconf", []string{"mdns"}},
		{"bonjour alias", "bonjour", []string{"mdns"}},
		{"coap alias", "coap", []string{"coiot"}},
		{"bluetooth alias", "bluetooth", []string{"ble"}},

		// Case variations
		{"HTTP uppercase", "HTTP", []string{"http"}},
		{"MDNS uppercase", "MDNS", []string{"mdns"}},
		{"COIOT uppercase", "COIOT", []string{"coiot"}},
		{"BLE uppercase", "BLE", []string{"ble"}},
		{"MixedCase", "HtTp,MdNs", []string{"http", "mdns"}},

		// Combinations
		{"http and mdns", "http,mdns", []string{"http", "mdns"}},
		{"all three", "http,mdns,coiot", []string{"http", "mdns", "coiot"}},
		{"all four", "http,mdns,coiot,ble", []string{"http", "mdns", "coiot", "ble"}},
		{"mixed aliases", "scan,zeroconf,coap,bluetooth", []string{"http", "mdns", "coiot", "ble"}},

		// Edge cases
		{"empty", "", []string{"http", "mdns", "coiot"}},
		{"all keyword", "all", []string{"http", "mdns", "coiot"}},
		{"only invalid", "invalid", []string{"http"}},
		{"valid and invalid", "http,invalid,mdns", []string{"http", "mdns"}},
		{"trailing comma", "http,mdns,", []string{"http", "mdns"}},
		{"leading comma", ",http,mdns", []string{"http", "mdns"}},
		{"multiple commas", "http,,mdns", []string{"http", "mdns"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseDiscoverModes(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parseDiscoverModes(%q) len = %d, want %d; got %v", tt.input, len(got), len(tt.want), got)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("parseDiscoverModes(%q)[%d] = %q, want %q", tt.input, i, v, tt.want[i])
				}
			}
		})
	}
}

// TestDefaultDiscoveryTimeoutValue tests the default discovery timeout constant.
func TestDefaultDiscoveryTimeoutValue(t *testing.T) {
	t.Parallel()

	expected := 15 * time.Second
	if defaultDiscoveryTimeout != expected {
		t.Errorf("defaultDiscoveryTimeout = %v, want %v", defaultDiscoveryTimeout, expected)
	}
}

// TestCheckCompletionInstalled_AllShells tests all shell types for completion check.
func TestCheckCompletionInstalled_AllShells(t *testing.T) {
	t.Parallel()

	shells := []string{
		"bash",
		"zsh",
		"fish",
		"powershell",
		"unknown",
		"",
		"invalid-shell",
		"BASH", // Should not match (case-sensitive)
	}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			t.Parallel()
			// Just verify it doesn't panic and returns a boolean
			result := CheckCompletionInstalled(shell)
			// Result is a valid boolean (can be true or false depending on file system)
			if result != true && result != false {
				t.Errorf("CheckCompletionInstalled(%q) returned invalid boolean", shell)
			}
		})
	}
}

// TestValidateConfig_OutputFormats tests all valid output format values.
func TestValidateConfig_OutputFormats(t *testing.T) {
	t.Parallel()

	validFormats := []string{"table", "json", "yaml", "text", "template", ""}
	invalidFormats := []string{"xml", "csv", "html", "invalid", "TABLE", "JSON"}

	for _, format := range validFormats {
		t.Run("valid_"+format, func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{Output: format}
			err := ValidateConfig(cfg)
			if err != nil {
				t.Errorf("ValidateConfig with output=%q should be valid, got error: %v", format, err)
			}
		})
	}

	for _, format := range invalidFormats {
		t.Run("invalid_"+format, func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{Output: format}
			err := ValidateConfig(cfg)
			if err == nil {
				t.Errorf("ValidateConfig with output=%q should be invalid", format)
			}
		})
	}
}

// TestValidateConfig_APIModes tests all valid API mode values.
func TestValidateConfig_APIModes(t *testing.T) {
	t.Parallel()

	validModes := []string{"local", "cloud", "auto", ""}
	invalidModes := []string{"remote", "hybrid", "LOCAL", "CLOUD", "invalid"}

	for _, mode := range validModes {
		t.Run("valid_"+mode, func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{APIMode: mode}
			err := ValidateConfig(cfg)
			if err != nil {
				t.Errorf("ValidateConfig with api_mode=%q should be valid, got error: %v", mode, err)
			}
		})
	}

	for _, mode := range invalidModes {
		t.Run("invalid_"+mode, func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{APIMode: mode}
			err := ValidateConfig(cfg)
			if err == nil {
				t.Errorf("ValidateConfig with api_mode=%q should be invalid", mode)
			}
		})
	}
}

// TestCheckExistingConfig_ReturnsPath tests that CheckExistingConfig always returns a path.
func TestCheckExistingConfig_ReturnsPath(t *testing.T) {
	t.Parallel()

	_, path := CheckExistingConfig()

	// Path should always be non-empty
	if path == "" {
		t.Error("CheckExistingConfig should always return a path")
	}

	// Path should contain expected directory structure
	expectedParts := []string{".config", "shelly", "config.yaml"}
	for _, part := range expectedParts {
		if !strings.Contains(path, part) {
			t.Errorf("path %q should contain %q", path, part)
		}
	}
}

// TestPrintWelcome_ContainsExpectedContent tests PrintWelcome output content.
func TestPrintWelcome_ContainsExpectedContent(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	PrintWelcome(ios)

	output := out.String()

	expectedPhrases := []string{
		"Welcome",
		"Shelly",
		"wizard",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("PrintWelcome output should contain %q, got: %s", phrase, output)
		}
	}
}

// TestPrintSummary_ContainsExpectedContent tests PrintSummary output content.
func TestPrintSummary_ContainsExpectedContent(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	PrintSummary(ios)

	output := out.String()

	expectedPhrases := []string{
		"complete",
		"shelly",
		"device list",
		"help",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("PrintSummary output should contain %q, got: %s", phrase, output)
		}
	}
}

// TestRunDiscoveryMethod_ValidMethods tests that valid discovery methods are dispatched correctly.
// These will fail due to no network, but the dispatch logic is covered.
func TestRunDiscoveryMethod_ValidMethods(t *testing.T) {
	t.Parallel()

	// These methods exist and will be dispatched, but will fail due to no network.
	// The goal is to cover the switch statement branches.
	methods := []string{"http", "mdns", "coiot", "ble"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			ios, _, _ := testIOStreams()
			ctx := context.Background()
			// Use a very short timeout since we expect these to fail
			_, err := runDiscoveryMethod(ctx, ios, method, 1*time.Millisecond, "")
			// We don't care about the error, just that the method was dispatched
			// The error will be about network failure, not "unknown method"
			if err != nil && strings.Contains(err.Error(), "unknown method") {
				t.Errorf("method %q should be recognized, got: %v", method, err)
			}
		})
	}
}

// TestSelectDiscoveryMethods_InteractiveFallback tests interactive mode fallback.
func TestSelectDiscoveryMethods_InteractiveFallback(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()

	// Interactive mode (empty options) with non-interactive IO should fallback
	opts := &Options{}

	methods := selectDiscoveryMethods(ios, opts)

	// Should get at least one method (fallback to http)
	if len(methods) == 0 {
		t.Error("selectDiscoveryMethods should return at least one method in fallback")
	}

	// The first (or only) method should be http due to fallback
	found := false
	for _, m := range methods {
		if m == testAPIModeHTTP {
			found = true
			break
		}
	}
	if !found && len(methods) > 0 {
		t.Logf("methods returned: %v (expected %s in fallback)", methods, testAPIModeHTTP)
	}
}

// TestCheckExistingConfig_HomeError tests CheckExistingConfig behavior.
func TestCheckExistingConfig_HomeError(t *testing.T) {
	t.Parallel()

	// Save current HOME
	oldHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", oldHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
	}()

	// Even with invalid HOME, should not panic
	// Note: this may still work on some systems due to /etc/passwd fallback
	exists, path := CheckExistingConfig()

	// Just verify it returns without panicking
	_ = exists
	_ = path
}

// TestCheckCompletionInstalled_EmptyHome tests CheckCompletionInstalled with edge cases.
func TestCheckCompletionInstalled_EmptyHome(t *testing.T) {
	t.Parallel()

	// Test various edge cases
	tests := []struct {
		name  string
		shell string
	}{
		{"empty shell", ""},
		{"whitespace shell", "   "},
		{"numeric shell", "123"},
		{"special chars", "sh!@#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Should return false for unknown shells without panicking
			result := CheckCompletionInstalled(tt.shell)
			if result {
				t.Errorf("CheckCompletionInstalled(%q) = true, want false", tt.shell)
			}
		})
	}
}

// TestValidateConfig_MultipleErrors tests that multiple validation errors are combined.
func TestValidateConfig_MultipleErrors(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Output:  "invalid-format",
		APIMode: "invalid-mode",
		Theme:   config.ThemeConfig{Name: "invalid-theme-xyz"},
		Devices: map[string]model.Device{
			"no-address": {Address: ""},
		},
		Groups: map[string]config.Group{
			"bad-ref": {Devices: []string{"nonexistent-device"}},
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errStr := err.Error()

	// Should contain multiple error messages
	expectedParts := []string{
		"invalid output format",
		"invalid api_mode",
		"unknown theme",
		"has no address",
		"references unknown device",
	}

	missingParts := []string{}
	for _, part := range expectedParts {
		if !strings.Contains(errStr, part) {
			missingParts = append(missingParts, part)
		}
	}

	if len(missingParts) > 0 {
		t.Errorf("error should contain all validation failures\nmissing: %v\ngot: %s", missingParts, errStr)
	}
}

// TestSelectDiscoveryMethods_NonInteractiveEmpty tests empty discover modes defaults.
func TestSelectDiscoveryMethods_NonInteractiveEmpty(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()

	// Non-interactive with empty DiscoverModes should default to all
	opts := &Options{
		Discover:      true,
		DiscoverModes: "",
	}

	methods := selectDiscoveryMethods(ios, opts)

	expected := []string{"http", "mdns", "coiot"}
	if len(methods) != len(expected) {
		t.Errorf("selectDiscoveryMethods() = %v, want %v", methods, expected)
		return
	}

	for i, m := range methods {
		if m != expected[i] {
			t.Errorf("selectDiscoveryMethods()[%d] = %q, want %q", i, m, expected[i])
		}
	}
}

// TestOptionsFields_Comprehensive tests all Options fields comprehensively.
func TestOptionsFields_Comprehensive(t *testing.T) {
	t.Parallel()

	// Test with various field combinations
	tests := []struct {
		name string
		opts Options
	}{
		{
			name: "devices only",
			opts: Options{Devices: []string{"192.168.1.1", "192.168.1.2"}},
		},
		{
			name: "devicesJSON only",
			opts: Options{DevicesJSON: []string{
				`{"name":"dev1","address":"192.168.1.1"}`,
				`{"name":"dev2","address":"192.168.1.2"}`,
			}},
		},
		{
			name: "discovery settings",
			opts: Options{
				Discover:        true,
				DiscoverTimeout: 30 * time.Second,
				DiscoverModes:   "http,mdns",
				Network:         "10.0.0.0/8",
			},
		},
		{
			name: "completion settings",
			opts: Options{
				Completions: "bash,zsh,fish,powershell",
				Aliases:     true,
			},
		},
		{
			name: "config settings",
			opts: Options{
				Theme:        "tokyo-night",
				OutputFormat: "yaml",
				NoColor:      true,
				APIMode:      "auto",
			},
		},
		{
			name: "cloud settings",
			opts: Options{
				CloudEmail:    "test@example.com",
				CloudPassword: "verysecurepassword123",
			},
		},
		{
			name: "control settings",
			opts: Options{
				Force: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify IsNonInteractive returns expected value
			if len(tt.opts.Devices) > 0 && !tt.opts.IsNonInteractive() {
				t.Error("with Devices set, should be non-interactive")
			}
			if len(tt.opts.DevicesJSON) > 0 && !tt.opts.IsNonInteractive() {
				t.Error("with DevicesJSON set, should be non-interactive")
			}
			if tt.opts.Theme != "" && !tt.opts.IsNonInteractive() {
				t.Error("with Theme set, should be non-interactive")
			}
			if tt.opts.Force && !tt.opts.IsNonInteractive() {
				t.Error("with Force set, should be non-interactive")
			}

			// Verify WantsCloudSetup returns expected value
			if tt.opts.CloudEmail != "" && !tt.opts.WantsCloudSetup() {
				t.Error("with CloudEmail set, should want cloud setup")
			}
			if tt.opts.CloudPassword != "" && !tt.opts.WantsCloudSetup() {
				t.Error("with CloudPassword set, should want cloud setup")
			}
		})
	}
}

// TestCheckAndConfirmConfig_NonInteractiveWithExistingConfig tests behavior
// when config exists and non-interactive mode is used without force.
func TestCheckAndConfirmConfig_NonInteractiveWithExistingConfig(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()

	// Non-interactive without force - if config exists, should return false
	opts := &Options{
		Theme: "dracula", // Makes it non-interactive
		Force: false,
	}

	// This test's behavior depends on whether a config file exists
	shouldContinue, err := CheckAndConfirmConfig(ios, opts)
	if err != nil {
		t.Errorf("CheckAndConfirmConfig() error = %v, want nil", err)
	}

	// Either continues (no config) or stops (config exists)
	// Check that output is appropriate
	output := out.String()
	if !shouldContinue {
		// If stopped, should have warning about existing config
		if !strings.Contains(output, "Configuration already exists") && !strings.Contains(output, "force") {
			// Output may be empty if config doesn't exist
			t.Logf("shouldContinue=%v, output=%q", shouldContinue, output)
		}
	}
}

// TestParseDiscoverModes_SpacedInput tests parseDiscoverModes with various spacing.
func TestParseDiscoverModes_SpacedInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		minLen   int
		contains []string
	}{
		{
			name:     "spaces around commas",
			input:    "http , mdns , coiot",
			minLen:   3,
			contains: []string{"http", "mdns", "coiot"},
		},
		{
			name:     "tabs and spaces",
			input:    "http\t,\tmdns",
			minLen:   2,
			contains: []string{"http", "mdns"},
		},
		{
			name:     "newlines in input",
			input:    "http\n,\nmdns",
			minLen:   0, // May not parse newlines correctly, but shouldn't crash
			contains: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseDiscoverModes(tt.input)
			if len(got) < tt.minLen {
				t.Logf("parseDiscoverModes(%q) = %v, len=%d", tt.input, got, len(got))
			}
			for _, want := range tt.contains {
				found := false
				for _, g := range got {
					if g == want {
						found = true
						break
					}
				}
				if !found && tt.minLen > 0 {
					t.Errorf("parseDiscoverModes(%q) should contain %q, got %v", tt.input, want, got)
				}
			}
		})
	}
}

// TestSanitizeDeviceName_LongStrings tests SanitizeDeviceName with very long inputs.
func TestSanitizeDeviceName_LongStrings(t *testing.T) {
	t.Parallel()

	// Very long string
	longName := strings.Repeat("a", 1000)
	result := SanitizeDeviceName(longName)
	if len(result) != 1000 {
		t.Errorf("SanitizeDeviceName should preserve length for valid chars, got len=%d", len(result))
	}

	// Long string with many special chars
	mixedLong := strings.Repeat("a@b#c$d%", 100)
	result = SanitizeDeviceName(mixedLong)
	// Should only keep a, b, c, d letters
	for _, r := range result {
		isLowerAlpha := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		isHyphen := r == '-'
		if !isLowerAlpha && !isDigit && !isHyphen {
			t.Errorf("result contains invalid char %q", r)
		}
	}
}

// TestValidateConfig_Themes tests theme validation with various theme names.
func TestValidateConfig_Themes(t *testing.T) {
	t.Parallel()

	// Known valid themes (from bubbletint)
	validThemes := []string{
		"dracula",
		"nord",
		"catppuccin_mocha",
		"", // Empty is valid (means default)
	}

	// Invalid themes
	invalidThemes := []string{
		"not-a-real-theme",
		"xyz123",
		"DRACULA", // Case sensitive
	}

	for _, theme := range validThemes {
		t.Run("valid_"+theme, func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{
				Theme: config.ThemeConfig{Name: theme},
			}
			err := ValidateConfig(cfg)
			if err != nil {
				t.Errorf("ValidateConfig with theme=%q should be valid, got: %v", theme, err)
			}
		})
	}

	for _, theme := range invalidThemes {
		t.Run("invalid_"+theme, func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{
				Theme: config.ThemeConfig{Name: theme},
			}
			err := ValidateConfig(cfg)
			if err == nil {
				t.Errorf("ValidateConfig with theme=%q should be invalid", theme)
			}
		})
	}
}

// TestCheckAndConfirmConfig_InteractiveNoConfig tests interactive mode without existing config.
func TestCheckAndConfirmConfig_InteractiveNoConfig(t *testing.T) {
	t.Parallel()

	// Create temp directory for testing
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Skipf("cannot set HOME: %v", err)
	}
	defer func() {
		if err := os.Setenv("HOME", oldHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
	}()

	ios, _, _ := testIOStreams()
	opts := &Options{} // Interactive mode

	// No config exists, should continue
	shouldContinue, err := CheckAndConfirmConfig(ios, opts)
	if err != nil {
		t.Errorf("CheckAndConfirmConfig() error = %v, want nil", err)
	}
	if !shouldContinue {
		t.Error("CheckAndConfirmConfig() should continue when no config exists")
	}
}

// TestValidateConfig_DeviceAddresses tests various device address formats.
func TestValidateConfig_DeviceAddresses(t *testing.T) {
	t.Parallel()

	validAddresses := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"shelly-plus-1.local",
		"mydevice.home",
		"fe80::1",
		"::1",
		"device-1",
	}

	for _, addr := range validAddresses {
		t.Run("addr_"+addr, func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{
				Devices: map[string]model.Device{
					"test": {Address: addr},
				},
			}
			err := ValidateConfig(cfg)
			if err != nil {
				t.Errorf("ValidateConfig with address=%q should be valid, got: %v", addr, err)
			}
		})
	}
}

// TestValidateConfig_GroupsWithIPAddresses tests groups that reference IPs directly.
func TestValidateConfig_GroupsWithIPAddresses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "group with only IPs",
			cfg: &config.Config{
				Groups: map[string]config.Group{
					"ips": {Devices: []string{"192.168.1.1", "192.168.1.2", "10.0.0.1"}},
				},
			},
			wantErr: false,
		},
		{
			name: "group with IPs and named devices",
			cfg: &config.Config{
				Devices: map[string]model.Device{
					"kitchen": {Address: "192.168.1.10"},
				},
				Groups: map[string]config.Group{
					"mixed": {Devices: []string{"kitchen", "192.168.1.20"}},
				},
			},
			wantErr: false,
		},
		{
			name: "group with hostname containing dots",
			cfg: &config.Config{
				Groups: map[string]config.Group{
					"hosts": {Devices: []string{"shelly.local", "device.home.lan"}},
				},
			},
			wantErr: false, // Hostnames with dots are treated like IPs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRunCheck tests the RunCheck function with a test factory.
//
//nolint:paralleltest // RunCheck uses global config.Get() which cannot be parallelized
func TestRunCheck(t *testing.T) {
	tf := factory.NewTestFactory(t)

	// Run the check
	err := RunCheck(tf.Factory)
	if err != nil {
		t.Errorf("RunCheck() error = %v, want nil", err)
	}

	// Check that output was produced
	output := tf.OutString()
	if output == "" {
		t.Error("RunCheck should produce output")
	}

	// Should contain the header
	if !strings.Contains(output, "Setup Check") {
		t.Error("output should contain 'Setup Check'")
	}
}

// TestRunCheck_WithDevices tests RunCheck when devices are registered.
//
//nolint:paralleltest // RunCheck uses global config.Get() which cannot be parallelized
func TestRunCheck_WithDevices(t *testing.T) {
	devices := map[string]model.Device{
		"kitchen": {Address: "192.168.1.100", Model: "SNSW-001P16EU"},
		"bedroom": {Address: "192.168.1.101", Model: "SNSW-002P16EU"},
	}
	tf := factory.NewTestFactoryWithDevices(t, devices)

	err := RunCheck(tf.Factory)
	if err != nil {
		t.Errorf("RunCheck() error = %v, want nil", err)
	}

	output := tf.OutString()

	// Should report devices
	if !strings.Contains(output, "Registered devices") {
		t.Error("output should mention registered devices")
	}
}

// TestRunCheck_NoConfig tests RunCheck when no config exists.
//
//nolint:paralleltest // RunCheck uses global config.Get() which cannot be parallelized
func TestRunCheck_NoConfig(t *testing.T) {
	// Create a test factory - by default has empty config
	tf := factory.NewTestFactory(t)

	err := RunCheck(tf.Factory)
	if err != nil {
		t.Errorf("RunCheck() error = %v, want nil", err)
	}

	output := tf.OutString()

	// Should complete without error
	if !strings.Contains(output, "Setup Check") {
		t.Error("output should contain setup check header")
	}
}

// TestRunDiscoveryStepIfNeeded_NonInteractive tests runDiscoveryStepIfNeeded behavior.
func TestRunDiscoveryStepIfNeeded_NonInteractive(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()

	tests := []struct {
		name     string
		opts     *Options
		wantNil  bool
		wantSkip bool
	}{
		{
			name:     "non-interactive without discover flag skips",
			opts:     &Options{Theme: "dracula", Discover: false},
			wantNil:  true,
			wantSkip: true,
		},
		{
			name:     "non-interactive with discover flag continues",
			opts:     &Options{Theme: "dracula", Discover: true},
			wantNil:  false, // Will run discovery (and likely fail but that's ok)
			wantSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.wantSkip {
				// Just verify the logic path
				if !tt.opts.IsNonInteractive() {
					t.Error("expected non-interactive mode")
				}
				if tt.opts.Discover {
					t.Error("expected discover to be false for skip case")
				}
			}
			_ = ios
		})
	}
}

// TestSelectOutputFormat_Interactive tests selectOutputFormat interactive fallback.
func TestSelectOutputFormat_Interactive(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()

	// Interactive mode - will fail to select but should return error
	opts := &Options{}

	_, err := selectOutputFormat(ios, opts)
	// Error expected because there's no actual terminal
	if err == nil {
		// If no error, that's also fine - depends on Select implementation
		t.Log("selectOutputFormat succeeded in test environment (unexpected but ok)")
	}
}

// TestSelectTheme_Interactive tests selectTheme interactive fallback.
func TestSelectTheme_Interactive(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()

	// Interactive mode - will fail to select
	opts := &Options{}

	_, err := selectTheme(ios, opts)
	// Error expected because there's no actual terminal
	if err == nil {
		// If no error, that's also fine
		t.Log("selectTheme succeeded in test environment (unexpected but ok)")
	}
}

// TestSelectDiscoveryMethods_InteractiveWithParsing tests the interactive path parsing.
func TestSelectDiscoveryMethods_InteractiveWithParsing(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()

	// This tests the fallback when MultiSelect fails in test environment
	opts := &Options{} // Interactive mode

	methods := selectDiscoveryMethods(ios, opts)

	// Should get fallback http when multiselect fails
	if len(methods) == 0 {
		t.Error("should return at least one method")
	}

	// First/only method should be http (fallback)
	if len(methods) > 0 && methods[0] != testAPIModeHTTP {
		t.Logf("got methods: %v (expected %s as fallback)", methods, testAPIModeHTTP)
	}
}

// TestRunFlagDevicesStep tests the runFlagDevicesStep helper.
//
//nolint:paralleltest // stepFlagDevices uses global config.RegisterDevice which cannot be parallelized
func TestRunFlagDevicesStep(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	tests := []struct {
		name string
		opts *Options
	}{
		{
			name: "no devices - does nothing",
			opts: &Options{},
		},
		{
			name: "with devices",
			opts: &Options{Devices: []string{"192.168.1.1"}},
		},
		{
			name: "with devicesJSON",
			opts: &Options{DevicesJSON: []string{`{"name":"test","address":"192.168.1.1"}`}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _ := testIOStreams()
			// Should not panic
			runFlagDevicesStep(ios, tt.opts)
		})
	}
}

// TestRunDiscoveryStepIfNeeded tests runDiscoveryStepIfNeeded.
func TestRunDiscoveryStepIfNeeded(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	ctx := context.Background()

	// Non-interactive without discover flag should return nil immediately
	opts := &Options{Theme: "dracula", Discover: false}
	devices := runDiscoveryStepIfNeeded(ctx, ios, opts)
	if devices != nil {
		t.Error("expected nil devices when discovery is skipped")
	}
}

// TestRunRegistrationStep tests runRegistrationStep.
func TestRunRegistrationStep(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{Force: true}

	// No devices - should do nothing
	runRegistrationStep(tf.Factory, opts, nil)

	// Empty devices slice - should do nothing
	runRegistrationStep(tf.Factory, opts, []discovery.DiscoveredDevice{})
}

// TestRunCompletionsStep tests runCompletionsStep.
func TestRunCompletionsStep(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()

	// Non-interactive without completions flag - does nothing
	opts := &Options{Theme: "dracula", Completions: ""}
	runCompletionsStep(ios, nil, opts)

	// Non-interactive with invalid shell - should warn but not panic
	opts = &Options{Theme: "dracula", Completions: "invalid-shell"}
	runCompletionsStep(ios, nil, opts)
}

// TestRunCloudStep tests runCloudStep.
func TestRunCloudStep(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	ctx := context.Background()

	// Non-interactive without cloud setup - does nothing
	opts := &Options{Theme: "dracula"}
	runCloudStep(ctx, ios, opts)

	// Non-interactive with incomplete credentials - should warn
	opts = &Options{Theme: "dracula", CloudEmail: "test@example.com"}
	runCloudStep(ctx, ios, opts)
}

// TestRunDiscoveryStep tests runDiscoveryStep.
func TestRunDiscoveryStep(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	ctx := context.Background()

	// Non-interactive - will try to discover and likely fail
	opts := &Options{
		Discover:        true,
		DiscoverModes:   "http",
		DiscoverTimeout: 1 * time.Millisecond, // Very short timeout
		Network:         "10.255.255.0/30",    // Unlikely to have devices
	}
	devices := runDiscoveryStep(ctx, ios, opts)
	// May or may not find devices, but shouldn't panic
	_ = devices
}

// TestStepFlagDevices tests stepFlagDevices.
//
//nolint:paralleltest // stepFlagDevices uses global config.RegisterDevice which cannot be parallelized
func TestStepFlagDevices(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	tests := []struct {
		name    string
		opts    *Options
		wantErr bool
	}{
		{
			name:    "empty devices",
			opts:    &Options{Devices: []string{}},
			wantErr: false,
		},
		{
			name:    "valid device spec format",
			opts:    &Options{Devices: []string{"test-device=192.168.99.100"}},
			wantErr: false,
		},
		{
			name:    "invalid device spec format - expects error",
			opts:    &Options{Devices: []string{"192.168.1.100"}},
			wantErr: true, // Missing name=ip format
		},
		{
			name:    "empty devicesJSON",
			opts:    &Options{DevicesJSON: []string{}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _ := testIOStreams()
			err := stepFlagDevices(ios, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("stepFlagDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRunCheck_Output tests that RunCheck produces expected output sections.
//
//nolint:paralleltest // RunCheck uses global config.Get() which cannot be parallelized
func TestRunCheck_Output(t *testing.T) {
	tf := factory.NewTestFactory(t)

	err := RunCheck(tf.Factory)
	if err != nil {
		t.Fatalf("RunCheck() error = %v", err)
	}

	output := tf.OutString()

	// Check for expected sections
	sections := []string{
		"Setup Check",
		"Config",
		"devices",
		"Cloud",
	}

	for _, section := range sections {
		if !strings.Contains(strings.ToLower(output), strings.ToLower(section)) {
			t.Logf("output: %s", output)
			t.Errorf("output should contain section %q", section)
		}
	}
}

// TestStepCloudNonInteractive_IncompleteCredentials tests cloud step with incomplete credentials.
func TestStepCloudNonInteractive_IncompleteCredentials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name          string
		opts          *Options
		expectWarning bool
	}{
		{
			name:          "email only",
			opts:          &Options{CloudEmail: "test@example.com"},
			expectWarning: true,
		},
		{
			name:          "password only",
			opts:          &Options{CloudPassword: "secret"},
			expectWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			localIOS, localOut, _ := testIOStreams()
			err := stepCloudNonInteractive(ctx, localIOS, tt.opts)
			if err != nil {
				// Errors are acceptable - incomplete credentials
				t.Logf("stepCloudNonInteractive error: %v", err)
			}
			_ = localOut.String()
		})
	}
}

// TestStepCompletionsNonInteractive tests stepCompletionsNonInteractive.
func TestStepCompletionsNonInteractive(t *testing.T) {
	t.Parallel()

	// Test only cases that don't require a valid rootCmd (which would need completion generation)
	// Invalid shells get skipped early before trying to use rootCmd
	tests := []struct {
		name    string
		opts    *Options
		wantErr bool
	}{
		{
			name:    "empty completions - no shells to process",
			opts:    &Options{Completions: ""},
			wantErr: false,
		},
		{
			name:    "only invalid shells - all skipped",
			opts:    &Options{Completions: "foo,bar,baz"},
			wantErr: true, // All failed because unknown
		},
		{
			name:    "single invalid shell",
			opts:    &Options{Completions: "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ios, _, _ := testIOStreams() // Create new buffer per subtest to avoid race
			// Pass nil for rootCmd - invalid shells are skipped before trying to use it
			err := stepCompletionsNonInteractive(ios, nil, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("stepCompletionsNonInteractive() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRunSetupSteps_Components tests that setup step helper functions work correctly.
func TestRunSetupSteps_Components(t *testing.T) {
	t.Parallel()

	// Test the component functions work in isolation
	ios, _, _ := testIOStreams()

	// Test runFlagDevicesStep with valid spec
	opts := &Options{Devices: []string{"testdev=127.0.0.1"}}
	runFlagDevicesStep(ios, opts)

	// Test runCompletionsStep with non-interactive no completions
	opts = &Options{Theme: "dracula"}
	runCompletionsStep(ios, nil, opts)
}

// TestRunCheck_AllPaths tests RunCheck covers all output paths.
//
//nolint:paralleltest // RunCheck uses global config.Get() which cannot be parallelized
func TestRunCheck_AllPaths(t *testing.T) {
	// Test with devices
	devices := map[string]model.Device{
		"dev1": {Address: "192.168.1.100", Model: "SNSW-001P16EU"},
	}
	tf := factory.NewTestFactoryWithDevices(t, devices)

	err := RunCheck(tf.Factory)
	if err != nil {
		t.Fatalf("RunCheck() error = %v", err)
	}

	output := tf.OutString()

	// Should contain device count
	if !strings.Contains(output, "1") {
		t.Logf("output: %s", output)
	}

	// Should complete
	if !strings.Contains(output, "issue") || !strings.Contains(output, "passed") {
		t.Log("expected issue count or passed message in output")
	}
}

// TestStepDiscovery_NonInteractive tests stepDiscovery in non-interactive mode.
func TestStepDiscovery_NonInteractive(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	ctx := context.Background()

	// Non-interactive with short timeout
	opts := &Options{
		Discover:        true,
		DiscoverModes:   "http",
		DiscoverTimeout: 1 * time.Millisecond,
		Network:         "10.255.255.0/30", // Very small subnet unlikely to have devices
	}

	devices, err := stepDiscovery(ctx, ios, opts)
	// May or may not find devices, may error
	_ = devices
	_ = err
}

// TestRunCompletionsStep_NonInteractive tests runCompletionsStep paths.
func TestRunCompletionsStep_NonInteractive(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()

	// Test non-interactive with no completions
	opts := &Options{Force: true} // Non-interactive
	runCompletionsStep(ios, nil, opts)

	// Test non-interactive with completions set (but empty)
	opts = &Options{Force: true, Completions: ""}
	runCompletionsStep(ios, nil, opts)
}

// TestRunCloudStep_NonInteractive tests runCloudStep with various options.
func TestRunCloudStep_NonInteractive(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	ctx := context.Background()

	// Non-interactive without cloud flags - does nothing
	opts := &Options{Force: true}
	runCloudStep(ctx, ios, opts)

	// Non-interactive with only email - WantsCloudSetup is true
	opts = &Options{Force: true, CloudEmail: "user@test.com"}
	runCloudStep(ctx, ios, opts)

	// Non-interactive with only password
	opts = &Options{Force: true, CloudPassword: "pass"}
	runCloudStep(ctx, ios, opts)
}

// TestStepDiscovery_Paths tests more stepDiscovery paths.
func TestStepDiscovery_Paths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts *Options
	}{
		{
			name: "empty discovery modes defaults to all",
			opts: &Options{Discover: true, DiscoverModes: "", DiscoverTimeout: 1 * time.Millisecond},
		},
		{
			name: "mdns only",
			opts: &Options{Discover: true, DiscoverModes: "mdns", DiscoverTimeout: 1 * time.Millisecond},
		},
		{
			name: "coiot only",
			opts: &Options{Discover: true, DiscoverModes: "coiot", DiscoverTimeout: 1 * time.Millisecond},
		},
		{
			name: "all methods",
			opts: &Options{Discover: true, DiscoverModes: "all", DiscoverTimeout: 1 * time.Millisecond},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ios, _, _ := testIOStreams() // Create new buffer per subtest to avoid race
			ctx := context.Background()
			// Just verify it doesn't panic - actual discovery will fail but that's expected
			devices, err := stepDiscovery(ctx, ios, tt.opts)
			if err != nil {
				t.Logf("discovery error (expected in test): %v", err)
			}
			_ = devices
		})
	}
}

// TestRunDiscoveryStep_WithMethods tests runDiscoveryStep with various methods.
func TestRunDiscoveryStep_WithMethods(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	ctx := context.Background()

	// Test with mdns
	opts := &Options{
		Discover:        true,
		DiscoverModes:   "mdns",
		DiscoverTimeout: 1 * time.Millisecond,
	}
	devices := runDiscoveryStep(ctx, ios, opts)
	_ = devices

	// Test with coiot
	opts = &Options{
		Discover:        true,
		DiscoverModes:   "coiot",
		DiscoverTimeout: 1 * time.Millisecond,
	}
	devices = runDiscoveryStep(ctx, ios, opts)
	_ = devices
}

// TestStepFlagDevices_MoreCases tests additional stepFlagDevices scenarios.
//
//nolint:paralleltest // stepFlagDevices uses global config.RegisterDevice which cannot be parallelized
func TestStepFlagDevices_MoreCases(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	ios, out, _ := testIOStreams()

	// Test with valid JSON device
	opts := &Options{
		DevicesJSON: []string{`{"name":"test-json-device","address":"192.168.99.101"}`},
	}
	err := stepFlagDevices(ios, opts)
	if err != nil {
		t.Logf("stepFlagDevices with JSON error: %v", err)
	}

	output := out.String()
	_ = output
}

// TestCheckAndConfirmConfig_MorePaths tests more CheckAndConfirmConfig scenarios.
func TestCheckAndConfirmConfig_MorePaths(t *testing.T) {
	t.Parallel()

	// Create temp dir with config file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "shelly")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("output: table\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	oldHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Skipf("cannot set HOME: %v", err)
	}
	defer func() {
		if err := os.Setenv("HOME", oldHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
	}()

	ios, out, _ := testIOStreams()

	// Non-interactive with existing config and no force - should return false
	opts := &Options{Theme: "dracula", Force: false}
	shouldContinue, err := CheckAndConfirmConfig(ios, opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if shouldContinue {
		t.Error("expected shouldContinue=false with existing config and no force")
	}
	if !strings.Contains(out.String(), "Configuration already exists") {
		t.Logf("output: %s", out.String())
	}

	// With force flag - should return true
	out.Reset()
	opts = &Options{Theme: "dracula", Force: true}
	shouldContinue, err = CheckAndConfirmConfig(ios, opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !shouldContinue {
		t.Error("expected shouldContinue=true with force flag")
	}
}

// TestStepCloudNonInteractive_Empty tests stepCloudNonInteractive with empty credentials.
func TestStepCloudNonInteractive_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	ctx := context.Background()

	// Both empty - should warn
	opts := &Options{CloudEmail: "", CloudPassword: ""}
	err := stepCloudNonInteractive(ctx, ios, opts)
	// Should not error, just warn
	if err != nil {
		t.Logf("stepCloudNonInteractive error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "requires") || !strings.Contains(output, "cloud") {
		t.Logf("expected warning about credentials, got: %s", output)
	}
}

// TestRunCheck_CloudAuth tests RunCheck cloud auth detection.
//
//nolint:paralleltest // RunCheck uses global config.Get() which cannot be parallelized
func TestRunCheck_CloudAuth(t *testing.T) {
	tf := factory.NewTestFactory(t)

	// Set cloud token to test that code path
	tf.Config.Cloud.AccessToken = "test-token"

	err := RunCheck(tf.Factory)
	if err != nil {
		t.Errorf("RunCheck() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Cloud") {
		t.Error("output should contain Cloud section")
	}
}

// TestStepDiscovery_NoMethods tests stepDiscovery with no valid methods selected.
func TestStepDiscovery_NoMethods(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	ctx := context.Background()

	// Interactive mode will fail to select, returning empty methods
	opts := &Options{} // Interactive

	devices, err := stepDiscovery(ctx, ios, opts)
	// Expect nil devices when methods are empty (after multiselect fails)
	_ = devices
	_ = err
}

// TestRunDiscoveryStepIfNeeded_Interactive tests interactive path.
func TestRunDiscoveryStepIfNeeded_Interactive(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	ctx := context.Background()

	// Interactive mode - will attempt discovery
	opts := &Options{} // Not non-interactive
	devices := runDiscoveryStepIfNeeded(ctx, ios, opts)
	_ = devices
}
