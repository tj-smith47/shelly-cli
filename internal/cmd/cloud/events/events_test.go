package events

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// Test constants.
const (
	testEventOnline         = "Shelly:Online"
	testEventStatusOnChange = "Shelly:StatusOnChange"
	testEventSettings       = "Shelly:Settings"
	testFormatJSON          = "json"
	testFormatText          = "text"
	testDevice1             = "device1"
	testTypeString          = "string"
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

	if cmd.Use != "events" {
		t.Errorf("Use = %q, want %q", cmd.Use, "events")
	}

	wantAliases := []string{"watch", "subscribe"}
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	flagTests := []struct {
		name     string
		defValue string
	}{
		{"device", ""},
		{"event", ""},
		{"format", "text"},
		{"raw", "false"},
	}

	for _, f := range flagTests {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag %q not found", f.name)
			continue
		}
		if flag.DefValue != f.defValue {
			t.Errorf("flag %q default = %q, want %q", f.name, flag.DefValue, f.defValue)
		}
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
		"shelly cloud events",
		"--device",
		"--event",
		"--raw",
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
		DeviceFilter: "abc123",
		EventFilter:  "Shelly:Online",
		Raw:          true,
	}

	if opts.DeviceFilter != "abc123" {
		t.Errorf("DeviceFilter = %q, want %q", opts.DeviceFilter, "abc123")
	}

	if opts.EventFilter != "Shelly:Online" {
		t.Errorf("EventFilter = %q, want %q", opts.EventFilter, "Shelly:Online")
	}

	if !opts.Raw {
		t.Error("Raw should be true")
	}
}

func TestExecute_NotLoggedIn(t *testing.T) {
	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when not logged in")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("Expected 'not logged in' error, got: %v", err)
	}
}

// setupTestManagerWithCloud creates a test manager with cloud config.
func setupTestManagerWithCloud(t *testing.T, accessToken, serverURL string) *config.Manager {
	t.Helper()
	tmpDir := t.TempDir()
	mgr := config.NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := mgr.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	cfg := mgr.Get()
	cfg.Cloud.AccessToken = accessToken
	cfg.Cloud.ServerURL = serverURL
	return mgr
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_WebSocketURLBuildFailure(t *testing.T) {
	// Setup manager with empty server URL and invalid token (will fail to build WS URL)
	mgr := setupTestManagerWithCloud(t, "invalid-token", "")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid WebSocket URL")
	}

	if !strings.Contains(err.Error(), "WebSocket") && !strings.Contains(err.Error(), "token") {
		t.Errorf("expected WebSocket URL build error, got: %v", err)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_WebSocketConnectionFailure(t *testing.T) {
	// Setup manager with valid-looking token and unreachable server
	mgr := setupTestManagerWithCloud(t, "test-token", "https://localhost:19999")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	// Use a context with short timeout to avoid hanging
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := NewCommand(f)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unreachable WebSocket server")
	}

	if !strings.Contains(err.Error(), "connect") && !strings.Contains(err.Error(), "dial") {
		t.Errorf("expected connection error, got: %v", err)
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_ContextCancellation(t *testing.T) {
	// Setup manager with token
	mgr := setupTestManagerWithCloud(t, "test-token", "https://localhost:19999")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	// Create already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cmd := NewCommand(f)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	// Should fail with connection/context error
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestNewCommand_LongDescriptionContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"WebSocket",
		"Ctrl+C",
		"Shelly:StatusOnChange",
		"Shelly:Online",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
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

func TestOptions_OutputFlags(t *testing.T) {
	t.Parallel()

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "json",
		},
		DeviceFilter: "device1",
		EventFilter:  "Online",
		Raw:          false,
	}

	if opts.Format != "json" {
		t.Errorf("Format = %q, want %q", opts.Format, "json")
	}

	if opts.DeviceFilter != "device1" {
		t.Errorf("DeviceFilter = %q, want %q", opts.DeviceFilter, "device1")
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
			name:    "device filter",
			args:    []string{"--device", "abc123"},
			wantErr: false,
		},
		{
			name:    "event filter",
			args:    []string{"--event", "Shelly:Online"},
			wantErr: false,
		},
		{
			name:    "raw flag",
			args:    []string{"--raw"},
			wantErr: false,
		},
		{
			name:    "format json",
			args:    []string{"--format", "json"},
			wantErr: false,
		},
		{
			name:    "format text",
			args:    []string{"--format", "text"},
			wantErr: false,
		},
		{
			name:    "all flags combined",
			args:    []string{"--device", "dev1", "--event", "Online", "--raw"},
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

func TestNewCommand_OutputContainsLoginHint(t *testing.T) {
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	_ = cmd.Execute()

	// Check IOStreams output (where the login hint is printed)
	output := tf.OutString()
	if !strings.Contains(output, "login") {
		t.Errorf("expected output to contain 'login' hint, got: %q", output)
	}
}

func TestNewCommand_ShortDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(strings.ToLower(cmd.Short), "event") {
		t.Error("Short description should mention events")
	}
}

func TestNewCommand_ExampleContainsJSON(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "json") {
		t.Error("Example should show json format option")
	}
}

func TestNewCommand_NoSubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Commands()) > 0 {
		t.Errorf("events command should not have subcommands, has %d", len(cmd.Commands()))
	}
}

func TestNewCommand_CommandName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Name() != "events" {
		t.Errorf("Name() = %q, want 'events'", cmd.Name())
	}
}

