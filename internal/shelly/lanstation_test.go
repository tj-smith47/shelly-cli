package shelly

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
	"github.com/tj-smith47/shelly-go/reprovision"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// testStation is a backup station, written in each generation's spelling.
type testStation struct {
	ssid     string
	disabled bool
	key      string // Gen1 only
	open     bool   // Gen2 only
	staticIP string // static address with no gateway or netmask
}

func (s testStation) json(generation int) string {
	if generation == 1 {
		if s.staticIP != "" {
			return fmt.Sprintf(`{"enabled":%t,"ssid":%q,"key":%q,"ipv4_method":"static","ip":%q}`,
				!s.disabled, s.ssid, s.key, s.staticIP)
		}
		return fmt.Sprintf(`{"enabled":%t,"ssid":%q,"key":%q}`, !s.disabled, s.ssid, s.key)
	}
	if s.staticIP != "" {
		return fmt.Sprintf(`{"enable":%t,"ssid":%q,"is_open":%t,"ipv4mode":"static","ip":%q}`,
			!s.disabled, s.ssid, s.open, s.staticIP)
	}
	return fmt.Sprintf(`{"enable":%t,"ssid":%q,"is_open":%t}`, !s.disabled, s.ssid, s.open)
}

func stationBackup(generation int, sta testStation, sta1 *testStation) *backup.DeviceBackup {
	wifi := `{"sta":` + sta.json(generation)
	if sta1 != nil {
		wifi += `,"sta1":` + sta1.json(generation)
	}
	wifi += "}"
	return &backup.DeviceBackup{Backup: &shellybackup.Backup{
		DeviceInfo: &shellybackup.DeviceInfo{Generation: generation},
		WiFi:       json.RawMessage(wifi),
	}}
}

func hostLookups(scanner *recordingScanner) []string {
	var out []string
	for _, c := range scanner.called() {
		if ssid, ok := strings.CutPrefix(c, "password:"); ok {
			out = append(out, ssid)
		}
	}
	return out
}

