package cmdutil

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// APRestorer restores a backup onto a device at its WiFi access point, hopping
// the host's WiFi to reach the device's AP. *shelly.Service satisfies it, as do
// the migrate and restore command service interfaces.
type APRestorer interface {
	RestoreToAP(ctx context.Context, apSSID, apHostIP, registryName string, bkp *backup.DeviceBackup, opts backup.RestoreOptions) (*backup.RestoreResult, string, error)
}

// RestoreAtAP runs a host-WiFi-hopping restore onto a device at its AP, reports
// the outcome with the supplied reporter, and surfaces the device's post-restore
// LAN address.
//
// It returns the reporter's error so a partial section rejection (a result with
// Success=false and a nil top-level error) still yields a non-zero exit — never a
// false success — while always printing the recovered LAN address when the device
// rejoined, so a partial failure can be finished by hand. errPrefix labels a
// transport-level failure of the restore call itself.
//
// migrate and restore share this exact AP-hop sequence; centralizing it keeps the
// partial-failure handling (which has a history of device-stranding bugs) in one
// place so a fix lands for both commands at once. Callers differ only in the
// reporter (migration vs restore messaging) and the error prefix.
func RestoreAtAP(
	ctx context.Context,
	ios *iostreams.IOStreams,
	svc APRestorer,
	apSSID, apHostIP, name string,
	bkp *backup.DeviceBackup,
	opts backup.RestoreOptions,
	errPrefix string,
	report func(ios *iostreams.IOStreams, target string, result *backup.RestoreResult) error,
) error {
	var (
		result  *backup.RestoreResult
		newAddr string
	)
	err := RunWithSpinner(ctx, ios,
		fmt.Sprintf("Restoring onto %s at AP %s (hopping host WiFi)...", name, apSSID),
		func(ctx context.Context) error {
			var restoreErr error
			result, newAddr, restoreErr = svc.RestoreToAP(ctx, apSSID, apHostIP, name, bkp, opts)
			return restoreErr
		})
	if err != nil {
		return fmt.Errorf("%s: %w", errPrefix, err)
	}

	reportErr := report(ios, name, result)
	if newAddr != "" {
		ios.Info("%s is live at %s", name, newAddr)
	}
	return reportErr
}
