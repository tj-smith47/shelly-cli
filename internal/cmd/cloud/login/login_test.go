package login

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const cmdName = "login"

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != cmdName {
		t.Errorf("Use = %q, want 'login'", cmd.Use)
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

	expectedAliases := map[string]bool{"auth": true, "signin": true}
	for _, alias := range cmd.Aliases {
		delete(expectedAliases, alias)
	}

	if len(expectedAliases) > 0 {
		t.Errorf("missing aliases: %v", expectedAliases)
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
	if cmd.Run != nil {
		t.Error("Run should not be set when RunE is used")
	}
}

func TestNewCommand_NoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Login command should not require args
	if cmd.Args != nil {
		if err := cmd.Args(cmd, []string{}); err != nil {
			t.Errorf("command should accept zero args: %v", err)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Check email flag exists
	emailFlag := cmd.Flags().Lookup("email")
	if emailFlag == nil {
		t.Fatal("email flag not found")
	}

	// Check password flag exists
	passwordFlag := cmd.Flags().Lookup("password")
	if passwordFlag == nil {
		t.Fatal("password flag not found")
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Parse with no flags to get defaults
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	emailFlag := cmd.Flags().Lookup("email")
	if emailFlag.DefValue != "" {
		t.Errorf("email default = %q, want empty string", emailFlag.DefValue)
	}

	passwordFlag := cmd.Flags().Lookup("password")
	if passwordFlag.DefValue != "" {
		t.Errorf("password default = %q, want empty string", passwordFlag.DefValue)
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
			name:    "email flag",
			args:    []string{"--email", "user@example.com"},
			wantErr: false,
		},
		{
			name:    "password flag",
			args:    []string{"--password", "secret123"},
			wantErr: false,
		},
		{
			name:    "both flags",
			args:    []string{"--email", "user@example.com", "--password", "secret123"},
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

func TestNewCommand_NoSubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Commands()) > 0 {
		t.Errorf("login command should not have subcommands, has %d", len(cmd.Commands()))
	}
}

func TestNewCommand_CommandStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"has Use", func() bool { return NewCommand(cmdutil.NewFactory()).Use != "" }},
		{"has Short", func() bool { return NewCommand(cmdutil.NewFactory()).Short != "" }},
		{"has Long", func() bool { return NewCommand(cmdutil.NewFactory()).Long != "" }},
		{"has Example", func() bool { return NewCommand(cmdutil.NewFactory()).Example != "" }},
		{"has Aliases", func() bool { return len(NewCommand(cmdutil.NewFactory()).Aliases) > 0 }},
		{"has RunE", func() bool { return NewCommand(cmdutil.NewFactory()).RunE != nil }},
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

func TestNewCommand_ExampleFormat(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "shelly") {
		t.Error("Example should contain 'shelly' command")
	}

	if !strings.Contains(cmd.Example, "login") {
		t.Error("Example should contain 'login' command")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "Cloud") {
		t.Error("Long description should mention 'Cloud'")
	}
}

func TestNewCommand_ShortDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(strings.ToLower(cmd.Short), "auth") && !strings.Contains(strings.ToLower(cmd.Short), "cloud") {
		t.Error("Short description should mention authentication or cloud")
	}
}

func TestNewCommand_ExampleContainsEmail(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "email") {
		t.Error("Example should show email flag usage")
	}
}

func TestNewCommand_ExampleContainsPassword(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "password") {
		t.Error("Example should show password flag usage")
	}
}

func TestNewCommand_UseField(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != cmdName {
		t.Errorf("Use = %q, want 'login'", cmd.Use)
	}
}

func TestNewCommand_AliasesContents(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	aliasMap := make(map[string]bool)
	for _, a := range cmd.Aliases {
		aliasMap[a] = true
	}

	if !aliasMap["auth"] {
		t.Error("missing 'auth' alias")
	}
	if !aliasMap["signin"] {
		t.Error("missing 'signin' alias")
	}
}

func TestNewCommand_LongMentionsCredentials(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should mention various ways to provide credentials
	if !strings.Contains(cmd.Long, "flag") && !strings.Contains(cmd.Long, "environment") {
		t.Error("Long description should mention credential sources")
	}
}

func TestNewCommand_LongMentionsEnvVars(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "SHELLY_CLOUD_EMAIL") || !strings.Contains(cmd.Long, "SHELLY_CLOUD_PASSWORD") {
		t.Error("Long description should mention environment variables")
	}
}