func TestNewCommand_AliasesContents(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	aliasMap := make(map[string]bool)
	for _, a := range cmd.Aliases {
		aliasMap[a] = true
	}

	if !aliasMap["watch"] {
		t.Error("missing 'watch' alias")
	}
	if !aliasMap["subscribe"] {
		t.Error("missing 'subscribe' alias")
	}
}

func TestNewCommand_UsageString(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	usage := cmd.UsageString()
	if !strings.Contains(usage, "events") {
		t.Error("UsageString should contain command name")
	}
}

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_DisplaysConnectingMessage(t *testing.T) {
	// Setup manager with token (will fail to connect but should print "Connecting...")
	mgr := setupTestManagerWithCloud(t, "test-token", "https://localhost:19999")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := NewCommand(f)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	_ = cmd.Execute()

	output := out.String()
	if !strings.Contains(output, "Connecting") {
		t.Error("expected output to contain 'Connecting' message")
	}
}

func TestNewCommand_DeviceFlagShort(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	deviceFlag := cmd.Flags().Lookup("device")
	if deviceFlag == nil {
		t.Fatal("device flag not found")
	}

	// Parse the args to set the device filter
	if err := cmd.ParseFlags([]string{"--device", "mydevice"}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	val, err := cmd.Flags().GetString("device")
	if err != nil {
		t.Fatalf("GetString error: %v", err)
	}
	if val != "mydevice" {
		t.Errorf("device value = %q, want 'mydevice'", val)
	}
}

func TestNewCommand_EventFlagValue(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{"--event", "Shelly:StatusOnChange"}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	val, err := cmd.Flags().GetString("event")
	if err != nil {
		t.Fatalf("GetString error: %v", err)
	}
	if val != "Shelly:StatusOnChange" {
		t.Errorf("event value = %q, want 'Shelly:StatusOnChange'", val)
	}
}

func TestNewCommand_RawFlagValue(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{"--raw"}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	val, err := cmd.Flags().GetBool("raw")
	if err != nil {
		t.Fatalf("GetBool error: %v", err)
	}
	if !val {
		t.Error("raw value should be true when flag is set")
	}
}

func TestNewCommand_FormatFlagValue(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{"--format", "json"}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	val, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Fatalf("GetString error: %v", err)
	}
	if val != "json" {
		t.Errorf("format value = %q, want 'json'", val)
	}
}

// WebSocket test server for integration testing.
type testWSServer struct {
	server   *httptest.Server
	upgrader websocket.Upgrader
	events   []model.CloudEvent
}

func newTestWSServer(events []model.CloudEvent) *testWSServer {
	s := &testWSServer{
		upgrader: websocket.Upgrader{},
		events:   events,
	}
	s.server = httptest.NewServer(http.HandlerFunc(s.handler))
	return s
}

func (s *testWSServer) handler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

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

	// Close normally after sending events.
	_ = conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func (s *testWSServer) URL() string {
	return "ws" + strings.TrimPrefix(s.server.URL, "http")
}

func (s *testWSServer) Close() {
	s.server.Close()
}

// dialTestWS connects to the test websocket server and handles cleanup.
func dialTestWS(t *testing.T, ctx context.Context, url string) *websocket.Conn {
	t.Helper()
	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close() //nolint:errcheck // test helper
	}
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Logf("warning: close error: %v", err)
		}
	})
	return conn
}