func TestPlanLANStation(t *testing.T) {
	t.Parallel()
	home := testStation{ssid: "home"}
	tests := []struct {
		name       string
		gen        int // 0: both generations
		sta        testStation
		sta1       *testStation
		deviceSSID string
		deviceSta1 string
		noStation  bool
		readFail   bool
		hostPass   map[string]string
		ov         *backup.NetworkOverride
		skip       bool

		wantKey     StationKey
		wantPass    string
		wantOpen    bool
		wantSameOv  bool // the override comes back as passed
		wantRead    bool
		wantLookups []string
		wantErr     error // matched with errors.Is
		wantSecond  StationKey
		wantSt1     backup.Station1Write
	}{
		{name: "same network keeps the key", sta: home, deviceSSID: "home", wantKey: StationKeyKept,
			wantSameOv: true, wantRead: true},
		{name: "changed network takes the host's key", sta: home, deviceSSID: "old", hostPass: map[string]string{"home": "stored"},
			wantKey: StationKeyHost, wantPass: "stored", wantRead: true, wantLookups: []string{"home"}},
		{name: "changed network without a key is refused", sta: home, deviceSSID: "old",
			wantErr: reprovision.ErrNoPassphrase, wantRead: true, wantLookups: []string{"home"}},
		{name: "ssid override changes the network", sta: home, deviceSSID: "home", ov: &backup.NetworkOverride{SSID: "guest"},
			hostPass: map[string]string{"guest": "gp"}, wantKey: StationKeyHost, wantPass: "gp", wantRead: true,
			wantLookups: []string{"guest"}},
		{name: "password flag", sta: home, deviceSSID: "old", ov: &backup.NetworkOverride{Password: "given"},
			wantKey: StationKeyGiven, wantPass: "given", wantSameOv: true},
		{name: "open flag", sta: home, deviceSSID: "old", ov: &backup.NetworkOverride{Open: true},
			wantKey: StationKeyOpen, wantOpen: true, wantSameOv: true},
		{name: "skip network", sta: home, deviceSSID: "old", skip: true, wantKey: StationKeySkipped, wantSameOv: true},
		{name: "device has no station", sta: home, noStation: true, hostPass: map[string]string{"home": "stored"},
			wantKey: StationKeyHost, wantPass: "stored", wantRead: true, wantLookups: []string{"home"}},
		{name: "device has no station and no key is known", sta: home, noStation: true,
			wantErr: reprovision.ErrNoPassphrase, wantRead: true, wantLookups: []string{"home"}},
		{name: "device read fails", sta: home, readFail: true, hostPass: map[string]string{"home": "stored"},
			wantErr: errAnyRead, wantRead: true},

		{name: "disabled station is written disabled", sta: testStation{ssid: "home", disabled: true}, deviceSSID: "old",
			wantKey: StationKeyDisabled, wantSameOv: true},
		{name: "disabled station with --ssid is planned", sta: testStation{ssid: "home", disabled: true}, deviceSSID: "old",
			ov: &backup.NetworkOverride{SSID: "guest"}, hostPass: map[string]string{"guest": "gp"},
			wantKey: StationKeyHost, wantPass: "gp", wantRead: true, wantLookups: []string{"guest"}},

		{name: "no network anywhere", sta: testStation{}, wantKey: StationKeyAbsent, wantSameOv: true},
		{name: "--open with no network", sta: testStation{}, ov: &backup.NetworkOverride{Open: true},
			wantErr: types.ErrInvalidParam},
		{name: "--password with no network", sta: testStation{}, ov: &backup.NetworkOverride{Password: "pw"},
			wantErr: types.ErrInvalidParam},
		{name: "--static-ip on a DHCP backup without --gateway/--netmask", sta: home, deviceSSID: "home",
			ov: &backup.NetworkOverride{StaticIP: "10.0.0.5"}, wantErr: shellybackup.ErrIncompleteStaticNetwork},
		{name: "backup static station without gateway or netmask", sta: testStation{ssid: "home", staticIP: "10.0.0.6"},
			deviceSSID: "home", wantErr: shellybackup.ErrIncompleteStaticNetwork},
		{name: "--static-ip with no network", sta: testStation{},
			ov:      &backup.NetworkOverride{StaticIP: "10.0.0.9", Gateway: testGateway, Netmask: testNetmask},
			wantErr: types.ErrInvalidParam},

		{name: "open backup station is joined open", gen: 2, sta: testStation{ssid: "home", open: true}, deviceSSID: "old",
			wantKey: StationKeyBackupOpen, wantOpen: true},
		{name: "open backup station, same --ssid", gen: 2, sta: testStation{ssid: "home", open: true}, deviceSSID: "old",
			ov: &backup.NetworkOverride{SSID: "home"}, wantKey: StationKeyBackupOpen, wantOpen: true},
		{name: "open backup station, other --ssid", gen: 2, sta: testStation{ssid: "home", open: true}, deviceSSID: "old",
			ov: &backup.NetworkOverride{SSID: "guest"}, hostPass: map[string]string{"guest": "gp"},
			wantKey: StationKeyHost, wantPass: "gp", wantRead: true, wantLookups: []string{"guest"}},
		{name: "open backup station, --password", gen: 2, sta: testStation{ssid: "home", open: true}, deviceSSID: "old",
			ov: &backup.NetworkOverride{Password: "pw"}, wantKey: StationKeyGiven, wantPass: "pw", wantSameOv: true},
		{name: "gen1 backup key is written", gen: 1, sta: testStation{ssid: "home", key: "bk"}, deviceSSID: "old",
			wantKey: StationKeyBackup, wantPass: "bk"},
		{name: "gen1 backup key, --ssid elsewhere: host key", gen: 1, sta: testStation{ssid: "home", key: "bk"},
			deviceSSID: "old", ov: &backup.NetworkOverride{SSID: "guest"}, hostPass: map[string]string{"guest": "gp"},
			wantKey: StationKeyHost, wantPass: "gp", wantRead: true, wantLookups: []string{"guest"}},
		{name: "gen1 backup key, --ssid elsewhere, no host key: refused", gen: 1,
			sta: testStation{ssid: "home", key: "bk"}, deviceSSID: "old", ov: &backup.NetworkOverride{SSID: "guest"},
			wantErr: reprovision.ErrNoPassphrase, wantRead: true, wantLookups: []string{"guest"}},
		{name: "gen1 backup key, --ssid of the device: kept", gen: 1, sta: testStation{ssid: "home", key: "bk"},
			deviceSSID: "guest", ov: &backup.NetworkOverride{SSID: "guest"}, wantKey: StationKeyKept,
			wantSameOv: true, wantRead: true},
		{name: "gen1 backup key loses to --password", gen: 1, sta: testStation{ssid: "home", key: "bk"}, deviceSSID: "old",
			ov: &backup.NetworkOverride{Password: "pw"}, wantKey: StationKeyGiven, wantPass: "pw", wantSameOv: true},

		{name: "secondary on the same network", sta: home, sta1: &testStation{ssid: "spare"}, deviceSSID: "home",
			deviceSta1: "spare", wantKey: StationKeyKept, wantSameOv: true, wantRead: true, wantSecond: StationKeyKept},
		{name: "secondary changed, host key", sta: home, sta1: &testStation{ssid: "spare"}, deviceSSID: "home",
			deviceSta1: "old", hostPass: map[string]string{"spare": "sp"}, wantKey: StationKeyKept, wantSameOv: true,
			wantRead: true, wantLookups: []string{"spare"}, wantSecond: StationKeyHost,
			wantSt1: backup.Station1Write{Password: "sp"}},
		{name: "secondary changed, no key: left out", sta: home, sta1: &testStation{ssid: "spare"}, deviceSSID: "home",
			deviceSta1: "old", wantKey: StationKeyKept, wantSameOv: true, wantRead: true, wantLookups: []string{"spare"},
			wantSecond: StationKeyLeftOut, wantSt1: backup.Station1Write{Omit: true, SSID: "spare"}},
		{name: "secondary, device read fails: left out", sta: home, sta1: &testStation{ssid: "spare"}, readFail: true,
			ov: &backup.NetworkOverride{Password: "pw"}, wantKey: StationKeyGiven, wantPass: "pw", wantSameOv: true,
			wantRead: true, wantSecond: StationKeyUnread,
			wantSt1: backup.Station1Write{Omit: true, SSID: "spare", Unread: true}},
		{name: "secondary disabled", sta: home, sta1: &testStation{ssid: "spare", disabled: true}, deviceSSID: "home",
			wantKey: StationKeyKept, wantSameOv: true, wantRead: true, wantSecond: StationKeyDisabled},
		{name: "secondary open", gen: 2, sta: home, sta1: &testStation{ssid: "spare", open: true}, deviceSSID: "home",
			wantKey: StationKeyKept, wantSameOv: true, wantRead: true, wantSecond: StationKeyBackupOpen,
			wantSt1: backup.Station1Write{Open: true}},
		{name: "secondary gen1 key", gen: 1, sta: home, sta1: &testStation{ssid: "spare", key: "k1"}, deviceSSID: "home",
			wantKey: StationKeyKept, wantSameOv: true, wantRead: true, wantSecond: StationKeyBackup},
	}
	for _, gen := range []int{1, 2} {
		for _, tt := range tests {
			if tt.gen != 0 && tt.gen != gen {
				continue
			}
			t.Run(fmt.Sprintf("gen%d/%s", gen, tt.name), func(t *testing.T) {
				t.Parallel()
				d := newAPDevServer(t, gen)
				d.staSSID, d.sta1SSID, d.noStation, d.wifiReadFail = tt.deviceSSID, tt.deviceSta1, tt.noStation, tt.readFail
				scanner := &recordingScanner{passwords: tt.hostPass}
				svc := New(d.resolver(gen), WithRateLimiter(ratelimit.New()), WithWiFiScanner(scanner))

				ov, plan, err := svc.PlanLANStation(context.Background(), "apdev", stationBackup(gen, tt.sta, tt.sta1), tt.ov, tt.skip)
				if got := hostLookups(scanner); !slices.Equal(got, tt.wantLookups) {
					t.Errorf("host lookups = %v, want %v", got, tt.wantLookups)
				}
				if read := d.wifiReads.Load() > 0; read != tt.wantRead {
					t.Errorf("device read = %v, want %v", read, tt.wantRead)
				}
				switch {
				case errors.Is(tt.wantErr, errAnyRead):
					if err == nil || errors.Is(err, reprovision.ErrNoPassphrase) || !strings.Contains(err.Error(), "read the device's WiFi station") {
						t.Fatalf("err = %v, want the device read error", err)
					}
					return
				case tt.wantErr != nil:
					if !errors.Is(err, tt.wantErr) {
						t.Fatalf("err = %v, want %v", err, tt.wantErr)
					}
					switch {
					case errors.Is(err, shellybackup.ErrIncompleteStaticNetwork):
						if !strings.Contains(err.Error(), "--gateway") || strings.Contains(err.Error(), "Network.") {
							t.Errorf("err = %v, want it to name --gateway and no SDK field", err)
						}
					case errors.Is(err, types.ErrInvalidParam) && !strings.Contains(err.Error(), "--ssid"):
						t.Errorf("err = %v, want it to name --ssid", err)
					}
					return
				}
				if err != nil {
					t.Fatalf("PlanLANStation() error = %v", err)
				}
				if plan.Key != tt.wantKey {
					t.Errorf("plan = %v, want key %d", plan, tt.wantKey)
				}
				if tt.wantSameOv && ov != tt.ov {
					t.Errorf("override = %+v, want the one passed (%+v)", ov, tt.ov)
				}
				var pass string
				var open bool
				if ov != nil {
					pass, open = ov.Password, ov.Open
				}
				if pass != tt.wantPass || open != tt.wantOpen {
					t.Errorf("override password %q open %v, want %q %v", pass, open, tt.wantPass, tt.wantOpen)
				}
				gotSecond := StationKey(0)
				if plan.Secondary != nil {
					gotSecond = plan.Secondary.Key
				}
				if (tt.sta1 != nil) != (plan.Secondary != nil) || gotSecond != tt.wantSecond {
					t.Errorf("secondary plan = %+v, want key %d", plan.Secondary, tt.wantSecond)
				}
				if got := plan.Station1Write(); got != tt.wantSt1 {
					t.Errorf("Station1Write() = %+v, want %+v", got, tt.wantSt1)
				}
				if (tt.wantSecond == StationKeyLeftOut || tt.wantSecond == StationKeyUnread) &&
					!strings.Contains(plan.String(), "left out") {
					t.Errorf("plan line %q does not say the secondary station is left out", plan.String())
				}
			})
		}
	}
}

