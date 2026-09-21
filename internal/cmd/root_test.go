// Package cmd provides the root command and command wiring for the CLI.
package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"regexp"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tj-smith47/shelly-go/transport"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// rootCmdMu guards concurrent access to the shared rootCmd singleton from
// parallel tests. cobra.Command.Commands() and .Find() both sort the
// command tree in place, which races when called from more than one
// goroutine at once.
var rootCmdMu sync.Mutex

func TestRootCommandStructure(t *testing.T) {
	t.Parallel()

	// Test that the root command is properly initialized
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}

	if rootCmd.Use != "shelly" {
		t.Errorf("Use = %q, want %q", rootCmd.Use, "shelly")
	}

	if rootCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if rootCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestRootCommand_GlobalFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{"output", "output", "o", "table"},
		{"template", "template", "", ""},
		{"verbose", "verbose", "v", "0"},
		{"quiet", "quiet", "q", "false"},
		{"config", "config", "", ""},
		{"no-color", "no-color", "", "false"},
	}

	for _, tt := range tests {
		flag := rootCmd.PersistentFlags().Lookup(tt.flagName)
		if flag == nil {
			t.Errorf("%s flag not found", tt.flagName)
			continue
		}
		if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
			t.Errorf("%s shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
		}
		if flag.DefValue != tt.defValue {
			t.Errorf("%s default = %q, want %q", tt.flagName, flag.DefValue, tt.defValue)
		}
	}
}

func TestRootCommand_RawFlag(t *testing.T) {
	t.Parallel()

	flag := rootCmd.PersistentFlags().Lookup("raw")
	if flag == nil {
		t.Fatal("raw flag not registered")
	}
	if flag.DefValue != "false" {
		t.Errorf("raw default = %q, want %q", flag.DefValue, "false")
	}
}

func TestEmitRawResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		responses []json.RawMessage
		want      string
	}{
		{
			name:      "no device calls yields empty array",
			responses: nil,
			want:      "[]\n",
		},
		{
			name:      "single Gen1 REST body",
			responses: []json.RawMessage{json.RawMessage(`{"ison":true}`)},
			want:      `[{"ison":true}]` + "\n",
		},
		{
			name: "multiple responses preserve order",
			responses: []json.RawMessage{
				json.RawMessage(`{"id":0}`),
				json.RawMessage(`{"was_on":false}`),
			},
			want: `[{"id":0},{"was_on":false}]` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var sink transport.RawCapture
			for _, r := range tt.responses {
				sink.Record(r)
			}

			var out bytes.Buffer
			if err := emitRawResponses(&out, &sink); err != nil {
				t.Fatalf("emitRawResponses() error = %v", err)
			}

			if got := out.String(); got != tt.want {
				t.Errorf("stdout = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRootCommand_Subcommands(t *testing.T) {
	t.Parallel()

	rootCmdMu.Lock()
	subcommands := rootCmd.Commands()
	rootCmdMu.Unlock()
	expectedSubcommands := map[string]bool{
		"discover": false,
		"switch":   false,
		"cover":    false,
		"light":    false,
		"rgb":      false,
		"input":    false,
		"version":  false,
	}

	for _, sub := range subcommands {
		if _, ok := expectedSubcommands[sub.Use]; ok {
			expectedSubcommands[sub.Use] = true
		}
	}

	for name, found := range expectedSubcommands {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}

func TestSwitchSubcommandsNotShadowedByToggleAlias(t *testing.T) {
	t.Parallel()

	// Regression: the root `toggle` quick command once aliased "switch", which
	// shadowed the `switch` command group (registered later). Cobra matched the
	// alias first, so `shelly switch status <dev>` mis-routed to toggle and died
	// with "accepts 1 arg(s), received 2". Every switch subcommand must resolve
	// to the switch group.
	for _, sub := range []string{"status", "on", "off", "toggle", "list"} {
		rootCmdMu.Lock()
		cmd, _, err := rootCmd.Find([]string{"switch", sub})
		rootCmdMu.Unlock()
		if err != nil {
			t.Fatalf("Find(switch %s): %v", sub, err)
		}
		if cmd.Name() != sub {
			t.Errorf("switch %s resolved to %q, want the switch-group subcommand", sub, cmd.Name())
		}
		if cmd.Parent() == nil || cmd.Parent().Name() != "switch" {
			t.Errorf("switch %s parent = %v, want the switch group", sub, cmd.Parent())
		}
	}
}

func TestMust_NilError(t *testing.T) {
	t.Parallel()

	// Should not panic with nil error
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("utils.Must(nil) panicked: %v", r)
		}
	}()

	utils.Must(nil)
}

func TestIsColorDisabled(t *testing.T) {
	// Cannot run in parallel due to env var and viper manipulation

	tests := []struct {
		name       string
		noColorEnv string
		shellyEnv  string
		viperVal   bool
		want       bool
	}{
		{"default", "", "", false, false},
		{"flag set", "", "", true, true},
		{"NO_COLOR set", "1", "", false, true},
		{"SHELLY_NO_COLOR set", "", "1", false, true},
		{"NO_COLOR empty string", "set", "", false, true},
		{"all set", "1", "1", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up env vars using t.Setenv (auto-restores after test)
			// First unset, then use t.Setenv to track
			if err := os.Unsetenv("NO_COLOR"); err != nil {
				t.Logf("warning: failed to unset NO_COLOR: %v", err)
			}
			if err := os.Unsetenv("SHELLY_NO_COLOR"); err != nil {
				t.Logf("warning: failed to unset SHELLY_NO_COLOR: %v", err)
			}
			viper.Set("no-color", false)

			// Set test conditions
			if tt.noColorEnv != "" {
				t.Setenv("NO_COLOR", tt.noColorEnv)
			}
			if tt.shellyEnv != "" {
				t.Setenv("SHELLY_NO_COLOR", tt.shellyEnv)
			}
			viper.Set("no-color", tt.viperVal)

			got := iostreams.IsColorDisabled()
			if got != tt.want {
				t.Errorf("iostreams.IsColorDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

// walkCommands visits c and every descendant. Callers hold rootCmdMu:
// Commands() sorts the tree in place.
func walkCommands(c *cobra.Command, visit func(*cobra.Command)) {
	visit(c)
	for _, sub := range c.Commands() {
		walkCommands(sub, visit)
	}
}

// bareDashExample matches a documented invocation whose last argument is a bare
// "-", e.g. "shelly backup create living-room -".
var bareDashExample = regexp.MustCompile(`(?m)^\s*shelly\s.*\s-\s*$`)

// TestDashIsOutput_AnnotationMatchesDocs keeps the stdin substitution and the
// docs in step: a command that documents a bare "-" argument must opt out of the
// root's stdin replacement, and only such commands may opt out.
func TestDashIsOutput_AnnotationMatchesDocs(t *testing.T) {
	t.Parallel()

	rootCmdMu.Lock()
	defer rootCmdMu.Unlock()

	var annotated int
	walkCommands(rootCmd, func(c *cobra.Command) {
		documented := bareDashExample.MatchString(c.Example)
		marked := cmdutil.DashIsOutput(c)
		if marked {
			annotated++
		}
		switch {
		case documented && !marked:
			t.Errorf("%s documents a bare \"-\" argument but lacks cmdutil.DashIsOutputAnnotation(); "+
				"the root would replace \"-\" with stdin before the command sees it", c.CommandPath())
		case marked && !documented:
			t.Errorf("%s carries cmdutil.DashIsOutputAnnotation() but no example shows the bare \"-\" form", c.CommandPath())
		}
	})
	if annotated == 0 {
		t.Error("no command carries the dash-is-output annotation; backup create and config export should")
	}
}

func TestDashIsOutput_ResolvesCommand(t *testing.T) {
	t.Parallel()

	rootCmdMu.Lock()
	defer rootCmdMu.Unlock()

	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"backup create to stdout", []string{"backup", "create", "living-room", "-"}, true},
		{"config export to stdout", []string{"device", "config", "export", "living-room", "-"}, true},
		{"device from stdin", []string{"status", "-"}, false},
		{"unknown command", []string{"no-such-command", "-"}, false},
	}
	for _, tt := range tests {
		if got := dashIsOutput(rootCmd, tt.args); got != tt.want {
			t.Errorf("%s: dashIsOutput(%v) = %v, want %v", tt.name, tt.args, got, tt.want)
		}
	}
}

// localShadowsOfGlobalFlags lists every command-local flag that reuses a global
// flag's name, with what the local flag means there. In cobra the local flag
// wins, so the global one is unreachable on these commands. The root reads each
// global through its own flag object (viper bindings, applyRawCapture), so a
// local flag never switches on the global behavior.
var localShadowsOfGlobalFlags = map[string]string{
	"shelly api --raw":                     "compact JSON of the single response",
	"shelly auth export --output":          "output file path",
	"shelly batch command --output":        "output format, limited to the formats the command supports",
	"shelly cloud events --raw":            "print each event message as received",
	"shelly debug coiot --raw":             "print each event message as received",
	"shelly debug websocket --raw":         "print each event message as received",
	"shelly device list --refresh":         "re-read device metadata from hardware",
	"shelly discover coiot --verbose":      "show Gen1-specific detail",
	"shelly energy export --output":        "output file path",
	"shelly firmware download --output":    "output file path",
	"shelly fleet status --offline":        "list only offline devices",
	"shelly group members --output":        "output format, limited to the formats the command supports",
	"shelly init --no-color":               "same meaning as the global flag",
	"shelly kvs get --raw":                 "print the stored value only",
	"shelly log export --output":           "output file path",
	"shelly metrics influxdb --output":     "output file path",
	"shelly metrics json --output":         "output file path",
	"shelly modbus status --output":        "output format, limited to the formats the command supports",
	"shelly plugin create --output":        "output directory",
	"shelly profile info --output":         "output format, limited to the formats the command supports",
	"shelly profile list --output":         "output format, limited to the formats the command supports",
	"shelly profile search --output":       "output format, limited to the formats the command supports",
	"shelly scene show --output":           "output format, limited to the formats the command supports",
	"shelly script template list --output": "output format, limited to the formats the command supports",
	"shelly script template show --output": "output format, limited to the formats the command supports",
	"shelly sensoraddon list --output":     "output format, limited to the formats the command supports",
	"shelly sensoraddon scan --output":     "output format, limited to the formats the command supports",
	"shelly virtual get --output":          "output format, limited to the formats the command supports",
	"shelly virtual list --output":         "output format, limited to the formats the command supports",
	"shelly webhook server --log-json":     "log received webhooks as JSON",
	"shelly zwave config --output":         "output format, limited to the formats the command supports",
	"shelly zwave info --output":           "output format, limited to the formats the command supports",
}

// TestLocalFlagsDoNotShadowGlobals stops new commands from redefining a global
// flag name: the local flag wins in cobra, silently disabling the global one.
func TestLocalFlagsDoNotShadowGlobals(t *testing.T) {
	t.Parallel()

	rootCmdMu.Lock()
	defer rootCmdMu.Unlock()

	seen := map[string]bool{}
	walkCommands(rootCmd, func(c *cobra.Command) {
		if c == rootCmd {
			return
		}
		c.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if rootCmd.PersistentFlags().Lookup(f.Name) == nil {
				return
			}
			key := c.CommandPath() + " --" + f.Name
			seen[key] = true
			if _, ok := localShadowsOfGlobalFlags[key]; !ok {
				t.Errorf("%s reuses the global --%s flag name; pick a different local flag name", key, f.Name)
			}
		})
	})
	for key := range localShadowsOfGlobalFlags {
		if !seen[key] {
			t.Errorf("%s is allowlisted but no longer shadows a global flag; drop it from the list", key)
		}
	}
}

// TestApplyRawCapture_IgnoresLocalRawFlag asserts a command-local --raw does not
// switch the command into global capture mode (which discarded its output and
// printed "[]").
func TestApplyRawCapture_IgnoresLocalRawFlag(t *testing.T) {
	t.Parallel()

	rootCmdMu.Lock()
	defer rootCmdMu.Unlock()

	apiCmd, _, err := rootCmd.Find([]string{"api"})
	if err != nil {
		t.Fatalf("find api: %v", err)
	}
	if err := apiCmd.Flags().Set("raw", "true"); err != nil {
		t.Fatalf("set local --raw: %v", err)
	}
	t.Cleanup(func() {
		if err := apiCmd.Flags().Set("raw", "false"); err != nil {
			t.Errorf("reset local --raw: %v", err)
		}
	})

	rawSink = nil
	applyRawCapture(apiCmd)
	if rawSink != nil {
		rawSink = nil
		t.Error("local --raw installed the global capture sink")
	}
}
