package addaction

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const testScene = "movie-night"

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Name() != "add-action" {
		t.Errorf("Name() = %q, want add-action", cmd.Name())
	}
	if len(cmd.Aliases) == 0 || cmd.Example == "" || cmd.Long == "" {
		t.Error("Aliases, Example and Long must be set")
	}
	for args, wantErr := range map[string]bool{"a b": true, "a b c": false, "a b c d": false, "a b c d e": true} {
		if err := cmd.Args(cmd, strings.Fields(args)); (err != nil) != wantErr {
			t.Errorf("Args(%q) error = %v, wantErr %v", args, err, wantErr)
		}
	}
}

//nolint:paralleltest // Test modifies global config state
func TestRun(t *testing.T) {
	config.ResetDefaultManagerForTesting()
	t.Cleanup(config.ResetDefaultManagerForTesting)
	config.SetDefaultManager(config.NewTestManager(&config.Config{}))
	if err := config.CreateScene(testScene, ""); err != nil {
		t.Fatalf("CreateScene: %v", err)
	}

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{testScene, "lamp", "Light.Set", `{"id":0,"on":true,"brightness":20}`})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if err := run(&Options{Factory: tf.Factory, Scene: testScene, Device: "hallway", Method: "Shelly.Reboot"}); err != nil {
		t.Fatalf("run() without params error = %v", err)
	}

	scene, _ := config.GetScene(testScene)
	if len(scene.Actions) != 2 {
		t.Fatalf("scene has %d actions, want 2", len(scene.Actions))
	}
	first := scene.Actions[0]
	if first.Device != "lamp" || first.Method != "Light.Set" || first.Params["brightness"] != float64(20) {
		t.Errorf("first action = %+v", first)
	}
	if scene.Actions[1].Params != nil {
		t.Errorf("second action params = %v, want none", scene.Actions[1].Params)
	}

	err := run(&Options{Factory: tf.Factory, Scene: testScene, Device: "lamp", Method: "Light.Set", Params: "on"})
	if err == nil || !strings.Contains(err.Error(), "JSON object") {
		t.Errorf("run() with bad params error = %v, want a JSON error", err)
	}
	err = run(&Options{Factory: tf.Factory, Scene: "missing", Device: "lamp", Method: "Light.Set"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("run() on a missing scene error = %v, want not found", err)
	}
}