// errAnyRead marks a row that expects the device read error.
var errAnyRead = errors.New("device read")

func TestLANStationPlan_String(t *testing.T) {
	t.Parallel()
	for key := StationKeySkipped; key <= StationKeyUnread; key++ {
		if (LANStationPlan{SSID: "home", Key: key}).String() == "" {
			t.Errorf("key %d has no description", key)
		}
	}
	notWritten := []StationKey{StationKeySkipped, StationKeyAbsent}
	for key := StationKeySkipped; key <= StationKeyOpen; key++ {
		says := strings.Contains(LANStationPlan{SSID: "home", Key: key}.String(), "not written")
		if says != slices.Contains(notWritten, key) {
			t.Errorf("key %d: %q, \"not written\" only for a plan that writes no station", key,
				LANStationPlan{SSID: "home", Key: key}.String())
		}
	}
}

// TestPlanLANStation_MatchesToAP checks the LAN plan against the SDK's
// reprovision.MergeNetwork, the rule a --to-ap restore joins by, for one
// backup and one set of flags: both must name the same network with the same
// key or open state. A merged network with no key falls to this host's stored
// passphrase on both paths.
func TestPlanLANStation_MatchesToAP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		gen  int
		sta  testStation
		ov   *backup.NetworkOverride
	}{
		{name: "open backup station", gen: 2, sta: testStation{ssid: "home", open: true}},
		{name: "open backup station, --ssid elsewhere", gen: 2, sta: testStation{ssid: "home", open: true},
			ov: &backup.NetworkOverride{SSID: "guest"}},
		{name: "gen1 backup key", gen: 1, sta: testStation{ssid: "home", key: "bk"}},
		{name: "gen1 backup key, --ssid elsewhere", gen: 1, sta: testStation{ssid: "home", key: "bk"},
			ov: &backup.NetworkOverride{SSID: "guest"}},
		{name: "gen1 backup key, --password", gen: 1, sta: testStation{ssid: "home", key: "bk"},
			ov: &backup.NetworkOverride{Password: "pw"}},
		{name: "--open", gen: 2, sta: testStation{ssid: "home"}, ov: &backup.NetworkOverride{Open: true}},
		{name: "changed network, host key", gen: 2, sta: testStation{ssid: "home"}},
		{name: "changed network, host key (gen1)", gen: 1, sta: testStation{ssid: "home"}},
	}
	hostPass := map[string]string{"home": "stored", "guest": "gp"}
	station := func(ssid, pass string, open bool) string {
		if open {
			return ssid + ":open"
		}
		return ssid + ":" + pass
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bkp := stationBackup(tt.gen, tt.sta, nil)

			d := newAPDevServer(t, tt.gen)
			d.staSSID = "elsewhere"
			lan := New(d.resolver(tt.gen), WithRateLimiter(ratelimit.New()),
				WithWiFiScanner(&recordingScanner{passwords: hostPass}))
			ov, plan, err := lan.PlanLANStation(context.Background(), "apdev", bkp, tt.ov, false)
			if err != nil {
				t.Fatalf("PlanLANStation() error = %v", err)
			}
			var lanPass string
			var lanOpen bool
			if ov != nil {
				lanPass, lanOpen = ov.Password, ov.Open
			}

			fromBackup := reprovision.NetworkFromBackup(bkp.Backup)
			given := tt.ov.Network()
			merged, err := reprovision.MergeNetwork(&fromBackup, &given)
			if err != nil {
				t.Fatalf("MergeNetwork() error = %v", err)
			}
			if merged.Password == "" && !merged.Open {
				merged.Password = hostPass[merged.SSID]
			}
			want := station(merged.SSID, merged.Password, merged.Open)
			if got := station(plan.SSID, lanPass, lanOpen); got != want {
				t.Errorf("LAN writes %q, --to-ap joins %q (plan %v)", got, want, plan)
			}
		})
	}
}

