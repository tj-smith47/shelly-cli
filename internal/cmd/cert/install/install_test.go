package install

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const testStringType = "string"

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "install <device>" {
		t.Errorf("Use = %q, want 'install <device>'", cmd.Use)
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
		t.Fatal("expected at least one alias")
	}

	expectedAliases := map[string]bool{"add": true, "set": true, "upload": true}
	for _, alias := range cmd.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("unexpected alias: %s", alias)
		}
		delete(expectedAliases, alias)
	}
	if len(expectedAliases) > 0 {
		t.Errorf("missing aliases: %v", expectedAliases)
	}
}

func TestNewCommand_RequiresOneArg(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should require exactly 1 argument
	tests := []struct {
		args      []string
		wantError bool
	}{
		{[]string{}, true},
		{[]string{"device"}, false},
		{[]string{"device", "extra"}, true},
	}

	for _, tt := range tests {
		err := cmd.Args(cmd, tt.args)
		gotError := err != nil
		if gotError != tt.wantError {
			t.Errorf("Args(%v) error = %v, want error = %v", tt.args, err, tt.wantError)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check ca flag
	caFlag := cmd.Flags().Lookup("ca")
	if caFlag == nil {
		t.Fatal("ca flag not found")
	}

	// Check client-cert flag
	clientCertFlag := cmd.Flags().Lookup("client-cert")
	if clientCertFlag == nil {
		t.Fatal("client-cert flag not found")
	}

	// Check client-key flag
	clientKeyFlag := cmd.Flags().Lookup("client-key")
	if clientKeyFlag == nil {
		t.Fatal("client-key flag not found")
	}
}

func TestNewCommand_CAFlagType(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	caFlag := cmd.Flags().Lookup("ca")
	if caFlag == nil {
		t.Fatal("ca flag not found")
	}

	// Flag type should be string
	if caFlag.Value.Type() != testStringType {
		t.Errorf("ca flag type = %q, want %q", caFlag.Value.Type(), testStringType)
	}
}

func TestNewCommand_ClientCertFlagType(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	clientCertFlag := cmd.Flags().Lookup("client-cert")
	if clientCertFlag == nil {
		t.Fatal("client-cert flag not found")
	}

	// Flag type should be string
	if clientCertFlag.Value.Type() != testStringType {
		t.Errorf("client-cert flag type = %q, want %q", clientCertFlag.Value.Type(), testStringType)
	}
}

func TestNewCommand_ClientKeyFlagType(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	clientKeyFlag := cmd.Flags().Lookup("client-key")
	if clientKeyFlag == nil {
		t.Fatal("client-key flag not found")
	}

	// Flag type should be string
	if clientKeyFlag.Value.Type() != testStringType {
		t.Errorf("client-key flag type = %q, want %q", clientKeyFlag.Value.Type(), testStringType)
	}
}

func TestNewCommand_RunESet(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_CommandStructure(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"has Use", func() bool { return cmd.Use != "" }},
		{"has Short", func() bool { return cmd.Short != "" }},
		{"has Long", func() bool { return cmd.Long != "" }},
		{"has Example", func() bool { return cmd.Example != "" }},
		{"has Aliases", func() bool { return len(cmd.Aliases) > 0 }},
		{"has RunE", func() bool { return cmd.RunE != nil }},
		{"has Args", func() bool { return cmd.Args != nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !tt.fn() {
				t.Errorf("command structure check failed: %s", tt.name)
			}
		})
	}
}

func TestNewCommand_ExampleContainsShelly(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "shelly") {
		t.Error("Example should contain 'shelly' command")
	}
}

func TestNewCommand_ExampleContainsCertInstall(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "cert") {
		t.Error("Example should contain 'cert' command")
	}
	if !strings.Contains(cmd.Example, "install") {
		t.Error("Example should contain 'install' command")
	}
}

func TestNewCommand_NoSubcommands(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Install command should not have subcommands
	if len(cmd.Commands()) > 0 {
		t.Errorf("install command should not have subcommands, has %d", len(cmd.Commands()))
	}
}

func TestNewCommand_HasRunE_NotRun(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// RunE should be set (not Run)
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
	if cmd.Run != nil {
		t.Error("Run should not be set when RunE is used")
	}
}

