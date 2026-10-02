package restore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	clibackup "github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// stubRestoreService is an in-memory restoreService for driving run() and
// restoreViaAP without reaching a device or hopping WiFi.
type stubRestoreService struct {
	restore     func(context.Context, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, error)
	plan        func(*clibackup.NetworkOverride) (*clibackup.NetworkOverride, shelly.LANStationPlan, error)
	apPlanErr   error
	restoreToAP func(context.Context, string, string, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error)
	targetMAC   string
}

func (s *stubRestoreService) RestoreBackup(ctx context.Context, id string, bkp *clibackup.DeviceBackup, opts clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
	if s.restore != nil {
		return s.restore(ctx, id, bkp, opts)
	}
	return &clibackup.RestoreResult{Success: true}, nil
}

func (s *stubRestoreService) RestoreToAP(ctx context.Context, ssid, apIP, name string, bkp *clibackup.DeviceBackup, opts clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error) {
	if s.restoreToAP != nil {
		return s.restoreToAP(ctx, ssid, apIP, name, bkp, opts)
	}
	return &clibackup.RestoreResult{Success: true}, "10.0.0.50", nil
}

var errStubRestore = errors.New("stub restore failure")

const testBackupPath = "/b.json"

// validBackupData marshals a minimal valid Gen2 backup the restore path accepts.
func validBackupData(t *testing.T) []byte {
	t.Helper()
	data, err := json.Marshal(shellybackup.Backup{
		Version: 1,
		DeviceInfo: &shellybackup.DeviceInfo{
			ID: "shellyplus1-test", Name: "Test Device", Model: "SNSW-001X16EU",
			Generation: 2, Version: "1.0.0", MAC: "AA:BB:CC:DD:EE:FF",
		},
		Config:    json.RawMessage(`{"sys":{"device":{"name":"Test Device"}}}`),
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("marshal backup: %v", err)
	}
	return data
}

// setupRestore installs a memmap FS holding a valid backup at testBackupPath and
// returns the captured output buffers plus a Factory bound to them.
func setupRestore(t *testing.T) (out, errOut *bytes.Buffer, f *cmdutil.Factory) {
	t.Helper()
	factory.SetupTestFs(t)

	out, errOut = &bytes.Buffer{}, &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f = cmdutil.NewFactory().SetIOStreams(ios)

	if err := afero.WriteFile(config.Fs(), testBackupPath, validBackupData(t), 0o600); err != nil {
		t.Fatalf("write backup file: %v", err)
	}
	return out, errOut, f
}

func restoreTestCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// TestRun_Restore_Success drives the on-LAN restore path through the service stub.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_Restore_Success(t *testing.T) {
	out, errOut, f := setupRestore(t)
	stub := &stubRestoreService{}
	opts := &Options{Factory: f, Device: "dev", FilePath: testBackupPath, svc: stub}

	if err := run(restoreTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if combined := out.String() + errOut.String(); !strings.Contains(combined, "Backup restored to dev") {
		t.Errorf("missing success message, got %q", combined)
	}
}

// TestRun_Restore_Fails covers the on-LAN restore failure path.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_Restore_Fails(t *testing.T) {
	_, _, f := setupRestore(t)
	stub := &stubRestoreService{
		restore: func(context.Context, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
			return nil, errStubRestore
		},
	}
	opts := &Options{Factory: f, Device: "dev", FilePath: testBackupPath, svc: stub}

	err := run(restoreTestCtx(t), opts)
	if err == nil || !strings.Contains(err.Error(), "failed to restore backup") {
		t.Fatalf("expected restore failure, got %v", err)
	}
}

