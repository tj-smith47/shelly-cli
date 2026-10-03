package provision

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-go/reprovision"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// stubProvisionService is an in-memory provisionService for driving run() and the
// onboarding orchestration without reaching a device over BLE or hopping the
// host's WiFi to a Gen1 AP. Each hook, when nil, returns a benign default so a
// test sets only the behavior it asserts on.
type stubProvisionService struct {
	loadSource    func(context.Context, string, string) (*shelly.ProvisionSource, error)
	getWiFiCreds  func(context.Context) *shelly.OnboardWiFiConfig
	hostWiFiCreds func(context.Context) *shelly.OnboardWiFiConfig
	hostPassword  func(context.Context, string) (string, error)
	discover      func(context.Context, *shelly.OnboardOptions, func(shelly.OnboardProgress)) ([]shelly.OnboardDevice, error)
	onboardBLE    func(context.Context, []*shelly.OnboardDevice, *shelly.OnboardWiFiConfig, *shelly.OnboardOptions) []*shelly.OnboardResult
	onboardAP     func(context.Context, *shelly.OnboardDevice, *shelly.OnboardWiFiConfig, *shelly.OnboardOptions) *shelly.OnboardResult
	applySource   func(context.Context, string, *shelly.ProvisionSource) error

	applySourceCalls []string
}

func (s *stubProvisionService) LoadProvisionSource(ctx context.Context, fromDevice, fromTemplate string) (*shelly.ProvisionSource, error) {
	if s.loadSource != nil {
		return s.loadSource(ctx, fromDevice, fromTemplate)
	}
	return &shelly.ProvisionSource{}, nil
}

func (s *stubProvisionService) GetWiFiCredentials(ctx context.Context) *shelly.OnboardWiFiConfig {
	if s.getWiFiCreds != nil {
		return s.getWiFiCreds(ctx)
	}
	return nil
}

func (s *stubProvisionService) HostWiFiCredentials(ctx context.Context) *shelly.OnboardWiFiConfig {
	if s.hostWiFiCreds != nil {
		return s.hostWiFiCreds(ctx)
	}
	return nil
}

func (s *stubProvisionService) HostWiFiPassword(ctx context.Context, ssid string) (string, error) {
	if s.hostPassword != nil {
		return s.hostPassword(ctx, ssid)
	}
	return "", errors.New("no stored passphrase")
}

func (s *stubProvisionService) DiscoverForOnboard(ctx context.Context, opts *shelly.OnboardOptions, progress func(shelly.OnboardProgress)) ([]shelly.OnboardDevice, error) {
	if s.discover != nil {
		return s.discover(ctx, opts, progress)
	}
	return nil, nil
}

func (s *stubProvisionService) OnboardBLEParallel(ctx context.Context, devices []*shelly.OnboardDevice, wifiCfg *shelly.OnboardWiFiConfig, opts *shelly.OnboardOptions) []*shelly.OnboardResult {
	if s.onboardBLE != nil {
		return s.onboardBLE(ctx, devices, wifiCfg, opts)
	}
	results := make([]*shelly.OnboardResult, len(devices))
	for i, d := range devices {
		results[i] = &shelly.OnboardResult{Device: d, NewAddress: "10.0.0.50", Method: string(shelly.OnboardSourceBLE)}
	}
	return results
}

func (s *stubProvisionService) OnboardViaAP(ctx context.Context, device *shelly.OnboardDevice, wifi *shelly.OnboardWiFiConfig, opts *shelly.OnboardOptions) *shelly.OnboardResult {
	if s.onboardAP != nil {
		return s.onboardAP(ctx, device, wifi, opts)
	}
	return &shelly.OnboardResult{Device: device, NewAddress: "10.0.0.51", Method: string(shelly.OnboardSourceWiFiAP)}
}

func (s *stubProvisionService) ApplyProvisionSource(ctx context.Context, deviceAddr string, source *shelly.ProvisionSource) error {
	s.applySourceCalls = append(s.applySourceCalls, deviceAddr)
	if s.applySource != nil {
		return s.applySource(ctx, deviceAddr, source)
	}
	return nil
}

func provisionTestCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}

var errStubProvision = errors.New("stub provision failure")