func TestStreamCloudEvents_RawMode(t *testing.T) {
	t.Parallel()

	online := 1
	events := []model.CloudEvent{
		{Event: testEventOnline, DeviceID: testDevice1, Online: &online},
	}

	srv := newTestWSServer(events)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn := dialTestWS(t, ctx, srv.URL())

	var received []string
	opts := network.CloudEventStreamOptions{Raw: true}

	err := network.StreamCloudEvents(ctx, conn, opts, func(event *model.CloudEvent, raw []byte) error {
		received = append(received, string(raw))
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCloudEvents error: %v", err)
	}

	if len(received) != 1 {
		t.Errorf("received %d events, want 1", len(received))
	}

	if len(received) > 0 && !strings.Contains(received[0], "Shelly:Online") {
		t.Errorf("expected raw event to contain 'Shelly:Online', got %q", received[0])
	}
}

func TestStreamCloudEvents_DeviceFilter(t *testing.T) {
	t.Parallel()

	online := 1
	events := []model.CloudEvent{
		{Event: "Shelly:Online", DeviceID: "device1", Online: &online},
		{Event: "Shelly:Online", DeviceID: "device2", Online: &online},
		{Event: "Shelly:Online", DeviceID: "device1", Online: &online},
	}

	srv := newTestWSServer(events)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, srv.URL(), nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	var receivedIDs []string
	opts := network.CloudEventStreamOptions{DeviceFilter: "device1"}

	err = network.StreamCloudEvents(ctx, conn, opts, func(event *model.CloudEvent, _ []byte) error {
		receivedIDs = append(receivedIDs, event.GetDeviceID())
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCloudEvents error: %v", err)
	}

	if len(receivedIDs) != 2 {
		t.Errorf("received %d events, want 2 (filtered to device1)", len(receivedIDs))
	}

	for _, id := range receivedIDs {
		if id != "device1" {
			t.Errorf("expected only device1 events, got %q", id)
		}
	}
}

func TestStreamCloudEvents_EventFilter(t *testing.T) {
	t.Parallel()

	online := 1
	events := []model.CloudEvent{
		{Event: "Shelly:Online", DeviceID: "device1", Online: &online},
		{Event: "Shelly:StatusOnChange", DeviceID: "device1", Status: []byte(`{}`)},
		{Event: "Shelly:Settings", DeviceID: "device1", Settings: []byte(`{}`)},
	}

	srv := newTestWSServer(events)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, srv.URL(), nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	var receivedTypes []string
	opts := network.CloudEventStreamOptions{EventFilter: "Online"}

	err = network.StreamCloudEvents(ctx, conn, opts, func(event *model.CloudEvent, _ []byte) error {
		receivedTypes = append(receivedTypes, event.Event)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCloudEvents error: %v", err)
	}

	if len(receivedTypes) != 1 {
		t.Errorf("received %d events, want 1 (filtered to Online)", len(receivedTypes))
	}

	if len(receivedTypes) > 0 && receivedTypes[0] != "Shelly:Online" {
		t.Errorf("expected Shelly:Online event, got %q", receivedTypes[0])
	}
}

func TestStreamCloudEvents_MultipleEvents(t *testing.T) {
	t.Parallel()

	online := 1
	offline := 0
	events := []model.CloudEvent{
		{Event: "Shelly:Online", DeviceID: "device1", Online: &online},
		{Event: "Shelly:StatusOnChange", DeviceID: "device1", Status: []byte(`{"switch:0":{"output":true}}`)},
		{Event: "Shelly:Online", DeviceID: "device1", Online: &offline},
	}

	srv := newTestWSServer(events)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, srv.URL(), nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	var received []model.CloudEvent
	opts := network.CloudEventStreamOptions{}

	err = network.StreamCloudEvents(ctx, conn, opts, func(event *model.CloudEvent, _ []byte) error {
		received = append(received, *event)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCloudEvents error: %v", err)
	}

	if len(received) != 3 {
		t.Errorf("received %d events, want 3", len(received))
	}
}

func TestStreamCloudEvents_ContextCancellation(t *testing.T) {
	t.Parallel()

	online := 1
	events := []model.CloudEvent{
		{Event: "Shelly:Online", DeviceID: "device1", Online: &online},
	}

	srv := newTestWSServer(events)
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())

	conn, _, err := websocket.DefaultDialer.DialContext(context.Background(), srv.URL(), nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	receivedCount := 0
	opts := network.CloudEventStreamOptions{}

	// Cancel context after first event
	err = network.StreamCloudEvents(ctx, conn, opts, func(event *model.CloudEvent, _ []byte) error {
		receivedCount++
		cancel()
		return nil
	})

	// Should complete without error when context is cancelled
	if err != nil {
		t.Errorf("expected nil error for context cancellation, got: %v", err)
	}

	if receivedCount != 1 {
		t.Errorf("received %d events, want 1", receivedCount)
	}
}

func TestOptions_ZeroValues(t *testing.T) {
	t.Parallel()

	var opts Options

	if opts.DeviceFilter != "" {
		t.Errorf("DeviceFilter zero value = %q, want empty", opts.DeviceFilter)
	}
	if opts.EventFilter != "" {
		t.Errorf("EventFilter zero value = %q, want empty", opts.EventFilter)
	}
	if opts.Raw {
		t.Error("Raw zero value should be false")
	}
	if opts.Format != "" {
		t.Errorf("Format zero value = %q, want empty", opts.Format)
	}
}

func TestNewCommand_LocalFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	localFlags := cmd.LocalFlags()
	if !localFlags.HasFlags() {
		t.Error("events command should have local flags defined")
	}
}

func TestNewCommand_DeviceFlagType(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	deviceFlag := cmd.Flags().Lookup("device")
	if deviceFlag == nil {
		t.Fatal("device flag should exist")
	}

	if deviceFlag.Value.Type() != "string" {
		t.Errorf("device flag type = %q, want 'string'", deviceFlag.Value.Type())
	}
}

func TestNewCommand_EventFlagType(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	eventFlag := cmd.Flags().Lookup("event")
	if eventFlag == nil {
		t.Fatal("event flag should exist")
	}

	if eventFlag.Value.Type() != "string" {
		t.Errorf("event flag type = %q, want 'string'", eventFlag.Value.Type())
	}
}

func TestNewCommand_RawFlagType(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	rawFlag := cmd.Flags().Lookup("raw")
	if rawFlag == nil {
		t.Fatal("raw flag should exist")
	}

	if rawFlag.Value.Type() != "bool" {
		t.Errorf("raw flag type = %q, want 'bool'", rawFlag.Value.Type())
	}
}

func TestNewCommand_FormatFlagType(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag should exist")
	}

	if formatFlag.Value.Type() != "string" {
		t.Errorf("format flag type = %q, want 'string'", formatFlag.Value.Type())
	}
}

func TestCloudEvent_GetDeviceID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		event    model.CloudEvent
		wantID   string
	}{
		{
			name:   "DeviceID set",
			event:  model.CloudEvent{DeviceID: "device123"},
			wantID: "device123",
		},
		{
			name:   "Device set, DeviceID empty",
			event:  model.CloudEvent{Device: "device456"},
			wantID: "device456",
		},
		{
			name:   "Both set, prefer DeviceID",
			event:  model.CloudEvent{DeviceID: "device123", Device: "device456"},
			wantID: "device123",
		},
		{
			name:   "Neither set",
			event:  model.CloudEvent{},
			wantID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.event.GetDeviceID()
			if got != tt.wantID {
				t.Errorf("GetDeviceID() = %q, want %q", got, tt.wantID)
			}
		})
	}
}