// TestRun_Restore_PartialFailure covers B4: shelly-go reports a per-section
// rejection as Success=false with a nil top-level error. run must NOT print a
// success line, must surface the rejected section, and must return an error so
// the exit code is non-zero.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_Restore_PartialFailure(t *testing.T) {
	out, errOut, f := setupRestore(t)
	stub := &stubRestoreService{
		restore: func(context.Context, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
			return &clibackup.RestoreResult{Success: false, Errors: []string{"wifi section rejected"}}, nil
		},
	}
	opts := &Options{Factory: f, Device: "dev", FilePath: testBackupPath, svc: stub}

	err := run(restoreTestCtx(t), opts)
	if err == nil {
		t.Fatal("a partial restore failure must return a non-nil error")
	}
	combined := out.String() + errOut.String()
	if strings.Contains(combined, "Backup restored to") {
		t.Errorf("must not print a success line on partial failure, got %q", combined)
	}
	if !strings.Contains(combined, "wifi section rejected") {
		t.Errorf("the rejected section must be surfaced, got %q", combined)
	}
}

// TestRun_RestoreViaAP_Success drives the --to-ap dispatch through restoreViaAP to
// a successful at-AP restore that reports the device's new LAN address.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_RestoreViaAP_Success(t *testing.T) {
	out, errOut, f := setupRestore(t)
	stub := &stubRestoreService{
		restoreToAP: func(context.Context, string, string, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error) {
			return &clibackup.RestoreResult{Success: true}, "10.23.47.227", nil
		},
	}
	opts := &Options{Factory: f, Device: "fr", FilePath: testBackupPath, ToAP: "ShellyBulbDuo-AABBCC", svc: stub}

	if err := run(restoreTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	combined := out.String() + errOut.String()
	if !strings.Contains(combined, "Backup restored to fr") || !strings.Contains(combined, "10.23.47.227") {
		t.Errorf("missing AP success / new address, got %q", combined)
	}
}

// TestRun_RestoreViaAP_Fails covers the at-AP restore failure path.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_RestoreViaAP_Fails(t *testing.T) {
	_, _, f := setupRestore(t)
	stub := &stubRestoreService{
		restoreToAP: func(context.Context, string, string, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error) {
			return nil, "", errStubRestore
		},
	}
	opts := &Options{Factory: f, Device: "fr", FilePath: testBackupPath, ToAP: "ShellyBulbDuo-AABBCC", svc: stub}

	err := run(restoreTestCtx(t), opts)
	if err == nil || !strings.Contains(err.Error(), "failed to restore via AP") {
		t.Fatalf("expected AP restore failure, got %v", err)
	}
}

// TestRun_RestoreViaAP_StaticIPWithoutGateway checks that a --to-ap restore
// accepts --static-ip alone and hands the empty gateway and netmask through, so
// the device takes them from the backup's static settings.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_RestoreViaAP_StaticIPWithoutGateway(t *testing.T) {
	_, _, f := setupRestore(t)
	data, err := json.Marshal(shellybackup.Backup{
		Version: 1,
		DeviceInfo: &shellybackup.DeviceInfo{
			ID: "shellyplus1-test", Model: "SNSW-001X16EU", Generation: 2, Version: "1.0.0", MAC: "AA:BB:CC:DD:EE:FF",
		},
		Config: json.RawMessage(`{"sys":{"device":{"name":"Test Device"}}}`),
		WiFi: json.RawMessage(`{"sta":{"ssid":"home","ipv4mode":"static","ip":"10.23.47.200",` +
			`"gw":"10.23.47.1","netmask":"255.255.254.0"}}`),
	})
	if err != nil {
		t.Fatalf("marshal backup: %v", err)
	}
	if err = afero.WriteFile(config.Fs(), testBackupPath, data, 0o600); err != nil {
		t.Fatalf("write backup file: %v", err)
	}
	var got *clibackup.NetworkOverride
	stub := &stubRestoreService{
		restoreToAP: func(_ context.Context, _, _, _ string, _ *clibackup.DeviceBackup, o clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error) {
			got = o.NetworkOverride
			return &clibackup.RestoreResult{Success: true}, "10.23.47.227", nil
		},
	}
	opts := &Options{Factory: f, Device: "fr", FilePath: testBackupPath, ToAP: "ShellyBulbDuo-AABBCC", StaticIP: "10.23.47.227", svc: stub}

	if err := run(restoreTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got == nil || got.StaticIP != "10.23.47.227" || got.Gateway != "" || got.Netmask != "" {
		t.Errorf("override = %+v, want the static IP with no gateway or netmask", got)
	}
}