func bleDevice(name string) shelly.OnboardDevice {
	return shelly.OnboardDevice{Name: name, Model: "SNDC-0D4P10WW", Source: shelly.OnboardSourceBLE, BLEAddress: "aa:bb:cc:dd:ee:ff", MACAddress: "AABBCCDDEEFF", Generation: 2}
}

func apDevice(name, ssid string) shelly.OnboardDevice {
	return shelly.OnboardDevice{Name: name, Model: "SHCB-1", Source: shelly.OnboardSourceWiFiAP, SSID: ssid, Address: "192.168.33.1", Generation: 1}
}

// TestRun_NoDevicesFound proves the no-op path: when discovery returns nothing,
// run completes without error and never calls an onboarding method.
func TestRun_NoDevicesFound(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubProvisionService{}
	opts := &Options{Factory: tf.Factory, SSID: testSSID, Yes: true, svc: stub}

	if err := run(provisionTestCtx(t), opts); err != nil {
		t.Fatalf("run with no devices should not error, got %v", err)
	}
}

// TestRun_DiscoverOnly_EmitsJSON proves --discover-only lists devices as JSON and
// stops before any credential resolution or onboarding.
func TestRun_DiscoverOnly_EmitsJSON(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubProvisionService{
		discover: func(context.Context, *shelly.OnboardOptions, func(shelly.OnboardProgress)) ([]shelly.OnboardDevice, error) {
			return []shelly.OnboardDevice{apDevice("bulb", "shellycolorbulb-AABBCC")}, nil
		},
		loadSource: func(context.Context, string, string) (*shelly.ProvisionSource, error) {
			t.Error("--discover-only must not load a provision source")
			return nil, errStubProvision
		},
	}
	opts := &Options{Factory: tf.Factory, DiscoverOnly: true, svc: stub}

	if err := run(provisionTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	out := tf.OutString()
	if !strings.Contains(out, "shellycolorbulb-AABBCC") || !strings.Contains(out, "\"ssid\"") {
		t.Errorf("expected JSON with device SSID, got %q", out)
	}
}

// TestRun_BLEDevice_Provisioned drives the BLE happy path end to end: a single
// discovered BLE device is auto-selected (--yes) and provisioned in parallel.
func TestRun_BLEDevice_Provisioned(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	var bleCalls int
	stub := &stubProvisionService{
		discover: func(context.Context, *shelly.OnboardOptions, func(shelly.OnboardProgress)) ([]shelly.OnboardDevice, error) {
			return []shelly.OnboardDevice{bleDevice("bulb1")}, nil
		},
		onboardBLE: func(_ context.Context, devices []*shelly.OnboardDevice, _ *shelly.OnboardWiFiConfig, _ *shelly.OnboardOptions) []*shelly.OnboardResult {
			bleCalls++
			return []*shelly.OnboardResult{{Device: devices[0], NewAddress: "10.0.0.50"}}
		},
	}
	opts := &Options{Factory: tf.Factory, SSID: testSSID, Password: "secret", Yes: true, svc: stub}

	if err := run(provisionTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if bleCalls != 1 {
		t.Errorf("expected OnboardBLEParallel called once, got %d", bleCalls)
	}
}

// TestRun_APDevice_Provisioned drives a single Gen1 AP device through the
// sequential WiFi-AP onboarding path.
func TestRun_APDevice_Provisioned(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	var apCalls int
	stub := &stubProvisionService{
		discover: func(context.Context, *shelly.OnboardOptions, func(shelly.OnboardProgress)) ([]shelly.OnboardDevice, error) {
			return []shelly.OnboardDevice{apDevice("bulb2", "shellycolorbulb-AABBCC")}, nil
		},
		onboardAP: func(_ context.Context, device *shelly.OnboardDevice, _ *shelly.OnboardWiFiConfig, _ *shelly.OnboardOptions) *shelly.OnboardResult {
			apCalls++
			return &shelly.OnboardResult{Device: device, NewAddress: "10.0.0.51"}
		},
	}
	opts := &Options{Factory: tf.Factory, SSID: testSSID, Password: "secret", Yes: true, svc: stub}

	if err := run(provisionTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if apCalls != 1 {
		t.Errorf("expected OnboardViaAP called once, got %d", apCalls)
	}
}

// TestRun_DiscoveryFails proves a discovery error is wrapped and propagated.
func TestRun_DiscoveryFails(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubProvisionService{
		discover: func(context.Context, *shelly.OnboardOptions, func(shelly.OnboardProgress)) ([]shelly.OnboardDevice, error) {
			return nil, errStubProvision
		},
	}
	opts := &Options{Factory: tf.Factory, SSID: testSSID, Yes: true, svc: stub}

	err := run(provisionTestCtx(t), opts)
	if err == nil || !strings.Contains(err.Error(), "discovery failed") {
		t.Fatalf("expected wrapped discovery failure, got %v", err)
	}
}

// TestRun_FromDevice_AppliesSource proves the --from-device clone path: the source
// is loaded, then its config is applied to each successfully provisioned device.
func TestRun_FromDevice_AppliesSource(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubProvisionService{
		loadSource: func(_ context.Context, fromDevice, _ string) (*shelly.ProvisionSource, error) {
			if fromDevice != "living-room" {
				t.Errorf("expected source device living-room, got %q", fromDevice)
			}
			return &shelly.ProvisionSource{WiFi: &shelly.OnboardWiFiConfig{SSID: testSSID, Password: "secret"}}, nil
		},
		discover: func(context.Context, *shelly.OnboardOptions, func(shelly.OnboardProgress)) ([]shelly.OnboardDevice, error) {
			return []shelly.OnboardDevice{bleDevice("bulb1")}, nil
		},
	}
	opts := &Options{Factory: tf.Factory, FromDevice: "living-room", Yes: true, svc: stub}

	if err := run(provisionTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(stub.applySourceCalls) != 1 || stub.applySourceCalls[0] != "10.0.0.50" {
		t.Errorf("expected source applied to the provisioned device 10.0.0.50, got %v", stub.applySourceCalls)
	}
}

// TestResolveSourceAndCreds_AdoptsSourceWiFi proves WiFi credentials from a
// --from-device source are adopted when no --ssid flag overrides them.
func TestResolveSourceAndCreds_AdoptsSourceWiFi(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubProvisionService{
		loadSource: func(context.Context, string, string) (*shelly.ProvisionSource, error) {
			return &shelly.ProvisionSource{WiFi: &shelly.OnboardWiFiConfig{SSID: "FromSource", Password: "sourcepass"}}, nil
		},
	}
	opts := &Options{Factory: tf.Factory, FromDevice: "living-room", svc: stub}

	if _, err := opts.resolveSourceAndCreds(provisionTestCtx(t), stub); err != nil {
		t.Fatalf("resolveSourceAndCreds: %v", err)
	}
	if opts.SSID != "FromSource" || opts.Password != "sourcepass" {
		t.Errorf("expected source WiFi creds adopted, got SSID=%q Password=%q", opts.SSID, opts.Password)
	}
}

// TestResolveSourceAndCreds_FlagOverridesSource proves an explicit --ssid is not
// clobbered by the source's WiFi credentials.
func TestResolveSourceAndCreds_FlagOverridesSource(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubProvisionService{
		loadSource: func(context.Context, string, string) (*shelly.ProvisionSource, error) {
			return &shelly.ProvisionSource{WiFi: &shelly.OnboardWiFiConfig{SSID: "FromSource", Password: "sourcepass"}}, nil
		},
	}
	opts := &Options{Factory: tf.Factory, FromDevice: "living-room", SSID: "FlagNet", Password: "flagpass", svc: stub}

	if _, err := opts.resolveSourceAndCreds(provisionTestCtx(t), stub); err != nil {
		t.Fatalf("resolveSourceAndCreds: %v", err)
	}
	if opts.SSID != "FlagNet" || opts.Password != "flagpass" {
		t.Errorf("flag creds must win over source, got SSID=%q Password=%q", opts.SSID, opts.Password)
	}
}

// TestResolveSourceAndCreds_LoadFails proves a source-load failure is propagated.
func TestResolveSourceAndCreds_LoadFails(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubProvisionService{
		loadSource: func(context.Context, string, string) (*shelly.ProvisionSource, error) {
			return nil, errStubProvision
		},
	}
	opts := &Options{Factory: tf.Factory, FromDevice: "living-room", svc: stub}

	if _, err := opts.resolveSourceAndCreds(provisionTestCtx(t), stub); err == nil {
		t.Fatal("expected source-load failure to propagate")
	}
}

// TestRunDiscovery_ProgressBranches drives the discovery progress callback through
// each status branch — in-progress scanning, a successful BLE finish, and a failed
// WiFi-AP finish — so the live status-line reporting is exercised, and confirms the
// discovered devices are returned.
func TestRunDiscovery_ProgressBranches(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubProvisionService{
		discover: func(_ context.Context, _ *shelly.OnboardOptions, progress func(shelly.OnboardProgress)) ([]shelly.OnboardDevice, error) {
			progress(shelly.OnboardProgress{Method: methodBLE})                                       // scanning
			progress(shelly.OnboardProgress{Method: methodBLE, Found: 1, Done: true})                 // BLE success
			progress(shelly.OnboardProgress{Method: methodWiFiAP, Done: true, Err: errStubProvision}) // AP error
			progress(shelly.OnboardProgress{Method: "network", Found: 0, Done: true})                 // zero-found → skipped
			return []shelly.OnboardDevice{bleDevice("b1")}, nil
		},
	}
	opts := &Options{Factory: tf.Factory, svc: stub}

	devices, err := opts.runDiscovery(provisionTestCtx(t), stub, opts.buildOnboardOptions())
	if err != nil {
		t.Fatalf("runDiscovery: %v", err)
	}
	if len(devices) != 1 || devices[0].Name != "b1" {
		t.Errorf("expected the discovered device returned, got %v", devices)
	}
}

// TestProvisionAll_MergesBLEAndAP proves a mixed selection provisions BLE devices
// in parallel and AP devices sequentially, returning a result per device.
func TestProvisionAll_MergesBLEAndAP(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	var bleCalls, apCalls int
	stub := &stubProvisionService{
		onboardBLE: func(_ context.Context, devices []*shelly.OnboardDevice, _ *shelly.OnboardWiFiConfig, _ *shelly.OnboardOptions) []*shelly.OnboardResult {
			bleCalls++
			return []*shelly.OnboardResult{{Device: devices[0], NewAddress: "10.0.0.50"}}
		},
		onboardAP: func(_ context.Context, device *shelly.OnboardDevice, _ *shelly.OnboardWiFiConfig, _ *shelly.OnboardOptions) *shelly.OnboardResult {
			apCalls++
			return &shelly.OnboardResult{Device: device, NewAddress: "10.0.0.51"}
		},
	}
	opts := &Options{Factory: tf.Factory, svc: stub}
	selected := []shelly.OnboardDevice{bleDevice("b1"), apDevice("a1", "shelly-AABBCC")}

	results := opts.provisionAll(provisionTestCtx(t), stub, selected, opts.buildWiFiConfig(), opts.buildOnboardOptions(), nil)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if bleCalls != 1 || apCalls != 1 {
		t.Errorf("expected one BLE and one AP onboard call, got ble=%d ap=%d", bleCalls, apCalls)
	}
}

// TestApplySourceConfig_SkipsFailed proves source config is applied only to
// devices that provisioned successfully (nil Error and a non-empty NewAddress),
// never to a failed or unlocated device.
func TestApplySourceConfig_SkipsFailed(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubProvisionService{}
	opts := &Options{Factory: tf.Factory, svc: stub}

	good := bleDevice("good")
	failed := bleDevice("failed")
	unlocated := bleDevice("unlocated")
	results := []*shelly.OnboardResult{
		{Device: &good, NewAddress: "10.0.0.50"},
		{Device: &failed, Error: errStubProvision},
		{Device: &unlocated, NewAddress: ""},
	}

	opts.applySourceConfig(provisionTestCtx(t), stub, results, &shelly.ProvisionSource{})

	if len(stub.applySourceCalls) != 1 || stub.applySourceCalls[0] != "10.0.0.50" {
		t.Errorf("source config must apply only to the successful device, got %v", stub.applySourceCalls)
	}
}

// TestPromptWiFiCredentials_AutoDetect proves credentials auto-detected from an
// existing device on the network are adopted without prompting.
func TestPromptWiFiCredentials_AutoDetect(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubProvisionService{
		getWiFiCreds: func(context.Context) *shelly.OnboardWiFiConfig {
			return &shelly.OnboardWiFiConfig{SSID: "DetectedNet", Password: "detectedpass"}
		},
	}
	opts := &Options{Factory: tf.Factory, svc: stub}

	if err := opts.promptWiFiCredentials(provisionTestCtx(t)); err != nil {
		t.Fatalf("promptWiFiCredentials: %v", err)
	}
	if opts.SSID != "DetectedNet" || opts.Password != "detectedpass" {
		t.Errorf("expected detected creds adopted, got SSID=%q Password=%q", opts.SSID, opts.Password)
	}
}

// TestPromptWiFiCredentials_HostFallback proves that when no registered device can
// supply credentials, the host's own current-network credentials are recovered and
// adopted — the parity that lets --ap-only run unattended like restore --to-ap.
func TestPromptWiFiCredentials_HostFallback(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubProvisionService{
		getWiFiCreds: func(context.Context) *shelly.OnboardWiFiConfig { return nil },
		hostWiFiCreds: func(context.Context) *shelly.OnboardWiFiConfig {
			return &shelly.OnboardWiFiConfig{SSID: "HostNet", Password: "hostpass"}
		},
	}
	opts := &Options{Factory: tf.Factory, svc: stub}

	if err := opts.promptWiFiCredentials(provisionTestCtx(t)); err != nil {
		t.Fatalf("promptWiFiCredentials: %v", err)
	}
	if opts.SSID != "HostNet" || opts.Password != "hostpass" {
		t.Errorf("expected host creds adopted, got SSID=%q Password=%q", opts.SSID, opts.Password)
	}
}

// TestSelectDevices_TargetAPMatch proves --target-ap selects exactly the device
// whose AP SSID matches, non-interactively.
func TestSelectDevices_TargetAPMatch(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory, TargetAP: "shelly-AABBCC"}
	devices := []shelly.OnboardDevice{apDevice("a1", "shelly-AABBCC"), apDevice("a2", "shelly-112233")}

	selected, err := opts.selectDevices(devices)
	if err != nil {
		t.Fatalf("selectDevices: %v", err)
	}
	if len(selected) != 1 || selected[0].SSID != "shelly-AABBCC" {
		t.Errorf("expected the single matching AP, got %v", selected)
	}
}

// TestSelectDevices_TargetAPNoMatch proves a --target-ap with no match is an error
// rather than a silent no-op that would provision nothing.
func TestSelectDevices_TargetAPNoMatch(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory, TargetAP: "shelly-NOPE"}
	devices := []shelly.OnboardDevice{apDevice("a1", "shelly-AABBCC")}

	if _, err := opts.selectDevices(devices); err == nil {
		t.Fatal("expected an error when no AP SSID matches --target-ap")
	}
}

// TestOutputDiscovered_JSON proves the --discover-only payload carries the fields
// the scan/diff flow keys on (ssid, mac, generation, source).
func TestOutputDiscovered_JSON(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory}

	if err := opts.outputDiscovered([]shelly.OnboardDevice{apDevice("bulb", "shellycolorbulb-AABBCC")}); err != nil {
		t.Fatalf("outputDiscovered: %v", err)
	}
	out := tf.OutString()
	for _, want := range []string{"\"ssid\": \"shellycolorbulb-AABBCC\"", "\"mac\":", "\"generation\": 1", "\"source\": \"WiFi AP\""} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q, got %q", want, out)
		}
	}
}

