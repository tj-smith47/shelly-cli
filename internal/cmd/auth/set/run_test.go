package set

import (
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/auth"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const (
	runTestPassword = "n3w-set-pass"
	runTestGen1User = "bob"
)

// TestRun_SetsStoresAndVerifies runs auth set against Gen1 and Gen2 mock
// devices that keep only the login they are given, as real devices do. It is
// the only test in this package that installs a global config manager.
func TestRun_SetsStoresAndVerifies(t *testing.T) {
	t.Parallel()
	demo, err := mock.StartWithFixtures(&mock.Fixtures{Version: "1", Config: mock.ConfigFixture{
		Devices: []mock.DeviceFixture{
			{Name: "bath", MAC: "AA:BB:CC:DD:EE:10", Type: "SNSW-001X16EU", Model: "SNSW-001X16EU", Generation: 2},
			{Name: "porch", MAC: "AA:BB:CC:DD:EE:11", Type: "SNSW-001X16EU", Model: "SNSW-001X16EU", Generation: 2},
			{Name: "duo", MAC: "AA:BB:CC:DD:EE:12", Type: "SHBDUO-1", Model: "SHBDUO-1", Generation: 1},
			{Name: "lobby", MAC: "AA:BB:CC:DD:EE:13", Type: "SNSW-001X16EU", Model: "SNSW-001X16EU", Generation: 2},
		},
	}})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	for _, tc := range []struct {
		device, user, stdin string
		args                []string
	}{
		{device: "bath", user: auth.DefaultUser},
		{device: "duo", user: runTestGen1User},
		{device: "porch", user: auth.DefaultUser, args: []string{"porch", "--password", runTestPassword}},
		{device: "lobby", user: auth.DefaultUser, stdin: runTestPassword + "\n", args: []string{"lobby", "--password-stdin"}},
	} {
		t.Run(tc.device, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			demo.InjectIntoFactory(tf.Factory)
			tf.TestIO.In.WriteString(tc.stdin)

			if tc.args != nil {
				cmd := NewCommand(tf.Factory)
				cmd.SetArgs(tc.args)
				cmd.SetContext(context.Background())
				if err := cmd.Execute(); err != nil {
					t.Fatalf("execute %v: %v", tc.args, err)
				}
			} else if err := run(context.Background(), &Options{Factory: tf.Factory, Device: tc.device, User: tc.user, Password: runTestPassword}); err != nil {
				t.Fatalf("run: %v", err)
			}
			out := tf.OutString() + tf.ErrString()
			for _, want := range []string{"Authentication enabled on " + tc.device, "Saved the new credentials for " + tc.device} {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q:\n%s", want, out)
				}
			}
			if strings.Contains(out, runTestPassword) {
				t.Errorf("password printed:\n%s", out)
			}

			dev, ok := demo.ConfigMgr.GetDevice(tc.device)
			if !ok || dev.Auth == nil || dev.Auth.Password != runTestPassword || dev.Auth.Username != tc.user {
				t.Fatalf("stored auth = %+v, want %s and the new password", dev.Auth, tc.user)
			}
			if err := tf.ShellyService().VerifyCredentials(context.Background(), dev); err != nil {
				t.Fatalf("device refuses the stored credentials: %v", err)
			}
			dev.Auth = &model.Auth{Username: tc.user, Password: "other"}
			if err := tf.ShellyService().VerifyCredentials(context.Background(), dev); err == nil {
				t.Fatal("device accepts any password after auth set")
			}
		})
	}
}
