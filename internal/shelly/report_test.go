package shelly_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Report fixture devices: a Gen2 plug that meters power, a Gen1 bulb that
// meters power and has a firmware update, and a Gen2 device that does not
// answer. Names are chosen so map order and sorted order differ.
const (
	reportPlug    = "plug"
	reportBulb    = "bulb"
	reportOffline = "attic"
)

// reportService starts the mock with the report fixtures and returns a service
// resolving through the mock's registry, and that registry's devices. It sets
// the default registry, so its callers cannot run in parallel; they set HOME
// so no path DeviceInfo resolves reaches the real one.
func reportService(t *testing.T) (*shelly.Service, *mock.Demo) {
	t.Helper()
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{Devices: []mock.DeviceFixture{
			{Name: reportPlug, MAC: "AA:BB:CC:00:01:01", Type: "SNSW-001P16EU", Model: "Shelly Plus 1PM", Generation: 2},
			{Name: reportBulb, MAC: "AA:BB:CC:00:01:02", Type: "SHBDUO-1", Model: "Shelly Duo", Generation: 1},
			{Name: reportOffline, MAC: "AA:BB:CC:00:01:03", Type: "SNSW-001X16EU", Model: "Shelly Plus 1", Generation: 2},
		}},
		DeviceStates: map[string]mock.DeviceState{
			reportPlug: {
				"switch:0": map[string]any{"id": 0, "output": true, "apower": 45.2},
				"cloud":    map[string]any{"connected": true},
			},
			reportBulb: {
				"lights": []any{map[string]any{"ison": true, "brightness": 60}},
				"meters": []any{map[string]any{"power": 7.5, "is_valid": true}},
				"cloud":  map[string]any{"enabled": true, "connected": false},
				"ota":    map[string]any{"status": "pending", "has_update": true, "new_version": "v1.14.0"},
			},
		},
	}
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)
	if err := demo.ConfigMgr.UpdateDeviceAddress(reportOffline, refusedAddress(t)); err != nil {
		t.Fatalf("UpdateDeviceAddress: %v", err)
	}
	config.SetDefaultManager(demo.ConfigMgr)
	t.Cleanup(config.ResetDefaultManagerForTesting)
	return shelly.New(shelly.NewConfigResolver()), demo
}

// refusedAddress returns the address of a closed port, so a connection to it
// is refused at once.
func refusedAddress(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.NotFoundHandler())
	addr := srv.URL
	srv.Close()
	return addr
}

// reportContext bounds a report run: the refused connection to the offline
// device is retried until the context ends, while the mock answers at once.
func reportContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func registered(t *testing.T, demo *mock.Demo) map[string]model.Device {
	t.Helper()
	return demo.ConfigMgr.Get().Devices
}

func TestGenerateDevicesReport_Gen1Gen2AndOffline(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	svc, demo := reportService(t)

	report := svc.GenerateDevicesReport(reportContext(t), registered(t, demo))

	if report.ReportType != model.ReportTypeDevices {
		t.Errorf("ReportType = %q", report.ReportType)
	}
	names := make([]string, 0, len(report.Devices))
	for _, row := range report.Devices {
		names = append(names, row.Name)
	}
	if got := strings.Join(names, ","); got != "attic,bulb,plug" {
		t.Fatalf("rows = %s, want attic,bulb,plug (sorted by name)", got)
	}
	attic, bulb, plug := report.Devices[0], report.Devices[1], report.Devices[2]
	if attic.Online || attic.Error == "" {
		t.Errorf("attic = %+v, want offline with an error", attic)
	}
	if attic.Model != "Shelly Plus 1" || attic.Generation != 2 {
		t.Errorf("attic = %+v, want the registered model and generation", attic)
	}
	if !bulb.Online || bulb.Generation != 1 || bulb.MAC == "" {
		t.Errorf("bulb = %+v, want an online Gen1 device with its MAC", bulb)
	}
	if !plug.Online || plug.Generation != 2 || plug.Firmware == "" || plug.MAC == "" {
		t.Errorf("plug = %+v, want an online Gen2 device with firmware and MAC", plug)
	}
	for _, row := range report.Devices {
		if row.MAC != model.NormalizeMAC(row.MAC) {
			t.Errorf("%s MAC = %q, want the colon-separated form online and offline alike", row.Name, row.MAC)
		}
	}
	want := model.DevicesReportSummary{Total: 3, Online: 2, Offline: 1}
	if report.Summary != want {
		t.Errorf("Summary = %+v, want %+v", report.Summary, want)
	}
}