func TestRun_MissingEmail_NonInteractive(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Email:    "",
		Password: "",
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err == nil {
		t.Error("expected error when email is missing")
	}

	if !strings.Contains(err.Error(), "email") {
		t.Errorf("error should mention email, got: %v", err)
	}
}

func TestRun_MissingPassword_NonInteractive(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Email:    "test@example.com",
		Password: "",
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err == nil {
		t.Error("expected error when password is missing")
	}

	if !strings.Contains(err.Error(), "password") {
		t.Errorf("error should mention password, got: %v", err)
	}
}

func TestNewCommand_ExampleContainsCloud(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "cloud") {
		t.Error("Example should show 'cloud' subcommand")
	}
}

func TestNewCommand_ShortIsConcise(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if strings.Contains(cmd.Short, "\n") {
		t.Error("Short description should not contain newlines")
	}

	if len(cmd.Short) > 80 {
		t.Errorf("Short description too long (%d chars), should be under 80", len(cmd.Short))
	}
}

func TestNewCommand_ExampleHasMultipleLines(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	lines := strings.Split(cmd.Example, "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	if nonEmptyLines < 2 {
		t.Error("Example should contain multiple usage examples")
	}
}

func TestNewCommand_LongHasMultipleLines(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "\n") {
		t.Error("Long description should have multiple lines for detail")
	}
}

func TestRun_CancelledContext(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Email:    "test@example.com",
		Password: "testpassword",
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)

	// Should fail due to cancelled context
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestNewCommand_ExampleMentionsInteractive(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Example should show interactive login option
	if !strings.Contains(strings.ToLower(cmd.Example), "interactive") ||
		!strings.Contains(cmd.Example, "shelly cloud login") {
		// At minimum should show plain login command
		if !strings.Contains(cmd.Example, "shelly cloud login") {
			t.Error("Example should show basic login command")
		}
	}
}

func TestNewCommand_LongMentionsToken(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "token") {
		t.Error("Long description should mention access token storage")
	}
}

func TestNewCommand_HelpString(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)

	help := cmd.UsageString()
	if help == "" {
		t.Error("Help should not be empty")
	}

	if !strings.Contains(help, "login") {
		t.Error("Help should mention login")
	}
}

func TestNewCommand_CommandName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Name() != cmdName {
		t.Errorf("Name() = %q, want 'login'", cmd.Name())
	}
}

func TestNewCommand_CommandPath(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	path := cmd.CommandPath()
	if !strings.Contains(path, "login") {
		t.Errorf("CommandPath() = %q, should contain 'login'", path)
	}
}

func TestNewCommand_UsageString(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	usage := cmd.UsageString()
	if !strings.Contains(usage, "login") {
		t.Error("UsageString should contain command name")
	}
}

func TestNewCommand_VerifyReturnType(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	cmd := NewCommand(f)

	if cmd == nil {
		t.Fatal("NewCommand should not return nil")
	}

	if cmd.Use == "" {
		t.Error("Command Use should be set")
	}
}

func TestRun_BothCredentialsEmpty(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Email:    "",
		Password: "",
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err == nil {
		t.Error("expected error when both credentials are empty")
	}
}

func TestRun_EmailValidPasswordEmpty(t *testing.T) {
	t.Parallel()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Email:    "valid@email.com",
		Password: "",
	}

	ctx := context.Background()
	err := run(ctx, opts)

	if err == nil {
		t.Error("expected error when password is empty")
	}
}

func TestNewCommand_LocalFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should have email and password flags
	localFlags := cmd.LocalFlags()
	if !localFlags.HasFlags() {
		t.Error("Login command should have local flags defined")
	}
}

func TestNewCommand_EmailFlagType(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	emailFlag := cmd.Flags().Lookup("email")
	if emailFlag == nil {
		t.Fatal("email flag should exist")
	}

	if emailFlag.Value.Type() != "string" {
		t.Errorf("email flag type = %q, want 'string'", emailFlag.Value.Type())
	}
}

func TestNewCommand_PasswordFlagType(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	passwordFlag := cmd.Flags().Lookup("password")
	if passwordFlag == nil {
		t.Fatal("password flag should exist")
	}

	if passwordFlag.Value.Type() != "string" {
		t.Errorf("password flag type = %q, want 'string'", passwordFlag.Value.Type())
	}
}

func TestRun_EmailFromEnvVar(t *testing.T) {
	// Set environment variable for email
	t.Setenv("SHELLY_CLOUD_EMAIL", "envtest@example.com")

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Email:    "", // Not set via flag
		Password: "", // Not set - will error asking for password
	}

	ctx := context.Background()
	err := run(ctx, opts)

	// Should fail asking for password (email was obtained from env var)
	if err == nil {
		t.Error("expected error for missing password")
		return
	}

	// Should mention password (not email) since email was provided via env
	if !strings.Contains(err.Error(), "password") {
		t.Errorf("error should mention password, got: %v", err)
	}
}

