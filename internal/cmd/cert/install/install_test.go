package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

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
	if caFlag.Value.Type() != "string" {
		t.Errorf("ca flag type = %q, want 'string'", caFlag.Value.Type())
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
	if clientCertFlag.Value.Type() != "string" {
		t.Errorf("client-cert flag type = %q, want 'string'", clientCertFlag.Value.Type())
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
	if clientKeyFlag.Value.Type() != "string" {
		t.Errorf("client-key flag type = %q, want 'string'", clientKeyFlag.Value.Type())
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

// Test Options.validate function
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

// Test Options.loadCertData function
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
	if err := os.WriteFile(caFile, caContent, 0600); err != nil {
		t.Fatalf("failed to create temp CA file: %v", err)
	}

	opts := &Options{
		CAFile: caFile,
	}
	data, err := opts.loadCertData()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(data.CAData) != string(caContent) {
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

	if err := os.WriteFile(certFile, certContent, 0600); err != nil {
		t.Fatalf("failed to create temp cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, keyContent, 0600); err != nil {
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

	if string(data.CertData) != string(certContent) {
		t.Errorf("Cert data mismatch: got %q, want %q", data.CertData, certContent)
	}
	if string(data.KeyData) != string(keyContent) {
		t.Errorf("Key data mismatch: got %q, want %q", data.KeyData, keyContent)
	}
}

func TestOptions_LoadCertData_NonexistentClientKey(t *testing.T) {
	t.Parallel()

	// Create temp cert file only
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "cert.pem")
	certContent := []byte("-----BEGIN CERTIFICATE-----\ncert\n-----END CERTIFICATE-----")

	if err := os.WriteFile(certFile, certContent, 0600); err != nil {
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