func TestNewCommand_AliasesContent(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	aliasMap := make(map[string]bool)
	for _, a := range cmd.Aliases {
		aliasMap[a] = true
	}

	// Verify expected aliases exist
	if !aliasMap["add"] {
		t.Error("missing 'add' alias")
	}
	if !aliasMap["set"] {
		t.Error("missing 'set' alias")
	}
	if !aliasMap["upload"] {
		t.Error("missing 'upload' alias")
	}
}

func TestNewCommand_LongMentionsTLS(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should mention TLS
	if !strings.Contains(cmd.Long, "TLS") {
		t.Error("Long description should mention TLS")
	}
}

func TestNewCommand_LongMentionsGen2(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should mention Gen2+
	if !strings.Contains(cmd.Long, "Gen2") {
		t.Error("Long description should mention Gen2+ devices")
	}
}

func TestNewCommand_LongMentionsMQTT(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should mention MQTT
	if !strings.Contains(cmd.Long, "MQTT") {
		t.Error("Long description should mention MQTT")
	}
}

func TestNewCommand_UseHasDeviceArg(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Use should show <device> argument
	if !strings.Contains(cmd.Use, "<device>") {
		t.Error("Use should show <device> argument")
	}
}

func TestNewCommand_ExampleIsValid(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Example should start with proper formatting (spaces for indentation)
	if !strings.HasPrefix(cmd.Example, "  ") {
		t.Error("Example should start with proper indentation")
	}
}

func TestNewCommand_FlagUsages(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	flags := []string{"ca", "client-cert", "client-key"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("flag %q not found", flagName)
			continue
		}
		if flag.Usage == "" {
			t.Errorf("flag %q should have usage description", flagName)
		}
	}
}

// Test Options.validate function.
func TestOptions_Validate_NoFlags(t *testing.T) {
	t.Parallel()

	opts := &Options{}
	err := opts.validate()

	if err == nil {
		t.Fatal("expected error when no flags provided")
	}

	if !strings.Contains(err.Error(), "--ca") || !strings.Contains(err.Error(), "--client-cert") {
		t.Errorf("error should mention --ca or --client-cert, got: %v", err)
	}
}

func TestOptions_Validate_CAOnly(t *testing.T) {
	t.Parallel()

	opts := &Options{
		CAFile: "/path/to/ca.pem",
	}
	err := opts.validate()

	if err != nil {
		t.Errorf("should not error with only --ca, got: %v", err)
	}
}

func TestOptions_Validate_ClientCertWithoutKey(t *testing.T) {
	t.Parallel()

	opts := &Options{
		ClientCert: "/path/to/cert.pem",
	}
	err := opts.validate()

	if err == nil {
		t.Fatal("expected error when --client-cert provided without --client-key")
	}

	if !strings.Contains(err.Error(), "--client-key") {
		t.Errorf("error should mention --client-key, got: %v", err)
	}
}

func TestOptions_Validate_ClientCertWithKey(t *testing.T) {
	t.Parallel()

	opts := &Options{
		ClientCert: "/path/to/cert.pem",
		ClientKey:  "/path/to/key.pem",
	}
	err := opts.validate()

	if err != nil {
		t.Errorf("should not error with --client-cert and --client-key, got: %v", err)
	}
}

func TestOptions_Validate_AllFlags(t *testing.T) {
	t.Parallel()

	opts := &Options{
		CAFile:     "/path/to/ca.pem",
		ClientCert: "/path/to/cert.pem",
		ClientKey:  "/path/to/key.pem",
	}
	err := opts.validate()

	if err != nil {
		t.Errorf("should not error with all flags, got: %v", err)
	}
}

// Test Options.loadCertData function.
func TestOptions_LoadCertData_NonexistentCAFile(t *testing.T) {
	t.Parallel()

	opts := &Options{
		CAFile: "/nonexistent/path/to/ca.pem",
	}
	_, err := opts.loadCertData()

	if err == nil {
		t.Fatal("expected error when CA file doesn't exist")
	}

	if !strings.Contains(err.Error(), "read CA file") {
		t.Errorf("error should mention 'read CA file', got: %v", err)
	}
}

func TestOptions_LoadCertData_NonexistentClientCert(t *testing.T) {
	t.Parallel()

	opts := &Options{
		ClientCert: "/nonexistent/path/to/cert.pem",
		ClientKey:  "/nonexistent/path/to/key.pem",
	}
	_, err := opts.loadCertData()

	if err == nil {
		t.Fatal("expected error when client cert file doesn't exist")
	}

	if !strings.Contains(err.Error(), "read client cert") {
		t.Errorf("error should mention 'read client cert', got: %v", err)
	}
}

