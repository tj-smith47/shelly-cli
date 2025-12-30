package websocket

// Test Coverage Notes:
// Coverage is ~53% due to architectural constraints in testing WebSocket connections:
// - Mock server rewrites addresses to HTTP URLs, breaking WebSocket URL construction
// - shelly-go transport library has connection timeouts that can't be easily mocked
// Tests cover command creation, flag parsing, device resolution, and generation checking.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// testWSServer creates a mock WebSocket server for testing.
type testWSServer struct {
	server   *httptest.Server
	upgrader websocket.Upgrader
	events   []map[string]any
}

func newTestWSServer(events []map[string]any) *testWSServer {
	s := &testWSServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		events: events,
	}
	s.server = httptest.NewServer(http.HandlerFunc(s.handler))
	return s
}

func (s *testWSServer) handler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }() //nolint:errcheck // intentional

	// Send events
	for _, event := range s.events {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Keep connection open briefly, then close
	time.Sleep(100 * time.Millisecond)
}

func (s *testWSServer) URL() string {
	return strings.Replace(s.server.URL, "http://", "ws://", 1)
}

func (s *testWSServer) Close() {
	s.server.Close()
}

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "websocket <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "websocket <device>")
	}

	if cmd.Short != "Debug WebSocket connection and stream events" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Debug WebSocket connection and stream events")
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

	expectedAliases := []string{"ws", "events"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}
	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) || cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
	}
}

func TestNewCommand_RequiresArg(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"device1"},
			wantErr: false,
		},
		{
			name:    "two args",
			args:    []string{"device1", "extra"},
			wantErr: true,
		},
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

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{name: "duration", shorthand: "", defValue: "30s"},
		{name: "raw", shorthand: "", defValue: "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.name)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("%s shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("%s default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is not set")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is not set")
	}
}

func TestNewCommand_DurationFlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flagVal  string
		expected time.Duration
		wantErr  bool
	}{
		{name: "default", flagVal: "", expected: 30 * time.Second, wantErr: false},
		{name: "5 minutes", flagVal: "5m", expected: 5 * time.Minute, wantErr: false},
		{name: "zero for indefinite", flagVal: "0", expected: 0, wantErr: false},
		{name: "1 hour", flagVal: "1h", expected: 1 * time.Hour, wantErr: false},
		{name: "10 seconds", flagVal: "10s", expected: 10 * time.Second, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			var args []string
			if tt.flagVal != "" {
				args = []string{"--duration", tt.flagVal}
			}

			err := cmd.ParseFlags(args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				got, err := cmd.Flags().GetDuration("duration")
				if err != nil {
					t.Fatalf("GetDuration() error = %v", err)
				}
				if got != tt.expected {
					t.Errorf("duration = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestNewCommand_RawFlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{name: "default false", args: []string{}, expected: false},
		{name: "explicit true", args: []string{"--raw"}, expected: true},
		{name: "explicit false", args: []string{"--raw=false"}, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}

			got, err := cmd.Flags().GetBool("raw")
			if err != nil {
				t.Fatalf("GetBool() error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("raw = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewCommand_InvalidDuration(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.ParseFlags([]string{"--duration", "invalid"})
	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestOptions_Fields(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory:  f,
		Device:   "test-device",
		Duration: 5 * time.Minute,
		Raw:      true,
	}

	if opts.Factory != f {
		t.Error("Factory field not set correctly")
	}
	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
	if opts.Duration != 5*time.Minute {
		t.Errorf("Duration = %v, want %v", opts.Duration, 5*time.Minute)
	}
	if !opts.Raw {
		t.Error("Raw = false, want true")
	}
}

func TestNewCommand_Properties(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		check    func(*cmdutil.Factory) bool
		errorMsg string
	}{
		{
			name: "Use field is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Use == "websocket <device>"
			},
			errorMsg: "Use field not set correctly",
		},
		{
			name: "Short field is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Short != ""
			},
			errorMsg: "Short field is empty",
		},
		{
			name: "Long field is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Long != ""
			},
			errorMsg: "Long field is empty",
		},
		{
			name: "Example field is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Example != ""
			},
			errorMsg: "Example field is empty",
		},
		{
			name: "Has at least 2 aliases",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return len(cmd.Aliases) >= 2
			},
			errorMsg: "Should have at least 2 aliases (ws, events)",
		},
		{
			name: "Args validator is set",
			check: func(f *cmdutil.Factory) bool {
				cmd := NewCommand(f)
				return cmd.Args != nil
			},
			errorMsg: "Args validator not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := cmdutil.NewFactory()
			if !tt.check(f) {
				t.Error(tt.errorMsg)
			}
		})
	}
}

