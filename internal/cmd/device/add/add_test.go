package add

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const (
	testAddress   = "192.168.1.100"
	testPassword  = "s3cret-pass"
	wrongPassword = "wr0ng-pass"
	testPlatform  = "tasmota"
)

func execute(t *testing.T, tf *factory.TestFactory, args ...string) error {
	t.Helper()
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs(args)
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)
	return cmd.Execute()
}

func storedDevice(t *testing.T, tf *factory.TestFactory, name string) model.Device {
	t.Helper()
	dev, ok := tf.Manager.GetDevice(name)
	if !ok {
		t.Fatalf("device %q was not registered", name)
	}
	return dev
}

func assertNoPasswordInOutput(t *testing.T, tf *factory.TestFactory) {
	t.Helper()
	out := tf.OutString() + tf.ErrString()
	if strings.Contains(out, testPassword) || strings.Contains(out, wrongPassword) {
		t.Error("the password was printed")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(factory.NewTestFactory(t).Factory)
	if cmd.Use != "add <name> <address>" {
		t.Errorf("Use = %q", cmd.Use)
	}
	if len(cmd.Aliases) == 0 || cmd.Example == "" || cmd.Long == "" {
		t.Error("Aliases, Example and Long are required")
	}
	for _, name := range []string{"auth", "user", "password", "password-stdin", "force", "platform", "generation", "no-verify"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("flag --%s is missing", name)
		}
	}
}

func TestRun_ConflictingFlagsRejected(t *testing.T) {
	t.Parallel()

	cases := [][]string{
		{"--auth", "admin:x", "--password", "y"},
		{"--auth", "admin:x", "--user", "bob"},
		{"--auth", "admin:x", "--password-stdin"},
		{"--password", "x", "--password-stdin"},
		{"--platform", testPlatform, "--generation", "2"},
	}
	for _, flags := range cases {
		t.Run(strings.Join(flags, " "), func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			args := append([]string{"kitchen", testAddress, "--no-verify"}, flags...)
			if err := execute(t, tf, args...); err == nil {
				t.Fatal("expected an error for conflicting flags")
			}
			if _, ok := tf.Manager.GetDevice("kitchen"); ok {
				t.Error("device registered despite the error")
			}
		})
	}
}

func TestRun_PasswordDefaultsUserToAdmin(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	if err := execute(t, tf, "kitchen", testAddress, "--no-verify", "--password", testPassword); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	dev := storedDevice(t, tf, "kitchen")
	if dev.Auth == nil || dev.Auth.Username != "admin" || dev.Auth.Password != testPassword {
		t.Errorf("stored auth = %+v, want admin with the given password", dev.Auth)
	}
	assertNoPasswordInOutput(t, tf)
}

func TestRun_UserAndPassword(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	if err := execute(t, tf, "kitchen", testAddress, "--no-verify", "--user", "operator", "--password", testPassword); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if auth := storedDevice(t, tf, "kitchen").Auth; auth == nil || auth.Username != "operator" {
		t.Errorf("stored auth = %+v, want user operator", auth)
	}
}