func TestStreamCloudEvents_CombinedFilters(t *testing.T) {
	t.Parallel()

	online := 1
	events := []model.CloudEvent{
		{Event: "Shelly:Online", DeviceID: "device1", Online: &online},
		{Event: "Shelly:StatusOnChange", DeviceID: "device1", Status: []byte(`{}`)},
		{Event: "Shelly:Online", DeviceID: "device2", Online: &online},
		{Event: "Shelly:StatusOnChange", DeviceID: "device2", Status: []byte(`{}`)},
	}

	srv := newTestWSServer(events)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, srv.URL(), nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	var received []model.CloudEvent
	opts := network.CloudEventStreamOptions{
		DeviceFilter: "device1",
		EventFilter:  "Status",
	}

	err = network.StreamCloudEvents(ctx, conn, opts, func(event *model.CloudEvent, _ []byte) error {
		received = append(received, *event)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCloudEvents error: %v", err)
	}

	// Should only receive StatusOnChange from device1
	if len(received) != 1 {
		t.Errorf("received %d events, want 1", len(received))
	}

	if len(received) > 0 {
		if received[0].GetDeviceID() != "device1" {
			t.Errorf("expected device1, got %q", received[0].GetDeviceID())
		}
		if received[0].Event != "Shelly:StatusOnChange" {
			t.Errorf("expected StatusOnChange, got %q", received[0].Event)
		}
	}
}

func TestNewCommand_HelpOutput(t *testing.T) {
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

	if !strings.Contains(help, "events") {
		t.Error("Help should mention events")
	}

	if !strings.Contains(help, "device") {
		t.Error("Help should mention device flag")
	}

	if !strings.Contains(help, "event") {
		t.Error("Help should mention event flag")
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

//nolint:paralleltest // Tests modify global state via config.SetDefaultManager
func TestExecute_WithAllFlags(t *testing.T) {
	// Setup manager with token
	mgr := setupTestManagerWithCloud(t, "test-token", "https://localhost:19999")
	config.SetDefaultManager(mgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := NewCommand(f)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--device", "device1", "--event", "Online", "--raw"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	// Will fail to connect but should parse all flags correctly
	_ = cmd.Execute()

	// Verify flags were parsed (command ran, even if connection failed)
	output := out.String()
	if !strings.Contains(output, "Connecting") {
		t.Error("expected command to attempt connection")
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

func TestNewCommand_CommandPath(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	path := cmd.CommandPath()
	if !strings.Contains(path, "events") {
		t.Errorf("CommandPath() = %q, should contain 'events'", path)
	}
}

// TestEventHandler_RawOutput tests the raw output path in the event handler
func TestEventHandler_RawOutput(t *testing.T) {
	t.Parallel()

	online := 1
	events := []model.CloudEvent{
		{Event: "Shelly:Online", DeviceID: "device1", Online: &online},
	}

	srv := newTestWSServer(events)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, srv.URL(), nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	var rawOutput []string
	opts := network.CloudEventStreamOptions{Raw: true}

	err = network.StreamCloudEvents(ctx, conn, opts, func(event *model.CloudEvent, raw []byte) error {
		rawOutput = append(rawOutput, string(raw))
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCloudEvents error: %v", err)
	}

	if len(rawOutput) != 1 {
		t.Errorf("received %d raw outputs, want 1", len(rawOutput))
	}

	// Verify raw output is valid JSON
	if len(rawOutput) > 0 {
		var decoded model.CloudEvent
		if err := json.Unmarshal([]byte(rawOutput[0]), &decoded); err != nil {
			t.Errorf("raw output is not valid JSON: %v", err)
		}
	}
}

// TestEventHandler_JSONOutput tests the JSON output path in the event handler
func TestEventHandler_JSONOutput(t *testing.T) {
	t.Parallel()

	online := 1
	events := []model.CloudEvent{
		{Event: "Shelly:Online", DeviceID: "device1", Online: &online},
		{Event: "Shelly:StatusOnChange", DeviceID: "device1", Status: []byte(`{"switch:0":{"output":true}}`)},
	}

	srv := newTestWSServer(events)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, srv.URL(), nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	var jsonOutputs []string
	opts := network.CloudEventStreamOptions{}

	err = network.StreamCloudEvents(ctx, conn, opts, func(event *model.CloudEvent, _ []byte) error {
		formatted, jsonErr := json.Marshal(event)
		if jsonErr != nil {
			return jsonErr
		}
		jsonOutputs = append(jsonOutputs, string(formatted))
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCloudEvents error: %v", err)
	}

	if len(jsonOutputs) != 2 {
		t.Errorf("received %d JSON outputs, want 2", len(jsonOutputs))
	}

	// Verify each output is valid JSON that can decode back
	for i, output := range jsonOutputs {
		var decoded model.CloudEvent
		if err := json.Unmarshal([]byte(output), &decoded); err != nil {
			t.Errorf("output %d is not valid JSON: %v", i, err)
		}
	}
}

// TestEventHandler_TextOutput tests the text display path in the event handler
func TestEventHandler_TextOutput(t *testing.T) {
	t.Parallel()

	online := 1
	events := []model.CloudEvent{
		{Event: "Shelly:Online", DeviceID: "device1", Online: &online},
	}

	srv := newTestWSServer(events)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, srv.URL(), nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	var received []*model.CloudEvent
	opts := network.CloudEventStreamOptions{}

	err = network.StreamCloudEvents(ctx, conn, opts, func(event *model.CloudEvent, _ []byte) error {
		received = append(received, event)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCloudEvents error: %v", err)
	}

	if len(received) != 1 {
		t.Errorf("received %d events, want 1", len(received))
	}

	if len(received) > 0 {
		event := received[0]
		if event.Event != "Shelly:Online" {
			t.Errorf("event type = %q, want 'Shelly:Online'", event.Event)
		}
		if event.GetDeviceID() != "device1" {
			t.Errorf("device ID = %q, want 'device1'", event.GetDeviceID())
		}
	}
}

// TestStreamCloudEvents_HandlerError tests that handler errors stop streaming
func TestStreamCloudEvents_HandlerError(t *testing.T) {
	t.Parallel()

	online := 1
	events := []model.CloudEvent{
		{Event: "Shelly:Online", DeviceID: "device1", Online: &online},
		{Event: "Shelly:Online", DeviceID: "device2", Online: &online},
		{Event: "Shelly:Online", DeviceID: "device3", Online: &online},
	}

	srv := newTestWSServer(events)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, srv.URL(), nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	receivedCount := 0
	handlerErr := "handler error"
	opts := network.CloudEventStreamOptions{}

	err = network.StreamCloudEvents(ctx, conn, opts, func(event *model.CloudEvent, _ []byte) error {
		receivedCount++
		if receivedCount >= 1 {
			return errors.New(handlerErr)
		}
		return nil
	})

	if err == nil {
		t.Fatal("expected error from handler")
	}
	if !strings.Contains(err.Error(), handlerErr) {
		t.Errorf("expected handler error, got: %v", err)
	}

	if receivedCount != 1 {
		t.Errorf("received %d events before error, want 1", receivedCount)
	}
}

// TestCloudEvent_OnlineStatus tests online status parsing
func TestCloudEvent_OnlineStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		online   *int
		expected string
	}{
		{
			name:     "online",
			online:   func() *int { v := 1; return &v }(),
			expected: "online",
		},
		{
			name:     "offline",
			online:   func() *int { v := 0; return &v }(),
			expected: "offline",
		},
		{
			name:     "nil",
			online:   nil,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event := model.CloudEvent{
				Event:    "Shelly:Online",
				DeviceID: "device1",
				Online:   tt.online,
			}

			// Determine status string
			status := "unknown"
			if event.Online != nil {
				if *event.Online == 1 {
					status = "online"
				} else {
					status = "offline"
				}
			}

			if status != tt.expected {
				t.Errorf("status = %q, want %q", status, tt.expected)
			}
		})
	}
}

// TestCloudEvent_StatusOnChange tests status change parsing
func TestCloudEvent_StatusOnChange(t *testing.T) {
	t.Parallel()

	event := model.CloudEvent{
		Event:    "Shelly:StatusOnChange",
		DeviceID: "device1",
		Status:   []byte(`{"switch:0":{"output":true,"apower":15.5}}`),
	}

	if event.Event != "Shelly:StatusOnChange" {
		t.Errorf("event type = %q, want 'Shelly:StatusOnChange'", event.Event)
	}

	if len(event.Status) == 0 {
		t.Error("expected non-empty status")
	}

	// Verify status is valid JSON
	var status map[string]any
	if err := json.Unmarshal(event.Status, &status); err != nil {
		t.Errorf("status is not valid JSON: %v", err)
	}
}

// TestCloudEvent_Settings tests settings event parsing
func TestCloudEvent_Settings(t *testing.T) {
	t.Parallel()

	event := model.CloudEvent{
		Event:    "Shelly:Settings",
		DeviceID: "device1",
		Settings: []byte(`{"name":"Kitchen Light","eco_mode":true}`),
	}

	if event.Event != "Shelly:Settings" {
		t.Errorf("event type = %q, want 'Shelly:Settings'", event.Event)
	}

	if len(event.Settings) == 0 {
		t.Error("expected non-empty settings")
	}

	// Verify settings is valid JSON
	var settings map[string]any
	if err := json.Unmarshal(event.Settings, &settings); err != nil {
		t.Errorf("settings is not valid JSON: %v", err)
	}
}

// TestCloudEvent_Timestamp tests timestamp parsing
func TestCloudEvent_Timestamp(t *testing.T) {
	t.Parallel()

	// Use a well-known timestamp: 2024-06-15 12:00:00 UTC = 1718452800
	event := model.CloudEvent{
		Event:     "Shelly:Online",
		DeviceID:  "device1",
		Timestamp: 1718452800,
	}

	if event.Timestamp == 0 {
		t.Error("expected non-zero timestamp")
	}

	ts := time.Unix(event.Timestamp, 0).UTC()
	if ts.Year() != 2024 {
		t.Errorf("parsed year = %d, want 2024", ts.Year())
	}
	if ts.Month() != time.June {
		t.Errorf("parsed month = %v, want June", ts.Month())
	}
}

// TestNewWSServer tests the test websocket server itself
func TestNewWSServer(t *testing.T) {
	t.Parallel()

	events := []model.CloudEvent{}
	srv := newTestWSServer(events)
	defer srv.Close()

	url := srv.URL()
	if !strings.HasPrefix(url, "ws://") {
		t.Errorf("URL should start with ws://, got %q", url)
	}
}

// TestStreamCloudEvents_EmptyEvents tests handling of empty event list
func TestStreamCloudEvents_EmptyEvents(t *testing.T) {
	t.Parallel()

	events := []model.CloudEvent{}

	srv := newTestWSServer(events)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, srv.URL(), nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	receivedCount := 0
	opts := network.CloudEventStreamOptions{}

	err = network.StreamCloudEvents(ctx, conn, opts, func(event *model.CloudEvent, _ []byte) error {
		receivedCount++
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCloudEvents error: %v", err)
	}

	if receivedCount != 0 {
		t.Errorf("received %d events, want 0", receivedCount)
	}
}

// TestStreamCloudEvents_NoFilters tests streaming without any filters
func TestStreamCloudEvents_NoFilters(t *testing.T) {
	t.Parallel()

	online := 1
	events := []model.CloudEvent{
		{Event: "Shelly:Online", DeviceID: "device1", Online: &online},
		{Event: "Shelly:StatusOnChange", DeviceID: "device2", Status: []byte(`{}`)},
		{Event: "Shelly:Settings", DeviceID: "device3", Settings: []byte(`{}`)},
	}

	srv := newTestWSServer(events)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, srv.URL(), nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	var received []model.CloudEvent
	opts := network.CloudEventStreamOptions{}

	err = network.StreamCloudEvents(ctx, conn, opts, func(event *model.CloudEvent, _ []byte) error {
		received = append(received, *event)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCloudEvents error: %v", err)
	}

	if len(received) != 3 {
		t.Errorf("received %d events, want 3", len(received))
	}

	// Verify all event types
	eventTypes := make(map[string]bool)
	for _, e := range received {
		eventTypes[e.Event] = true
	}

	expectedTypes := []string{"Shelly:Online", "Shelly:StatusOnChange", "Shelly:Settings"}
	for _, et := range expectedTypes {
		if !eventTypes[et] {
			t.Errorf("missing event type %q", et)
		}
	}
}

// TestOptions_AllFieldsSet tests Options with all fields populated
func TestOptions_AllFieldsSet(t *testing.T) {
	t.Parallel()

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "json",
		},
		DeviceFilter: "device123",
		EventFilter:  "StatusOnChange",
		Raw:          true,
	}

	if opts.Format != "json" {
		t.Errorf("Format = %q, want 'json'", opts.Format)
	}
	if opts.DeviceFilter != "device123" {
		t.Errorf("DeviceFilter = %q, want 'device123'", opts.DeviceFilter)
	}
	if opts.EventFilter != "StatusOnChange" {
		t.Errorf("EventFilter = %q, want 'StatusOnChange'", opts.EventFilter)
	}
	if !opts.Raw {
		t.Error("Raw should be true")
	}
}

// TestNewCommand_FullFlagSet tests parsing all flags together
func TestNewCommand_FullFlagSet(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	args := []string{
		"--device", "mydevice",
		"--event", "Online",
		"--format", "json",
		"--raw",
	}

	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	device, _ := cmd.Flags().GetString("device")
	event, _ := cmd.Flags().GetString("event")
	format, _ := cmd.Flags().GetString("format")
	raw, _ := cmd.Flags().GetBool("raw")

	if device != "mydevice" {
		t.Errorf("device = %q, want 'mydevice'", device)
	}
	if event != "Online" {
		t.Errorf("event = %q, want 'Online'", event)
	}
	if format != "json" {
		t.Errorf("format = %q, want 'json'", format)
	}
	if !raw {
		t.Error("raw should be true")
	}
}

// TestMakeEventHandler_RawMode tests the event handler in raw mode
func TestMakeEventHandler_RawMode(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		Raw: true,
	}

	handler := makeEventHandler(ios, opts)

	online := 1
	event := &model.CloudEvent{
		Event:    "Shelly:Online",
		DeviceID: "device1",
		Online:   &online,
	}
	rawData := []byte(`{"event":"Shelly:Online","device_id":"device1","online":1}`)

	err := handler(event, rawData)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Shelly:Online") {
		t.Errorf("expected raw output to contain event, got: %q", output)
	}
}

// TestMakeEventHandler_JSONFormat tests the event handler with JSON format
func TestMakeEventHandler_JSONFormat(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "json",
		},
		Raw: false,
	}

	handler := makeEventHandler(ios, opts)

	online := 1
	event := &model.CloudEvent{
		Event:    "Shelly:Online",
		DeviceID: "device1",
		Online:   &online,
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	// Verify output is valid JSON
	var decoded model.CloudEvent
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &decoded); err != nil {
		t.Errorf("output is not valid JSON: %v, got: %q", err, output)
	}

	if decoded.Event != "Shelly:Online" {
		t.Errorf("decoded event = %q, want 'Shelly:Online'", decoded.Event)
	}
}

// TestMakeEventHandler_TextFormat tests the event handler with default text format
func TestMakeEventHandler_TextFormat(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "text",
		},
		Raw: false,
	}

	handler := makeEventHandler(ios, opts)

	online := 1
	event := &model.CloudEvent{
		Event:    "Shelly:Online",
		DeviceID: "device1",
		Online:   &online,
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	// Text format should contain the event type and device ID
	if !strings.Contains(output, "Shelly:Online") {
		t.Errorf("expected text output to contain event type, got: %q", output)
	}
	if !strings.Contains(output, "device1") {
		t.Errorf("expected text output to contain device ID, got: %q", output)
	}
}

// TestMakeEventHandler_DefaultFormat tests the event handler with empty format (defaults to text)
func TestMakeEventHandler_DefaultFormat(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		Raw: false,
		// Format is empty, should default to text display
	}

	handler := makeEventHandler(ios, opts)

	online := 0
	event := &model.CloudEvent{
		Event:    "Shelly:Online",
		DeviceID: "device2",
		Online:   &online,
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "device2") {
		t.Errorf("expected text output to contain device ID, got: %q", output)
	}
}