// TestRun_Restore_StaticIPWithoutGateway checks that a LAN restore accepts
// --static-ip alone and hands the empty gateway and netmask to the restore, which
// takes them from the backup's static settings.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_Restore_StaticIPWithoutGateway(t *testing.T) {
	_, _, f := setupRestore(t)
	var got *clibackup.NetworkOverride
	stub := &stubRestoreService{
		restore: func(_ context.Context, _ string, _ *clibackup.DeviceBackup, o clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
			got = o.NetworkOverride
			return &clibackup.RestoreResult{Success: true}, nil
		},
	}
	opts := &Options{Factory: f, Device: "fr", FilePath: testBackupPath, StaticIP: "10.23.47.227", svc: stub}

	if err := run(restoreTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got == nil || got.StaticIP != "10.23.47.227" || got.Gateway != "" || got.Netmask != "" {
		t.Errorf("override = %+v, want the static IP with no gateway or netmask", got)
	}
}

// TestRun_OpenReachesRestore checks that --open reaches the restore as an open
// network override.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_OpenReachesRestore(t *testing.T) {
	_, _, f := setupRestore(t)
	var got *clibackup.NetworkOverride
	stub := &stubRestoreService{
		restore: func(_ context.Context, _ string, _ *clibackup.DeviceBackup, o clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
			got = o.NetworkOverride
			return &clibackup.RestoreResult{Success: true}, nil
		},
	}
	opts := &Options{Factory: f, Device: "fr", FilePath: testBackupPath, SSID: "Guest", Open: true, svc: stub}

	if err := run(restoreTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got == nil || *got != (clibackup.NetworkOverride{SSID: "Guest", Open: true}) {
		t.Errorf("override = %+v, want an open Guest network", got)
	}
}

func (s *stubRestoreService) DescribeRestoreName(_ context.Context, _ string, bkp *clibackup.DeviceBackup, o clibackup.RestoreOptions) (string, error) {
	return clibackup.DescribeName(o.Name, o.AliasName, bkp.RecordedName(), bkp.Device().MAC, s.targetMAC), nil
}

func (s *stubRestoreService) PlanLANStation(_ context.Context, _ string, _ *clibackup.DeviceBackup, ov *clibackup.NetworkOverride, _ bool) (*clibackup.NetworkOverride, shelly.LANStationPlan, error) {
	if s.plan != nil {
		return s.plan(ov)
	}
	return ov, shelly.LANStationPlan{SSID: "home", Key: shelly.StationKeyKept}, nil
}

func (s *stubRestoreService) PlanAPStation(context.Context, *clibackup.DeviceBackup, *clibackup.NetworkOverride) (shelly.LANStationPlan, error) {
	return shelly.LANStationPlan{SSID: "home", Key: shelly.StationKeyGiven}, s.apPlanErr
}

