package rotate

import (
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/auth"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// TestRun_GeneratedPasswordIsStoredAndWorks runs auth rotate --generate
// against Gen1 and Gen2 mock devices that keep only the login they are given.
// It is the only test in this package that installs a global config manager.
func TestRun_GeneratedPasswordIsStoredAndWorks(t *testing.T) {
	t.Parallel()
	demo, err := mock.StartWithFixtures(&mock.Fixtures{Version: "1", Config: mock.ConfigFixture{
		Devices: []mock.DeviceFixture{
			{
				Name: "bath", MAC: "AA:BB:CC:DD:EE:11", Type: "SNSW-001X16EU", Model: "SNSW-001X16EU", Generation: 2,
				AuthEnabled: true, AuthUser: auth.DefaultUser, AuthPass: "old-pass",
			},
			{
				Name: "duo", MAC: "AA:BB:CC:DD:EE:12", Type: "SHBDUO-1", Model: "SHBDUO-1", Generation: 1,
				AuthEnabled: true, AuthUser: "bob", AuthPass: "old-pass",
			},
			{
				Name: "porch", MAC: "AA:BB:CC:DD:EE:13", Type: "SNSW-001X16EU", Model: "SNSW-001X16EU", Generation: 2,
				AuthEnabled: true, AuthUser: auth.DefaultUser, AuthPass: "old-pass",
			},
		},
	}})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)
	t.Run("porch password-stdin", func(t *testing.T) {
		t.Parallel()
		tf := factory.NewTestFactory(t)
		demo.InjectIntoFactory(tf.Factory)
		tf.TestIO.In.WriteString("stdin-pass\r\n")
		cmd := NewCommand(tf.Factory)
		cmd.SetArgs([]string{"porch", "--password-stdin"})
		cmd.SetContext(context.Background())
		if err := cmd.Execute(); err != nil {
			t.Fatalf("execute: %v", err)
		}
		dev, ok := demo.ConfigMgr.GetDevice("porch")
		if !ok || dev.Auth == nil || dev.Auth.Password != "stdin-pass" {
			t.Fatal("password read from stdin was not stored exactly (line ending must be dropped)")
		}
		if out := tf.OutString() + tf.ErrString(); strings.Contains(out, "stdin-pass") {
			t.Errorf("password printed:\n%s", out)
		}
		if err := tf.ShellyService().VerifyCredentials(context.Background(), dev); err != nil {
			t.Fatalf("device refuses the password read from stdin: %v", err)
		}
	})
	for device, user := range map[string]string{"bath": auth.DefaultUser, "duo": "bob"} {
		t.Run(device, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			demo.InjectIntoFactory(tf.Factory)

			if err := run(context.Background(), &Options{Factory: tf.Factory, Device: device, Generate: true, Length: 20}); err != nil {
				t.Fatalf("run: %v", err)
			}
			dev, ok := demo.ConfigMgr.GetDevice(device)
			// Without --user the stored user is kept: bob on the Gen1 device.
			if !ok || dev.Auth == nil || len(dev.Auth.Password) != 20 || dev.Auth.Username != user {
				t.Fatalf("stored auth = %+v, want user %s and the generated 20-character password", dev.Auth, user)
			}
			if out := tf.OutString(); !strings.Contains(out, "User: "+user) {
				t.Errorf("output does not name user %s:\n%s", user, out)
			}
			out := tf.OutString() + tf.ErrString()
			if strings.Contains(out, dev.Auth.Password) || strings.Contains(out, "old-pass") {
				t.Errorf("password printed without --show:\n%s", out)
			}
			if !strings.Contains(out, "Saved the new credentials for "+device) || strings.Contains(out, "Update your stored credentials") {
				t.Errorf("output does not report the stored credentials:\n%s", out)
			}
			if err := tf.ShellyService().VerifyCredentials(context.Background(), dev); err != nil {
				t.Fatalf("device refuses the stored credentials: %v", err)
			}
			old := dev
			old.Auth = &model.Auth{Username: user, Password: "old-pass"}
			if err := tf.ShellyService().VerifyCredentials(context.Background(), old); err == nil {
				t.Fatal("device still accepts the old password")
			}
		})
	}
}
