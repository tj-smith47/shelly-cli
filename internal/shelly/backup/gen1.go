package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tj-smith47/shelly-go/backup"
	"github.com/tj-smith47/shelly-go/gen1"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// exportGen1 creates a backup of a Gen1 device by reading its settings and actions.
// The backup is stored in the same Backup struct used for Gen2, with DeviceInfo.Generation=1.
func exportGen1(ctx context.Context, conn *client.Gen1Client) (*DeviceBackup, error) {
	bkp := &backup.Backup{
		Version:   backup.BackupVersion,
		CreatedAt: time.Now().UTC(),
	}

	// Get device info
	info := conn.Info()
	bkp.DeviceInfo = &backup.DeviceInfo{
		ID:         info.ID,
		MAC:        info.MAC,
		Model:      info.Model,
		Generation: info.Generation,
		Version:    info.Firmware,
		App:        info.App,
	}

	// Get full settings (this is Gen1's equivalent of Shelly.GetConfig)
	settings, err := conn.GetSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	// Store full settings as Config
	configData, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}
	bkp.Config = configData

	// Extract WiFi settings
	bkp.WiFi = marshalGen1WiFi(settings)

	// Extract MQTT settings
	if settings.MQTT != nil {
		bkp.MQTT = mustMarshal(settings.MQTT)
	}

	// Extract Cloud settings
	if settings.Cloud != nil {
		bkp.Cloud = mustMarshal(settings.Cloud)
	}

	// Extract Auth info
	if settings.Login != nil {
		bkp.Auth = &backup.AuthInfo{
			User:   settings.Login.Username,
			Enable: settings.Login.Enabled,
		}
	}

	// Extract component settings
	bkp.Components = marshalGen1Components(settings)

	// Extract schedule rules from relay/light settings
	bkp.Schedules = marshalGen1Schedules(settings)

	// Get action URLs (Gen1's equivalent of webhooks)
	actions, err := conn.GetActions(ctx)
	if err == nil && actions != nil {
		bkp.Webhooks = mustMarshal(actions)
	}

	return &DeviceBackup{Backup: bkp}, nil
}

// restoreGen1 restores a backup to a Gen1 device via individual HTTP settings calls.
func restoreGen1(ctx context.Context, conn *client.Gen1Client, bkp *DeviceBackup, opts RestoreOptions) (*RestoreResult, error) {
	result := &RestoreResult{
		Success: true,
	}
	dev := conn.Device()

	// Parse the full settings from backup Config
	var settings gen1.Settings
	if bkp.Config != nil {
		if err := json.Unmarshal(bkp.Config, &settings); err != nil {
			return nil, fmt.Errorf("failed to parse backup config: %w", err)
		}
	}

	// Restore device-level settings
	restoreGen1DeviceSettings(ctx, dev, &settings, result)

	// Restore WiFi (if not skipped)
	if !opts.SkipNetwork {
		restoreGen1WiFi(ctx, dev, bkp, result)
	}

	// Restore MQTT
	restoreGen1MQTT(ctx, dev, &settings, result)

	// Restore Cloud
	restoreGen1Cloud(ctx, dev, &settings, result)

	// Restore CoIoT
	restoreGen1CoIoT(ctx, dev, &settings, result)

	// Restore SNTP
	restoreGen1SNTP(ctx, dev, &settings, result)

	// Restore Auth (if not skipped)
	if !opts.SkipAuth {
		restoreGen1Auth(ctx, dev, bkp, result)
	}

	// Restore component configs (if not skipped via schedules/scripts)
	restoreGen1Components(ctx, dev, &settings, result)

	// Restore action URLs / webhooks (if not skipped)
	if !opts.SkipWebhooks {
		restoreGen1Actions(ctx, dev, bkp, result)
	}

	result.ConfigRestored = true
	return result, nil
}