func TestOptions_LoadCertData_ValidCAFile(t *testing.T) {
	t.Parallel()

	// Create temp CA file
	tmpDir := t.TempDir()
	caFile := filepath.Join(tmpDir, "ca.pem")
	caContent := []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----")
	if err := os.WriteFile(caFile, caContent, 0o600); err != nil {
		t.Fatalf("failed to create temp CA file: %v", err)
	}

	opts := &Options{
		CAFile: caFile,
	}
	data, err := opts.loadCertData()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(data.CAData, caContent) {
		t.Errorf("CA data mismatch: got %q, want %q", data.CAData, caContent)
	}
}

func TestOptions_LoadCertData_ValidClientCertAndKey(t *testing.T) {
	t.Parallel()

	// Create temp cert and key files
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "cert.pem")
	keyFile := filepath.Join(tmpDir, "key.pem")
	certContent := []byte("-----BEGIN CERTIFICATE-----\ncert\n-----END CERTIFICATE-----")
	keyContent := []byte("-----BEGIN PRIVATE KEY-----\nkey\n-----END PRIVATE KEY-----")

	if err := os.WriteFile(certFile, certContent, 0o600); err != nil {
		t.Fatalf("failed to create temp cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, keyContent, 0o600); err != nil {
		t.Fatalf("failed to create temp key file: %v", err)
	}

	opts := &Options{
		ClientCert: certFile,
		ClientKey:  keyFile,
	}
	data, err := opts.loadCertData()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(data.CertData, certContent) {
		t.Errorf("Cert data mismatch: got %q, want %q", data.CertData, certContent)
	}
	if !bytes.Equal(data.KeyData, keyContent) {
		t.Errorf("Key data mismatch: got %q, want %q", data.KeyData, keyContent)
	}
}

func TestOptions_LoadCertData_NonexistentClientKey(t *testing.T) {
	t.Parallel()

	// Create temp cert file only
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "cert.pem")
	certContent := []byte("-----BEGIN CERTIFICATE-----\ncert\n-----END CERTIFICATE-----")

	if err := os.WriteFile(certFile, certContent, 0o600); err != nil {
		t.Fatalf("failed to create temp cert file: %v", err)
	}

	opts := &Options{
		ClientCert: certFile,
		ClientKey:  "/nonexistent/path/to/key.pem",
	}
	_, err := opts.loadCertData()

	if err == nil {
		t.Fatal("expected error when client key file doesn't exist")
	}

	if !strings.Contains(err.Error(), "read client key") {
		t.Errorf("error should mention 'read client key', got: %v", err)
	}
}

func TestOptions_LoadCertData_NoFiles(t *testing.T) {
	t.Parallel()

	opts := &Options{}
	data, err := opts.loadCertData()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return empty data when no files specified
	if len(data.CAData) != 0 {
		t.Errorf("expected empty CA data, got %d bytes", len(data.CAData))
	}
	if len(data.CertData) != 0 {
		t.Errorf("expected empty cert data, got %d bytes", len(data.CertData))
	}
	if len(data.KeyData) != 0 {
		t.Errorf("expected empty key data, got %d bytes", len(data.KeyData))
	}
}

// Execute-based tests for the run function.

func TestExecute_ValidationError_NoFlags(t *testing.T) {
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected validation error when no flags provided")
	}

	if !strings.Contains(err.Error(), "--ca") && !strings.Contains(err.Error(), "--client-cert") {
		t.Errorf("error should mention --ca or --client-cert, got: %v", err)
	}
}

func TestExecute_ValidationError_ClientCertWithoutKey(t *testing.T) {
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "cert.pem")
	if err := os.WriteFile(certFile, []byte("test cert"), 0o600); err != nil {
		t.Fatalf("failed to create cert file: %v", err)
	}

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "--client-cert", certFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected validation error when --client-cert provided without --client-key")
	}

	if !strings.Contains(err.Error(), "--client-key") {
		t.Errorf("error should mention --client-key, got: %v", err)
	}
}

func TestExecute_LoadCertDataError_NonexistentCAFile(t *testing.T) {
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "--ca", "/nonexistent/path/to/ca.pem"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when CA file doesn't exist")
	}

	if !strings.Contains(err.Error(), "read CA file") {
		t.Errorf("error should mention 'read CA file', got: %v", err)
	}
}