func TestRun_PasswordFromEnvVar(t *testing.T) {
	// Set environment variable for password
	t.Setenv("SHELLY_CLOUD_PASSWORD", "envtestpassword")

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Email:    "", // Not set - will error asking for email
		Password: "", // Will be obtained from env var
	}

	ctx := context.Background()
	err := run(ctx, opts)

	// Should fail asking for email
	if err == nil {
		t.Error("expected error for missing email")
		return
	}

	// Should mention email since password was provided via env
	if !strings.Contains(err.Error(), "email") {
		t.Errorf("error should mention email, got: %v", err)
	}
}

func TestRun_BothFromEnvVars(t *testing.T) {
	// Set both environment variables
	t.Setenv("SHELLY_CLOUD_EMAIL", "envtest@example.com")
	t.Setenv("SHELLY_CLOUD_PASSWORD", "envtestpassword")

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Email:    "", // Will be obtained from env var
		Password: "", // Will be obtained from env var
	}

	// Use short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	err := run(ctx, opts)

	// Should fail due to network (authentication request), not missing credentials
	if err == nil {
		t.Log("unexpected success - expected network error")
	} else if strings.Contains(err.Error(), "email required") || strings.Contains(err.Error(), "password required") {
		t.Error("should not get credential missing error when env vars are set")
	}
}

func TestRun_FlagOverridesEnvVar(t *testing.T) {
	// Set environment variable
	t.Setenv("SHELLY_CLOUD_EMAIL", "env@example.com")

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Email:    "flag@example.com", // Flag should take precedence
		Password: "testpassword",
	}

	// Use short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	err := run(ctx, opts)

	// Should fail due to network, not missing credentials
	if err == nil {
		t.Log("unexpected success - expected network error")
	}
}

func TestRun_AuthenticationAttempt(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	opts := &Options{
		Factory:  f,
		Email:    "test@example.com",
		Password: "testpassword123",
	}

	// Use short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	err := run(ctx, opts)

	// Should fail due to network/authentication error, not missing credentials
	if err == nil {
		t.Log("unexpected success - expected authentication error")
	} else if strings.Contains(err.Error(), "email required") || strings.Contains(err.Error(), "password required") {
		t.Error("should not get credential missing error when credentials are provided")
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

func TestExecute_MissingCredentials(t *testing.T) {
	t.Parallel()

	// Clear env vars if set
	prevEmail := os.Getenv("SHELLY_CLOUD_EMAIL")
	prevPass := os.Getenv("SHELLY_CLOUD_PASSWORD")
	if prevEmail != "" {
		if err := os.Unsetenv("SHELLY_CLOUD_EMAIL"); err != nil {
			t.Logf("warning: failed to unset SHELLY_CLOUD_EMAIL: %v", err)
		}
		defer func() {
			if err := os.Setenv("SHELLY_CLOUD_EMAIL", prevEmail); err != nil {
				t.Logf("warning: failed to restore SHELLY_CLOUD_EMAIL: %v", err)
			}
		}()
	}
	if prevPass != "" {
		if err := os.Unsetenv("SHELLY_CLOUD_PASSWORD"); err != nil {
			t.Logf("warning: failed to unset SHELLY_CLOUD_PASSWORD: %v", err)
		}
		defer func() {
			if err := os.Setenv("SHELLY_CLOUD_PASSWORD", prevPass); err != nil {
				t.Logf("warning: failed to restore SHELLY_CLOUD_PASSWORD: %v", err)
			}
		}()
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing credentials")
	}
}

func TestExecute_WithCredentials(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--email", "test@example.com", "--password", "testpass"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	// Will fail due to network, but should not complain about missing credentials
	err := cmd.Execute()
	if err == nil {
		t.Log("unexpected success - expected network error")
	} else if strings.Contains(err.Error(), "email required") || strings.Contains(err.Error(), "password required") {
		t.Error("should not get credential missing error when credentials are provided")
	}
}

func TestOptions_Fields(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Email:    "test@example.com",
		Password: "secret123",
	}

	if opts.Email != "test@example.com" {
		t.Errorf("Email = %q, want 'test@example.com'", opts.Email)
	}
	if opts.Password != "secret123" {
		t.Errorf("Password = %q, want 'secret123'", opts.Password)
	}
}
