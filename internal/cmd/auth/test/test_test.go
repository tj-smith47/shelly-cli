package test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
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

	// Test Use
	if cmd.Use != "test <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "test <device>")
	}

	// Test Aliases
	wantAliases := []string{"verify", "check"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg valid", []string{"device"}, false},
		{"two args", []string{"device1", "device2"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name     string
		defValue string
	}{
		{"user", ""},
		{"password", ""},
		{"password-stdin", "false"},
		{"timeout", "10s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
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

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is nil")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly auth test",
		"--user",
		"--password",
		"--timeout",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

const (
	rightPassword = "r1ght-pass"
	wrongPassword = "wr0ng-pass"
	lockedDevice  = "locked"
	openDevice    = "open"
)

var macCounter atomic.Int32

// startDevices serves a password-protected and an open mock device of the
// given generation. The registry stores storedPassword for the protected one.
func startDevices(t *testing.T, generation int, storedPassword string) (*factory.TestFactory, *mock.Demo) {
	t.Helper()
	deviceType := "SNSW-001P16EU"
	if generation == 1 {
		deviceType = "SHSW-1"
	}
	// A probe refreshes the registry entry with the probed MAC in the
	// process-wide default registry, which a parallel test may have installed.
	// MACs unique to each call keep one test from rewriting another's device.
	n := macCounter.Add(1)
	lockedMAC := fmt.Sprintf("AA:BB:CC:%02X:%02X:01", generation, n)
	openMAC := fmt.Sprintf("AA:BB:CC:%02X:%02X:02", generation, n)
	demo, err := mock.StartWithFixtures(&mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name: lockedDevice, MAC: lockedMAC, Type: deviceType, Model: deviceType,
					Generation: generation, AuthEnabled: true, AuthUser: "admin", AuthPass: rightPassword,
				},
				{
					Name: openDevice, MAC: openMAC, Type: deviceType, Model: deviceType,
					Generation: generation,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	if err := demo.ConfigMgr.SetDeviceAuth(lockedDevice, "admin", storedPassword); err != nil {
		t.Fatalf("SetDeviceAuth: %v", err)
	}

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)
	return tf, demo
}

func execute(t *testing.T, tf *factory.TestFactory, args ...string) error {
	t.Helper()
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	return cmd.Execute()
}

func assertNoPasswordInOutput(t *testing.T, tf *factory.TestFactory) {
	t.Helper()
	out := tf.OutString() + tf.ErrString()
	if strings.Contains(out, rightPassword) || strings.Contains(out, wrongPassword) {
		t.Error("the password was printed")
	}
}

func TestRun_StoredPassword(t *testing.T) {
	t.Parallel()

	for _, generation := range []int{1, 2} {
		t.Run(fmt.Sprintf("gen%d right", generation), func(t *testing.T) {
			t.Parallel()
			tf, _ := startDevices(t, generation, rightPassword)

			if err := execute(t, tf, lockedDevice); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if out := tf.OutString(); !strings.Contains(out, "accepted the password for user admin") {
				t.Errorf("output = %q, want the accepted message", out)
			}
			assertNoPasswordInOutput(t, tf)
		})

		t.Run(fmt.Sprintf("gen%d wrong", generation), func(t *testing.T) {
			t.Parallel()
			tf, _ := startDevices(t, generation, wrongPassword)

			err := execute(t, tf, lockedDevice)
			if !errors.Is(err, shelly.ErrCredentialsRejected) {
				t.Fatalf("Execute() error = %v, want ErrCredentialsRejected", err)
			}
			if !strings.Contains(err.Error(), "device rejected the password for user admin") {
				t.Errorf("error = %q, want it to name the rejected user", err)
			}
			if strings.Contains(tf.OutString(), "accepted") {
				t.Errorf("output = %q, want no success message", tf.OutString())
			}
			assertNoPasswordInOutput(t, tf)
		})
	}
}

func TestRun_GivenPassword(t *testing.T) {
	t.Parallel()

	for _, generation := range []int{1, 2} {
		t.Run(fmt.Sprintf("gen%d right overrides a wrong stored one", generation), func(t *testing.T) {
			t.Parallel()
			tf, demo := startDevices(t, generation, wrongPassword)

			if err := execute(t, tf, lockedDevice, "--password", rightPassword); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if dev, ok := demo.ConfigMgr.GetDevice(lockedDevice); !ok || dev.Auth == nil || dev.Auth.Password != wrongPassword {
				t.Error("the tested password was stored; auth test must not change the registry")
			}
			assertNoPasswordInOutput(t, tf)
		})

		t.Run(fmt.Sprintf("gen%d wrong overrides a right stored one", generation), func(t *testing.T) {
			t.Parallel()
			tf, _ := startDevices(t, generation, rightPassword)

			err := execute(t, tf, lockedDevice, "--user", "admin", "--password", wrongPassword)
			if !errors.Is(err, shelly.ErrCredentialsRejected) {
				t.Fatalf("Execute() error = %v, want ErrCredentialsRejected", err)
			}
			assertNoPasswordInOutput(t, tf)
		})

		t.Run(fmt.Sprintf("gen%d from stdin", generation), func(t *testing.T) {
			t.Parallel()
			tf, _ := startDevices(t, generation, wrongPassword)
			tf.TestIO.In.WriteString(rightPassword + "\n")

			if err := execute(t, tf, lockedDevice, "--password-stdin"); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			assertNoPasswordInOutput(t, tf)
		})
	}
}

func TestRun_NoCredentialsForProtectedDevice(t *testing.T) {
	t.Parallel()

	for _, generation := range []int{1, 2} {
		t.Run(fmt.Sprintf("gen%d", generation), func(t *testing.T) {
			t.Parallel()
			tf, _ := startDevices(t, generation, "")

			if err := execute(t, tf, lockedDevice); !errors.Is(err, model.ErrAuthRequired) {
				t.Fatalf("Execute() error = %v, want ErrAuthRequired", err)
			}
		})
	}
}

func TestRun_AuthDisabledIsNotReportedAsCorrect(t *testing.T) {
	t.Parallel()

	for _, generation := range []int{1, 2} {
		t.Run(fmt.Sprintf("gen%d", generation), func(t *testing.T) {
			t.Parallel()
			tf, _ := startDevices(t, generation, rightPassword)

			if err := execute(t, tf, openDevice, "--password", wrongPassword); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			out := tf.OutString() + tf.ErrString()
			if !strings.Contains(out, "Authentication is not enabled") || strings.Contains(out, "accepted the password") {
				t.Errorf("output = %q, want the not-enabled warning and no accepted message", out)
			}
			assertNoPasswordInOutput(t, tf)
		})
	}
}

func TestRun_UserWithoutPassword(t *testing.T) {
	t.Parallel()

	tf, _ := startDevices(t, 2, rightPassword)
	if err := execute(t, tf, lockedDevice, "--user", "admin"); err == nil {
		t.Fatal("expected an error for --user without a password")
	}
}

func TestRun_DeviceNotFound(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{Version: "1", Config: mock.ConfigFixture{}}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "nonexistent-device",
		Timeout: 500 * time.Millisecond,
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}
