package create

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const testAllDir = "/out/nested"

func twoDeviceFixtures() *mock.Fixtures {
	return &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{Name: "alpha", Address: "192.168.1.10", MAC: "AA:BB:CC:00:00:01", Type: "SNSW-001P16EU", Model: "Shelly Plus 1PM", Generation: 2},
				{Name: "beta", Address: "192.168.1.11", MAC: "AA:BB:CC:00:00:02", Type: "SNSW-001P16EU", Model: "Shelly Plus 1PM", Generation: 2},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"alpha": {"switch:0": map[string]any{"output": false}},
			"beta":  {"switch:0": map[string]any{"output": true}},
		},
	}
}

func startAllDemo(t *testing.T) (*mock.Demo, *factory.TestFactory) {
	t.Helper()
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	demo, err := mock.StartWithFixtures(twoDeviceFixtures())
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)
	return demo, tf
}

func backupFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := afero.ReadDir(config.Fs(), dir)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

//nolint:paralleltest // modifies global state via config.SetFs and the default config manager
func TestAll_BacksUpEveryDeviceIntoDir(t *testing.T) {
	_, tf := startAllDemo(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--all", "--dir", testAllDir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v\nstderr: %s", err, tf.ErrString())
	}

	files := backupFiles(t, testAllDir)
	if len(files) != 2 {
		t.Fatalf("got files %v, want 2", files)
	}
	for _, want := range []string{"alpha-aabbcc000001-", "beta-aabbcc000002-"} {
		found := false
		for _, f := range files {
			if strings.HasPrefix(f, want) && strings.HasSuffix(f, ".json") {
				found = true
			}
		}
		if !found {
			t.Errorf("no auto-named file with prefix %q in %v", want, files)
		}
	}

	out := tf.OutString()
	for _, want := range []string{"alpha: " + testAllDir, "beta: " + testAllDir, "Backed up 2 of 2 devices to " + testAllDir} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

//nolint:paralleltest // modifies global state via config.SetFs and the default config manager
func TestAll_OneDeviceFailingDoesNotStopOthers(t *testing.T) {
	demo, tf := startAllDemo(t)
	if err := demo.ConfigMgr.RegisterDevice("gamma", "127.0.0.1:1", 2, "SNSW-001P16EU", "Shelly Plus 1PM", nil); err != nil {
		t.Fatalf("RegisterDevice: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--all", "--dir", testAllDir})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "1 of 3 device backups failed") {
		t.Fatalf("Execute error = %v, want '1 of 3 device backups failed'", err)
	}

	if files := backupFiles(t, testAllDir); len(files) != 2 {
		t.Errorf("got files %v, want the 2 reachable devices", files)
	}
	if !strings.Contains(tf.ErrString(), "gamma: ") {
		t.Errorf("stderr missing the failed device line:\n%s", tf.ErrString())
	}
	if !strings.Contains(tf.OutString(), "Backed up 2 of 3 devices") {
		t.Errorf("stdout missing real counts:\n%s", tf.OutString())
	}
}

//nolint:paralleltest // modifies global state via config.SetFs and the default config manager
func TestAll_EncryptsEveryFile(t *testing.T) {
	_, tf := startAllDemo(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--all", "--dir", testAllDir, "--encrypt", "pw"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	files := backupFiles(t, testAllDir)
	if len(files) != 2 {
		t.Fatalf("got files %v, want 2", files)
	}
	for _, f := range files {
		data, err := afero.ReadFile(config.Fs(), testAllDir+"/"+f)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if !backup.IsEncrypted(data) {
			t.Errorf("%s is not encrypted", f)
		}
	}
}

//nolint:paralleltest // modifies global state via config.SetFs and the default config manager
func TestAll_DefaultsToBackupsDir(t *testing.T) {
	_, tf := startAllDemo(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--all"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	dir, err := config.BackupsDir()
	if err != nil {
		t.Fatalf("BackupsDir: %v", err)
	}
	if files := backupFiles(t, dir); len(files) != 2 {
		t.Errorf("got files %v in %s, want 2", files, dir)
	}
}

//nolint:paralleltest // modifies global state via config.SetFs and the default config manager
func TestDir_SingleDeviceAutoName(t *testing.T) {
	_, tf := startAllDemo(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"alpha", "--dir", testAllDir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	files := backupFiles(t, testAllDir)
	if len(files) != 1 || !strings.HasPrefix(files[0], "alpha-aabbcc000001-") {
		t.Errorf("got files %v, want one alpha auto-named backup", files)
	}
}

func TestArgs_AllAndDirCombinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"all with device", []string{"--all", "alpha"}, "do not pass a device"},
		{"no device no all", nil, "requires a device argument, or --all"},
		{"dir with file", []string{"alpha", "out.json", "--dir", "x"}, "do not pass a file as well"},
		{"too many args", []string{"a", "b", "c"}, "accepts at most 2 arg(s)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			cmd := NewCommand(tf.Factory)
			cmd.SetArgs(tt.args)
			cmd.SetOut(&strings.Builder{})
			cmd.SetErr(&strings.Builder{})
			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("err = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