// ttyFactory returns a test factory whose terminal can prompt.
func ttyFactory(t *testing.T) *factory.TestFactory {
	t.Helper()
	tf := factory.NewTestFactory(t)
	tf.TestIO.SetStdinTTY(true)
	tf.TestIO.SetStdoutTTY(true)
	return tf
}

// TestPromptWiFiCredentials_EmptyPassword covers a named network with no
// password at a terminal: the host's stored passphrase is used when there is
// one; otherwise an empty answer joins it as an open network only when the user
// confirms, and declining refuses with the passphrase error.
func TestPromptWiFiCredentials_EmptyPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		hostPassword string
		answer       string
		confirm      bool
		wantPassword string
		wantOpen     bool
		wantErr      string
		wantAsked    bool
	}{
		{name: "host passphrase", hostPassword: "stored", wantPassword: "stored"},
		{name: "typed password", answer: "typed", wantPassword: "typed", wantAsked: true},
		{name: "empty answer confirmed", confirm: true, wantOpen: true, wantAsked: true},
		{name: "empty answer declined", wantAsked: true,
			wantErr: `no WiFi passphrase for "` + testSSID + `":`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := ttyFactory(t)
			stub := &stubProvisionService{hostPassword: func(_ context.Context, ssid string) (string, error) {
				if ssid != testSSID {
					t.Errorf("host lookup for %q, want %q", ssid, testSSID)
				}
				if tt.hostPassword == "" {
					return "", errors.New("none stored")
				}
				return tt.hostPassword, nil
			}}
			asked := false
			var confirmMsg string
			opts := &Options{
				Factory: tf.Factory, SSID: testSSID, svc: stub,
				askPassword: func(string) (string, error) {
					asked = true
					return tt.answer, nil
				},
				askConfirm: func(msg string, def bool) (bool, error) {
					confirmMsg = msg
					if def {
						t.Error("the open-network confirmation must default to no")
					}
					return tt.confirm, nil
				},
			}

			err := opts.promptWiFiCredentials(context.Background())

			if tt.wantErr != "" {
				if err == nil || !strings.HasPrefix(err.Error(), tt.wantErr) || !errors.Is(err, reprovision.ErrNoPassphrase) {
					t.Fatalf("err = %v, want prefix %s", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if asked != tt.wantAsked {
				t.Errorf("password prompt asked = %v, want %v", asked, tt.wantAsked)
			}
			if opts.Password != tt.wantPassword || opts.Open != tt.wantOpen {
				t.Errorf("Password=%q Open=%v, want %q %v", opts.Password, opts.Open, tt.wantPassword, tt.wantOpen)
			}
			if tt.wantAsked && tt.answer == "" && confirmMsg != "Join "+testSSID+" as an open network with no password?" {
				t.Errorf("confirmation = %q", confirmMsg)
			}
			if warned := strings.Contains(tf.TestIO.ErrString(), "open network"); warned != tt.wantOpen {
				t.Errorf("open-network warning shown = %v, want %v", warned, tt.wantOpen)
			}
		})
	}
}

// TestPromptWiFiCredentials_NoPromptLeavesPasswordUnknown proves that without a
// terminal, or with --yes, a missing password is neither prompted for nor
// treated as open: it is left for provisioning to look up.
func TestPromptWiFiCredentials_NoPromptLeavesPasswordUnknown(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		tty  bool
		yes  bool
	}{{name: "no terminal"}, {name: "yes at a terminal", tty: true, yes: true}} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			tf.TestIO.SetStdinTTY(tt.tty)
			tf.TestIO.SetStdoutTTY(tt.tty)
			stub := &stubProvisionService{hostPassword: func(context.Context, string) (string, error) {
				t.Error("no host lookup expected")
				return "", nil
			}}
			opts := &Options{
				Factory: tf.Factory, SSID: testSSID, Yes: tt.yes, svc: stub,
				askPassword: func(string) (string, error) {
					t.Error("no prompt expected")
					return "", nil
				},
			}
			if err := opts.promptWiFiCredentials(context.Background()); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opts.Password != "" || opts.Open {
				t.Errorf("Password=%q Open=%v, want unknown and not open", opts.Password, opts.Open)
			}
			if strings.Contains(tf.TestIO.ErrString(), "open network") {
				t.Errorf("open-network warning shown for a secured network: %q", tf.TestIO.ErrString())
			}
		})
	}
}