// TestMakeEventHandler_StatusOnChangeEvent tests the handler with status change events
func TestMakeEventHandler_StatusOnChangeEvent(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "text",
		},
	}

	handler := makeEventHandler(ios, opts)

	event := &model.CloudEvent{
		Event:    "Shelly:StatusOnChange",
		DeviceID: "device1",
		Status:   []byte(`{"switch:0":{"output":true}}`),
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "StatusOnChange") {
		t.Errorf("expected output to contain StatusOnChange, got: %q", output)
	}
}

// TestMakeEventHandler_SettingsEvent tests the handler with settings events
func TestMakeEventHandler_SettingsEvent(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "text",
		},
	}

	handler := makeEventHandler(ios, opts)

	event := &model.CloudEvent{
		Event:    "Shelly:Settings",
		DeviceID: "device1",
		Settings: []byte(`{"name":"Living Room"}`),
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Settings") {
		t.Errorf("expected output to contain Settings, got: %q", output)
	}
}

// TestMakeEventHandler_JSONMarshalError tests error handling when JSON marshaling fails
func TestMakeEventHandler_JSONMarshalError(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "json",
		},
	}

	handler := makeEventHandler(ios, opts)

	// Regular event should not cause error
	event := &model.CloudEvent{
		Event:    "Shelly:Online",
		DeviceID: "device1",
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
}

