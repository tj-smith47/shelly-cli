package ui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/browser"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const testURL = "http://192.168.1.100"

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "ui <device>" {
		t.Errorf("Use = %q, want 'ui <device>'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Aliases) == 0 {
		t.Fatal("Aliases should not be empty")
	}

	aliasSet := make(map[string]bool)
	for _, alias := range cmd.Aliases {
		aliasSet[alias] = true
	}

	if !aliasSet["web"] {
		t.Error("Expected 'web' alias")
	}
	if !aliasSet["open"] {
		t.Error("Expected 'open' alias")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	copyURLFlag := cmd.Flags().Lookup("copy-url")
	if copyURLFlag == nil {
		t.Fatal("copy-url flag not found")
	}
	if copyURLFlag.DefValue != "false" {
		t.Errorf("copy-url default = %q, want false", copyURLFlag.DefValue)
	}
}

func TestNewCommand_RequiresArg(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should require exactly 1 argument
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}

	err = cmd.Args(cmd, []string{"device1", "device2"})
	if err == nil {
		t.Error("Expected error when multiple args provided")
	}
}

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for device completion")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(*cobra.Command) bool
		wantOK    bool
		errMsg    string
	}{
		{
			name:      "has use",
			checkFunc: func(c *cobra.Command) bool { return c.Use != "" },
			wantOK:    true,
			errMsg:    "Use should not be empty",
		},
		{
			name:      "has short",
			checkFunc: func(c *cobra.Command) bool { return c.Short != "" },
			wantOK:    true,
			errMsg:    "Short should not be empty",
		},
		{
			name:      "has long",
			checkFunc: func(c *cobra.Command) bool { return c.Long != "" },
			wantOK:    true,
			errMsg:    "Long should not be empty",
		},
		{
			name:      "has example",
			checkFunc: func(c *cobra.Command) bool { return c.Example != "" },
			wantOK:    true,
			errMsg:    "Example should not be empty",
		},
		{
			name:      "has aliases",
			checkFunc: func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			wantOK:    true,
			errMsg:    "Aliases should not be empty",
		},
		{
			name:      "has RunE",
			checkFunc: func(c *cobra.Command) bool { return c.RunE != nil },
			wantOK:    true,
			errMsg:    "RunE should be set",
		},
		{
			name:      "uses ExactArgs(1)",
			checkFunc: func(c *cobra.Command) bool { return c.Args != nil },
			wantOK:    true,
			errMsg:    "Args should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			if tt.checkFunc(cmd) != tt.wantOK {
				t.Error(tt.errMsg)
			}
		})
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "copy-url flag",
			args:    []string{"--copy-url"},
			wantErr: false,
		},
		{
			name:    "no flags",
			args:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	if opts.CopyURL {
		t.Error("Default CopyURL should be false")
	}
}

func TestRun_Success(t *testing.T) {
	t.Parallel()

	tf, mb := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"192.168.1.100"})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify browser was called
	if !mb.BrowseCalled {
		t.Error("Expected browser.Browse to be called")
	}

	if mb.LastURL != testURL {
		t.Errorf("Browser opened URL = %q, want %q", mb.LastURL, testURL)
	}

	// Verify output
	output := tf.OutString()
	if !strings.Contains(output, "Opening") {
		t.Errorf("Expected 'Opening' in output, got: %q", output)
	}
	if !strings.Contains(output, "192.168.1.100") {
		t.Errorf("Expected IP address in output, got: %q", output)
	}
}

func TestRun_WithNamedDevice(t *testing.T) {
	t.Parallel()

	tf, mb := factory.NewTestFactoryWithMockBrowser(t)

	// Add a device to the config
	tf.Config.Devices["living-room"] = model.Device{
		Address: "192.168.1.50",
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"living-room"})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify browser was called with resolved address
	if !mb.BrowseCalled {
		t.Error("Expected browser.Browse to be called")
	}

	expectedURL := "http://192.168.1.50"
	if mb.LastURL != expectedURL {
		t.Errorf("Browser opened URL = %q, want %q", mb.LastURL, expectedURL)
	}
}

func TestRun_CopyURL_Success(t *testing.T) {
	t.Parallel()

	tf, mb := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"192.168.1.100", "--copy-url"})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify clipboard was called instead of browser
	if mb.BrowseCalled {
		t.Error("Browser.Browse should not be called when --copy-url is used")
	}
	if !mb.CopyToClipboardCalled {
		t.Error("Expected CopyToClipboard to be called")
	}

	if mb.LastURL != testURL {
		t.Errorf("Clipboard URL = %q, want %q", mb.LastURL, testURL)
	}

	// Verify success output
	output := tf.OutString()
	if !strings.Contains(output, "URL copied to clipboard") {
		t.Errorf("Expected 'URL copied to clipboard' in output, got: %q", output)
	}
}

