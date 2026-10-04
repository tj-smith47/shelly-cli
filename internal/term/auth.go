package term

import (
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// DisplayAuthTestPassed reports a device that answered an authenticated
// request made with dev's credentials. For a device with authentication
// disabled it says so, because such a device accepts any credentials.
func DisplayAuthTestPassed(ios *iostreams.IOStreams, dev model.Device, info *shelly.DeviceInfo) {
	switch {
	case !info.AuthEn:
		ios.Warning("Authentication is not enabled on %s; it accepts any credentials", dev.DisplayName())
	case dev.HasAuth():
		ios.Success("%s accepted the password for user %s", dev.DisplayName(), dev.Auth.Username)
	default:
		ios.Success("%s answered without credentials", dev.DisplayName())
	}
	ios.Println("")
	ios.Printf("Device: %s\n", dev.DisplayName())
	ios.Printf("ID: %s\n", info.ID)
	ios.Printf("MAC: %s\n", model.NormalizeMAC(info.MAC))
	ios.Printf("Generation: %d\n", info.Generation)
}

// DisplayStoredCredentials says whether the new credentials for device were
// saved to the config. A device that is not registered was reached by
// address, so nothing could be stored and every following command must be
// given the password.
func DisplayStoredCredentials(ios *iostreams.IOStreams, device string, stored bool) {
	if stored {
		ios.Info("Saved the new credentials for %s to the config", device)
		return
	}
	ios.Warning("%s is not a registered device, so the new password was not saved; register it with `shelly device add` and --password", device)
}
