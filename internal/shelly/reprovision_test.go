package shelly

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/afero"
	shellybackup "github.com/tj-smith47/shelly-go/backup"
	"github.com/tj-smith47/shelly-go/discovery"
	"github.com/tj-smith47/shelly-go/reprovision"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/netguard"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
)

const (
	testAPSSID   = "ShellyBulbDuo-D12965"
	testRegName  = "fr"
	testLANAddr  = "10.23.47.219"
	testHomeSSID = "home"
)

// sdkOptionNames are the reprovision option names the SDK's error text uses
// where the CLI has flags; none may reach a CLI user.
var sdkOptionNames = []string{
	"Network.Password", "Network.Gateway", "Network.Netmask", "FirmwareURL", "AllowFirmwareDowngrade", "StepTrace",
}

// fakeScanner is a WiFi backend that is only compared by identity; the fake
// flows never call it.
type fakeScanner struct{ discovery.WiFiScanner }

// withIsolatedConfig points the default config manager at an in-memory filesystem
// for the duration of the test, so registry writes never touch the user's real
// config. The default manager and the package fs are process-global, so a test
// using it must not run in parallel.
func withIsolatedConfig(t *testing.T) {
	t.Helper()
	config.SetFs(afero.NewMemMapFs())
	config.ResetDefaultManagerForTesting()
	t.Cleanup(func() {
		config.SetFs(nil)
		config.ResetDefaultManagerForTesting()
	})
}

// restoreService returns a service whose Restore is fn.
func restoreService(fn func(context.Context, *reprovision.RestoreOptions) (*reprovision.RestoreResult, error)) *Service {
	s := New(NewConfigResolver())
	s.ap = &apFlows{restore: fn}
	return s
}

// gen1Backup is a Gen1 backup with two action URLs and a home network.
func gen1Backup() *backup.DeviceBackup {
	return &backup.DeviceBackup{Backup: &shellybackup.Backup{
		DeviceInfo: &shellybackup.DeviceInfo{
			Model:      "SHBDUO-1",
			Generation: 1,
			Version:    "20230913-111821/v1.14.0-gcb84623",
			MAC:        "AABBCCD12965",
		},
		WiFi: json.RawMessage(`{"sta":{"ssid":"` + testHomeSSID + `"}}`),
		Webhooks: json.RawMessage(`{"actions":{
			"out_on_url":[{"index":0,"urls":["http://localhost/on"],"enabled":true}],
			"out_off_url":[{"index":0,"urls":[],"enabled":false}]}}`),
	}}
}

// gen2Backup is a Gen2 backup with one script, two schedules and one webhook.
func gen2Backup() *backup.DeviceBackup {
	return &backup.DeviceBackup{Backup: &shellybackup.Backup{
		DeviceInfo: &shellybackup.DeviceInfo{Model: "SNSW-001P16EU", Generation: 2, Version: "1.4.4"},
		Scripts:    []*shellybackup.Script{{ID: 1, Name: "s"}},
		Schedules:  json.RawMessage(`{"jobs":[{},{}]}`),
		Webhooks:   json.RawMessage(`{"hooks":[{}]}`),
	}}
}

func TestRestoreToAP_OptionMapping(t *testing.T) {
	t.Parallel()

	scanner := fakeScanner{}
	var trace bytes.Buffer
	bkp := gen1Backup()

	tests := []struct {
		name     string
		apSSID   string // default testAPSSID, whose suffix is the backup MAC's last three bytes
		opts     backup.RestoreOptions
		want     reprovision.Network
		wantName string
	}{
		{
			name:     "every flag",
			wantName: "front-room",
			opts: backup.RestoreOptions{
				Name:                   "front-room",
				AliasName:              "fr",
				SkipAuth:               true,
				SkipScripts:            true,
				SkipSchedules:          true,
				SkipWebhooks:           true,
				SkipKVS:                true,
				SkipState:              true,
				SkipMeters:             true,
				AllowFirmwareDowngrade: true,
				FirmwareURL:            "http://fw.example/duo.zip",
				StepTrace:              &trace,
				NetworkOverride: &backup.NetworkOverride{
					SSID: "iot", Password: "secret", StaticIP: testLANAddr,
					Gateway: "10.23.47.1", Netmask: "255.255.254.0", DNS: "10.23.47.2",
				},
			},
			want: reprovision.Network{
				SSID: "iot", Password: "secret", StaticIP: testLANAddr,
				Gateway: "10.23.47.1", Netmask: "255.255.254.0", DNS: "10.23.47.2",
			},
		},
		{name: "no override", opts: backup.RestoreOptions{}},
		{name: "alias, AP of the backup's device: backup name kept", opts: backup.RestoreOptions{AliasName: "fr"}},
		{name: "alias, AP of another device: alias written", apSSID: "ShellyBulbDuo-0A0B0C",
			opts: backup.RestoreOptions{AliasName: "fr"}, wantName: "fr"},
		{name: "alias, non-Shelly AP: alias written", apSSID: "guest-net",
			opts: backup.RestoreOptions{AliasName: "fr"}, wantName: "fr"},
		{name: "backup of another device: foreign backup allowed", apSSID: "ShellyBulbDuo-0A0B0C",
			opts: backup.RestoreOptions{}},
		{
			name: "open network",
			opts: backup.RestoreOptions{NetworkOverride: &backup.NetworkOverride{SSID: "guest", Open: true}},
			want: reprovision.Network{SSID: "guest", Open: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got *reprovision.RestoreOptions
			svc := restoreService(func(_ context.Context, o *reprovision.RestoreOptions) (*reprovision.RestoreResult, error) {
				got = o
				return &reprovision.RestoreResult{}, nil
			})
			svc.wifiScanner = scanner

			var steps []string
			ctx := WithStepReporter(context.Background(), func(s string) { steps = append(steps, s) })
			apSSID := tt.apSSID
			if apSSID == "" {
				apSSID = testAPSSID
			}
			if _, _, err := svc.RestoreToAP(ctx, apSSID, "192.168.33.150", testRegName, bkp, tt.opts); err != nil {
				t.Fatalf("RestoreToAP: %v", err)
			}

			if got.Logger == nil {
				t.Error("Logger is nil; debug traces would be lost")
			}
			if got.OnStep == nil {
				t.Fatal("OnStep is nil; the spinner would get no stages")
			}
			got.OnStep("joining access point")
			if len(steps) != 1 || steps[0] != "joining access point" {
				t.Errorf("OnStep reached reporter with %v", steps)
			}

			want := reprovision.RestoreOptions{
				Scanner:                scanner,
				StepTrace:              tt.opts.StepTrace,
				Backup:                 bkp.Backup,
				APSSID:                 apSSID,
				Name:                   tt.wantName,
				APHostIP:               "192.168.33.150",
				FirmwareURL:            tt.opts.FirmwareURL,
				Network:                tt.want,
				AllowFirmwareDowngrade: tt.opts.AllowFirmwareDowngrade,
				AllowForeignBackup:     true,
				SkipAuth:               tt.opts.SkipAuth,
				SkipScripts:            tt.opts.SkipScripts,
				SkipSchedules:          tt.opts.SkipSchedules,
				SkipKVS:                tt.opts.SkipKVS,
				SkipWebhooks:           tt.opts.SkipWebhooks,
				SkipState:              tt.opts.SkipState,
				SkipMeters:             tt.opts.SkipMeters,
			}
			cmp := *got
			cmp.Logger, cmp.OnStep = nil, nil
			if !reflect.DeepEqual(cmp, want) {
				t.Errorf("options =\n%+v\nwant\n%+v", cmp, want)
			}
		})
	}
}

