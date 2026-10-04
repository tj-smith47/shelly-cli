package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/errutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
)

// powerCarrierCheck is what one power or energy subcommand must do against a
// fleet whose only power metering is a switch and a Gen1 meter.
type powerCarrierCheck struct {
	args []string
	// want are substrings of the compact JSON output; empty with wantErr set.
	want []string
	// wantErr is a substring of the error, for commands that cannot use the
	// carrier; the error must name the command that can.
	wantErr string
}

// powerCarrierChecks lists every subcommand of `shelly power` and `shelly
// energy`. A new subcommand fails TestPowerCommandsReadEveryCarrier until it
// is added here.
var powerCarrierChecks = map[string][]powerCarrierCheck{
	"shelly power status": {
		{args: []string{"power", "status", "office"}, want: []string{`"type":"switch"`, `"power":48.5`}},
		{args: []string{"power", "status", "closet"}, want: []string{`"type":"meter"`, `"power":60.5`}},
		{args: []string{"power", "status", "--all"}, want: []string{`"name":"office"`, `"name":"closet"`}},
	},
	"shelly power list": {
		{args: []string{"power", "list", "office"}, want: []string{`"type":"switch"`, `"power":48.5`}},
		{args: []string{"power", "list", "closet"}, want: []string{`"type":"meter"`, `"power":60.5`}},
	},
	"shelly energy status": {
		{args: []string{"energy", "status", "office"}, want: []string{`"type":"switch"`, `"power":48.5`}},
		{args: []string{"energy", "status", "closet"}, want: []string{`"type":"meter"`, `"power":60.5`}},
		{args: []string{"energy", "status", "--all"}, want: []string{`"name":"office"`, `"name":"closet"`}},
	},
	"shelly energy list": {
		{args: []string{"energy", "list", "office"}, want: []string{`"type":"switch"`, `"power":48.5`}},
		{args: []string{"energy", "list", "closet"}, want: []string{`"type":"meter"`, `"power":60.5`}},
	},
	"shelly energy dashboard": {
		{args: []string{"energy", "dashboard"}, want: []string{`"type":"switch"`, `"power_w":48.5`, `"type":"meter"`, `"power_w":60.5`}},
	},
	"shelly energy compare": {
		{args: []string{"energy", "compare"}, want: []string{`"avg_power_w":48.5`, `"avg_power_w":60.5`}},
	},
	"shelly energy history": {
		{args: []string{"energy", "history", "office"}, wantErr: "shelly energy status office"},
		{args: []string{"energy", "history", "closet"}, wantErr: "shelly energy status closet"},
	},
	"shelly energy export": {
		{args: []string{"energy", "export", "office"}, wantErr: "shelly energy status office"},
		{args: []string{"energy", "export", "closet"}, wantErr: "shelly energy status closet"},
	},
	"shelly energy reset": {
		{args: []string{"energy", "reset", "office"}},
		{args: []string{"energy", "reset", "closet"}, wantErr: "energy reset works on"},
	},
}

// TestPowerCommandsReadEveryCarrier runs every power and energy subcommand
// against a Plus 1PM, whose only meter is its switch, and a Gen1 1PM, whose
// meter is in /status meters[]. A subcommand that reads only EM, EM1, PM or
// PM1 components finds nothing on these devices and fails here.
func TestPowerCommandsReadEveryCarrier(t *testing.T) {
	t.Setenv("HOME", "/testhome")
	rootCmdMu.Lock()
	defer rootCmdMu.Unlock()

	h := newHarness(t, []harnessDevice{
		{name: "office", gen: 2, model: "SNSW-001P16EU", state: mock.DeviceState{
			"switch:0": map[string]any{"id": 0, "output": true, "apower": 48.5, "voltage": 121.0, "current": 0.4,
				"aenergy": map[string]any{"total": 300.0}},
		}},
		{name: "closet", gen: 1, model: "SHSW-PM", state: mock.DeviceState{
			"relays": []any{map[string]any{"ison": true}},
			"meters": []any{map[string]any{"power": 60.5, "total": 600, "is_valid": true}},
		}},
	}, false)

	for _, parent := range []string{"power", "energy"} {
		cmd, _, err := rootCmd.Find([]string{parent})
		if err != nil {
			t.Fatalf("find %s: %v", parent, err)
		}
		for _, sub := range cmd.Commands() {
			if _, ok := powerCarrierChecks[sub.CommandPath()]; !ok && sub.IsAvailableCommand() {
				t.Errorf("%s is not in powerCarrierChecks: say what it does with a switch or Gen1 meter", sub.CommandPath())
			}
		}
	}

	for path, checks := range powerCarrierChecks {
		for _, c := range checks {
			args := append(append([]string{}, c.args...), "-o", "json")
			stdout, err := h.run(args)
			invocation := "shelly " + strings.Join(args, " ")
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Errorf("%s: err = %v, want containing %q", invocation, err, c.wantErr)
				}
				continue
			}
			if err != nil {
				t.Errorf("%s (%s): %v", invocation, path, err)
				continue
			}
			if len(c.want) == 0 {
				continue
			}
			var doc any
			if jsonErr := json.Unmarshal([]byte(stdout), &doc); jsonErr != nil {
				t.Errorf("%s: stdout is not JSON: %s", invocation, firstLines(stdout, 3))
				continue
			}
			compact, marshalErr := json.Marshal(doc)
			if marshalErr != nil {
				t.Fatalf("marshal: %v", marshalErr)
			}
			for _, w := range c.want {
				if !strings.Contains(string(compact), w) {
					t.Errorf("%s: output has no %s:\n%s", invocation, w, compact)
				}
			}
		}
	}
}

// TestMissingComponentErrorsSayNotAvailable runs every `<x> status <device>`
// and `<x> list <device>` command against a device with no components. When
// the device answers "No handler for X.Y", the error the user sees must read
// "X not available on this device", which needs the RPC error kept in the
// chain (%w, never %v).
func TestMissingComponentErrorsSayNotAvailable(t *testing.T) {
	t.Setenv("HOME", "/testhome")
	rootCmdMu.Lock()
	defer rootCmdMu.Unlock()

	h := newContractHarness(t, nil, false)
	var invocations [][]string
	walkCommands(rootCmd, func(c *cobra.Command) {
		name := c.Name()
		if (name != "status" && name != "list") || !c.Runnable() || c.Hidden {
			return
		}
		if !strings.Contains(c.Use, "<device>") && !strings.Contains(c.Use, "[device]") {
			return
		}
		// monitor status refreshes until interrupted, so it would only time out.
		if c.CommandPath() == "shelly monitor status" {
			return
		}
		args := strings.Fields(strings.TrimPrefix(c.CommandPath(), "shelly "))
		invocations = append(invocations, append(args, bareDevice))
	})
	if len(invocations) < 20 {
		t.Fatalf("found only %d status/list commands; the walk is broken", len(invocations))
	}
	answered := 0
	for _, args := range invocations {
		_, err := h.run(args)
		if err == nil || !strings.Contains(err.Error(), "No handler for") {
			continue
		}
		answered++
		if shown := errutil.NotAvailable("", err).Error(); !strings.Contains(shown, "not available on this device") {
			t.Errorf("shelly %s: the user sees %q; wrap the device error with %%w", strings.Join(args, " "), shown)
		}
	}
	if answered == 0 {
		t.Fatalf("none of %d commands got \"No handler for\" from the mock; the test checks nothing", len(invocations))
	}
}