// TestRun_StationPlan covers the LAN station decision: a refusal stops the run
// before any write, the planned override (a host passphrase) reaches the
// restore, and the decision is printed.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_StationPlan(t *testing.T) {
	t.Run("refused before any write", func(t *testing.T) {
		_, _, f := setupRestore(t)
		restored := false
		stub := &stubRestoreService{
			plan: func(*clibackup.NetworkOverride) (*clibackup.NetworkOverride, shelly.LANStationPlan, error) {
				return nil, shelly.LANStationPlan{}, shelly.PassphraseError("home")
			},
			restore: func(context.Context, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
				restored = true
				return &clibackup.RestoreResult{Success: true}, nil
			},
		}
		err := run(restoreTestCtx(t), &Options{Factory: f, Device: "dev", FilePath: testBackupPath, svc: stub})
		if err == nil || !strings.Contains(err.Error(), `no WiFi passphrase for "home"`) {
			t.Fatalf("err = %v, want the passphrase error", err)
		}
		if restored {
			t.Error("the restore ran after the refusal")
		}
	})
	t.Run("host passphrase reaches the restore", func(t *testing.T) {
		out, errOut, f := setupRestore(t)
		var got *clibackup.NetworkOverride
		stub := &stubRestoreService{
			plan: func(*clibackup.NetworkOverride) (*clibackup.NetworkOverride, shelly.LANStationPlan, error) {
				return &clibackup.NetworkOverride{Password: "stored"}, shelly.LANStationPlan{SSID: "home", Key: shelly.StationKeyHost}, nil
			},
			restore: func(_ context.Context, _ string, _ *clibackup.DeviceBackup, o clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
				got = o.NetworkOverride
				return &clibackup.RestoreResult{Success: true}, nil
			},
		}
		if err := run(restoreTestCtx(t), &Options{Factory: f, Device: "dev", FilePath: testBackupPath, svc: stub}); err != nil {
			t.Fatalf("run: %v", err)
		}
		if got == nil || got.Password != "stored" {
			t.Errorf("override = %+v, want the host passphrase", got)
		}
		if !strings.Contains(out.String()+errOut.String(), "joined with the password stored on this host") {
			t.Errorf("plan not printed: %q", out.String()+errOut.String())
		}
		if strings.Contains(out.String()+errOut.String(), "stored\n") {
			t.Error("the passphrase was printed")
		}
	})
}

// TestRun_RestoreName covers the name a restore of the backup (MAC
// AA:BB:CC:DD:EE:FF, name "Test Device") onto the device "gb" writes after the
// backup's own: none for the device's own backup, the alias for another
// device's, and --name always; and the dry-run line stating that decision.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_RestoreName(t *testing.T) {
	const sameMAC, otherMAC = "AABBCCDDEEFF", "112233445566"
	tests := []struct {
		name      string
		targetMAC string
		flag      string
		want      []string
		wantPlan  string
	}{
		{name: "own backup keeps its name", targetMAC: sameMAC,
			wantPlan: `name: keeps the backup's "Test Device" (same device)`},
		{name: "another device's backup takes the alias", targetMAC: otherMAC, want: []string{"gb"},
			wantPlan: `name: "gb" from the alias (different device)`},
		{name: "--name on the own backup", targetMAC: sameMAC, flag: "Kitchen", want: []string{"Kitchen"},
			wantPlan: `name: "Kitchen" from --name`},
		{name: "--name on another device's backup", targetMAC: otherMAC, flag: "Kitchen", want: []string{"Kitchen"},
			wantPlan: `name: "Kitchen" from --name`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, dryRun := range []bool{false, true} {
				out, errOut, f := setupRestore(t)
				dev := testutil.NewGen2NameDevice(t, tt.targetMAC)
				svc := shelly.New(testutil.Gen2At(dev.Addr),
					shelly.WithRateLimiter(ratelimit.New(ratelimit.WithGen2MinInterval(0))))
				opts := &Options{Factory: f, Device: "gb", FilePath: testBackupPath, Name: tt.flag,
					SkipNetwork: true, DryRun: dryRun, svc: svc}
				if err := run(restoreTestCtx(t), opts); err != nil {
					t.Fatalf("dry run %v: run: %v", dryRun, err)
				}
				got := dev.Written()
				if dryRun {
					if sets := dev.SetCalls(); len(sets) != 0 {
						t.Errorf("dry run called %q", sets)
					}
					if combined := out.String() + errOut.String(); !strings.Contains(combined, tt.wantPlan) {
						t.Errorf("dry run output %q, want the line %q", combined, tt.wantPlan)
					}
					continue
				}
				if !slices.Contains(dev.SetCalls(), "Sys.SetConfig") {
					t.Errorf("restore called %q, want Sys.SetConfig among them", dev.SetCalls())
				}
				// The backup's Sys config, its recorded name included, is restored
				// first; only a resolved name is written after it.
				if want := append([]string{"Test Device"}, tt.want...); !slices.Equal(got, want) {
					t.Errorf("names written = %q, want %q", got, want)
				}
			}
		})
	}
}