// TestRestoreToAP_EachSwitchReachesItsOption sets one boolean flag at a time,
// so two flags wired to each other's option fail.
func TestRestoreToAP_EachSwitchReachesItsOption(t *testing.T) {
	t.Parallel()

	flags := []struct {
		name string
		cli  func(*backup.RestoreOptions)
		sdk  func(*reprovision.RestoreOptions)
	}{
		{"--skip-auth", func(o *backup.RestoreOptions) { o.SkipAuth = true }, func(o *reprovision.RestoreOptions) { o.SkipAuth = true }},
		{"--skip-scripts", func(o *backup.RestoreOptions) { o.SkipScripts = true }, func(o *reprovision.RestoreOptions) { o.SkipScripts = true }},
		{"--skip-schedules", func(o *backup.RestoreOptions) { o.SkipSchedules = true }, func(o *reprovision.RestoreOptions) { o.SkipSchedules = true }},
		{"--skip-webhooks", func(o *backup.RestoreOptions) { o.SkipWebhooks = true }, func(o *reprovision.RestoreOptions) { o.SkipWebhooks = true }},
		{"--skip-kvs", func(o *backup.RestoreOptions) { o.SkipKVS = true }, func(o *reprovision.RestoreOptions) { o.SkipKVS = true }},
		{"--skip-state", func(o *backup.RestoreOptions) { o.SkipState = true }, func(o *reprovision.RestoreOptions) { o.SkipState = true }},
		{"--skip-meters", func(o *backup.RestoreOptions) { o.SkipMeters = true }, func(o *reprovision.RestoreOptions) { o.SkipMeters = true }},
		{"--allow-firmware-downgrade", func(o *backup.RestoreOptions) { o.AllowFirmwareDowngrade = true }, func(o *reprovision.RestoreOptions) { o.AllowFirmwareDowngrade = true }},
	}
	bkp := gen1Backup()
	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			var got *reprovision.RestoreOptions
			svc := restoreService(func(_ context.Context, o *reprovision.RestoreOptions) (*reprovision.RestoreResult, error) {
				got = o
				return &reprovision.RestoreResult{}, nil
			})
			var opts backup.RestoreOptions
			f.cli(&opts)
			if _, _, err := svc.RestoreToAP(context.Background(), testAPSSID, "", testRegName, bkp, opts); err != nil {
				t.Fatalf("RestoreToAP: %v", err)
			}
			want := reprovision.RestoreOptions{
				Backup: bkp.Backup, APSSID: testAPSSID, Scanner: OfflineWiFiScanner{}, AllowForeignBackup: true,
			}
			f.sdk(&want)
			cmp := *got
			cmp.Logger, cmp.OnStep = nil, nil
			if !reflect.DeepEqual(cmp, want) {
				t.Errorf("options =\n%+v\nwant\n%+v", cmp, want)
			}
		})
	}
}