func TestRun_PasswordStdin(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	tf.TestIO.In.WriteString(testPassword + "\r\n")
	if err := execute(t, tf, "kitchen", testAddress, "--no-verify", "--user", "admin", "--password-stdin"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if auth := storedDevice(t, tf, "kitchen").Auth; auth == nil || auth.Password != testPassword {
		t.Error("password read from stdin was not stored exactly (line ending must be dropped)")
	}
	assertNoPasswordInOutput(t, tf)
}

func TestRun_PasswordStdinEmpty(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	err := execute(t, tf, "kitchen", testAddress, "--no-verify", "--password-stdin")
	if err == nil || !strings.Contains(err.Error(), "stdin is empty") {
		t.Fatalf("error = %v, want stdin is empty", err)
	}
}

func TestRun_UserWithoutPassword(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	err := execute(t, tf, "kitchen", testAddress, "--no-verify", "--user", "bob")
	if err == nil || !strings.Contains(err.Error(), "--password") {
		t.Fatalf("error = %v, want a hint to add --password", err)
	}
}

func TestRun_AuthPair(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	if err := execute(t, tf, "kitchen", testAddress, "--no-verify", "--auth", "admin:"+testPassword); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if auth := storedDevice(t, tf, "kitchen").Auth; auth == nil || auth.Username != "admin" || auth.Password != testPassword {
		t.Errorf("stored auth = %+v", auth)
	}

	if err := execute(t, tf, "other", "192.168.1.101", "--no-verify", "--auth", "nocolon"); err == nil {
		t.Error("expected an error for --auth without a colon")
	}
}

func TestRun_NoCredentialsStoresNoAuth(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	if err := execute(t, tf, "kitchen", testAddress, "--no-verify", "--generation", "2"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	dev := storedDevice(t, tf, "kitchen")
	if dev.Auth != nil || dev.Generation != 2 || dev.Platform != "" {
		t.Errorf("stored device = %+v", dev)
	}
}

func TestRun_ExistingNameRefusedWithoutForce(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"kitchen": {Name: "kitchen", Address: testAddress},
	})
	err := execute(t, tf, "kitchen", "192.168.1.200", "--no-verify")
	if err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("error = %v, want a hint to use --force", err)
	}
	if dev := storedDevice(t, tf, "kitchen"); dev.Address != testAddress {
		t.Errorf("address changed to %q without --force", dev.Address)
	}
}

func TestRun_ForceReplacesAndKeepsAliases(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactoryWithDevices(t, map[string]model.Device{
		"kitchen": {
			Name:    "kitchen",
			Address: testAddress,
			Aliases: []string{"kit", "cook"},
			Auth:    &model.Auth{Username: "admin", Password: "old"},
		},
	})
	err := execute(t, tf, "kitchen", "192.168.1.200", "--no-verify", "--force", "--user", "admin", "--password", testPassword)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	dev := storedDevice(t, tf, "kitchen")
	if dev.Address != "192.168.1.200" {
		t.Errorf("Address = %q, want the new address", dev.Address)
	}
	if dev.Auth == nil || dev.Auth.Password != testPassword {
		t.Error("password was not replaced")
	}
	if strings.Join(dev.Aliases, ",") != "kit,cook" {
		t.Errorf("Aliases = %v, want [kit cook]", dev.Aliases)
	}
	if !strings.Contains(tf.OutString(), "Replaced") {
		t.Errorf("output = %q, want Replaced", tf.OutString())
	}
	assertNoPasswordInOutput(t, tf)
}

func TestRun_PlatformNoVerify(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	if err := execute(t, tf, "garage-plug", "192.168.1.50", "--no-verify", "--platform", testPlatform); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if dev := storedDevice(t, tf, "garage-plug"); dev.Platform != testPlatform {
		t.Errorf("Platform = %q, want %q", dev.Platform, testPlatform)
	}

	if err := execute(t, tf, "kitchen", testAddress, "--no-verify", "--platform", model.PlatformShelly); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if dev := storedDevice(t, tf, "kitchen"); dev.Platform != "" {
		t.Errorf("Platform = %q, want native devices stored without a platform", dev.Platform)
	}
}

func TestRun_PlatformWithoutPlugin(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	err := execute(t, tf, "garage-plug", "192.168.1.50", "--platform", testPlatform)
	if !errors.Is(err, shelly.ErrPluginNotFound) {
		t.Fatalf("error = %v, want plugin not found", err)
	}
	if _, ok := tf.Manager.GetDevice("garage-plug"); ok {
		t.Error("device registered although its plugin is missing")
	}
}

// installDetectPlugin installs a shelly-tasmota plugin whose detect hook
// prints detectJSON and records its arguments in the returned file.
func installDetectPlugin(t *testing.T, tf *factory.TestFactory, detectJSON string) string {
	t.Helper()

	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "shelly-"+testPlatform)
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatal(err)
	}
	argsFile := filepath.Join(dir, "args")
	testutil.WriteTestScript(t, filepath.Join(pluginDir, "detect"),
		"#!/bin/sh\necho \"$@\" > '"+argsFile+"'\necho '"+detectJSON+"'\n")
	testutil.WriteTestScript(t, filepath.Join(pluginDir, "shelly-"+testPlatform), "#!/bin/sh\n")
	manifest, err := json.Marshal(plugins.Manifest{
		SchemaVersion: "1",
		Name:          testPlatform,
		Version:       "1.0.0",
		Capabilities:  &plugins.Capabilities{Platform: testPlatform, DeviceDetection: true},
		Hooks:         &plugins.Hooks{Detect: "./detect"},
		Binary:        plugins.Binary{Name: "shelly-" + testPlatform},
	})
	if err != nil {
		t.Fatal(err)
	}
	testutil.WriteTestFile(t, filepath.Join(pluginDir, "manifest.json"), manifest)
	tf.ShellyService().SetPluginRegistry(plugins.NewRegistryWithDir(dir))
	return argsFile
}