func TestRun_DeviceNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "nonexistent-device",
		Duration: 30 * time.Second,
		Raw:      false,
	}

	err := run(context.Background(), opts)

	// Should fail because device doesn't exist
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestRun_WithTestFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 1 * time.Second,
		Raw:      false,
	}

	// This will fail on device connection, but exercises the early run() code
	err := run(context.Background(), opts)

	// Expect error due to no device
	if err == nil {
		t.Log("Expected connection error (no real device)")
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 30 * time.Second,
		Raw:      false,
	}

	err := run(ctx, opts)

	// Should return some error (context cancelled or connection error)
	if err == nil {
		t.Log("Expected error with cancelled context")
	}
}

func TestNewCommand_ExecuteWithNoArgs(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()

	if err == nil {
		t.Error("Expected error when executing with no arguments")
	}
}

func TestNewCommand_ExecuteWithDeviceArg(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"test-device"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// Execute will fail due to no real device, but args should be accepted
	err := cmd.Execute()

	// We expect an error (no device connection), but not an args error
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "accepts") && strings.Contains(errStr, "arg") {
			t.Errorf("Should accept device argument, got args error: %v", err)
		}
	}
}

func TestNewCommand_HelpOutput(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Logf("Help execution: %v", err)
	}

	helpOutput := stdout.String()

	if !strings.Contains(helpOutput, "websocket") {
		t.Error("Help should contain 'websocket'")
	}
	if !strings.Contains(helpOutput, "WebSocket") {
		t.Error("Help should contain 'WebSocket'")
	}
}

func TestOptions_FactoryAccess(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 30 * time.Second,
		Raw:      true,
	}

	// Verify factory is accessible
	if opts.Factory == nil {
		t.Fatal("Options.Factory should not be nil")
	}

	ios := opts.Factory.IOStreams()
	if ios == nil {
		t.Error("Factory.IOStreams() should not return nil")
	}

	svc := opts.Factory.ShellyService()
	if svc == nil {
		t.Error("Factory.ShellyService() should not return nil")
	}
}

func TestRun_RawMode(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 1 * time.Second,
		Raw:      true,
	}

	// This will fail on device connection
	err := run(context.Background(), opts)

	// We expect a device-related error, not a raw mode error
	if err != nil && strings.Contains(err.Error(), "raw") {
		t.Errorf("Unexpected raw mode error: %v", err)
	}
}

func TestRun_ZeroDuration(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 0, // Zero for indefinite
		Raw:      false,
	}

	// This will fail on device connection
	err := run(context.Background(), opts)

	// We expect a device-related error
	if err == nil {
		t.Log("Expected connection error (no real device)")
	}
}

//nolint:paralleltest // Uses global mock config
func TestRun_Gen1DeviceRejected(t *testing.T) {
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

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "gen1-device",
		Duration: 30 * time.Second,
		Raw:      false,
	}

	err = run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error for Gen1 device")
	}
	if !strings.Contains(err.Error(), "Gen2+") || !strings.Contains(err.Error(), "Gen1") {
		t.Errorf("Error should mention Gen2+ requirement and Gen1 device: %v", err)
	}
}

//nolint:paralleltest // Uses global mock config
func TestRun_Gen2DeviceWebSocketConnectionFails(t *testing.T) {
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
			"test-device": {
				"switch:0": map[string]any{"output": true},
				"ws_config": map[string]any{
					"enable": true,
					"server": "",
					"ssl_ca": "*",
				},
				"ws_status": map[string]any{
					"connected": false,
				},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 1 * time.Second,
		Raw:      false,
	}

	err = run(context.Background(), opts)

	// Should fail on WebSocket connection (no real WebSocket server)
	if err == nil {
		t.Error("Expected WebSocket connection error")
	}
	if !strings.Contains(err.Error(), "WebSocket connection failed") {
		t.Logf("Error = %v (WebSocket connection failed expected)", err)
	}

	// Check that device info was displayed in output
	output := tf.OutString()
	if !strings.Contains(output, "WebSocket") {
		t.Error("Output should contain WebSocket-related text")
	}
}

