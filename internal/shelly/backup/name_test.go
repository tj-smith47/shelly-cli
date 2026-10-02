package backup

import (
	"encoding/json"
	"testing"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
)

func TestResolveName(t *testing.T) {
	t.Parallel()
	const mac = "AA:BB:CC:DD:EE:FF"
	tests := []struct {
		name                        string
		explicit, alias, bMAC, tMAC string
		want                        string
	}{
		{name: "same MAC keeps the backup's name", alias: "gb", bMAC: mac, tMAC: "AABBCCDDEEFF"},
		{name: "case and separators ignored", alias: "gb", bMAC: "aa-bb-cc-dd-ee-ff", tMAC: "AABB.CCDD.EEFF"},
		{name: "different MAC takes the alias", alias: "gb", bMAC: mac, tMAC: "112233445566", want: "gb"},
		{name: "unknown target MAC counts as different", alias: "gb", bMAC: mac, want: "gb"},
		{name: "unknown backup MAC counts as different", alias: "gb", tMAC: mac, want: "gb"},
		{name: "explicit wins on the same device", explicit: "Kitchen", alias: "gb", bMAC: mac, tMAC: mac, want: "Kitchen"},
		{name: "explicit wins on another device", explicit: "Kitchen", alias: "gb", bMAC: mac, tMAC: "112233445566", want: "Kitchen"},
		{name: "no alias writes nothing", bMAC: mac, tMAC: "112233445566"},
		{name: "AP suffix of the same device", alias: "gb", bMAC: mac, tMAC: "ddeeff"},
		{name: "AP suffix of another device", alias: "gb", bMAC: mac, tMAC: "d12965", want: "gb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ResolveName(tt.explicit, tt.alias, tt.bMAC, tt.tMAC); got != tt.want {
				t.Errorf("ResolveName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDescribeName(t *testing.T) {
	t.Parallel()
	const mac = "AA:BB:CC:DD:EE:FF"
	tests := []struct {
		explicit, alias, bName, tMAC, want string
	}{
		{bName: "GB", alias: "gb", tMAC: mac, want: `name: keeps the backup's "GB" (same device)`},
		{bName: "GB", alias: "gb", tMAC: "112233445566", want: `name: "gb" from the alias (different device)`},
		{bName: "GB", explicit: "X", alias: "gb", tMAC: mac, want: `name: "X" from --name`},
		{bName: "GB", tMAC: "112233445566", want: `name: keeps the backup's "GB" (the target has no alias)`},
		{alias: "gb", tMAC: mac, want: nameUnchanged},
		{tMAC: "112233445566", want: nameUnchanged},
		{alias: "gb", tMAC: "112233445566", want: `name: "gb" from the alias (different device)`},
	}
	for _, tt := range tests {
		if got := DescribeName(tt.explicit, tt.alias, tt.bName, mac, tt.tMAC); got != tt.want {
			t.Errorf("DescribeName(%q, %q, %q, %q) = %q, want %q", tt.explicit, tt.alias, tt.bName, tt.tMAC, got, tt.want)
		}
	}
}

func TestDeviceBackup_RecordedName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		info   string
		config string
		want   string
	}{
		{name: "device info name", info: "GB", config: `{"sys":{"device":{"name":"other"}}}`, want: "GB"},
		{name: "gen2 sys.device.name", config: `{"sys":{"device":{"name":"GB"}}}`, want: "GB"},
		{name: "gen1 settings name", config: `{"name":"GB","fw":"x"}`, want: "GB"},
		{name: "none recorded", config: `{"fw":"x"}`},
		{name: "unparseable config", config: `[1]`},
	}
	for _, tt := range tests {
		b := &DeviceBackup{Backup: &shellybackup.Backup{
			DeviceInfo: &shellybackup.DeviceInfo{Name: tt.info},
			Config:     json.RawMessage(tt.config),
		}}
		if got := b.RecordedName(); got != tt.want {
			t.Errorf("%s: RecordedName() = %q, want %q", tt.name, got, tt.want)
		}
	}
}