func TestRestoreToAP_ResultMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		bkp  *backup.DeviceBackup
		opts backup.RestoreOptions
		lib  *shellybackup.RestoreResult
		want *backup.RestoreResult
	}{
		{
			name: "gen1 success counts action URLs",
			bkp:  gen1Backup(),
			lib:  &shellybackup.RestoreResult{Success: true, RestartRequired: true, Warnings: []string{"w"}},
			want: &backup.RestoreResult{
				Success: true, ConfigRestored: true, RestartRequired: true,
				Warnings: []string{"w"}, WebhooksRestored: 2,
			},
		},
		{
			name: "gen1 skip webhooks counts none",
			bkp:  gen1Backup(),
			opts: backup.RestoreOptions{SkipWebhooks: true},
			lib:  &shellybackup.RestoreResult{Success: true},
			want: &backup.RestoreResult{Success: true, ConfigRestored: true},
		},
		{
			name: "gen2 success counts scripts schedules webhooks",
			bkp:  gen2Backup(),
			lib:  &shellybackup.RestoreResult{Success: true},
			want: &backup.RestoreResult{
				Success: true, ConfigRestored: true,
				ScriptsRestored: 1, SchedulesRestored: 2, WebhooksRestored: 1,
			},
		},
		{
			name: "rejected sections keep errors and count nothing",
			bkp:  gen2Backup(),
			lib: &shellybackup.RestoreResult{
				Warnings: []string{"set device name: refused"},
				Errors:   []error{errors.New("mqtt: refused"), errors.New("cloud: refused")},
			},
			want: &backup.RestoreResult{
				Warnings: []string{"set device name: refused"},
				Errors:   []string{"mqtt: refused", "cloud: refused"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := restoreService(func(context.Context, *reprovision.RestoreOptions) (*reprovision.RestoreResult, error) {
				return &reprovision.RestoreResult{Restore: tt.lib}, nil
			})
			got, addr, err := svc.RestoreToAP(context.Background(), testAPSSID, "", testRegName, tt.bkp, tt.opts)
			if err != nil {
				t.Fatalf("RestoreToAP: %v", err)
			}
			if addr != "" {
				t.Errorf("addr = %q, want empty for a device not seen on the LAN", addr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("result = %+v, want %+v", got, tt.want)
			}
		})
	}
}

//nolint:paralleltest // withIsolatedConfig swaps the process-global config
func TestRestoreToAP_RegistryUpdate(t *testing.T) {
	lanErr := errors.New("set mqtt: timeout")
	noRoute := fmt.Errorf("%w: restore applied at AP %q and the device rejoined the LAN at %s (seen via mdns), "+
		"but this host cannot reach it to write the full configuration; run the restore from a host on the "+
		"device's subnet", reprovision.ErrNoRoute, testAPSSID, testLANAddr)

	tests := []struct {
		name     string
		res      *reprovision.RestoreResult
		err      error
		wantAddr string
		wantErr  string
	}{
		{
			name:     "success records the address",
			res:      &reprovision.RestoreResult{Address: testLANAddr, Reachable: true, Restore: &shellybackup.RestoreResult{Success: true}},
			wantAddr: testLANAddr,
		},
		{
			name: "LAN failure still records the address",
			res:  &reprovision.RestoreResult{Address: testLANAddr, Reachable: true},
			err: fmt.Errorf("the device joined the LAN at %s but the full configuration restore failed: %w",
				testLANAddr, lanErr),
			wantAddr: testLANAddr,
			wantErr: testRegName + " joined the LAN at " + testLANAddr +
				" but the full configuration restore failed: set mqtt: timeout",
		},
		{
			name: "no route records nothing",
			res:  &reprovision.RestoreResult{Address: testLANAddr, SeenVia: "mdns"},
			err:  noRoute,
			wantErr: `restore applied at AP "` + testAPSSID + `" and ` + testRegName + ` rejoined the LAN at ` +
				testLANAddr + ` (seen via mdns), but this host has no route to it to write the full configuration — ` +
				`run the restore from a host on the device's subnet`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withIsolatedConfig(t)
			svc := restoreService(func(context.Context, *reprovision.RestoreOptions) (*reprovision.RestoreResult, error) {
				return tt.res, tt.err
			})
			_, addr, err := svc.RestoreToAP(context.Background(), testAPSSID, "", testRegName, gen1Backup(), backup.RestoreOptions{})
			if addr != tt.wantAddr {
				t.Errorf("addr = %q, want %q", addr, tt.wantAddr)
			}
			dev, ok := config.GetDevice(testRegName)
			switch {
			case tt.wantAddr == "" && ok:
				t.Errorf("registry entry %+v written for a device this host cannot reach", dev)
			case tt.wantAddr != "" && (!ok || dev.Address != tt.wantAddr):
				t.Errorf("registry entry = %+v (ok=%v), want address %s", dev, ok, tt.wantAddr)
			}
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("RestoreToAP: %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Errorf("err = %v\nwant %s", err, tt.wantErr)
			}
			if !errors.Is(err, tt.err) {
				t.Error("errors.Is no longer reaches the SDK error")
			}
		})
	}

	t.Run("existing entry is moved", func(t *testing.T) {
		withIsolatedConfig(t)
		updateRegistryAddress(testRegName, "10.0.0.1", gen1Backup())
		svc := restoreService(func(context.Context, *reprovision.RestoreOptions) (*reprovision.RestoreResult, error) {
			return &reprovision.RestoreResult{Address: testLANAddr, Reachable: true}, nil
		})
		if _, _, err := svc.RestoreToAP(context.Background(), testAPSSID, "", testRegName, gen1Backup(), backup.RestoreOptions{}); err != nil {
			t.Fatalf("RestoreToAP: %v", err)
		}
		if dev, _ := config.GetDevice(testRegName); dev.Address != testLANAddr {
			t.Errorf("address = %q, want %s", dev.Address, testLANAddr)
		}
	})
}