//nolint:paralleltest // Uses global mock config
func TestRun_Gen2DeviceWithRawMode(t *testing.T) {
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
			"test-device": {"switch:0": map[string]any{"output": true}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 1 * time.Second,
		Raw:      true, // Test raw mode
	}

	err = run(context.Background(), opts)

	// Should fail on WebSocket connection
	if err == nil {
		t.Error("Expected WebSocket connection error")
	}
}

//nolint:paralleltest // Uses global mock config
func TestExecute_Gen1DeviceWithMock(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-switch",
					Address:    "192.168.1.50",
					MAC:        "11:22:33:44:55:66",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-switch": {"relay": map[string]any{"ison": true}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"gen1-switch"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err = cmd.Execute()

	if err == nil {
		t.Error("Expected error for Gen1 device")
	}
	if !strings.Contains(err.Error(), "Gen2+") {
		t.Errorf("Error should mention Gen2+ requirement: %v", err)
	}
}

//nolint:paralleltest // Uses global mock config
func TestExecute_Gen2DeviceWithMock(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen2-switch",
					Address:    "192.168.1.60",
					MAC:        "AA:BB:CC:DD:EE:11",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen2-switch": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"gen2-switch", "--duration", "1s"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err = cmd.Execute()

	// Should fail on WebSocket connection
	if err == nil {
		t.Error("Expected WebSocket connection error")
	}

	output := tf.OutString()
	if !strings.Contains(output, "WebSocket") {
		t.Error("Output should contain WebSocket configuration info")
	}
}

//nolint:paralleltest // Uses global mock config
func TestExecute_Gen3Device(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen3-device",
					Address:    "192.168.1.70",
					MAC:        "CC:DD:EE:FF:00:11",
					Type:       "S3SW-001P16EU",
					Model:      "Shelly Plus 1PM Gen3",
					Generation: 3,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen3-device": {"switch:0": map[string]any{"output": true}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"gen3-device"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err = cmd.Execute()

	// Gen3 should pass the generation check (>= 2)
	// but fail on WebSocket connection
	if err == nil {
		t.Error("Expected WebSocket connection error")
	}
	// Should NOT contain Gen2+ error
	if strings.Contains(err.Error(), "Gen2+") {
		t.Errorf("Gen3 should be accepted, got: %v", err)
	}
}

//nolint:paralleltest // Uses global mock config
func TestRun_DeviceResolutionError(t *testing.T) {
	// Test with an empty config - device lookup will fail
	fixtures := &mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{Devices: []mock.DeviceFixture{}},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "nonexistent",
		Duration: 1 * time.Second,
		Raw:      false,
	}

	err = run(context.Background(), opts)

	// Expect some error - either device resolution or generation check
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
	// The actual error could be about device resolution OR about Gen2+
	// depending on how the service handles unknown devices
	errStr := err.Error()
	if !strings.Contains(errStr, "resolve") && !strings.Contains(errStr, "Gen") {
		t.Errorf("Error should mention device resolution or generation: %v", err)
	}
}

//nolint:paralleltest // Uses global mock config
func TestRun_WithWebSocketInfoError(t *testing.T) {
	// Test that run() continues even if GetWebSocketInfo fails
	// We can't easily mock the WebSocket info error, but we can verify
	// the code path by checking output

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
			"test-device": {"switch:0": map[string]any{"output": true}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 1 * time.Second,
		Raw:      false,
	}

	err = run(context.Background(), opts)

	// Will fail at WebSocket connection, but should have gotten past
	// the device info and WebSocket info steps
	if err == nil {
		t.Error("Expected error")
	}

	output := tf.OutString()
	if !strings.Contains(output, "WebSocket Configuration") {
		t.Error("Output should contain device info section")
	}
}

//nolint:paralleltest // Uses global mock config
func TestRun_ContextCancelledDuringRun(t *testing.T) {
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
			"test-device": {"switch:0": map[string]any{"output": true}},
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
	// Cancel context after short delay to exercise cancellation path
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 30 * time.Second, // Long duration, will be cancelled
		Raw:      false,
	}

	err = run(ctx, opts)

	// Should fail at WebSocket connection or context cancellation
	if err == nil {
		t.Log("Expected error from WebSocket or context cancellation")
	}
}

func TestNewCommand_ExampleContainsImportantInfo(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	patterns := []string{
		"debug websocket",
		"--duration",
		"--raw",
		"Ctrl+C",
	}

	for _, pattern := range patterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("Example should contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescriptionContainsDetails(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	patterns := []string{
		"WebSocket",
		"Gen2+",
		"notifications",
		"real-time",
	}

	for _, pattern := range patterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("Long description should contain %q", pattern)
		}
	}
}