// marshalGen1WiFi extracts WiFi settings from Gen1 Settings into a JSON blob.
func marshalGen1WiFi(settings *gen1.Settings) json.RawMessage {
	wifi := map[string]any{}
	if settings.WiFiSta != nil {
		wifi["sta"] = settings.WiFiSta
	}
	if settings.WiFiSta1 != nil {
		wifi["sta1"] = settings.WiFiSta1
	}
	if settings.WiFiAp != nil {
		wifi["ap"] = settings.WiFiAp
	}
	if settings.ApRoaming != nil {
		wifi["ap_roaming"] = settings.ApRoaming
	}
	if len(wifi) == 0 {
		return nil
	}
	return mustMarshal(wifi)
}

// marshalGen1Components extracts component settings into the Components map.
func marshalGen1Components(settings *gen1.Settings) map[string]json.RawMessage {
	components := map[string]json.RawMessage{}
	if len(settings.Lights) > 0 {
		components["lights"] = mustMarshal(settings.Lights)
	}
	if len(settings.Relays) > 0 {
		components["relays"] = mustMarshal(settings.Relays)
	}
	if len(settings.Rollers) > 0 {
		components["rollers"] = mustMarshal(settings.Rollers)
	}
	if len(settings.Meters) > 0 {
		components["meters"] = mustMarshal(settings.Meters)
	}
	if len(settings.EMeters) > 0 {
		components["emeters"] = mustMarshal(settings.EMeters)
	}
	if len(components) == 0 {
		return nil
	}
	return components
}

// marshalGen1Schedules extracts schedule rules from relay and light settings.
func marshalGen1Schedules(settings *gen1.Settings) json.RawMessage {
	type scheduleEntry struct {
		Component string   `json:"component"`
		ID        int      `json:"id"`
		Enabled   bool     `json:"enabled"`
		Rules     []string `json:"rules"`
	}

	var entries []scheduleEntry
	for i, relay := range settings.Relays {
		if len(relay.ScheduleRules) > 0 {
			entries = append(entries, scheduleEntry{
				Component: "relay",
				ID:        i,
				Enabled:   relay.Schedule,
				Rules:     relay.ScheduleRules,
			})
		}
	}
	for i, light := range settings.Lights {
		if len(light.ScheduleRules) > 0 {
			entries = append(entries, scheduleEntry{
				Component: "light",
				ID:        i,
				Enabled:   light.Schedule,
				Rules:     light.ScheduleRules,
			})
		}
	}
	if len(entries) == 0 {
		return nil
	}
	return mustMarshal(entries)
}

// restoreGen1DeviceSettings restores device-level settings (name, timezone, etc.).
func restoreGen1DeviceSettings(ctx context.Context, dev *gen1.Device, settings *gen1.Settings, result *RestoreResult) {
	if settings.Name != "" {
		if err := dev.SetName(ctx, settings.Name); err != nil {
			addWarning(result, "set name: %v", err)
		}
	}
	if settings.Tz != "" {
		if err := dev.SetTimezone(ctx, settings.Tz); err != nil {
			addWarning(result, "set timezone: %v", err)
		}
	}
	if settings.Lat != 0 || settings.Lng != 0 {
		if err := dev.SetLocation(ctx, settings.Lat, settings.Lng); err != nil {
			addWarning(result, "set location: %v", err)
		}
	}
	if settings.Mode != "" {
		if err := dev.SetMode(ctx, settings.Mode); err != nil {
			addWarning(result, "set mode: %v", err)
		}
	}
	if err := dev.SetDiscoverable(ctx, settings.Discoverable); err != nil {
		addWarning(result, "set discoverable: %v", err)
	}
	if settings.MaxPower > 0 {
		if err := dev.SetMaxPower(ctx, settings.MaxPower); err != nil {
			addWarning(result, "set max power: %v", err)
		}
	}
}