func TestRun_CopyURL_Error(t *testing.T) {
	t.Parallel()

	tf, mb := factory.NewTestFactoryWithMockBrowser(t)
	mb.Err = errors.New("clipboard unavailable")

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"192.168.1.100", "--copy-url"})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error when clipboard fails")
	}

	if !strings.Contains(err.Error(), "failed to copy URL to clipboard") {
		t.Errorf("Expected 'failed to copy URL to clipboard' error, got: %v", err)
	}
}

func TestRun_BrowserError(t *testing.T) {
	t.Parallel()

	tf, mb := factory.NewTestFactoryWithMockBrowser(t)
	mb.Err = errors.New("browser not found")

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"192.168.1.100"})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error when browser fails")
	}

	if !strings.Contains(err.Error(), "failed to open browser") {
		t.Errorf("Expected 'failed to open browser' error, got: %v", err)
	}
}

func TestRun_ClipboardFallback(t *testing.T) {
	t.Parallel()

	tf, mb := factory.NewTestFactoryWithMockBrowser(t)
	// Return a ClipboardFallbackError to simulate browser failure with clipboard fallback
	mb.Err = &browser.ClipboardFallbackError{URL: testURL}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"192.168.1.100"})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v, expected nil (fallback should succeed)", err)
	}

	// Verify warning output about fallback (warnings go to stderr)
	errOutput := tf.ErrString()
	if !strings.Contains(errOutput, "Could not open browser") || !strings.Contains(errOutput, "clipboard") {
		t.Errorf("Expected clipboard fallback warning in stderr, got: %q", errOutput)
	}
}

func TestRun_AcceptsIPAddress(t *testing.T) {
	t.Parallel()

	tf, mb := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"192.168.1.100"})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if mb.LastURL != "http://192.168.1.100" {
		t.Errorf("Expected URL http://192.168.1.100, got: %q", mb.LastURL)
	}
}

func TestRun_AcceptsHostname(t *testing.T) {
	t.Parallel()

	tf, mb := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"shelly-device.local"})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if mb.LastURL != "http://shelly-device.local" {
		t.Errorf("Expected URL http://shelly-device.local, got: %q", mb.LastURL)
	}
}

func TestRun_AcceptsIPv4WithPort(t *testing.T) {
	t.Parallel()

	tf, mb := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"192.168.1.100:8080"})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if mb.LastURL != "http://192.168.1.100:8080" {
		t.Errorf("Expected URL http://192.168.1.100:8080, got: %q", mb.LastURL)
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf, _ := factory.NewTestFactoryWithMockBrowser(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"192.168.1.100"})

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// The command should still execute because the browser.OpenDeviceUI
	// may not honor the cancelled context (mock doesn't)
	// This test verifies the command doesn't panic with a cancelled context
	err := cmd.Execute()
	// The mock browser doesn't check context, so Execute should succeed
	if err != nil {
		t.Logf("Execute returned error (expected with cancelled context): %v", err)
	}
}

func TestNewCommand_AcceptsDeviceName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command accepts named devices
	err := cmd.Args(cmd, []string{"living-room"})
	if err != nil {
		t.Errorf("Command should accept device name, got error: %v", err)
	}
}

func TestNewCommand_RejectsMultipleArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the command rejects multiple devices
	err := cmd.Args(cmd, []string{"device1", "device2"})
	if err == nil {
		t.Error("Command should reject multiple device arguments")
	}
}

func TestRun_URLFormation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		device      string
		expectedURL string
	}{
		{
			name:        "ipv4",
			device:      "192.168.1.1",
			expectedURL: "http://192.168.1.1",
		},
		{
			name:        "ipv4_with_port",
			device:      "192.168.1.1:80",
			expectedURL: "http://192.168.1.1:80",
		},
		{
			name:        "localhost",
			device:      "localhost",
			expectedURL: "http://localhost",
		},
		{
			name:        "localhost_with_port",
			device:      "localhost:8080",
			expectedURL: "http://localhost:8080",
		},
		{
			name:        "mdns_hostname",
			device:      "shelly-device.local",
			expectedURL: "http://shelly-device.local",
		},
		{
			name:        "private_10_network",
			device:      "10.0.0.100",
			expectedURL: "http://10.0.0.100",
		},
		{
			name:        "private_172_network",
			device:      "172.16.0.1",
			expectedURL: "http://172.16.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tf, mb := factory.NewTestFactoryWithMockBrowser(t)

			cmd := NewCommand(tf.Factory)
			cmd.SetArgs([]string{tt.device})
			cmd.SetContext(t.Context())

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if mb.LastURL != tt.expectedURL {
				t.Errorf("URL = %q, want %q", mb.LastURL, tt.expectedURL)
			}
		})
	}
}