// gen1BackupData marshals a Gen1 backup of the device with MAC
// AA:BB:CC:DD:EE:FF, recording no WiFi network.
func gen1BackupData(t *testing.T) []byte {
	t.Helper()
	data, err := json.Marshal(shellybackup.Backup{
		Version: 1,
		DeviceInfo: &shellybackup.DeviceInfo{
			ID: "shellybulbduo-ddeeff", Name: "Test Bulb", Model: "SHBDUO-1",
			Generation: 1, Version: "20230913-111821/v1.14.0-gcb84623", MAC: "AA:BB:CC:DD:EE:FF",
		},
		Config:    json.RawMessage(`{"name":"Test Bulb"}`),
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("marshal backup: %v", err)
	}
	return data
}

// TestRun_ToAPDryRun drives `backup restore --to-ap --dry-run` through the real
// service: it prints the plan a hop would carry out, reading this host's WiFi
// state but never scanning, joining or leaving a network or writing to any
// device.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_ToAPDryRun(t *testing.T) {
	const hopLine = "nothing was written"
	tests := []struct {
		name    string
		gen1    bool
		opts    Options
		want    []string
		notWant []string
	}{
		{
			name: "own device, password given",
			opts: Options{ToAP: "ShellyPlus1-AABBCCDDEEFF", SSID: "home", Password: "pw"},
			want: []string{
				`name: keeps the backup's "Test Device" (same device)`,
				`WiFi station "home": joined with the password from --password`,
				`Would hop onto AP "ShellyPlus1-AABBCCDDEEFF" from this host; ` + hopLine,
			},
			notWant: []string{"firmware"},
		},
		{
			name: "another device takes the alias, host password, static address",
			opts: Options{ToAP: "ShellyPlus1-112233445566", SSID: "home",
				StaticIP: "10.0.0.9", Gateway: "10.0.0.1", Netmask: "255.255.255.0"},
			want: []string{
				`name: "gb" from the alias (different device)`,
				`WiFi station "home": joined with the password stored on this host; ` +
					"static address 10.0.0.9 (gateway 10.0.0.1, netmask 255.255.255.0) from --static-ip",
				"WiFi station IP will be overridden to 10.0.0.9",
				`Would hop onto AP "ShellyPlus1-112233445566" from this host; ` + hopLine,
			},
		},
		{
			name: "gen1, --name, no network named",
			gen1: true,
			opts: Options{ToAP: "ShellyBulbDuo-DDEEFF", Name: "Kitchen"},
			want: []string{
				`name: "Kitchen" from --name`,
				`WiFi station "home": joined with the password stored on this host`,
				"Would update Gen1 firmware to match the backup (20230913-111821/v1.14.0-gcb84623) " +
					"if the device runs older firmware",
				`Would hop onto AP "ShellyBulbDuo-DDEEFF" from this host; ` + hopLine,
			},
		},
		{
			name: "gen1, --allow-firmware-downgrade",
			gen1: true,
			opts: Options{ToAP: "ShellyBulbDuo-DDEEFF", AllowFirmwareDowngrade: true, SSID: "home", Open: true},
			want: []string{
				`WiFi station "home": joined as an open network (--open)`,
				`Would hop onto AP "ShellyBulbDuo-DDEEFF" from this host; ` + hopLine,
			},
			notWant: []string{"firmware"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, errOut, f := setupRestore(t)
			if tt.gen1 {
				if err := afero.WriteFile(config.Fs(), testBackupPath, gen1BackupData(t), 0o600); err != nil {
					t.Fatalf("write backup file: %v", err)
				}
			}
			dev := testutil.NewGen2NameDevice(t, "112233445566")
			scanner := &testutil.CountingWiFiScanner{CurrentSSID: "home", Passwords: map[string]string{"home": "hpw"}}
			svc := shelly.New(testutil.Gen2At(dev.Addr), shelly.WithWiFiScanner(scanner))
			opts := tt.opts
			opts.Factory, opts.Device, opts.FilePath, opts.DryRun, opts.svc = f, "gb", testBackupPath, true, svc

			if err := run(restoreTestCtx(t), &opts); err != nil {
				t.Fatalf("run: %v", err)
			}
			if n := scanner.Moves(); n != 0 {
				t.Errorf("dry run made %d WiFi scans, joins or leaves, want 0", n)
			}
			if sets := dev.SetCalls(); len(sets) != 0 {
				t.Errorf("dry run wrote to the device: %q", sets)
			}
			combined := out.String() + errOut.String()
			for _, w := range tt.want {
				if !strings.Contains(combined, w) {
					t.Errorf("output missing %q:\n%s", w, combined)
				}
			}
			for _, w := range tt.notWant {
				if strings.Contains(combined, w) {
					t.Errorf("output has %q:\n%s", w, combined)
				}
			}
		})
	}
}