func TestGenerateEnergyReport_CountsGen1Meters(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	svc, demo := reportService(t)

	report := svc.GenerateEnergyReport(reportContext(t), registered(t, demo))

	byName := map[string]model.EnergyReportRow{}
	for _, row := range report.Devices {
		byName[row.Name] = row
	}
	if row := byName[reportOffline]; row.Online || row.Reporting || row.PowerW != 0 {
		t.Errorf("attic = %+v, want offline and not reporting", row)
	}
	if row := byName[reportBulb]; !row.Online || !row.Reporting || row.PowerW != 7.5 {
		t.Errorf("bulb = %+v, want online, reporting 7.5 W", row)
	}
	if row := byName[reportPlug]; !row.Online || !row.Reporting || row.PowerW != 45.2 {
		t.Errorf("plug = %+v, want online, reporting 45.2 W", row)
	}
	s := report.Summary
	if s.Total != 3 || s.Online != 2 || s.Offline != 1 || s.DevicesReporting != 2 {
		t.Errorf("Summary = %+v, want 3 total, 2 online, 1 offline, 2 reporting", s)
	}
	if s.TotalPowerW != 52.7 {
		t.Errorf("TotalPowerW = %v, want 52.7", s.TotalPowerW)
	}
}

func TestGenerateAuditReport_SharesAuditDevices(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	svc, demo := reportService(t)

	report := svc.GenerateAuditReport(reportContext(t), registered(t, demo))

	if len(report.Devices) != 3 || report.Devices[0].Name != reportOffline {
		t.Fatalf("rows = %+v, want 3 rows sorted by name", report.Devices)
	}
	attic, bulb, plug := report.Devices[0], report.Devices[1], report.Devices[2]
	if attic.Reachable || attic.AuthEnabled != nil || attic.CloudConnected != nil || attic.FirmwareOutdated != nil {
		t.Errorf("attic = %+v, want unreachable with every check null", attic)
	}
	if !bulb.Reachable || bulb.AuthEnabled == nil || *bulb.AuthEnabled {
		t.Errorf("bulb auth = %+v, want reachable with auth disabled", bulb)
	}
	if bulb.CloudConnected == nil || *bulb.CloudConnected {
		t.Errorf("bulb cloud_connected = %v, want false from the Gen1 /status cloud section", bulb.CloudConnected)
	}
	if bulb.FirmwareOutdated == nil || !*bulb.FirmwareOutdated || bulb.FirmwareAvailable != "v1.14.0" {
		t.Errorf("bulb firmware = %+v, want outdated with v1.14.0 available", bulb)
	}
	if plug.CloudConnected == nil || !*plug.CloudConnected {
		t.Errorf("plug cloud_connected = %v, want true", plug.CloudConnected)
	}
	if plug.FirmwareCurrent == "" || plug.FirmwareOutdated == nil {
		t.Errorf("plug firmware = %+v, want the current version and an outdated flag", plug)
	}

	s := report.Summary
	if s.DevicesScanned != 3 || s.Reachable != 2 || s.Unreachable != 1 {
		t.Errorf("Summary = %+v, want 3 scanned, 2 reachable, 1 unreachable", s)
	}
	if s.AuthEnabled != 0 || s.AuthDisabled != 2 || s.CloudConnected != 1 {
		t.Errorf("Summary = %+v, want auth 0/2 and 1 cloud connected", s)
	}
	outdated := 0
	for _, row := range report.Devices {
		if row.FirmwareOutdated != nil && *row.FirmwareOutdated {
			outdated++
		}
	}
	if s.OutdatedFirmware != outdated || outdated == 0 {
		t.Errorf("OutdatedFirmware = %d, rows say %d (want at least the bulb)", s.OutdatedFirmware, outdated)
	}

	// `shelly audit` runs AuditDevices; the report's rows must carry its results.
	results := svc.AuditDevices(reportContext(t), []string{reportOffline, reportBulb, reportPlug}, model.AllAuditChecks())
	for i, result := range results {
		row := model.NewAuditReportRow(result)
		if row.Name != report.Devices[i].Name || row.Reachable != report.Devices[i].Reachable ||
			len(row.Issues) != len(report.Devices[i].Issues) {
			t.Errorf("AuditDevices[%d] = %+v, report row = %+v", i, row, report.Devices[i])
		}
	}
}

