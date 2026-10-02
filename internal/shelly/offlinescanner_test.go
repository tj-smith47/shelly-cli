package shelly_test

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// TestTestServicesAreOffline checks that the services the demo and the test
// factory build carry the offline WiFi scanner, so a demo or a test reaching an
// access point flow never touches the host's WiFi.
//
//nolint:paralleltest // demo.InjectIntoFactory sets the global config manager
func TestTestServicesAreOffline(t *testing.T) {
	demo, err := mock.StartWithFixtures(&mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{Devices: []mock.DeviceFixture{
			{Name: "dev", Address: "192.168.1.101", MAC: "AA:BB:CC:DD:EE:02", Type: "SNSW-001X16EU", Model: "Shelly Plus 1", Generation: 2},
		}},
	})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)
	f := cmdutil.NewFactory()
	demo.InjectIntoFactory(f)

	services := map[string]*shelly.Service{
		"demo":         f.ShellyService(),
		"test factory": factory.NewTestFactory(t).ShellyService(),
	}
	for name, svc := range services {
		if _, ok := shelly.InjectedWiFiScanner(svc).(shelly.OfflineWiFiScanner); !ok {
			t.Errorf("%s service scanner = %T, want shelly.OfflineWiFiScanner", name, shelly.InjectedWiFiScanner(svc))
		}
	}
}