// restoreGen1WiFi restores WiFi settings from the backup.
func restoreGen1WiFi(ctx context.Context, dev *gen1.Device, bkp *DeviceBackup, result *RestoreResult) {
	if bkp.WiFi == nil {
		return
	}
	var wifi struct {
		Sta       *gen1.WiFiStaSettings   `json:"sta,omitempty"`
		Sta1      *gen1.WiFiStaSettings   `json:"sta1,omitempty"`
		Ap        *gen1.WiFiApSettings    `json:"ap,omitempty"`
		ApRoaming *gen1.ApRoamingSettings `json:"ap_roaming,omitempty"`
	}
	if err := json.Unmarshal(bkp.WiFi, &wifi); err != nil {
		addWarning(result, "parse WiFi config: %v", err)
		return
	}

	if wifi.Sta != nil {
		restoreGen1WiFiStation(ctx, dev, wifi.Sta, result)
		result.RestartRequired = true
	}
	if wifi.Ap != nil {
		if err := dev.SetWiFiAP(ctx, wifi.Ap.Enabled, wifi.Ap.SSID, wifi.Ap.Key); err != nil {
			addWarning(result, "set WiFi AP: %v", err)
		}
	}
}

// restoreGen1WiFiStation restores a WiFi station configuration.
func restoreGen1WiFiStation(ctx context.Context, dev *gen1.Device, sta *gen1.WiFiStaSettings, result *RestoreResult) {
	if sta.Ipv4Method == "static" {
		err := dev.SetWiFiStationStatic(ctx, sta.SSID, sta.Key, sta.IP, sta.Gw, sta.Mask, sta.DNS)
		if err != nil {
			addWarning(result, "set WiFi station static: %v", err)
		}
		return
	}
	if err := dev.SetWiFiStation(ctx, sta.Enabled, sta.SSID, sta.Key); err != nil {
		addWarning(result, "set WiFi station: %v", err)
	}
}

// restoreGen1MQTT restores MQTT settings.
func restoreGen1MQTT(ctx context.Context, dev *gen1.Device, settings *gen1.Settings, result *RestoreResult) {
	if settings.MQTT == nil {
		return
	}
	cfg := &gen1.MQTTConfig{
		Enable:              settings.MQTT.Enable,
		Server:              settings.MQTT.Server,
		User:                settings.MQTT.User,
		Password:            settings.MQTT.Pass,
		ID:                  settings.MQTT.ID,
		KeepAlive:           settings.MQTT.KeepAlive,
		MaxQos:              settings.MQTT.MaxQos,
		CleanSession:        settings.MQTT.CleanSession,
		Retain:              settings.MQTT.Retain,
		UpdatePeriod:        settings.MQTT.UpdatePeriod,
		ReconnectTimeoutMax: settings.MQTT.ReconnectTimeoutMax,
		ReconnectTimeoutMin: settings.MQTT.ReconnectTimeoutMin,
	}
	if err := dev.SetMQTTConfig(ctx, cfg); err != nil {
		addWarning(result, "set MQTT: %v", err)
	}
}

// restoreGen1Cloud restores Cloud settings.
func restoreGen1Cloud(ctx context.Context, dev *gen1.Device, settings *gen1.Settings, result *RestoreResult) {
	if settings.Cloud == nil {
		return
	}
	if err := dev.SetCloud(ctx, settings.Cloud.Enabled); err != nil {
		addWarning(result, "set cloud: %v", err)
	}
}

// restoreGen1CoIoT restores CoIoT protocol settings.
func restoreGen1CoIoT(ctx context.Context, dev *gen1.Device, settings *gen1.Settings, result *RestoreResult) {
	if settings.CoIoT == nil {
		return
	}
	if err := dev.SetCoIoT(ctx, settings.CoIoT.Enabled, settings.CoIoT.UpdatePeriod, settings.CoIoT.Peer); err != nil {
		addWarning(result, "set CoIoT: %v", err)
	}
}

// restoreGen1SNTP restores time server settings.
func restoreGen1SNTP(ctx context.Context, dev *gen1.Device, settings *gen1.Settings, result *RestoreResult) {
	if settings.SNTP == nil {
		return
	}
	if settings.SNTP.Server != "" {
		if err := dev.SetTimeServer(ctx, settings.SNTP.Server); err != nil {
			addWarning(result, "set time server: %v", err)
		}
	}
}