// TestGenerateReports_QueryDevicesConcurrently serves every device through a
// gate that holds each device's first request until all devices have sent
// one. Asked one at a time, the first device waits for the others until the
// gate gives up and answers 503, so a device reads as offline.
func TestGenerateReports_QueryDevicesConcurrently(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	svc, demo := reportService(t)
	devices := []string{reportPlug, reportBulb}
	inner := demo.DeviceServer.Config.Handler

	for _, generate := range []struct {
		name   string
		online func(map[string]model.Device) int
	}{
		{"devices", func(d map[string]model.Device) int {
			return svc.GenerateDevicesReport(context.Background(), d).Summary.Online
		}},
		{"energy", func(d map[string]model.Device) int {
			return svc.GenerateEnergyReport(context.Background(), d).Summary.Online
		}},
		{"audit", func(d map[string]model.Device) int {
			return svc.GenerateAuditReport(context.Background(), d).Summary.Reachable
		}},
	} {
		gate := newArrivalGate(len(devices), 5*time.Second, inner)
		srv := httptest.NewServer(gate)
		for _, name := range devices {
			if err := demo.ConfigMgr.UpdateDeviceAddress(name, srv.URL+"/devices/"+name); err != nil {
				t.Fatalf("UpdateDeviceAddress: %v", err)
			}
		}
		subset := map[string]model.Device{}
		for _, name := range devices {
			subset[name] = registered(t, demo)[name]
		}

		if online := generate.online(subset); online != len(devices) {
			t.Errorf("%s report: %d of %d devices answered; the devices were not queried at the same time",
				generate.name, online, len(devices))
		}
		srv.Close()
	}
}

// arrivalGate holds the first request of each device until want devices have
// sent one, then passes every request to next. A device whose first request
// waits longer than timeout gets 503 for that and every later request.
type arrivalGate struct {
	next    http.Handler
	want    int
	timeout time.Duration
	all     chan struct{}

	mu      sync.Mutex
	arrived map[string]bool
	failed  map[string]bool
}

func newArrivalGate(want int, timeout time.Duration, next http.Handler) *arrivalGate {
	return &arrivalGate{
		next: next, want: want, timeout: timeout, all: make(chan struct{}),
		arrived: map[string]bool{}, failed: map[string]bool{},
	}
}

func (g *arrivalGate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	device, _, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/devices/"), "/")
	g.mu.Lock()
	first := !g.arrived[device]
	if first {
		g.arrived[device] = true
		if len(g.arrived) == g.want {
			close(g.all)
		}
	}
	failed := g.failed[device]
	g.mu.Unlock()

	if failed {
		http.Error(w, "gate closed", http.StatusServiceUnavailable)
		return
	}
	if first {
		select {
		case <-g.all:
		case <-time.After(g.timeout):
			g.mu.Lock()
			g.failed[device] = true
			g.mu.Unlock()
			http.Error(w, "gate closed", http.StatusServiceUnavailable)
			return
		}
	}
	g.next.ServeHTTP(w, r)
}

func TestForEachDevice_GivesEachDeviceItsOwnTimeout(t *testing.T) {
	t.Parallel()

	devices := []string{"a", "b", "c"}
	deadlines := make([]time.Time, len(devices))
	got := make([]string, len(devices))
	shelly.ForEachDevice(context.Background(), devices, time.Minute, func(ctx context.Context, i int, device string) {
		deadlines[i], _ = ctx.Deadline()
		got[i] = device
	})
	for i, device := range devices {
		if got[i] != device {
			t.Errorf("slot %d = %q, want %q", i, got[i], device)
		}
		if deadlines[i].IsZero() {
			t.Errorf("device %q ran without a deadline", device)
		}
	}
}