// TestRun_SourceWithoutKey_NotOpen drives a --from-device source whose backup
// names a network but holds no key: the AP onboard receives that network with
// an empty password and Open unset, for the SDK to look up the passphrase.
func TestRun_SourceWithoutKey_NotOpen(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		open     bool
		wantWarn bool
	}{{name: "secured station"}, {name: "open station", open: true, wantWarn: true}} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			var got *shelly.OnboardWiFiConfig
			stub := &stubProvisionService{
				loadSource: func(context.Context, string, string) (*shelly.ProvisionSource, error) {
					return &shelly.ProvisionSource{WiFi: &shelly.OnboardWiFiConfig{SSID: "FromSource", Open: tt.open}}, nil
				},
				discover: func(context.Context, *shelly.OnboardOptions, func(shelly.OnboardProgress)) ([]shelly.OnboardDevice, error) {
					return []shelly.OnboardDevice{apDevice("bulb2", "shellycolorbulb-AABBCC")}, nil
				},
				onboardAP: func(_ context.Context, d *shelly.OnboardDevice, w *shelly.OnboardWiFiConfig, _ *shelly.OnboardOptions) *shelly.OnboardResult {
					got = w
					return &shelly.OnboardResult{Device: d}
				},
			}
			opts := &Options{Factory: tf.Factory, FromDevice: "living-room", Yes: true, svc: stub}

			if err := run(provisionTestCtx(t), opts); err != nil {
				t.Fatalf("run: %v", err)
			}
			want := shelly.OnboardWiFiConfig{SSID: "FromSource", Open: tt.open}
			if got == nil || *got != want {
				t.Errorf("onboard WiFi = %+v, want %+v", got, want)
			}
			if warned := strings.Contains(tf.TestIO.ErrString(), "open network"); warned != tt.wantWarn {
				t.Errorf("open-network warning shown = %v, want %v", warned, tt.wantWarn)
			}
		})
	}
}