func TestRestoreAPError_Sentinels(t *testing.T) {
	t.Parallel()

	apPrefix := fmt.Sprintf("restore at AP %q failed: ", testAPSSID)
	hopPrefix := fmt.Sprintf("AP hop for %q failed: ", testAPSSID)
	halted := `restore halted: device became unstable after the "mqtt" step — a write drove it into a reboot ` +
		`loop; capture the per-step trace with --trace-file to confirm`
	sdkHalt := fmt.Errorf("%w: restore halted: the device became unstable after the %q step, a write drove it "+
		"into a reboot loop; set StepTrace to capture the per-step trace", reprovision.ErrUnstable, "mqtt")
	haltedRes := func(reachable bool) *reprovision.RestoreResult {
		r := &reprovision.RestoreResult{Restore: &shellybackup.RestoreResult{DestabilizedStep: "mqtt"}}
		if reachable {
			r.Address, r.Reachable = testLANAddr, true
		}
		return r
	}
	identity := fmt.Errorf("%w: AP %q is serving device %s, whose MAC does not match the AP name's suffix %q; "+
		"nothing was written; confirm the AP name and that the intended device is the one in AP mode, then retry",
		reprovision.ErrIdentityMismatch, testAPSSID, "AABBCC6645B6", "D12965")
	unreachable := fmt.Errorf("%w: failed to connect to Shelly AP %q: %w",
		reprovision.ErrAPUnreachable, testAPSSID, errors.New("no such network"))
	fwUpdate := fmt.Errorf("%w: %w", reprovision.ErrFirmwareUpdate, errors.New("OTA did not take"))
	unstableGate := fmt.Errorf("%w: refusing to write the station config: the device is not holding a stable "+
		"uptime at its factory AP (highest uptime %ds, need %ds held, the signature of a reboot loop)",
		reprovision.ErrUnstable, 4, 12)

	tests := []struct {
		name     string
		sentinel error
		err      error
		res      *reprovision.RestoreResult
		want     string
	}{
		{
			name:     "no passphrase names the SDK's network and the flags",
			sentinel: reprovision.ErrNoPassphrase,
			err:      &reprovision.NoPassphraseError{SSID: testHomeSSID},
			want: `no WiFi passphrase for "` + testHomeSSID + `": Shelly devices return no station key and none ` +
				`was found in this host's stored credentials — pass --password, or --open for a network that has ` +
				`no password`,
		},
		{
			name:     "no passphrase with no network named",
			sentinel: reprovision.ErrNoPassphrase,
			err:      &reprovision.NoPassphraseError{},
			want: `no WiFi passphrase: no network was named and this host is not on a WiFi network — pass --ssid ` +
				`with --password, or with --open for a network that has no password`,
		},
		{
			name:     "incomplete static network names the flags",
			sentinel: reprovision.ErrIncompleteStaticNetwork,
			err: fmt.Errorf("%w: %w: static address 10.0.0.9 needs a gateway and a netmask, and the backup has "+
				"none; set Network.Gateway and Network.Netmask", types.ErrInvalidParam,
				reprovision.ErrIncompleteStaticNetwork),
			want: backup.MsgIncompleteBackupStatic,
		},
		{
			name:     "identity mismatch keeps SDK text",
			sentinel: reprovision.ErrIdentityMismatch,
			err:      fmt.Errorf("restore at AP %q failed: %w", testAPSSID, identity),
			want:     apPrefix + identity.Error(),
		},
		{
			name:     "AP unreachable keeps SDK text",
			sentinel: reprovision.ErrAPUnreachable,
			err:      fmt.Errorf("AP hop for %q failed: %w", testAPSSID, unreachable),
			want:     hopPrefix + unreachable.Error(),
		},
		{
			name:     "firmware unavailable names the flags",
			sentinel: reprovision.ErrFirmwareUnavailable,
			err: fmt.Errorf("AP hop for %q failed: %w", testAPSSID, &reprovision.FirmwareUnavailableError{
				Current: "20210101-000000/v1.9.0", Required: "20230913-111821/v1.14.0-gcb84623",
			}),
			// The old CLI's text, byte for byte.
			want: hopPrefix + `device on firmware "20210101-000000/v1.9.0" needs an update to the backup's ` +
				`"20230913-111821/v1.14.0-gcb84623" before restore, but no firmware image is available (the factory ` +
				`AP has no internet, so the image is prefetched before the hop — its URL was underivable or the ` +
				`download failed); retry with connectivity, pass --firmware-url, or --allow-firmware-downgrade to ` +
				`force the downgrade and accept the reboot-loop risk`,
		},
		{
			name:     "firmware update keeps SDK text",
			sentinel: reprovision.ErrFirmwareUpdate,
			err:      fmt.Errorf("AP hop for %q failed: %w", testAPSSID, fwUpdate),
			want:     hopPrefix + fwUpdate.Error(),
		},
		{
			name:     "unstable at the AP gate keeps SDK text",
			sentinel: reprovision.ErrUnstable,
			err:      fmt.Errorf("AP hop for %q failed: %w", testAPSSID, unstableGate),
			want:     hopPrefix + unstableGate.Error(),
		},
		{
			name:     "restore halted at the AP names --trace-file",
			sentinel: reprovision.ErrUnstable,
			err:      fmt.Errorf("restore at AP %q failed: %w", testAPSSID, sdkHalt),
			res:      haltedRes(false),
			want:     apPrefix + halted,
		},
		{
			name:     "restore halted on the LAN names --trace-file",
			sentinel: reprovision.ErrUnstable,
			err: fmt.Errorf("the device joined the LAN at %s but the full configuration restore failed: %w",
				testLANAddr, sdkHalt),
			res:  haltedRes(true),
			want: testRegName + " joined the LAN at " + testLANAddr + " but the full configuration restore failed: " + halted,
		},
		{
			name:     "not rejoined names the target",
			sentinel: reprovision.ErrNotRejoined,
			err: fmt.Errorf("restore applied at AP %q but the device was not seen back on the LAN (%w); the "+
				"device may still be on its factory AP; if this host is not on the device's subnet, restore from "+
				"one that is", testAPSSID, fmt.Errorf("%w within %s", reprovision.ErrNotRejoined, "2m0s")),
			res: &reprovision.RestoreResult{MAC: "AABBCCD12965"},
			want: `restore applied at AP "` + testAPSSID + `" but ` + testRegName + ` was not seen back on the LAN ` +
				`(device not seen on the network within 2m0s); the device may still be on its factory AP — if ` +
				`this host is not on the device's subnet, restore from one that is`,
		},
		{
			name:     "no route names the target",
			sentinel: reprovision.ErrNoRoute,
			err:      fmt.Errorf("%w: rejoined but unreachable", reprovision.ErrNoRoute),
			res:      &reprovision.RestoreResult{Address: testLANAddr, SeenVia: "coiot"},
			// The old CLI's text, byte for byte.
			want: `restore applied at AP "` + testAPSSID + `" and ` + testRegName + ` rejoined the LAN at ` +
				testLANAddr + ` (seen via coiot), but this host has no route to it to write the full configuration ` +
				`— run the restore from a host on the device's subnet`,
		},
	}
	target := restoreTarget{apSSID: testAPSSID, name: testRegName}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := restoreAPError(tt.err, target, tt.res)
			if got.Error() != tt.want {
				t.Errorf("message =\n%s\nwant\n%s", got.Error(), tt.want)
			}
			if !errors.Is(got, tt.sentinel) {
				t.Errorf("errors.Is(%v) = false", tt.sentinel)
			}
			for _, name := range sdkOptionNames {
				if strings.Contains(got.Error(), name) {
					t.Errorf("message names SDK option %s: %s", name, got)
				}
			}
		})
	}
}