func TestRun_PlatformVerifiedThroughPlugin(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	argsFile := installDetectPlugin(t, tf, `{"detected":true,"platform":"tasmota","model":"Sonoff Basic"}`)

	if err := execute(t, tf, "garage-plug", "192.168.1.50", "--platform", testPlatform, "--password", testPassword); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	dev := storedDevice(t, tf, "garage-plug")
	if dev.Platform != testPlatform || dev.Type != "Sonoff Basic" || dev.Generation != 0 {
		t.Errorf("stored device = %+v", dev)
	}
	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(args), "--address 192.168.1.50 --auth-user admin") {
		t.Errorf("detect hook args = %q, want the address and credentials", args)
	}
	if !strings.Contains(tf.OutString(), "(tasmota)") {
		t.Errorf("output = %q, want the platform shown", tf.OutString())
	}
	assertNoPasswordInOutput(t, tf)
}

func TestRun_PlatformNotRecognised(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	installDetectPlugin(t, tf, `{"detected":false}`)

	err := execute(t, tf, "garage-plug", "192.168.1.50", "--platform", testPlatform)
	if err == nil || !strings.Contains(err.Error(), "does not recognise") {
		t.Fatalf("error = %v, want not recognised", err)
	}
	if _, ok := tf.Manager.GetDevice("garage-plug"); ok {
		t.Error("device registered although the plugin did not recognise it")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_VerifiesShellyDevice(t *testing.T) {
	demo, err := mock.StartWithFixtures(&mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{{
				Name:       "probe-target",
				MAC:        "AA:BB:CC:DD:EE:FF",
				Type:       "SNSW-001P16EU",
				Model:      "SNSW-001P16EU",
				Generation: 2,
			}},
		},
	})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)
	address := demo.DeviceServer.DeviceURL("probe-target")

	if err := execute(t, tf, "new-device", address, "--password", testPassword); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	dev, ok := demo.ConfigMgr.GetDevice("new-device")
	if !ok {
		t.Fatal("device was not registered")
	}
	if dev.Generation != 2 || dev.Type != "SNSW-001P16EU" || dev.Auth == nil || dev.Auth.Password != testPassword {
		t.Errorf("stored device = %+v", dev)
	}
	if !strings.Contains(tf.OutString(), "Gen2") {
		t.Errorf("output = %q, want the generation shown", tf.OutString())
	}
	assertNoPasswordInOutput(t, tf)

	// Each case installs its own mock registry as the process-wide default,
	// so they run here, in sequence, instead of as parallel tests.
	t.Run("wrong password rejected", checkWrongPasswordRejected)
	t.Run("wrong password keeps existing registration", checkWrongPasswordKeepsExistingRegistration)
	t.Run("right password accepted", checkRightPasswordAccepted)
	t.Run("no-verify skips the credentials check", checkNoVerifySkipsCredentialsCheck)
	t.Run("warns when auth is enabled and no credentials given", checkWarnsWhenAuthEnabledAndNoCredentials)
}

// startProtectedDevice serves one mock device of the given generation that
// requires testPassword for user admin, and returns the factory wired to the
// mock registry together with the device's address.
func startProtectedDevice(t *testing.T, generation int) (tf *factory.TestFactory, demo *mock.Demo, address string) {
	t.Helper()
	deviceType := "SNSW-001P16EU"
	if generation == 1 {
		deviceType = "SHSW-1"
	}
	demo, err := mock.StartWithFixtures(&mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{{
				Name:        "probe-target",
				MAC:         "AA:BB:CC:DD:EE:FF",
				Type:        deviceType,
				Model:       deviceType,
				Generation:  generation,
				AuthEnabled: true,
				AuthUser:    "admin",
				AuthPass:    testPassword,
			}},
		},
	})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf = factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)
	return tf, demo, demo.DeviceServer.DeviceURL("probe-target")
}