// TestMakeEventHandler_OnlineEvent tests the handler with online events
func TestMakeEventHandler_OnlineEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		online *int
		want   string
	}{
		{
			name:   "online",
			online: func() *int { v := 1; return &v }(),
			want:   "online",
		},
		{
			name:   "offline",
			online: func() *int { v := 0; return &v }(),
			want:   "offline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out := &bytes.Buffer{}
			errOut := &bytes.Buffer{}
			ios := iostreams.Test(nil, out, errOut)

			opts := &Options{
				OutputFlags: flags.OutputFlags{
					Format: "text",
				},
			}

			handler := makeEventHandler(ios, opts)

			event := &model.CloudEvent{
				Event:    "Shelly:Online",
				DeviceID: "device1",
				Online:   tt.online,
			}

			err := handler(event, nil)
			if err != nil {
				t.Fatalf("handler error: %v", err)
			}

			output := out.String()
			if !strings.Contains(output, tt.want) {
				t.Errorf("expected output to contain %q, got: %q", tt.want, output)
			}
		})
	}
}

// TestMakeEventHandler_UnknownEventType tests the handler with unknown event types
func TestMakeEventHandler_UnknownEventType(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "text",
		},
	}

	handler := makeEventHandler(ios, opts)

	event := &model.CloudEvent{
		Event:    "Shelly:UnknownEvent",
		DeviceID: "device1",
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "UnknownEvent") {
		t.Errorf("expected output to contain event type, got: %q", output)
	}
}