// TestRun_ToAPRefusals covers --to-ap runs refused before any hop, alike and
// with the same error by the dry run and the real run: network flags with no
// network named, and a network this host has no stored passphrase for (named,
// or the host's own when none is named); and, on a dry run, a static address
// the backup cannot complete.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_ToAPRefusals(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		dryOnly bool
		wantErr string
	}{
		{name: "--password with no network named", opts: Options{Password: "pw"}, wantErr: "need a network"},
		{name: "--open with no network named", opts: Options{Open: true}, wantErr: "need a network"},
		{name: "incomplete static address", opts: Options{SSID: "home", StaticIP: "10.0.0.9"}, dryOnly: true,
			wantErr: "needs a gateway and a netmask"},
		{name: "no stored password for the named network", opts: Options{SSID: "home"},
			wantErr: `no WiFi passphrase for "home"`},
		{name: "no stored password for the host's network", opts: Options{},
			wantErr: `no WiFi passphrase for "home"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errs []string
			for _, dryRun := range []bool{true, false} {
				if !dryRun && tt.dryOnly {
					continue
				}
				_, _, f := setupRestore(t)
				dev := testutil.NewGen2NameDevice(t, "112233445566")
				scanner := &testutil.CountingWiFiScanner{CurrentSSID: "home"}
				svc := shelly.New(testutil.Gen2At(dev.Addr), shelly.WithWiFiScanner(scanner))
				opts := tt.opts
				opts.Factory, opts.Device, opts.FilePath, opts.svc = f, "gb", testBackupPath, svc
				opts.ToAP, opts.DryRun = "ShellyPlus1-112233445566", dryRun

				err := run(restoreTestCtx(t), &opts)
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("dry run %v: err = %v, want %q", dryRun, err, tt.wantErr)
				}
				errs = append(errs, err.Error())
				if n := scanner.Moves(); n != 0 {
					t.Errorf("dry run %v: %d WiFi scans, joins or leaves, want 0", dryRun, n)
				}
				if sets := dev.SetCalls(); len(sets) != 0 {
					t.Errorf("dry run %v: wrote to the device: %q", dryRun, sets)
				}
			}
			if len(errs) == 2 && errs[0] != errs[1] {
				t.Errorf("dry run refused with %q, real run with %q", errs[0], errs[1])
			}
		})
	}
}