func TestExecute_Gen1DeviceError(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {"relay": map[string]any{"ison": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create temp CA file
	tmpDir := t.TempDir()
	caFile := filepath.Join(tmpDir, "ca.pem")
	if err := os.WriteFile(caFile, []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----"), 0o600); err != nil {
		t.Fatalf("failed to create CA file: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"gen1-device", "--ca", caFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for Gen1 device")
	}

	if !strings.Contains(err.Error(), "Gen2") {
		t.Errorf("error should mention Gen2+ requirement, got: %v", err)
	}
}

func TestExecute_Gen2DeviceWithCA(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen2-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen2-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create temp CA file
	tmpDir := t.TempDir()
	caFile := filepath.Join(tmpDir, "ca.pem")
	if err := os.WriteFile(caFile, []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----"), 0o600); err != nil {
		t.Fatalf("failed to create CA file: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"gen2-device", "--ca", caFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check output contains success message
	output := tf.OutString()
	if !strings.Contains(output, "Installed CA certificate") {
		t.Errorf("output should contain success message, got: %s", output)
	}
}

func TestExecute_Gen2DeviceWithClientCert(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen2-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen2-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create temp cert and key files
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "cert.pem")
	keyFile := filepath.Join(tmpDir, "key.pem")
	if err := os.WriteFile(certFile, []byte("-----BEGIN CERTIFICATE-----\ncert\n-----END CERTIFICATE-----"), 0o600); err != nil {
		t.Fatalf("failed to create cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, []byte("-----BEGIN PRIVATE KEY-----\nkey\n-----END PRIVATE KEY-----"), 0o600); err != nil {
		t.Fatalf("failed to create key file: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"gen2-device", "--client-cert", certFile, "--client-key", keyFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check output contains success message
	output := tf.OutString()
	if !strings.Contains(output, "Installed client certificate") {
		t.Errorf("output should contain success message, got: %s", output)
	}
}

func TestExecute_Gen2DeviceWithBothCerts(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen2-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen2-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create temp files
	tmpDir := t.TempDir()
	caFile := filepath.Join(tmpDir, "ca.pem")
	certFile := filepath.Join(tmpDir, "cert.pem")
	keyFile := filepath.Join(tmpDir, "key.pem")
	if err := os.WriteFile(caFile, []byte("-----BEGIN CERTIFICATE-----\nca\n-----END CERTIFICATE-----"), 0o600); err != nil {
		t.Fatalf("failed to create CA file: %v", err)
	}
	if err := os.WriteFile(certFile, []byte("-----BEGIN CERTIFICATE-----\ncert\n-----END CERTIFICATE-----"), 0o600); err != nil {
		t.Fatalf("failed to create cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, []byte("-----BEGIN PRIVATE KEY-----\nkey\n-----END PRIVATE KEY-----"), 0o600); err != nil {
		t.Fatalf("failed to create key file: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"gen2-device", "--ca", caFile, "--client-cert", certFile, "--client-key", keyFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check output contains both success messages
	output := tf.OutString()
	if !strings.Contains(output, "Installed CA certificate") {
		t.Errorf("output should contain CA success message, got: %s", output)
	}
	if !strings.Contains(output, "Installed client certificate") {
		t.Errorf("output should contain client cert success message, got: %s", output)
	}
}

func TestExecute_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "install") {
		t.Error("help output should contain 'install'")
	}
	if !strings.Contains(output, "--ca") {
		t.Error("help output should mention --ca flag")
	}
	if !strings.Contains(output, "--client-cert") {
		t.Error("help output should mention --client-cert flag")
	}
}

func TestExecute_DeviceNotFound(t *testing.T) {
	tf := factory.NewTestFactory(t)

	// Create temp CA file
	tmpDir := t.TempDir()
	caFile := filepath.Join(tmpDir, "ca.pem")
	if err := os.WriteFile(caFile, []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----"), 0o600); err != nil {
		t.Fatalf("failed to create CA file: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistent-device", "--ca", caFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when device not found")
	}
}

func TestExecute_LoadCertDataError_NonexistentClientKey(t *testing.T) {
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "cert.pem")
	if err := os.WriteFile(certFile, []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----"), 0o600); err != nil {
		t.Fatalf("failed to create cert file: %v", err)
	}

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "--client-cert", certFile, "--client-key", "/nonexistent/key.pem"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when client key file doesn't exist")
	}

	if !strings.Contains(err.Error(), "read client key") {
		t.Errorf("error should mention 'read client key', got: %v", err)
	}
}