// TestPlanLANStation_Address asserts the plan line names the address the
// restore writes: a static backup station hands its address to the target
// unless --static-ip replaces it.
func TestPlanLANStation_Address(t *testing.T) {
	t.Parallel()
	gen1Static := `{"sta":{"enabled":true,"ssid":"home","ipv4_method":"static","ip":"10.0.0.5","gw":"10.0.0.1","mask":"255.255.255.0"}}`
	gen2Static := `{"sta":{"enable":true,"ssid":"home","ipv4mode":"static","ip":"10.0.0.5","gw":"10.0.0.1","netmask":"255.255.255.0"},` +
		`"sta1":{"enable":true,"ssid":"spare","ipv4mode":"dhcp"}}`
	tests := []struct {
		name string
		gen  int
		wifi string
		ov   *backup.NetworkOverride
		want string
	}{
		{name: "gen1 static backup, no flags", gen: 1, wifi: gen1Static,
			want: "static address 10.0.0.5 (gateway 10.0.0.1, netmask 255.255.255.0) from the backup"},
		{name: "gen2 static backup, no flags", gen: 2, wifi: gen2Static,
			want: "static address 10.0.0.5 (gateway 10.0.0.1, netmask 255.255.255.0) from the backup; " +
				`secondary WiFi station "spare"`},
		{name: "gen2 static backup, --static-ip", gen: 2, wifi: gen2Static, ov: &backup.NetworkOverride{StaticIP: "10.0.0.9"},
			want: "static address 10.0.0.9 (gateway 10.0.0.1, netmask 255.255.255.0) from --static-ip"},
		{name: "gen2 dhcp backup", gen: 2, wifi: `{"sta":{"enable":true,"ssid":"home","ipv4mode":"dhcp"}}`, want: "; DHCP"},
		{name: "gen2 backup with no mode", gen: 2, wifi: `{"sta":{"enable":true,"ssid":"home"}}`, want: "; address unchanged"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newAPDevServer(t, tt.gen)
			d.staSSID, d.sta1SSID = "home", "spare"
			svc := New(d.resolver(tt.gen), WithRateLimiter(ratelimit.New()), WithWiFiScanner(&recordingScanner{}))
			bkp := &backup.DeviceBackup{Backup: &shellybackup.Backup{
				DeviceInfo: &shellybackup.DeviceInfo{Generation: tt.gen}, WiFi: json.RawMessage(tt.wifi)}}
			_, plan, err := svc.PlanLANStation(context.Background(), "apdev", bkp, tt.ov, false)
			if err != nil {
				t.Fatalf("PlanLANStation() error = %v", err)
			}
			if !strings.Contains(plan.String(), tt.want) {
				t.Errorf("plan = %q, want it to contain %q", plan.String(), tt.want)
			}
		})
	}
}
