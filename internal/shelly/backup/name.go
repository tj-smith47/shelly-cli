package backup

import (
	"cmp"
	"encoding/json"
	"fmt"
	"strings"
)

// RecordedName returns the device name the backup recorded: its device info
// name, else the configured name (Gen1 settings "name", Gen2+
// sys.device.name), or "" when it records none.
func (b *DeviceBackup) RecordedName() string {
	if name := b.Device().Name; name != "" || b.Backup == nil {
		return name
	}
	var cfg struct {
		Name string `json:"name"`
		Sys  struct {
			Device struct {
				Name string `json:"name"`
			} `json:"device"`
		} `json:"sys"`
	}
	if json.Unmarshal(b.Config, &cfg) != nil {
		return ""
	}
	return cmp.Or(cfg.Sys.Device.Name, cfg.Name)
}

// ResolveName returns the device name a restore writes, or "" to leave the
// name the backup recorded. An explicit name (--name) always wins. The alias
// (the target's registry name) is written only when the backup came from a
// different device: its MAC differs from targetMAC, or either MAC is unknown.
// MACs are compared case- and separator-insensitively; targetMAC may also be
// the 6-hex suffix of a factory AP SSID, which is the last three bytes of the
// device's MAC.
func ResolveName(explicit, alias, backupMAC, targetMAC string) string {
	if explicit != "" {
		return explicit
	}
	if sameDevice(backupMAC, targetMAC) {
		return ""
	}
	return alias
}

// nameUnchanged is the dry-run line when no name will be written.
const nameUnchanged = "name: left unchanged (the backup records none)"

// DescribeName returns the dry-run line for the name decision ResolveName
// makes. backupName is the name the backup recorded.
func DescribeName(explicit, alias, backupName, backupMAC, targetMAC string) string {
	switch {
	case explicit != "":
		return fmt.Sprintf("name: %q from --name", explicit)
	case backupName == "" && (alias == "" || sameDevice(backupMAC, targetMAC)):
		return nameUnchanged
	case sameDevice(backupMAC, targetMAC):
		return fmt.Sprintf("name: keeps the backup's %q (same device)", backupName)
	case alias != "":
		return fmt.Sprintf("name: %q from the alias (different device)", alias)
	default:
		return fmt.Sprintf("name: keeps the backup's %q (the target has no alias)", backupName)
	}
}

// sameDevice reports whether a backup's MAC names the target: equal MACs, or
// a 6-hex target that matches the backup MAC's last three bytes.
func sameDevice(backupMAC, targetMAC string) bool {
	b, t := macHex(backupMAC), macHex(targetMAC)
	if len(b) != 12 {
		return false
	}
	switch len(t) {
	case 12:
		return b == t
	case 6:
		return strings.HasSuffix(b, t)
	default:
		return false
	}
}

// macHex returns mac lowercased with its separators removed.
func macHex(mac string) string {
	return strings.ToLower(strings.NewReplacer(":", "", "-", "", ".", "").Replace(mac))
}