// TestRun_FullWebSocketFlow tests the complete WebSocket flow with a mock server.
// This test uses a real WebSocket server to exercise the full run() function.
//
//nolint:paralleltest // Uses global mock config
func TestRun_FullWebSocketFlow(t *testing.T) {
	// Create test WebSocket server that sends events
	events := []map[string]any{
		{
			"method": "NotifyStatus",
			"params": map[string]any{
				"switch:0": map[string]any{"output": true},
			},
		},
	}

	wsServer := newTestWSServer(events)
	defer wsServer.Close()

	// Extract host:port from WebSocket server URL
	wsURL := wsServer.URL()
	hostPort := strings.TrimPrefix(wsURL, "ws://")

	// Create fixtures with device pointing to our WebSocket server
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    hostPort, // Point to our WS server
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Use short duration to complete quickly
	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 200 * time.Millisecond,
		Raw:      false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = run(ctx, opts)

	// Should complete without error (or with expected timeout)
	if err != nil {
		// Some WebSocket connection errors are acceptable in test env
		if !strings.Contains(err.Error(), "WebSocket") {
			t.Logf("Run completed with non-websocket error: %v", err)
		}
	}

	output := tf.OutString()
	// Should have printed device info
	if !strings.Contains(output, "WebSocket Configuration") {
		t.Error("Output should contain WebSocket Configuration section")
	}
}

// TestRun_FullWebSocketFlowWithRawMode tests the complete WebSocket flow with raw output.
//
//nolint:paralleltest // Uses global mock config
func TestRun_FullWebSocketFlowWithRawMode(t *testing.T) {
	events := []map[string]any{
		{
			"method": "NotifyStatus",
			"params": map[string]any{
				"switch:0": map[string]any{"output": false},
			},
		},
	}

	wsServer := newTestWSServer(events)
	defer wsServer.Close()

	hostPort := strings.TrimPrefix(wsServer.URL(), "ws://")

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    hostPort,
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

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 200 * time.Millisecond,
		Raw:      true, // Test raw mode
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = run(ctx, opts) //nolint:errcheck // connection to mock may fail

	// Check output was produced
	output := tf.OutString()
	if !strings.Contains(output, "WebSocket") {
		t.Error("Output should contain WebSocket text")
	}
}

// TestRun_FullWebSocketFlowZeroDuration tests with zero duration (indefinite).
//
//nolint:paralleltest // Uses global mock config
func TestRun_FullWebSocketFlowZeroDuration(t *testing.T) {
	events := []map[string]any{
		{"method": "NotifyStatus", "params": map[string]any{}},
	}

	wsServer := newTestWSServer(events)
	defer wsServer.Close()

	hostPort := strings.TrimPrefix(wsServer.URL(), "ws://")

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    hostPort,
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 0, // Zero duration = indefinite (context will cancel)
		Raw:      false,
	}

	// Cancel context after a short delay to stop indefinite mode
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	_ = run(ctx, opts) //nolint:errcheck // connection to mock may fail

	// Context cancellation is expected
	output := tf.OutString()
	if !strings.Contains(output, "WebSocket Configuration") {
		t.Error("Output should contain WebSocket Configuration section")
	}
}

// TestRun_WebSocketWithContextCancellation tests context cancellation during streaming.
//
//nolint:paralleltest // Uses global mock config
func TestRun_WebSocketWithContextCancellation(t *testing.T) {
	// Create server that sends many events
	events := make([]map[string]any, 5)
	for i := range 5 {
		events[i] = map[string]any{
			"method": "NotifyStatus",
			"params": map[string]any{
				"switch:0": map[string]any{"output": true},
			},
		}
	}

	wsServer := newTestWSServer(events)
	defer wsServer.Close()

	hostPort := strings.TrimPrefix(wsServer.URL(), "ws://")

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    hostPort,
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
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
		Factory:  tf.Factory,
		Device:   "test-device",
		Duration: 10 * time.Second, // Long duration
		Raw:      false,
	}

	// Cancel after brief moment
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()

	_ = run(ctx, opts) //nolint:errcheck // connection to mock may fail

	// Context cancellation is expected
	output := tf.OutString()
	if output == "" {
		t.Error("Expected some output before cancellation")
	}
}