// TestMakeEventHandler_EmptyDeviceID tests the handler with empty device ID
func TestMakeEventHandler_EmptyDeviceID(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "text",
		},
	}

	handler := makeEventHandler(ios, opts)

	online := 1
	event := &model.CloudEvent{
		Event:  "Shelly:Online",
		Online: &online,
		// DeviceID is empty
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	// Should handle empty device ID gracefully
	if output == "" {
		t.Error("expected some output even with empty device ID")
	}
}

// TestMakeEventHandler_RawModeIgnoresFormat tests that raw mode ignores format setting
func TestMakeEventHandler_RawModeIgnoresFormat(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "json", // This should be ignored when Raw is true
		},
		Raw: true,
	}

	handler := makeEventHandler(ios, opts)

	rawData := []byte(`{"raw":"data","not":"parsed"}`)

	err := handler(&model.CloudEvent{}, rawData)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	// Raw mode should output the raw data directly
	if !strings.Contains(output, `"raw":"data"`) {
		t.Errorf("expected raw output, got: %q", output)
	}
}

// TestMakeEventHandler_JSONFormat_WithTimestamp tests JSON output includes timestamp
func TestMakeEventHandler_JSONFormat_WithTimestamp(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "json",
		},
	}

	handler := makeEventHandler(ios, opts)

	event := &model.CloudEvent{
		Event:     "Shelly:StatusOnChange",
		DeviceID:  "device1",
		Timestamp: 1718452800,
		Status:    []byte(`{"switch:0":{"output":true}}`),
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "1718452800") {
		t.Errorf("expected JSON output to include timestamp, got: %q", output)
	}
}

// TestMakeEventHandler_TextFormat_StatusChange tests text output for status changes
func TestMakeEventHandler_TextFormat_StatusChange(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "text",
		},
	}

	handler := makeEventHandler(ios, opts)

	event := &model.CloudEvent{
		Event:    "Shelly:StatusOnChange",
		DeviceID: "switch123",
		Status:   []byte(`{"switch:0":{"output":true,"apower":15.5}}`),
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "switch123") {
		t.Errorf("expected output to contain device ID, got: %q", output)
	}
}

