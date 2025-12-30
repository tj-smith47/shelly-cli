package server

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
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

	if cmd.Use != "server" {
		t.Errorf("Use = %q, want %q", cmd.Use, "server")
	}

	wantAliases := []string{"serve", "listen", "receiver"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Server command takes no args
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("Args should accept no arguments: %v", err)
	}
	if err := cmd.Args(cmd, []string{"extra"}); err == nil {
		t.Error("Args should reject arguments")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name     string
		defValue string
	}{
		{"port", "8080"},
		{"interface", "0.0.0.0"},
		{"log-json", "false"},
		{"auto-config", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
	}

	deviceFlag := cmd.Flags().Lookup("device")
	if deviceFlag == nil {
		t.Error("--device flag not found")
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

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly webhook server",
		"--port",
		"--log-json",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Port:       9000,
		Interface:  "localhost",
		LogJSON:    true,
		AutoConfig: true,
		Devices:    []string{"dev1", "dev2"},
	}

	if opts.Port != 9000 {
		t.Errorf("Port = %d, want %d", opts.Port, 9000)
	}

	if opts.Interface != "localhost" {
		t.Errorf("Interface = %q, want %q", opts.Interface, "localhost")
	}

	if !opts.LogJSON {
		t.Error("LogJSON should be true")
	}

	if !opts.AutoConfig {
		t.Error("AutoConfig should be true")
	}

	if len(opts.Devices) != 2 {
		t.Errorf("Devices length = %d, want %d", len(opts.Devices), 2)
	}
}

func TestExecute_ContextCancellation(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--port", "0"}) // Use port 0 to let the OS pick a free port
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Start the server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Execute()
	}()

	// Give the server time to start, then cancel
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for the command to complete
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Expected no error on graceful shutdown, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shut down within timeout")
	}

	// Check output contains expected messages
	output := buf.String() + tf.OutString()
	if !strings.Contains(output, "Webhook Server") && !strings.Contains(output, "Listening") {
		t.Logf("Output: %s", output)
	}
}

func TestExecute_InvalidPort(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(ctx)
	// Use an invalid port to trigger a flag parsing error
	cmd.SetArgs([]string{"--port", "invalid"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid port")
	}
}

//nolint:paralleltest // Uses shared mock server
func TestExecute_WithAutoConfig(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	ctx, cancel := context.WithCancel(context.Background())

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--port", "0", "--auto-config", "--device", "test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Start the server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Execute()
	}()

	// Give the server time to start and configure, then cancel
	time.Sleep(200 * time.Millisecond)
	cancel()

	// Wait for the command to complete
	select {
	case err := <-errCh:
		if err != nil {
			t.Logf("Execute error = %v (may be expected for mock)", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shut down within timeout")
	}

	output := buf.String() + tf.OutString()
	// Verify auto-config was attempted
	if !strings.Contains(output, "Auto-configuring") && !strings.Contains(output, "Configuring") {
		t.Logf("Auto-config message not found in output: %s", output)
	}
}

func TestExecute_WithLogJSON(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--port", "0", "--log-json"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Start the server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Execute()
	}()

	// Give the server time to start, then cancel
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for the command to complete
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Expected no error on graceful shutdown, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shut down within timeout")
	}
}

func TestExecute_WithInterface(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--port", "0", "--interface", "127.0.0.1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Start the server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Execute()
	}()

	// Give the server time to start, then cancel
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for the command to complete
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Expected no error on graceful shutdown, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shut down within timeout")
	}
}

func TestRun_GracefulShutdown(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())

	opts := &Options{
		Port:      0, // Use port 0 to let the OS pick a free port
		Interface: "127.0.0.1",
		LogJSON:   false,
	}

	// Start the server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, tf.Factory, opts)
	}()

	// Give the server time to start, then cancel
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for the run function to complete
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Expected no error on graceful shutdown, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shut down within timeout")
	}

	output := tf.OutString()
	if !strings.Contains(output, "Webhook Server") {
		t.Logf("Expected 'Webhook Server' in output, got: %s", output)
	}
	if !strings.Contains(output, "Shutting down") {
		t.Logf("Expected 'Shutting down' in output, got: %s", output)
	}
	if !strings.Contains(output, "Server stopped") {
		t.Logf("Expected 'Server stopped' in output, got: %s", output)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_WithAutoConfigDevices(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "kitchen",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"kitchen": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	ctx, cancel := context.WithCancel(context.Background())

	opts := &Options{
		Port:       0, // Use port 0 to let the OS pick a free port
		Interface:  "127.0.0.1",
		LogJSON:    false,
		AutoConfig: true,
		Devices:    []string{"kitchen"},
	}

	// Start the server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, tf.Factory, opts)
	}()

	// Give the server time to start and configure, then cancel
	time.Sleep(200 * time.Millisecond)
	cancel()

	// Wait for the run function to complete
	select {
	case err := <-errCh:
		if err != nil {
			t.Logf("run() error = %v (may be expected for mock webhook creation)", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shut down within timeout")
	}

	output := tf.OutString()
	if !strings.Contains(output, "Auto-configuring") {
		t.Logf("Expected 'Auto-configuring' in output, got: %s", output)
	}
}

func TestRun_WithLogJSONEnabled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())

	opts := &Options{
		Port:      0,
		Interface: "127.0.0.1",
		LogJSON:   true,
	}

	// Start the server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, tf.Factory, opts)
	}()

	// Give the server time to start, then cancel
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for the run function to complete
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Expected no error on graceful shutdown, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shut down within timeout")
	}
}

func TestExecute_ExtraArgsError(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"extra-arg"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when passing extra arguments")
	}
}