func TestRestoreToAP_IncompleteStaticAddress(t *testing.T) {
	t.Parallel()

	refused := func(ip string) error {
		return fmt.Errorf("%w: %w: static address %s needs a gateway and a netmask, and the backup has none; "+
			"set Network.Gateway and Network.Netmask", types.ErrInvalidParam, reprovision.ErrIncompleteStaticNetwork, ip)
	}
	rejected := fmt.Errorf("restore at AP %q failed: %w", testAPSSID,
		fmt.Errorf("%w: wifi.sta: invalid argument", types.ErrInvalidParam))
	tests := []struct {
		name     string
		override *backup.NetworkOverride
		res      *reprovision.RestoreResult
		err      error
		want     string
	}{
		{
			name:     "--static-ip without a gateway on a DHCP backup names the flags",
			override: &backup.NetworkOverride{StaticIP: testLANAddr},
			err:      refused(testLANAddr),
			want: "static address " + testLANAddr + " needs a gateway and a netmask, and the backup has none — " +
				"pass --gateway and --netmask",
		},
		{
			name: "backup static address without a gateway names the flags",
			err:  refused("10.0.0.9"),
			want: backup.MsgIncompleteBackupStatic,
		},
		{
			name:     "a device rejecting a write after the hop keeps the SDK text",
			override: &backup.NetworkOverride{StaticIP: testLANAddr},
			res:      &reprovision.RestoreResult{},
			err:      rejected,
			want:     rejected.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got *reprovision.RestoreOptions
			svc := restoreService(func(_ context.Context, o *reprovision.RestoreOptions) (*reprovision.RestoreResult, error) {
				got = o
				return tt.res, tt.err
			})
			_, _, err := svc.RestoreToAP(context.Background(), testAPSSID, "", testRegName, gen1Backup(),
				backup.RestoreOptions{NetworkOverride: tt.override})
			if err == nil || err.Error() != tt.want {
				t.Errorf("err =\n%v\nwant\n%s", err, tt.want)
			}
			if !errors.Is(err, types.ErrInvalidParam) {
				t.Error("errors.Is(types.ErrInvalidParam) = false")
			}
			if errors.Is(err, reprovision.ErrIncompleteStaticNetwork) != errors.Is(tt.err, reprovision.ErrIncompleteStaticNetwork) {
				t.Error("the reworded error lost or gained ErrIncompleteStaticNetwork")
			}
			if tt.override != nil && (got.Network.Gateway != "" || got.Network.Netmask != "") {
				t.Errorf("the CLI filled in gateway %q / netmask %q; the SDK takes them from the backup",
					got.Network.Gateway, got.Network.Netmask)
			}
		})
	}
}

// TestRestoreToAP_PassphraseNamesJoinNetwork checks that the refusal names the
// network the SDK resolved, whatever the backup or flags held.
func TestRestoreToAP_PassphraseNamesJoinNetwork(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		sdkSSID string
		want    string
	}{
		{name: "named network", sdkSSID: "iot", want: `no WiFi passphrase for "iot":`},
		{name: "no network named", sdkSSID: "", want: "no WiFi passphrase: no network was named"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := restoreService(func(context.Context, *reprovision.RestoreOptions) (*reprovision.RestoreResult, error) {
				return nil, &reprovision.NoPassphraseError{SSID: tt.sdkSSID}
			})
			_, _, err := svc.RestoreToAP(context.Background(), testAPSSID, "", testRegName, gen2Backup(),
				backup.RestoreOptions{NetworkOverride: &backup.NetworkOverride{SSID: "flag-ssid"}})
			if err == nil || !strings.HasPrefix(err.Error(), tt.want) {
				t.Errorf("err = %v, want prefix %s", err, tt.want)
			}
			var pwErr *reprovision.NoPassphraseError
			if !errors.As(err, &pwErr) || pwErr.SSID != tt.sdkSSID {
				t.Errorf("errors.As lost the SDK error: %v", err)
			}
		})
	}
}

func TestOnboardViaAP_OptionMapping(t *testing.T) {
	t.Parallel()

	scanner := fakeScanner{}
	var got *reprovision.OnboardOptions
	svc := New(NewConfigResolver(), WithWiFiScanner(scanner))
	svc.ap = &apFlows{onboard: func(_ context.Context, o *reprovision.OnboardOptions) (*reprovision.OnboardResult, error) {
		got = o
		return &reprovision.OnboardResult{Note: "provisioned but device not seen"}, nil
	}}
	dev := &OnboardDevice{Name: "duo", SSID: testAPSSID, Generation: 1, Model: "SHBDUO-1"}
	wifi := &OnboardWiFiConfig{
		SSID: "iot", Password: "secret", StaticIP: testLANAddr,
		Gateway: "10.23.47.1", Netmask: "255.255.254.0", DNS: "10.23.47.2",
	}
	var steps []string
	ctx := WithStepReporter(context.Background(), func(s string) { steps = append(steps, s) })

	res := svc.OnboardViaAP(ctx, dev, wifi, &OnboardOptions{})

	if got.Logger == nil || got.OnStep == nil {
		t.Fatal("Logger and OnStep must be set")
	}
	got.OnStep("writing WiFi settings")
	if len(steps) != 1 {
		t.Errorf("OnStep reached reporter with %v", steps)
	}
	want := reprovision.OnboardOptions{
		Scanner: scanner,
		APSSID:  testAPSSID,
		Network: reprovision.Network{
			SSID: "iot", Password: "secret", StaticIP: testLANAddr,
			Gateway: "10.23.47.1", Netmask: "255.255.254.0", DNS: "10.23.47.2",
		},
		Generation: 1,
	}
	cmp := *got
	cmp.Logger, cmp.OnStep = nil, nil
	if !reflect.DeepEqual(cmp, want) {
		t.Errorf("options =\n%+v\nwant\n%+v", cmp, want)
	}
	if res.Error != nil || res.Note != "provisioned but device not seen" || res.NewAddress != "" || res.Registered {
		t.Errorf("result = %+v, want the note with no address and no registration", res)
	}
	if res.Method != string(OnboardSourceWiFiAP) || res.Device != dev {
		t.Errorf("result method/device = %q/%p, want %q/%p", res.Method, res.Device, OnboardSourceWiFiAP, dev)
	}
}

//nolint:paralleltest // withIsolatedConfig swaps the process-global config
func TestOnboardViaAP_RegistersWithDeviceGeneration(t *testing.T) {
	for _, reachable := range []bool{true, false} {
		t.Run(fmt.Sprintf("reachable=%v", reachable), func(t *testing.T) {
			withIsolatedConfig(t)
			svc := New(NewConfigResolver())
			svc.ap = &apFlows{onboard: func(context.Context, *reprovision.OnboardOptions) (*reprovision.OnboardResult, error) {
				return &reprovision.OnboardResult{Address: testLANAddr, Generation: 2, Reachable: reachable}, nil
			}}
			dev := &OnboardDevice{Name: "plug", SSID: "ShellyPlusPlugS-AABBCC", Generation: 1, Model: "SNPL-00112EU"}

			res := svc.OnboardViaAP(context.Background(), dev, &OnboardWiFiConfig{SSID: "iot", Password: "p"}, nil)

			if res.Error != nil || res.NewAddress != testLANAddr || !res.Registered {
				t.Fatalf("result = %+v, want registered at %s", res, testLANAddr)
			}
			got, ok := config.GetDevice("plug")
			if !ok || got.Address != testLANAddr || got.Generation != 2 {
				t.Errorf("registry entry = %+v (ok=%v), want %s with the device's generation 2", got, ok, testLANAddr)
			}
			if dev.Generation != 1 {
				t.Error("the caller's device was modified")
			}
		})
	}
}

