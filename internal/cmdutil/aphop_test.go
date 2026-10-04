package cmdutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// stubAPRestorer drives RestoreAtAP without hopping the host's WiFi.
type stubAPRestorer struct {
	result  *backup.RestoreResult
	newAddr string
	err     error
}

func (s stubAPRestorer) RestoreToAP(_ context.Context, _, _, _ string, _ *backup.DeviceBackup, _ backup.RestoreOptions) (*backup.RestoreResult, string, error) {
	return s.result, s.newAddr, s.err
}

func newTestIOS() (*iostreams.IOStreams, *bytes.Buffer) {
	out := &bytes.Buffer{}
	return iostreams.Test(&bytes.Buffer{}, out, &bytes.Buffer{}), out
}

// okReport reports success (nil) for any result; partialReport rejects.
func okReport(*iostreams.IOStreams, string, *backup.RestoreResult) error { return nil }

var errSectionRejected = errors.New("cloud section rejected")

func rejectReport(*iostreams.IOStreams, string, *backup.RestoreResult) error {
	return errSectionRejected
}

// TestRestoreAtAP_Success: a clean restore returns nil and surfaces the new LAN
// address so the user knows where the device landed.
func TestRestoreAtAP_Success(t *testing.T) {
	t.Parallel()
	ios, out := newTestIOS()
	svc := stubAPRestorer{result: &backup.RestoreResult{Success: true}, newAddr: "10.23.47.227"}

	err := RestoreAtAP(context.Background(), ios, svc, "ShellyBulbDuo-AABBCC", "192.168.33.1", "fr",
		&backup.DeviceBackup{}, backup.RestoreOptions{}, "restore via AP failed", okReport)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if o := out.String(); !strings.Contains(o, "10.23.47.227") {
		t.Errorf("new LAN address should be surfaced, got %q", o)
	}
}

// TestRestoreAtAP_PartialRejection: a reporter error (a section rejected with no
// top-level transport error) propagates so the exit code is non-zero, yet the new
// address is STILL surfaced so a partial failure can be finished by hand.
func TestRestoreAtAP_PartialRejection(t *testing.T) {
	t.Parallel()
	ios, out := newTestIOS()
	svc := stubAPRestorer{result: &backup.RestoreResult{Success: false}, newAddr: "10.23.47.227"}

	err := RestoreAtAP(context.Background(), ios, svc, "ap", "192.168.33.1", "fr",
		&backup.DeviceBackup{}, backup.RestoreOptions{}, "restore via AP failed", rejectReport)
	if !errors.Is(err, errSectionRejected) {
		t.Fatalf("reporter error must propagate, got %v", err)
	}
	if o := out.String(); !strings.Contains(o, "10.23.47.227") {
		t.Errorf("address must still be surfaced on partial failure, got %q", o)
	}
}

// TestRestoreAtAP_TransportFailure: a RestoreToAP error is wrapped with the
// caller's prefix and reported before any result reporting.
func TestRestoreAtAP_TransportFailure(t *testing.T) {
	t.Parallel()
	ios, _ := newTestIOS()
	svc := stubAPRestorer{err: errors.New("connection refused")}
	reportCalled := false
	report := func(*iostreams.IOStreams, string, *backup.RestoreResult) error {
		reportCalled = true
		return nil
	}

	err := RestoreAtAP(context.Background(), ios, svc, "ap", "192.168.33.1", "fr",
		&backup.DeviceBackup{}, backup.RestoreOptions{}, "migration via AP failed", report)
	if err == nil || !strings.Contains(err.Error(), "migration via AP failed") {
		t.Fatalf("expected wrapped transport failure, got %v", err)
	}
	if reportCalled {
		t.Error("reporter must not run when the restore call itself failed")
	}
}

