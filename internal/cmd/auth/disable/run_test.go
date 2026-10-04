package disable

import (
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// TestRun_TurnsAuthOff runs auth disable against Gen1 and Gen2 mock devices
// that require the credentials stored for them, then checks the device
// reports authentication off and answers requests without credentials.
func TestRun_TurnsAuthOff(t *testing.T) {
	t.Parallel()
	demo, err := mock.StartWithFixtures(&mock.Fixtures{Version: "1", Config: mock.ConfigFixture{
		Devices: []mock.DeviceFixture{
			{
				Name: "bath", MAC: "AA:BB:CC:DD:EE:20", Type: "SNSW-001X16EU", Model: "SNSW-001X16EU", Generation: 2,
				AuthEnabled: true, AuthUser: "admin", AuthPass: "gen2-pass",
			},
			{
				Name: "duo", MAC: "AA:BB:CC:DD:EE:21", Type: "SHBDUO-1", Model: "SHBDUO-1", Generation: 1,
				AuthEnabled: true, AuthUser: "bob", AuthPass: "gen1-pass",
			},
		},
	}})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	for _, device := range []string{"bath", "duo"} {
		t.Run(device, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			demo.InjectIntoFactory(tf.Factory)
			ctx := context.Background()
			svc := tf.ShellyService()

			if st, err := svc.GetAuthStatus(ctx, device); err != nil || !st.Enabled {
				t.Fatalf("before: status = %+v, err = %v, want enabled", st, err)
			}
			if err := run(ctx, &Options{Factory: tf.Factory, Device: device, ConfirmFlags: flags.ConfirmFlags{Yes: true}}); err != nil {
				t.Fatalf("run: %v", err)
			}
			if out := tf.OutString(); !strings.Contains(out, "Authentication disabled on "+device) {
				t.Errorf("output = %q", out)
			}
			if st, err := svc.GetAuthStatus(ctx, device); err != nil || st.Enabled {
				t.Fatalf("after: status = %+v, err = %v, want disabled", st, err)
			}
			dev, ok := demo.ConfigMgr.GetDevice(device)
			if !ok {
				t.Fatal("device left the config")
			}
			dev.Auth = nil
			if err := svc.VerifyCredentials(ctx, dev); err != nil {
				t.Fatalf("device still requires credentials: %v", err)
			}
		})
	}
}
