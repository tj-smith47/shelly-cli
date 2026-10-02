package shelly

import (
	"context"
	"fmt"
	"maps"
	"strings"
)

// stationMode says how a Shelly.SetConfig wifi section is treated before it
// is sent.
type stationMode int

const (
	// stationsCrossDevice is a template from another device: its station
	// address is that device's identity and is never copied.
	stationsCrossDevice stationMode = iota
	// stationsSameDevice is a capture of the same device, or a config the
	// user hands over as written (sync push, SetConfig).
	stationsSameDevice
	// stationsOmit drops every station; the device's network was set by
	// another step.
	stationsOmit
	// stationsMatchMAC is a config file of unknown origin (config import): it
	// is treated as the same device only when its sys.device.mac is the
	// target's, and as another device's otherwise, so a file from another
	// device never hands its station address to the target.
	stationsMatchMAC
)

const componentWiFi = "wifi"

// resolve settles stationsMatchMAC against the device's current config.
func (m stationMode) resolve(cfg, current map[string]any) stationMode {
	if m != stationsMatchMAC {
		return m
	}
	mac := configMAC(cfg)
	if mac != "" && mac == configMAC(current) {
		return stationsSameDevice
	}
	return stationsCrossDevice
}

// configMAC returns a config's sys.device.mac, upper case without separators,
// or "" when it has none.
func configMAC(cfg map[string]any) string {
	mac := field[string](field[map[string]any](field[map[string]any](cfg, "sys"), "device"), "mac")
	return strings.ToUpper(strings.NewReplacer(":", "", "-", "").Replace(mac))
}

var (
	stationNames   = []string{fieldSTA, fieldSTA1}
	stationAddress = []string{"ip", fieldNetmask, "gw", fieldNameserver, fieldIPv4Mode}
)

// field returns m[key] as a T, or the zero T when it is missing or of
// another type.
func field[T any](m map[string]any, key string) T {
	v, ok := m[key].(T)
	if !ok {
		var zero T
		return zero
	}
	return v
}

// configHasStation reports whether cfg carries a wifi sta or sta1 object.
func configHasStation(cfg map[string]any) bool {
	wifi, ok := cfg[componentWiFi].(map[string]any)
	if !ok {
		return false
	}
	for _, name := range stationNames {
		if _, ok := wifi[name].(map[string]any); ok {
			return true
		}
	}
	return false
}

// planConfigStations returns cfg with its wifi stations reduced to what may
// be written to a device whose current config is current, and a warning for
// each station left out. A pass the config gives is written as given. Without
// one, a station is written with no key when the device is already on its
// network (the device keeps its own), with this host's stored passphrase when
// the network changes, and left out when no passphrase is found: a changed
// network written with no pass keeps the old network's key and never joins.
// cfg is not modified.
func (s *Service) planConfigStations(ctx context.Context, cfg, current map[string]any, mode stationMode) (planned map[string]any, warnings []string) {
	src, ok := cfg[componentWiFi].(map[string]any)
	if !ok || !configHasStation(cfg) {
		return cfg, nil
	}
	out := maps.Clone(cfg)
	wifi := maps.Clone(src)
	out[componentWiFi] = wifi
	currentWiFi := field[map[string]any](current, componentWiFi)
	mode = mode.resolve(cfg, current)

	for _, name := range stationNames {
		sta, ok := wifi[name].(map[string]any)
		if !ok {
			continue
		}
		if mode == stationsOmit {
			delete(wifi, name)
			continue
		}
		station, warning := s.planConfigStation(ctx, name, sta, currentWiFi, mode)
		if station == nil {
			delete(wifi, name)
		} else {
			wifi[name] = station
		}
		if warning != "" {
			warnings = append(warnings, warning)
		}
	}
	if len(wifi) == 0 {
		delete(out, componentWiFi)
	}
	return out, warnings
}

// planConfigStation plans one station; a nil result leaves it out.
func (s *Service) planConfigStation(ctx context.Context, name string, sta, currentWiFi map[string]any, mode stationMode) (planned map[string]any, warning string) {
	planned = maps.Clone(sta)
	given := field[string](planned, "pass")
	delete(planned, "pass")
	open := field[bool](planned, "is_open")
	delete(planned, "is_open")
	if mode == stationsCrossDevice {
		for _, key := range stationAddress {
			delete(planned, key)
		}
	}
	ssid := field[string](planned, "ssid")
	curSta := field[map[string]any](currentWiFi, name)
	curSSID := field[string](curSta, "ssid")
	keepsAddress := func() string {
		if ssid == "" || ssid == curSSID || planned[fieldIPv4Mode] != nil {
			return ""
		}
		return keepsStaticWarning(name, ssid, curSta)
	}
	if given != "" {
		planned["pass"] = given
		return planned, keepsAddress()
	}

	switch {
	case ssid == curSSID && mode == stationsCrossDevice:
		return nil, ""
	case ssid == curSSID:
		return planned, ""
	case ssid == "":
		// A station that names no network would erase the device's one.
		return nil, ""
	case open && mode == stationsSameDevice:
		planned["pass"] = ""
		return planned, keepsAddress()
	}
	pass, err := s.hostPassphrase(ctx, ssid)
	if err != nil {
		return nil, fmt.Sprintf("WiFi station %s (%q) was not written: the device is on another network and no "+
			"password for %q was found on this host; set it with `shelly wifi set`", name, ssid, ssid)
	}
	planned["pass"] = pass
	return planned, keepsAddress()
}

// keepsStaticWarning warns when a station moves to ssid while the device's
// current station cur keeps a static address the write does not replace: the
// address was chosen for the old network and may not route on the new one.
func keepsStaticWarning(name, ssid string, cur map[string]any) string {
	if field[string](cur, fieldIPv4Mode) != ipv4Static {
		return ""
	}
	return staticKeptWarning(name, ssid, field[string](cur, "ip"))
}

// staticKeptWarning is the warning for a station that joins ssid with the
// static address ip it had on its previous network.
func staticKeptWarning(name, ssid, ip string) string {
	return fmt.Sprintf("WiFi station %s joins %q but keeps its static address %s from its previous network; "+
		"set an address for %q with `shelly wifi set --static-ip`", name, ssid, ip, ssid)
}

// describeStations returns one dry-run line per station cfg writes, saying
// which network it joins, whether a password is sent and what address it
// gets, without ever showing the password.
func describeStations(cfg map[string]any) []string {
	wifi := field[map[string]any](cfg, componentWiFi)
	var lines []string
	for _, name := range stationNames {
		sta, ok := wifi[name].(map[string]any)
		if !ok {
			continue
		}
		key := "no password sent, the device keeps its own"
		if pass, sent := sta["pass"].(string); sent && pass == "" {
			key = "open, no password"
		} else if sent {
			key = "password sent (not shown)"
		}
		address := "address unchanged"
		switch field[string](sta, fieldIPv4Mode) {
		case ipv4Static:
			address = "static " + field[string](sta, "ip")
		case ipv4DHCP:
			address = "DHCP"
		}
		lines = append(lines, fmt.Sprintf("wifi.%s: %q, %s, %s", name, field[string](sta, "ssid"), key, address))
	}
	return lines
}