// TestRestoreAtAP_NoAddress: when the device did not rejoin (empty newAddr), no
// "is live at" line is printed.
func TestRestoreAtAP_NoAddress(t *testing.T) {
	t.Parallel()
	ios, out := newTestIOS()
	svc := stubAPRestorer{result: &backup.RestoreResult{Success: true}, newAddr: ""}

	if err := RestoreAtAP(context.Background(), ios, svc, "ap", "192.168.33.1", "fr",
		&backup.DeviceBackup{}, backup.RestoreOptions{}, "restore via AP failed", okReport); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if o := out.String(); strings.Contains(o, "is live at") {
		t.Errorf("no address line should print when the device did not rejoin, got %q", o)
	}
}

func TestValidateStaticIPFlags(t *testing.T) {
	t.Parallel()

	const ip, gw, mask, dns = "10.0.0.5", "10.0.0.1", "255.255.255.0", "10.0.0.53"
	const refusal = "--gateway, --netmask and --dns only apply with --static-ip"
	tests := []struct {
		name                            string
		staticIP, gateway, netmask, dns string
		wantErr                         string
	}{
		{name: "no static flags"},
		{name: "full static", staticIP: ip, gateway: gw, netmask: mask},
		{name: "static alone", staticIP: ip},
		{name: "static and gateway", staticIP: ip, gateway: gw},
		{name: "gateway without static", gateway: gw, wantErr: refusal},
		{name: "netmask without static", netmask: mask, wantErr: refusal},
		{name: "dns without static", dns: dns, wantErr: refusal},
		{name: "static and dns", staticIP: ip, dns: dns},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateStaticIPFlags(tt.staticIP, tt.gateway, tt.netmask, tt.dns)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Errorf("err = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestNetworkFlags_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		flags       NetworkFlags
		skipNetwork bool
		wantErr     string
	}{
		{name: "nothing set", skipNetwork: true},
		{name: "every flag without skip", flags: NetworkFlags{SSID: "n", Password: "p", StaticIP: "10.0.0.5"}},
		{name: "static with skip", flags: NetworkFlags{StaticIP: "10.0.0.5", SSID: "n"}, skipNetwork: true,
			wantErr: "--static-ip cannot be used with --skip-network"},
		{name: "ssid with skip", flags: NetworkFlags{SSID: "n"}, skipNetwork: true,
			wantErr: "--ssid cannot be used with --skip-network"},
		{name: "password with skip", flags: NetworkFlags{Password: "p"}, skipNetwork: true,
			wantErr: "--password cannot be used with --skip-network"},
		{name: "open with skip", flags: NetworkFlags{Open: true}, skipNetwork: true,
			wantErr: "--open cannot be used with --skip-network"},
		{name: "dns without static", flags: NetworkFlags{DNS: "10.0.0.53"},
			wantErr: "--gateway, --netmask and --dns only apply with --static-ip"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.flags.Validate(tt.skipNetwork)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Errorf("err = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestNetworkFlags_Override(t *testing.T) {
	t.Parallel()

	if got := (&NetworkFlags{}).Override(); got != nil {
		t.Errorf("no flags must give no override, got %+v", got)
	}
	got := (&NetworkFlags{SSID: "Guest", Open: true}).Override()
	want := &backup.NetworkOverride{SSID: "Guest", Open: true}
	if got == nil || *got != *want {
		t.Errorf("Override() = %+v, want %+v", got, want)
	}
	full := NetworkFlags{SSID: "n", Password: "p", StaticIP: "10.0.0.5", Gateway: "10.0.0.1", Netmask: "255.0.0.0", DNS: "1.1.1.1"}
	got = full.Override()
	want = &backup.NetworkOverride{SSID: "n", Password: "p", StaticIP: "10.0.0.5", Gateway: "10.0.0.1",
		Netmask: "255.0.0.0", DNS: "1.1.1.1"}
	if got == nil || *got != *want {
		t.Errorf("Override() = %+v, want %+v", got, want)
	}
}

func TestAddOpenFlag(t *testing.T) {
	t.Parallel()

	newCmd := func() (*cobra.Command, *bool) {
		var open bool
		var password string
		var passwordStdin bool
		cmd := &cobra.Command{Use: "x", RunE: func(*cobra.Command, []string) error { return nil }}
		AddWiFiPasswordFlag(cmd, &password, &passwordStdin)
		AddOpenFlag(cmd, &open)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd, &open
	}

	cmd, open := newCmd()
	cmd.SetArgs([]string{"--open"})
	if err := cmd.Execute(); err != nil || !*open {
		t.Fatalf("--open alone: err=%v open=%v", err, *open)
	}
	if usage := cmd.Flags().Lookup("open").Usage; usage != "Join a network that has no password" {
		t.Errorf("usage = %q", usage)
	}

	cmd, _ = newCmd()
	cmd.SetArgs([]string{"--open", "--password", "secret"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "none of the others can be") {
		t.Errorf("--open with --password must be refused, got %v", err)
	}

	cmd, _ = newCmd()
	cmd.SetArgs([]string{"--open", "--password-stdin"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "none of the others can be") {
		t.Errorf("--open with --password-stdin must be refused, got %v", err)
	}
}

// TestRunAtAP_ReportsSteps proves the operation runs under the labelled
// progress line with a step reporter in its context.
func TestRunAtAP_ReportsSteps(t *testing.T) {
	t.Parallel()

	var errOut bytes.Buffer
	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &errOut)
	var reporter func(string)
	err := RunAtAP(context.Background(), ios, "Onboarding bulb at AP ShellyBulb-1", func(ctx context.Context) error {
		reporter = shelly.StepReporter(ctx)
		return nil
	})
	if err != nil {
		t.Fatalf("RunAtAP: %v", err)
	}
	if reporter == nil {
		t.Fatal("action context carries no step reporter")
	}
	reporter("joining access point")
	if want := "Onboarding bulb at AP ShellyBulb-1 (hopping host WiFi)..."; !strings.Contains(errOut.String(), want) {
		t.Errorf("progress line %q missing; stderr = %q", want, errOut.String())
	}
}

func TestResolveWiFiPassword(t *testing.T) {
	t.Parallel()

	noHost := func(context.Context, string) (string, error) { return "", errors.New("not stored") }
	answer := func(s string) func(string) (string, error) {
		return func(string) (string, error) { return s, nil }
	}
	confirm := func(ok bool) func(string, bool) (bool, error) {
		return func(msg string, def bool) (bool, error) {
			if def || !strings.Contains(msg, "home") {
				t.Errorf("confirm(%q, %v): want the SSID and default no", msg, def)
			}
			return ok, nil
		}
	}

	tests := []struct {
		name     string
		lookup   func(context.Context, string) (string, error)
		prompts  WiFiPasswordPrompts
		wantPass string
		wantOpen bool
		wantErr  bool
	}{
		{"host passphrase", func(context.Context, string) (string, error) { return "stored", nil }, WiFiPasswordPrompts{}, "stored", false, false},
		{"typed", noHost, WiFiPasswordPrompts{Password: answer("typed")}, "typed", false, false},
		{"empty confirmed", noHost, WiFiPasswordPrompts{Password: answer(""), Confirm: confirm(true)}, "", true, false},
		{"empty declined", noHost, WiFiPasswordPrompts{Password: answer(""), Confirm: confirm(false)}, "", false, true},
		{"cannot prompt", noHost, WiFiPasswordPrompts{}, "", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ios, _ := newTestIOS()
			pass, open, err := ResolveWiFiPassword(context.Background(), ios, tt.lookup, "home", tt.prompts)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), `"home"`) || !strings.Contains(err.Error(), "--open") {
					t.Fatalf("err = %v, want the passphrase error naming home", err)
				}
				return
			}
			if err != nil || pass != tt.wantPass || open != tt.wantOpen {
				t.Errorf("got (%q, %v, %v), want (%q, %v, nil)", pass, open, err, tt.wantPass, tt.wantOpen)
			}
		})
	}
}

func TestResolveWiFiPassword_NeverEchoesPassphrase(t *testing.T) {
	t.Parallel()

	const secret = "s3cr3t-pass"
	hostHas := func(context.Context, string) (string, error) { return secret, nil }
	hostFails := func(context.Context, string) (string, error) { return secret, errors.New("keyring locked " + secret) }
	typed := func(string) (string, error) { return secret, nil }
	empty := func(string) (string, error) { return "", nil }
	decline := func(string, bool) (bool, error) { return false, nil }

	tests := []struct {
		name    string
		lookup  func(context.Context, string) (string, error)
		prompts WiFiPasswordPrompts
	}{
		{"host passphrase", hostHas, WiFiPasswordPrompts{}},
		{"typed", hostFails, WiFiPasswordPrompts{Password: typed}},
		{"open declined", hostFails, WiFiPasswordPrompts{Password: empty, Confirm: decline}},
		{"cannot prompt", hostFails, WiFiPasswordPrompts{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
			ios := iostreams.Test(in, out, errOut)
			_, _, err := ResolveWiFiPassword(context.Background(), ios, tt.lookup, "home", tt.prompts)
			if strings.Contains(out.String(), secret) || strings.Contains(errOut.String(), secret) {
				t.Errorf("passphrase written to output: stdout=%q stderr=%q", out.String(), errOut.String())
			}
			if err != nil && strings.Contains(err.Error(), secret) {
				t.Errorf("passphrase in error: %v", err)
			}
		})
	}
}

type stubPlanner struct {
	ov   *backup.NetworkOverride
	plan shelly.LANStationPlan
	err  error
}

func (s stubPlanner) PlanLANStation(
	context.Context, string, *backup.DeviceBackup, *backup.NetworkOverride, bool,
) (*backup.NetworkOverride, shelly.LANStationPlan, error) {
	return s.ov, s.plan, s.err
}

// TestPlanLANStation_DryRun checks which planning errors a dry run downgrades
// to a warning: only a device that cannot be read. A flag with no network to
// apply it to, and a missing passphrase, are refused on a dry run too.
func TestPlanLANStation_DryRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{name: "device read fails", err: errors.New("read the device's WiFi station: timeout")},
		{name: "no network for --open", err: fmt.Errorf("%w: pass --ssid", types.ErrInvalidParam), wantErr: true},
		{name: "no passphrase", err: shelly.PassphraseError("home"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ios, _ := newTestIOS()
			opts := &backup.RestoreOptions{DryRun: true}
			err := PlanLANStation(context.Background(), ios, stubPlanner{err: tt.err}, "dev", &backup.DeviceBackup{}, opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, want error %v", err, tt.wantErr)
			}
		})
	}
}

// TestPlanLANStation_StoresTheWrite checks that the planned override and the
// secondary-station write reach the restore options and the plan is printed.
func TestPlanLANStation_StoresTheWrite(t *testing.T) {
	t.Parallel()
	ios, out := newTestIOS()
	ov := &backup.NetworkOverride{Password: "stored"}
	plan := shelly.LANStationPlan{SSID: "home", Key: shelly.StationKeyKept}
	opts := &backup.RestoreOptions{}
	if err := PlanLANStation(context.Background(), ios, stubPlanner{ov: ov, plan: plan}, "dev", &backup.DeviceBackup{}, opts); err != nil {
		t.Fatalf("PlanLANStation() error = %v", err)
	}
	if opts.NetworkOverride != ov {
		t.Errorf("override = %+v, want the planned one", opts.NetworkOverride)
	}
	if !strings.Contains(out.String(), plan.String()) {
		t.Errorf("output %q lacks the plan line", out.String())
	}
}