// restoreGen1Auth restores authentication settings.
func restoreGen1Auth(ctx context.Context, dev *gen1.Device, bkp *DeviceBackup, result *RestoreResult) {
	if bkp.Auth == nil {
		return
	}
	// Note: Gen1 auth restore enables/disables auth but cannot restore the password
	// (passwords are never exported). User must set password separately if needed.
	if err := dev.SetAuth(ctx, bkp.Auth.Enable, bkp.Auth.User, ""); err != nil {
		addWarning(result, "set auth: %v", err)
	}
}

// restoreGen1Components restores component-specific configurations.
func restoreGen1Components(ctx context.Context, dev *gen1.Device, settings *gen1.Settings, result *RestoreResult) {
	for i, light := range settings.Lights {
		cfg := gen1.LightConfig{
			Name:         light.Name,
			DefaultState: light.DefaultState,
			BtnType:      light.BtnType,
			AutoOn:       light.AutoOn,
			AutoOff:      light.AutoOff,
			BtnReverse:   light.BtnReverse != 0,
			Schedule:     light.Schedule,
		}
		if err := dev.SetLightConfig(ctx, i, cfg); err != nil {
			addWarning(result, "set light %d config: %v", i, err)
		}
	}

	for i, relay := range settings.Relays {
		cfg := &gen1.RelayConfig{
			Name:         relay.Name,
			DefaultState: relay.DefaultState,
			BtnType:      relay.BtnType,
			AutoOn:       relay.AutoOn,
			AutoOff:      relay.AutoOff,
			MaxPower:     relay.MaxPower,
			BtnReverse:   relay.IsBtnReverse(),
			Schedule:     relay.Schedule,
		}
		if err := dev.SetRelayConfig(ctx, i, cfg); err != nil {
			addWarning(result, "set relay %d config: %v", i, err)
		}
	}

	for i, roller := range settings.Rollers {
		cfg := &gen1.RollerConfig{
			MaxTimeOpen:    roller.MaxTimeOpen,
			MaxTimeClose:   roller.MaxTimeClose,
			DefaultState:   roller.DefaultState,
			Swap:           roller.Swap,
			SwapInputs:     roller.SwapInputs,
			InputMode:      roller.InputMode,
			BtnType:        roller.BtnType,
			BtnReverse:     roller.BtnReverse != 0,
			SafetyMode:     roller.SafetyMode,
			SafetyAction:   roller.SafetyAction,
			ObstacleMode:   roller.ObstacleMode,
			ObstacleAction: roller.ObstacleAction,
			ObstaclePower:  roller.ObstaclePower,
			ObstacleDelay:  roller.ObstacleDelay,
			Positioning:    roller.Positioning,
		}
		if err := dev.SetRollerConfig(ctx, i, cfg); err != nil {
			addWarning(result, "set roller %d config: %v", i, err)
		}
	}
}

// restoreGen1Actions restores action URLs (Gen1's equivalent of webhooks).
func restoreGen1Actions(ctx context.Context, dev *gen1.Device, bkp *DeviceBackup, result *RestoreResult) {
	if bkp.Webhooks == nil {
		return
	}
	var actions gen1.ActionSettings
	if err := json.Unmarshal(bkp.Webhooks, &actions); err != nil {
		addWarning(result, "parse actions: %v", err)
		return
	}

	for _, action := range actions.Actions {
		if len(action.URLs) > 0 {
			if err := dev.SetAction(ctx, action.Index, action.Event, action.URLs, action.Enabled); err != nil {
				addWarning(result, "set action %s: %v", action.Event, err)
			}
		}
	}
	result.WebhooksRestored = len(actions.Actions)
}

// addWarning adds a warning message to the restore result.
func addWarning(result *RestoreResult, format string, args ...any) {
	result.Warnings = append(result.Warnings, fmt.Sprintf(format, args...))
}

// mustMarshal marshals a value to JSON, panicking on error.
// Only used for values that are known to be valid JSON-serializable types.
func mustMarshal(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("backup: failed to marshal %T: %v", v, err))
	}
	return data
}