func checkWrongPasswordRejected(t *testing.T) {
	t.Helper()
	for _, generation := range []int{1, 2} {
		t.Run(fmt.Sprintf("gen%d", generation), func(t *testing.T) {
			tf, demo, address := startProtectedDevice(t, generation)

			err := execute(t, tf, "new-device", address, "--password", wrongPassword)
			if !errors.Is(err, shelly.ErrCredentialsRejected) {
				t.Fatalf("Execute() error = %v, want ErrCredentialsRejected", err)
			}
			if !strings.Contains(err.Error(), "device rejected the password for user admin") {
				t.Errorf("error = %q, want it to name the rejected user", err)
			}
			if strings.Contains(err.Error(), wrongPassword) {
				t.Error("the error contains the password")
			}
			if _, ok := demo.ConfigMgr.GetDevice("new-device"); ok {
				t.Error("device registered although its password was rejected")
			}
			assertNoPasswordInOutput(t, tf)
		})
	}
}

func checkWrongPasswordKeepsExistingRegistration(t *testing.T) {
	t.Helper()
	tf, demo, address := startProtectedDevice(t, 2)

	err := execute(t, tf, "probe-target", address, "--force", "--password", wrongPassword)
	if !errors.Is(err, shelly.ErrCredentialsRejected) {
		t.Fatalf("Execute() error = %v, want ErrCredentialsRejected", err)
	}
	dev, ok := demo.ConfigMgr.GetDevice("probe-target")
	if !ok || dev.Auth == nil || dev.Auth.Password != testPassword {
		t.Error("the existing registration was changed although the new password was rejected")
	}
	assertNoPasswordInOutput(t, tf)
}

func checkRightPasswordAccepted(t *testing.T) {
	t.Helper()
	for _, generation := range []int{1, 2} {
		t.Run(fmt.Sprintf("gen%d", generation), func(t *testing.T) {
			tf, demo, address := startProtectedDevice(t, generation)

			if err := execute(t, tf, "new-device", address, "--password", testPassword); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			dev, ok := demo.ConfigMgr.GetDevice("new-device")
			if !ok {
				t.Fatal("device was not registered")
			}
			if dev.Generation != generation || dev.Auth == nil || dev.Auth.Username != "admin" || dev.Auth.Password != testPassword {
				t.Errorf("stored device = generation %d, auth set %t; want generation %d with the given credentials",
					dev.Generation, dev.Auth != nil, generation)
			}
			if strings.Contains(tf.ErrString(), "requires authentication") {
				t.Errorf("stderr = %q, want no warning when credentials were given", tf.ErrString())
			}
			assertNoPasswordInOutput(t, tf)
		})
	}
}

func checkNoVerifySkipsCredentialsCheck(t *testing.T) {
	t.Helper()
	tf, demo, address := startProtectedDevice(t, 2)

	if err := execute(t, tf, "new-device", address, "--no-verify", "--password", wrongPassword); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if _, ok := demo.ConfigMgr.GetDevice("new-device"); !ok {
		t.Error("device was not registered with --no-verify")
	}
	assertNoPasswordInOutput(t, tf)
}

func checkWarnsWhenAuthEnabledAndNoCredentials(t *testing.T) {
	t.Helper()
	for _, generation := range []int{1, 2} {
		t.Run(fmt.Sprintf("gen%d", generation), func(t *testing.T) {
			tf, demo, address := startProtectedDevice(t, generation)

			if err := execute(t, tf, "new-device", address); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			dev, ok := demo.ConfigMgr.GetDevice("new-device")
			if !ok {
				t.Fatal("device was not registered")
			}
			if dev.Auth != nil {
				t.Error("credentials were stored although none were given")
			}
			if out := tf.OutString() + tf.ErrString(); !strings.Contains(out, "requires authentication") {
				t.Errorf("output = %q, want the requires-authentication warning", out)
			}
		})
	}
}