// TestRun_APOnboardShowsSteps proves AP onboarding runs under the labelled
// progress line with a step reporter in its context, as restore and inspect do.
func TestRun_APOnboardShowsSteps(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	hasReporter := false
	stub := &stubProvisionService{
		discover: func(context.Context, *shelly.OnboardOptions, func(shelly.OnboardProgress)) ([]shelly.OnboardDevice, error) {
			return []shelly.OnboardDevice{apDevice("bulb2", "shellycolorbulb-AABBCC")}, nil
		},
		onboardAP: func(ctx context.Context, d *shelly.OnboardDevice, _ *shelly.OnboardWiFiConfig, _ *shelly.OnboardOptions) *shelly.OnboardResult {
			hasReporter = shelly.StepReporter(ctx) != nil
			return &shelly.OnboardResult{Device: d}
		},
	}
	opts := &Options{Factory: tf.Factory, SSID: testSSID, Password: "secret", Yes: true, svc: stub}

	if err := run(provisionTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !hasReporter {
		t.Error("OnboardViaAP ran without a step reporter")
	}
	if want := "Onboarding bulb2 at AP shellycolorbulb-AABBCC (hopping host WiFi)..."; !strings.Contains(tf.TestIO.ErrString(), want) {
		t.Errorf("progress line %q missing; stderr = %q", want, tf.TestIO.ErrString())
	}
}

// TestRun_DisableAP checks that --disable-ap reaches the WiFi AP onboarding and
// that the result names the access point that was turned off.
func TestRun_DisableAP(t *testing.T) {
	t.Parallel()
	for _, disable := range []bool{true, false} {
		t.Run(fmt.Sprintf("disable=%v", disable), func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			var got *shelly.OnboardOptions
			stub := &stubProvisionService{
				discover: func(context.Context, *shelly.OnboardOptions, func(shelly.OnboardProgress)) ([]shelly.OnboardDevice, error) {
					return []shelly.OnboardDevice{apDevice("plug", "ShellyPlusPlugS-AABBCC")}, nil
				},
				onboardAP: func(_ context.Context, device *shelly.OnboardDevice, _ *shelly.OnboardWiFiConfig, o *shelly.OnboardOptions) *shelly.OnboardResult {
					got = o
					return &shelly.OnboardResult{Device: device, NewAddress: "10.0.0.51", APDisabled: o.DisableAP}
				},
			}
			cmd := NewCommand(tf.Factory)
			opts := &Options{Factory: tf.Factory, SSID: testSSID, Password: "secret", Yes: true, DisableAP: disable, svc: stub}
			if cmd.Flags().Lookup("disable-ap") == nil {
				t.Fatal("--disable-ap flag is not defined")
			}

			if err := run(provisionTestCtx(t), opts); err != nil {
				t.Fatalf("run: %v", err)
			}
			if got == nil || got.DisableAP != disable {
				t.Fatalf("onboard options = %+v, want DisableAP=%v", got, disable)
			}
			if shown := strings.Contains(tf.OutString(), "access point turned off"); shown != disable {
				t.Errorf("\"access point turned off\" shown = %v, want %v; output:\n%s", shown, disable, tf.OutString())
			}
		})
	}
}

// TestNewCommand_DisableAPExcludesBLEOnly checks that --disable-ap, which acts
// only on WiFi AP onboarding, is refused together with --ble-only.
func TestNewCommand_DisableAPExcludesBLEOnly(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--ble-only", "--disable-ap", "--discover-only"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.ExecuteContext(provisionTestCtx(t))
	if err == nil || !strings.Contains(err.Error(), "disable-ap") {
		t.Fatalf("error = %v, want a mutually-exclusive error naming disable-ap", err)
	}
}