// TestOnboardViaAP_AppliesOptions checks that the name, timezone and cloud
// choice reach the device over the LAN after the AP write, on both generations,
// and that a device with no route from this host gets a note naming them.
//
//nolint:paralleltest // withIsolatedConfig swaps the process-global config
func TestOnboardViaAP_AppliesOptions(t *testing.T) {
	opts := &OnboardOptions{DeviceName: "guest bath", Timezone: "America/Los_Angeles", NoCloud: true}
	for _, gen := range []int{1, 2} {
		t.Run(fmt.Sprintf("gen%d", gen), func(t *testing.T) {
			withIsolatedConfig(t)
			d := newAPDevServer(t, gen)
			svc := New(d.resolver(gen), WithRateLimiter(ratelimit.New(ratelimit.WithGen1MinInterval(0))))
			svc.ap = &apFlows{onboard: func(context.Context, *reprovision.OnboardOptions) (*reprovision.OnboardResult, error) {
				return &reprovision.OnboardResult{Address: d.addr(), Generation: gen, Reachable: true}, nil
			}}
			dev := &OnboardDevice{Name: "duo", SSID: testAPSSID, Generation: gen}

			res := svc.OnboardViaAP(context.Background(), dev, &OnboardWiFiConfig{SSID: "iot", Password: "p"}, opts)

			if res.Error != nil || res.Note != "" || !res.Registered {
				t.Fatalf("result = %+v, want registered with no note", res)
			}
			var got []string
			for _, w := range d.written() {
				switch gen {
				case 1:
					got = append(got, w.method+"?"+w.query.Encode())
				default:
					b, err := json.Marshal(w.params)
					if err != nil {
						t.Fatal(err)
					}
					got = append(got, w.method+" "+string(b))
				}
			}
			want := map[int][]string{
				1: {
					"/settings?name=guest+bath",
					"/settings?timezone=America%2FLos_Angeles",
					"/settings/cloud?enabled=false",
				},
				2: {
					`Sys.SetConfig {"config":{"device":{"name":"guest bath"}}}`,
					`Sys.SetConfig {"config":{"location":{"tz":"America/Los_Angeles"}}}`,
					`Cloud.SetConfig {"config":{"enable":false}}`,
				},
			}[gen]
			if !reflect.DeepEqual(got, want) {
				t.Errorf("device writes =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
			}
		})
	}

	t.Run("no route to the device", func(t *testing.T) {
		withIsolatedConfig(t)
		d := newAPDevServer(t, 2)
		svc := New(d.resolver(2))
		svc.ap = &apFlows{onboard: func(context.Context, *reprovision.OnboardOptions) (*reprovision.OnboardResult, error) {
			return &reprovision.OnboardResult{Address: d.addr(), Generation: 2, Reachable: false, SeenVia: "mdns"}, nil
		}}

		res := svc.OnboardViaAP(context.Background(), &OnboardDevice{Name: "plug", SSID: "ShellyPlusPlugS-AABBCC"},
			&OnboardWiFiConfig{SSID: "iot", Password: "p"}, opts)

		want := "name, timezone, cloud not applied: the device announced itself but this host has no route to it"
		if res.Error != nil || !res.Registered || res.Note != want {
			t.Errorf("result = %+v, want registered with note %q", res, want)
		}
		if n := len(d.written()); n != 0 {
			t.Errorf("%d writes reached a device this host cannot route to", n)
		}
	})

	t.Run("nothing asked writes nothing", func(t *testing.T) {
		withIsolatedConfig(t)
		d := newAPDevServer(t, 1)
		svc := New(d.resolver(1))
		svc.ap = &apFlows{onboard: func(context.Context, *reprovision.OnboardOptions) (*reprovision.OnboardResult, error) {
			return &reprovision.OnboardResult{Address: d.addr(), Generation: 1, Reachable: true}, nil
		}}
		for _, o := range []*OnboardOptions{nil, {}} {
			res := svc.OnboardViaAP(context.Background(), &OnboardDevice{Name: "duo", SSID: testAPSSID},
				&OnboardWiFiConfig{SSID: "iot", Password: "p"}, o)
			if res.Error != nil || res.Note != "" {
				t.Errorf("opts=%+v: result = %+v, want clean", o, res)
			}
		}
		if n := len(d.written()); n != 0 {
			t.Errorf("%d writes with nothing asked for", n)
		}
	})
}

// TestOnboardViaAP_OpenNetwork checks that a network is joined as an open one
// only when Open is set: an empty password without it goes to the SDK as an
// unknown password, for the SDK to look up.
func TestOnboardViaAP_OpenNetwork(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		wifi OnboardWiFiConfig
		want reprovision.Network
	}{
		{name: "open set", wifi: OnboardWiFiConfig{SSID: "guest", Open: true},
			want: reprovision.Network{SSID: "guest", Open: true}},
		{name: "open unset, no password", wifi: OnboardWiFiConfig{SSID: "guest"},
			want: reprovision.Network{SSID: "guest"}},
		{name: "open unset, password", wifi: OnboardWiFiConfig{SSID: "guest", Password: "secret"},
			want: reprovision.Network{SSID: "guest", Password: "secret"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got reprovision.Network
			svc := New(NewConfigResolver())
			svc.ap = &apFlows{onboard: func(_ context.Context, o *reprovision.OnboardOptions) (*reprovision.OnboardResult, error) {
				got = o.Network
				return &reprovision.OnboardResult{}, nil
			}}
			svc.OnboardViaAP(context.Background(), &OnboardDevice{SSID: testAPSSID}, &tt.wifi, nil)
			if got != tt.want {
				t.Errorf("network = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestOnboardViaAP_Errors(t *testing.T) {
	t.Parallel()

	hop := fmt.Errorf("AP hop for %q failed: %w", testAPSSID,
		fmt.Errorf("%w: failed to connect", reprovision.ErrAPUnreachable))
	tests := []struct {
		name     string
		err      error
		sentinel error
		want     string
	}{
		{name: "other errors keep SDK text", err: hop, sentinel: reprovision.ErrAPUnreachable, want: hop.Error()},
		{
			name: "no passphrase names the network and the flags", err: &reprovision.NoPassphraseError{SSID: "iot"},
			sentinel: reprovision.ErrNoPassphrase,
			want: `no WiFi passphrase for "iot": Shelly devices return no station key and none was found in this ` +
				`host's stored credentials — pass --password, or --open for a network that has no password`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := New(NewConfigResolver())
			svc.ap = &apFlows{onboard: func(context.Context, *reprovision.OnboardOptions) (*reprovision.OnboardResult, error) {
				return &reprovision.OnboardResult{Address: testLANAddr}, tt.err
			}}
			res := svc.OnboardViaAP(context.Background(), &OnboardDevice{SSID: testAPSSID},
				&OnboardWiFiConfig{SSID: "iot"}, nil)
			if res.Error == nil || res.Error.Error() != tt.want {
				t.Errorf("error = %v\nwant %s", res.Error, tt.want)
			}
			if !errors.Is(res.Error, tt.sentinel) {
				t.Errorf("errors.Is(%v) = false", tt.sentinel)
			}
			if res.NewAddress != "" || res.Registered {
				t.Errorf("a failed onboard reported address %q registered=%v", res.NewAddress, res.Registered)
			}
		})
	}
}

func TestInspectAtAP(t *testing.T) {
	t.Parallel()

	scanner := fakeScanner{}
	insp := &reprovision.Inspection{Model: "SHBDUO-1", Generation: 1, StaSSID: "iot", StaKeySet: true}

	t.Run("maps options and returns the inspection", func(t *testing.T) {
		t.Parallel()
		var got *reprovision.InspectOptions
		svc := New(NewConfigResolver(), WithWiFiScanner(scanner))
		svc.ap = &apFlows{inspect: func(_ context.Context, o *reprovision.InspectOptions) (*reprovision.Inspection, error) {
			got = o
			return insp, nil
		}}
		var steps []string
		ctx := WithStepReporter(context.Background(), func(s string) { steps = append(steps, s) })

		res, err := svc.InspectAtAP(ctx, testAPSSID, "192.168.33.150")
		if err != nil || res != insp {
			t.Fatalf("InspectAtAP = %+v, %v", res, err)
		}
		if got.Logger == nil || got.OnStep == nil {
			t.Fatal("Logger and OnStep must be set")
		}
		got.OnStep("reading the device")
		if len(steps) != 1 {
			t.Errorf("OnStep reached reporter with %v", steps)
		}
		if got.Scanner != scanner || got.APSSID != testAPSSID || got.APHostIP != "192.168.33.150" {
			t.Errorf("options = %+v", got)
		}
	})

	t.Run("an error returns no partial inspection", func(t *testing.T) {
		t.Parallel()
		svc := New(NewConfigResolver())
		readErr := errors.New("read device at AP: read Gen1 WiFi config: timeout")
		svc.ap = &apFlows{inspect: func(context.Context, *reprovision.InspectOptions) (*reprovision.Inspection, error) {
			return insp, readErr
		}}
		res, err := svc.InspectAtAP(context.Background(), testAPSSID, "")
		if res != nil || !errors.Is(err, readErr) {
			t.Errorf("InspectAtAP = %+v, %v; want nil, %v", res, err, readErr)
		}
	})
}

func TestStepReporter(t *testing.T) {
	t.Parallel()

	if fn := StepReporter(context.Background()); fn != nil {
		t.Error("a context without a reporter returned one")
	}
	called := ""
	fn := StepReporter(WithStepReporter(context.Background(), func(s string) { called = s }))
	if fn == nil {
		t.Fatal("the reporter was lost")
	}
	fn("checking firmware")
	if called != "checking firmware" {
		t.Errorf("reporter got %q", called)
	}
}

func TestTraceLogger(t *testing.T) {
	t.Parallel()

	n, err := traceWriter{}.Write([]byte("level=DEBUG msg=x\n"))
	if err != nil || n != len("level=DEBUG msg=x\n") {
		t.Errorf("Write = %d, %v", n, err)
	}
	if !traceLogger.Enabled(context.Background(), -4) {
		t.Error("traceLogger drops debug records")
	}

	var buf bytes.Buffer
	slog.New(slog.NewTextHandler(&buf, traceOptions)).Debug("joined", "ssid", "iot")
	if got := buf.String(); got != "level=DEBUG msg=joined ssid=iot\n" {
		t.Errorf("trace line = %q, want no time attribute", got)
	}
}

func TestFlowsDefaultToReprovision(t *testing.T) {
	t.Parallel()

	f := New(NewConfigResolver()).flows()
	if f != defaultAPFlows || f.restore == nil || f.onboard == nil || f.inspect == nil {
		t.Fatal("a service without injected flows must run the reprovision package")
	}
	ctx := context.Background()
	if _, err := f.restore(ctx, &reprovision.RestoreOptions{}); !errors.Is(err, netguard.ErrBlocked) {
		t.Errorf("default restore under test = %v, want netguard.ErrBlocked", err)
	}
	if _, err := f.onboard(ctx, &reprovision.OnboardOptions{}); !errors.Is(err, netguard.ErrBlocked) {
		t.Errorf("default onboard under test = %v, want netguard.ErrBlocked", err)
	}
	if _, err := f.inspect(ctx, &reprovision.InspectOptions{}); !errors.Is(err, netguard.ErrBlocked) {
		t.Errorf("default inspect under test = %v, want netguard.ErrBlocked", err)
	}
}

func TestPlanAPStation(t *testing.T) {
	t.Parallel()
	const hostSSID = "upstairs"
	withWiFi := func(sta string) *backup.DeviceBackup {
		return &backup.DeviceBackup{Backup: &shellybackup.Backup{
			DeviceInfo: &shellybackup.DeviceInfo{Generation: 1},
			WiFi:       json.RawMessage(`{"sta":` + sta + `}`),
		}}
	}
	stored := map[string]string{"home": "hpw", "other": "opw", hostSSID: "cpw"}
	tests := []struct {
		name      string
		bkp       *backup.DeviceBackup
		ov        *backup.NetworkOverride
		current   string
		passwords map[string]string
		want      StationKey
		wantSSID  string
		wantErr   error
		noReads   bool
	}{
		{name: "backup holds the key", bkp: withWiFi(`{"ssid":"home","key":"k"}`), want: StationKeyBackup, noReads: true},
		{name: "backup records an open network", bkp: withWiFi(`{"ssid":"home","is_open":true}`),
			want: StationKeyBackupOpen, noReads: true},
		{name: "--password", bkp: withWiFi(`{"ssid":"home"}`), ov: &backup.NetworkOverride{Password: "p"},
			want: StationKeyGiven, noReads: true},
		{name: "--open", bkp: withWiFi(`{"ssid":"home","key":"k"}`), ov: &backup.NetworkOverride{Open: true},
			want: StationKeyOpen, noReads: true},
		{name: "host has the password", bkp: withWiFi(`{"ssid":"home"}`), passwords: stored,
			want: StationKeyHost, wantSSID: "home"},
		{name: "host has no password", bkp: withWiFi(`{"ssid":"home"}`), wantErr: reprovision.ErrNoPassphrase},
		{name: "no network named takes the host's network", bkp: gen2Backup(), current: hostSSID, passwords: stored,
			want: StationKeyHost, wantSSID: hostSSID},
		{name: "no network named, host on no network", bkp: gen2Backup(), passwords: stored,
			wantErr: reprovision.ErrNoPassphrase},
		{name: "no network named, no password for the host's network", bkp: gen2Backup(), current: hostSSID,
			wantErr: reprovision.ErrNoPassphrase},
		{name: "--open with no network", bkp: gen2Backup(), ov: &backup.NetworkOverride{Open: true},
			current: hostSSID, passwords: stored, wantErr: types.ErrInvalidParam, noReads: true},
		{name: "--password with no network", bkp: gen2Backup(), ov: &backup.NetworkOverride{Password: "p"},
			current: hostSSID, passwords: stored, wantErr: types.ErrInvalidParam, noReads: true},
		{name: "changed SSID drops the backup's key", bkp: withWiFi(`{"ssid":"home","key":"k"}`),
			ov: &backup.NetworkOverride{SSID: "other"}, passwords: stored, want: StationKeyHost, wantSSID: "other"},
		{name: "changed SSID drops the backup's open state", bkp: withWiFi(`{"ssid":"home","is_open":true}`),
			ov: &backup.NetworkOverride{SSID: "other"}, passwords: stored, want: StationKeyHost, wantSSID: "other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			scanner := &testutil.CountingWiFiScanner{CurrentSSID: tt.current, Passwords: tt.passwords}
			svc := New(NewConfigResolver(), WithWiFiScanner(scanner))
			plan, err := svc.PlanAPStation(context.Background(), tt.bkp, tt.ov)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && plan.Key != tt.want {
				t.Errorf("key = %d, want %d (%s)", plan.Key, tt.want, plan)
			}
			if tt.wantSSID != "" && plan.SSID != tt.wantSSID {
				t.Errorf("SSID = %q, want %q", plan.SSID, tt.wantSSID)
			}
			if n := scanner.Moves(); n != 0 {
				t.Errorf("%d WiFi scans, joins or leaves, want 0", n)
			}
			if n := scanner.Reads(); tt.noReads && n != 0 {
				t.Errorf("%d host WiFi reads, want 0", n)
			}
		})
	}
}

// TestPlanAPStation_HopError checks that the plan refuses a missing passphrase
// with the error the hop returns for it.
func TestPlanAPStation_HopError(t *testing.T) {
	t.Parallel()
	svc := New(NewConfigResolver(), WithWiFiScanner(&testutil.CountingWiFiScanner{}))
	for _, ssid := range []string{"", "home"} {
		bkp := gen2Backup()
		var ov *backup.NetworkOverride
		if ssid != "" {
			ov = &backup.NetworkOverride{SSID: ssid}
		}
		_, err := svc.PlanAPStation(context.Background(), bkp, ov)
		want := restoreAPError(&reprovision.NoPassphraseError{SSID: ssid}, restoreTarget{}, nil)
		if err == nil || err.Error() != want.Error() {
			t.Errorf("ssid %q: err = %v, want the hop's %v", ssid, err, want)
		}
	}
}

func TestDescribeAPRestoreName(t *testing.T) {
	t.Parallel()
	bkp := gen1Backup()
	if got, want := DescribeAPRestoreName(testAPSSID, bkp, backup.RestoreOptions{AliasName: "fr"}),
		backup.DescribeName("", "fr", bkp.RecordedName(), bkp.Device().MAC, "d12965"); got != want {
		t.Errorf("own AP: %q, want %q", got, want)
	}
	if got := DescribeAPRestoreName("ShellyBulbDuo-0A0B0C", bkp, backup.RestoreOptions{AliasName: "fr"}); got != `name: "fr" from the alias (different device)` {
		t.Errorf("another device's AP: %q", got)
	}
}

// TestOnboardViaAP_DisableAP checks that DisableAP reaches the SDK onboard and
// that the SDK's report of a turned-off access point reaches the result.
func TestOnboardViaAP_DisableAP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts *OnboardOptions
		want bool
	}{
		{"asked", &OnboardOptions{DisableAP: true}, true},
		{"not asked", &OnboardOptions{}, false},
		{"nil options", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got *reprovision.OnboardOptions
			svc := New(NewConfigResolver(), WithWiFiScanner(fakeScanner{}))
			svc.ap = &apFlows{onboard: func(_ context.Context, o *reprovision.OnboardOptions) (*reprovision.OnboardResult, error) {
				got = o
				// No address, so the result is returned before any registry write.
				return &reprovision.OnboardResult{APDisabled: o.DisableAP}, nil
			}}

			res := svc.OnboardViaAP(context.Background(), &OnboardDevice{Name: "plug", SSID: "ShellyPlusPlugS-AABBCC"},
				&OnboardWiFiConfig{SSID: "iot", Password: "p"}, tt.opts)

			if got.DisableAP != tt.want || res.APDisabled != tt.want {
				t.Errorf("SDK DisableAP = %v, result APDisabled = %v, want both %v", got.DisableAP, res.APDisabled, tt.want)
			}
		})
	}
}