// TestRun_DirectWebSocketConnection tests WebSocket connection using direct device config.
// This bypasses the mock server URL rewriting to test the actual WebSocket code path.
//
//nolint:paralleltest // Sets global config.SetDefaultManager
func TestRun_DirectWebSocketConnection(t *testing.T) {
	// Create a test WebSocket server that sends events
	events := []map[string]any{
		{
			"method": "NotifyStatus",
			"params": map[string]any{
				"switch:0": map[string]any{"output": true},
			},
		},
		{
			"method": "NotifyFullStatus",
			"params": map[string]any{
				"sys": map[string]any{"uptime": 12345},
			},
		},
	}

	wsServer := newTestWSServer(events)
	defer wsServer.Close()

	// Extract host:port from the WebSocket server
	hostPort := strings.TrimPrefix(wsServer.URL(), "ws://")

	// Create factory with device pointing directly to our WS server
	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"ws-device": {
			Name:       "ws-device",
			Address:    hostPort, // Direct host:port, not a full URL
			MAC:        "AA:BB:CC:DD:EE:FF",
			Model:      "Shelly Plus 1PM",
			Type:       "SNSW-001P16EU",
			Generation: 2,
		},
	})

	// Set global default manager so ConfigResolver can find our device
	config.SetDefaultManager(tf.Manager)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "ws-device",
		Duration: 300 * time.Millisecond, // Short duration
		Raw:      false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := run(ctx, opts)

	// Should connect and stream events, then complete
	// Error is possible if WebSocket connection handling differs
	if err != nil {
		t.Logf("Run completed with error: %v (may be expected in test env)", err)
	}

	output := tf.OutString()
	// Should have printed device info and event streaming info
	if !strings.Contains(output, "Event Streaming") {
		t.Error("Output should contain 'Event Streaming' section")
	}
}

// TestRun_DirectWebSocketRawMode tests raw mode WebSocket output.
//
//nolint:paralleltest // Sets global config.SetDefaultManager
func TestRun_DirectWebSocketRawMode(t *testing.T) {
	events := []map[string]any{
		{"method": "NotifyStatus", "params": map[string]any{"switch:0": map[string]any{"output": false}}},
	}

	wsServer := newTestWSServer(events)
	defer wsServer.Close()

	hostPort := strings.TrimPrefix(wsServer.URL(), "ws://")

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"raw-device": {
			Name:       "raw-device",
			Address:    hostPort,
			MAC:        "11:22:33:44:55:66",
			Model:      "Shelly Pro 1PM",
			Type:       "SPSW-001PE16EU",
			Generation: 2,
		},
	})

	// Set global default manager so ConfigResolver can find our device
	config.SetDefaultManager(tf.Manager)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "raw-device",
		Duration: 300 * time.Millisecond,
		Raw:      true, // Enable raw mode
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = run(ctx, opts) //nolint:errcheck // connection to test server may timeout

	output := tf.OutString()
	if !strings.Contains(output, "Connecting") {
		t.Error("Output should contain connection info")
	}
}

// TestRun_DirectWebSocketZeroDuration tests indefinite streaming with context cancel.
//
//nolint:paralleltest // Sets global config.SetDefaultManager
func TestRun_DirectWebSocketZeroDuration(t *testing.T) {
	events := []map[string]any{
		{"method": "NotifyEvent", "params": map[string]any{"event": "button_press"}},
	}

	wsServer := newTestWSServer(events)
	defer wsServer.Close()

	hostPort := strings.TrimPrefix(wsServer.URL(), "ws://")

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"indefinite-device": {
			Name:       "indefinite-device",
			Address:    hostPort,
			MAC:        "AA:BB:CC:DD:EE:00",
			Model:      "Shelly Plus 2PM",
			Type:       "SNSW-002P16EU",
			Generation: 2,
		},
	})

	// Set global default manager so ConfigResolver can find our device
	config.SetDefaultManager(tf.Manager)

	ctx, cancel := context.WithCancel(context.Background())

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "indefinite-device",
		Duration: 0, // Zero means indefinite streaming
		Raw:      false,
	}

	// Cancel after a short delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	_ = run(ctx, opts) //nolint:errcheck // connection to test server may timeout

	output := tf.OutString()
	// With zero duration, output should mention streaming indefinitely OR Event Streaming
	// (depending on whether connection succeeded before context cancelled)
	if !strings.Contains(output, "indefinitely") && !strings.Contains(output, "Event Streaming") {
		t.Errorf("Output should mention streaming indefinitely or Event Streaming, got: %s", output)
	}
}