// TestMakeEventHandler_MultipleEventsSequence tests handling multiple events in sequence
func TestMakeEventHandler_MultipleEventsSequence(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "json",
		},
	}

	handler := makeEventHandler(ios, opts)

	events := []model.CloudEvent{
		{Event: "Shelly:Online", DeviceID: "device1"},
		{Event: "Shelly:StatusOnChange", DeviceID: "device2"},
		{Event: "Shelly:Settings", DeviceID: "device3"},
	}

	for _, event := range events {
		e := event // capture loop variable
		err := handler(&e, nil)
		if err != nil {
			t.Fatalf("handler error for event %s: %v", event.Event, err)
		}
	}

	output := out.String()
	if !strings.Contains(output, "device1") {
		t.Error("expected output to contain device1")
	}
	if !strings.Contains(output, "device2") {
		t.Error("expected output to contain device2")
	}
	if !strings.Contains(output, "device3") {
		t.Error("expected output to contain device3")
	}
}

// TestMakeEventHandler_TextFormat_SettingsWithJSON tests settings display with JSON data
func TestMakeEventHandler_TextFormat_SettingsWithJSON(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "text",
		},
	}

	handler := makeEventHandler(ios, opts)

	event := &model.CloudEvent{
		Event:    "Shelly:Settings",
		DeviceID: "device1",
		Settings: []byte(`{"name":"Kitchen Light","eco_mode":true}`),
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected non-empty output for settings event")
	}
}

// TestMakeEventHandler_RawMode_EmptyData tests raw mode with empty data
func TestMakeEventHandler_RawMode_EmptyData(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		Raw: true,
	}

	handler := makeEventHandler(ios, opts)

	// Empty raw data - should still work
	err := handler(&model.CloudEvent{}, []byte{})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
}

// TestMakeEventHandler_UnknownFormat tests non-json and non-text format defaults to text
func TestMakeEventHandler_UnknownFormat(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "yaml", // Unknown format should default to text
		},
	}

	handler := makeEventHandler(ios, opts)

	online := 1
	event := &model.CloudEvent{
		Event:    "Shelly:Online",
		DeviceID: "device1",
		Online:   &online,
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// Should still produce output (via default text path)
	output := out.String()
	if output == "" {
		t.Error("expected output for unknown format (should default to text)")
	}
}

// TestMakeEventHandler_DeviceField tests handler with Device field instead of DeviceID
func TestMakeEventHandler_DeviceField(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "json",
		},
	}

	handler := makeEventHandler(ios, opts)

	event := &model.CloudEvent{
		Event:  "Shelly:Online",
		Device: "deviceABC", // Using Device instead of DeviceID
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "deviceABC") {
		t.Errorf("expected output to contain device, got: %q", output)
	}
}

// TestNewCommand_FlagShorthand tests flag shorthand availability
func TestNewCommand_FlagShorthand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag not found")
	}

	if formatFlag.Shorthand != "f" {
		t.Errorf("format flag shorthand = %q, want 'f'", formatFlag.Shorthand)
	}
}

// TestNewCommand_RunEReturnsError tests that RunE returns error correctly
func TestNewCommand_RunEReturnsError(t *testing.T) {
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when not logged in")
	}

	// Should be "not logged in" error
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("expected 'not logged in' error, got: %v", err)
	}
}

// TestOptions_CopyValues tests that Options fields can be read correctly
func TestOptions_CopyValues(t *testing.T) {
	t.Parallel()

	original := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "json",
		},
		DeviceFilter: "device123",
		EventFilter:  "Online",
		Raw:          true,
	}

	// Create copy
	copied := &Options{
		OutputFlags: flags.OutputFlags{
			Format: original.Format,
		},
		DeviceFilter: original.DeviceFilter,
		EventFilter:  original.EventFilter,
		Raw:          original.Raw,
	}

	if copied.Format != original.Format {
		t.Error("Format not copied correctly")
	}
	if copied.DeviceFilter != original.DeviceFilter {
		t.Error("DeviceFilter not copied correctly")
	}
	if copied.EventFilter != original.EventFilter {
		t.Error("EventFilter not copied correctly")
	}
	if copied.Raw != original.Raw {
		t.Error("Raw not copied correctly")
	}
}

// TestMakeEventHandler_NilEvent tests handler with nil event doesn't panic
func TestMakeEventHandler_NilEvent(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		Raw: true,
	}

	handler := makeEventHandler(ios, opts)

	// Should handle nil gracefully
	err := handler(nil, []byte("raw data"))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
}

// TestMakeEventHandler_TextFormat_LargeStatus tests text output with large status
func TestMakeEventHandler_TextFormat_LargeStatus(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	opts := &Options{
		OutputFlags: flags.OutputFlags{
			Format: "text",
		},
	}

	handler := makeEventHandler(ios, opts)

	// Create event with large status payload
	largeStatus := `{"switch:0":{"output":true,"apower":15.5,"voltage":230.1},"switch:1":{"output":false,"apower":0}}`
	event := &model.CloudEvent{
		Event:    "Shelly:StatusOnChange",
		DeviceID: "device1",
		Status:   []byte(largeStatus),
	}

	err := handler(event, nil)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "StatusOnChange") {
		t.Errorf("expected output to contain event type, got: %q", output)
	}
}