func TestRun_WithWebAlias(t *testing.T) {
	t.Parallel()

	tf, mb := factory.NewTestFactoryWithMockBrowser(t)

	uiCmd := NewCommand(tf.Factory)

	// The command structure should support using 'web' alias
	// This verifies the alias is properly configured
	foundAlias := false
	for _, alias := range uiCmd.Aliases {
		if alias == "web" {
			foundAlias = true
			break
		}
	}
	if !foundAlias {
		t.Error("Expected 'web' alias to be defined")
	}

	// Execute via normal path (aliases don't change execution, just how cmd is invoked)
	uiCmd.SetArgs([]string{"192.168.1.100"})
	uiCmd.SetContext(t.Context())

	err := uiCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !mb.BrowseCalled {
		t.Error("Browser should have been called")
	}
}

func TestRun_WithOpenAlias(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify the 'open' alias exists
	foundAlias := false
	for _, alias := range cmd.Aliases {
		if alias == "open" {
			foundAlias = true
			break
		}
	}
	if !foundAlias {
		t.Error("Expected 'open' alias to be defined")
	}
}

func TestRun_OutputMessages(t *testing.T) {
	t.Parallel()

	t.Run("opening_message", func(t *testing.T) {
		t.Parallel()

		tf, _ := factory.NewTestFactoryWithMockBrowser(t)

		cmd := NewCommand(tf.Factory)
		cmd.SetArgs([]string{"192.168.1.100"})
		cmd.SetContext(t.Context())

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := tf.OutString()
		if !strings.Contains(output, "Opening") {
			t.Errorf("Expected 'Opening' in output, got: %q", output)
		}
		if !strings.Contains(output, "http://192.168.1.100") {
			t.Errorf("Expected URL in output, got: %q", output)
		}
		if !strings.Contains(output, "browser") {
			t.Errorf("Expected 'browser' in output, got: %q", output)
		}
	})

	t.Run("copy_success_message", func(t *testing.T) {
		t.Parallel()

		tf, _ := factory.NewTestFactoryWithMockBrowser(t)

		cmd := NewCommand(tf.Factory)
		cmd.SetArgs([]string{"192.168.1.100", "--copy-url"})
		cmd.SetContext(t.Context())

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := tf.OutString()
		if !strings.Contains(output, "URL copied to clipboard") {
			t.Errorf("Expected 'URL copied to clipboard' in output, got: %q", output)
		}
	})

	t.Run("fallback_warning_message", func(t *testing.T) {
		t.Parallel()

		tf, mb := factory.NewTestFactoryWithMockBrowser(t)
		mb.Err = &browser.ClipboardFallbackError{URL: testURL}

		cmd := NewCommand(tf.Factory)
		cmd.SetArgs([]string{"192.168.1.100"})
		cmd.SetContext(t.Context())

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		// Warning goes to stderr
		errOutput := tf.ErrString()
		if !strings.Contains(errOutput, "Could not open browser") {
			t.Errorf("Expected 'Could not open browser' in stderr, got: %q", errOutput)
		}
	})
}

func TestRun_ResolveDeviceAddress(t *testing.T) {
	t.Parallel()

	tf, mb := factory.NewTestFactoryWithMockBrowser(t)

	// Add multiple devices to config
	tf.Config.Devices["kitchen"] = model.Device{Address: "192.168.1.10"}
	tf.Config.Devices["bedroom"] = model.Device{Address: "192.168.1.20"}

	// Test with kitchen device
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"kitchen"})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if mb.LastURL != "http://192.168.1.10" {
		t.Errorf("Expected resolved address for kitchen, got: %q", mb.LastURL)
	}
}

func TestRun_UnknownDeviceUsedAsAddress(t *testing.T) {
	t.Parallel()

	tf, mb := factory.NewTestFactoryWithMockBrowser(t)

	// Unknown device name should be used as-is (treated as address/hostname)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"unknown-device.local"})
	cmd.SetContext(t.Context())

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if mb.LastURL != "http://unknown-device.local" {
		t.Errorf("Expected unknown device to be used as address, got: %q", mb.LastURL)
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Parse with no flags to get defaults
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	copyURLFlag := cmd.Flags().Lookup("copy-url")
	if copyURLFlag.DefValue != "false" {
		t.Errorf("copy-url default = %q, want false", copyURLFlag.DefValue)
	}
}
