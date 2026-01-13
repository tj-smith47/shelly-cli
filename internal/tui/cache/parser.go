// Package cache provides a shared device data cache for the TUI.
// This file contains unified status parsing for both HTTP and WebSocket sources.
package cache

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Component type constants for parsing and matching.
const (
	ComponentSwitch      = "switch"
	ComponentLight       = "light"
	ComponentCover       = "cover"
	ComponentInput       = "input"
	ComponentPM          = "pm"
	ComponentPM1         = "pm1"
	ComponentEM          = "em"
	ComponentEM1         = "em1"
	ComponentTemperature = "temperature"
	ComponentSys         = "sys"
	ComponentWiFi        = "wifi"
)

// Gen1 status field names.
const (
	Gen1Relays         = "relays"
	Gen1Meters         = "meters"
	Gen1EMeters        = "emeters"
	Gen1Lights         = "lights"
	Gen1Rollers        = "rollers"
	Gen1Inputs         = "inputs"
	Gen1Temperature    = "temperature"
	Gen1ExtTemperature = "ext_temperature"
	Gen1WiFiSta        = "wifi_sta"
)

// Cover state constants.
const (
	CoverStateOpen    = "open"
	CoverStateClosed  = "closed"
	CoverStateStopped = "stopped"
)

// ParsedStatus holds all possible status fields from a device.
// This is the unified intermediate representation used by both HTTP and WebSocket parsing.
type ParsedStatus struct {
	// Device identification (from sys or device info)
	DeviceID   string
	Generation int
	Model      string
	MAC        string

	// Component states
	Switches []SwitchState
	Lights   []LightState
	Covers   []CoverState
	Inputs   []InputState

	// Per-component power tracking for accurate WebSocket aggregation
	SwitchPowers map[int]float64
	LightPowers  map[int]float64
	CoverPowers  map[int]float64
	PMPowers     map[int]float64
	EMPowers     map[int]float64
	EM1Powers    map[int]float64

	// Power/energy metrics
	Power       float64
	Voltage     float64
	Current     float64
	TotalEnergy float64
	Temperature float64

	// Power monitoring components (for snapshot)
	PM  []model.PMStatus
	EM  []model.EMStatus
	EM1 []model.EM1Status

	// System info (optional - only if present in response)
	WiFi *WiFiInfo
	Sys  *SysInfo

	// Raw component map for extensibility (keeps unparsed components)
	Components map[string]json.RawMessage
}

// WiFiInfo holds parsed WiFi status from Shelly.GetStatus response.
type WiFiInfo struct {
	StaIP  string `json:"sta_ip"`
	Status string `json:"status"`
	SSID   string `json:"ssid"`
	RSSI   int    `json:"rssi"`
}

// SysInfo holds parsed system status from Shelly.GetStatus response.
type SysInfo struct {
	MAC             string `json:"mac"`
	Uptime          int    `json:"uptime"`
	Time            string `json:"time,omitempty"`
	RAMFree         int    `json:"ram_free"`
	RAMSize         int    `json:"ram_size"`
	FSFree          int    `json:"fs_free"`
	FSSize          int    `json:"fs_size"`
	RestartRequired bool   `json:"restart_required"`
}

// ComponentCapabilities describes what components a device has.
// This allows optimized fetching - we only try to parse components the device has.
type ComponentCapabilities struct {
	HasSwitches bool
	HasLights   bool
	HasCovers   bool
	HasPM       bool // Power monitoring (switches with apower)
	HasEM       bool // Energy metering (Pro 3EM, etc.)
	NumSwitches int  // Expected count (0 = unknown, detect from response)
	NumLights   int
	NumCovers   int
}

// unmarshalJSON is a generic helper to unmarshal JSON into any type.
func unmarshalJSON[T any](raw json.RawMessage) (*T, error) {
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// ParseFullStatus parses a complete Shelly.GetStatus response (Gen2+) into ParsedStatus.
// The input map contains component names as keys (e.g., "switch:0", "wifi", "sys")
// and their status as json.RawMessage for flexible parsing.
func ParseFullStatus(deviceID string, statusMap map[string]json.RawMessage) (*ParsedStatus, error) {
	result := &ParsedStatus{
		DeviceID:     deviceID,
		Components:   statusMap,
		SwitchPowers: make(map[int]float64),
		LightPowers:  make(map[int]float64),
		CoverPowers:  make(map[int]float64),
		PMPowers:     make(map[int]float64),
		EMPowers:     make(map[int]float64),
		EM1Powers:    make(map[int]float64),
	}

	// Parse each component by type
	for key, rawStatus := range statusMap {
		parseGen2Component(result, key, rawStatus)
	}

	return result, nil
}

// parseGen2Component routes parsing to the appropriate handler based on component prefix.
func parseGen2Component(result *ParsedStatus, key string, rawStatus json.RawMessage) {
	prefix, _ := ParseComponentName(key)
	switch prefix {
	case ComponentSwitch:
		parseSwitchComponent(result, rawStatus)
	case ComponentLight:
		parseLightComponent(result, rawStatus)
	case ComponentCover:
		parseCoverComponent(result, rawStatus)
	case ComponentInput:
		parseInputComponent(result, rawStatus)
	case ComponentPM, ComponentPM1:
		parsePMComponent(result, rawStatus)
	case ComponentEM:
		parseEMComponent(result, rawStatus)
	case ComponentEM1:
		parseEM1Component(result, rawStatus)
	case ComponentTemperature:
		parseTemperatureComponent(result, rawStatus)
	case ComponentSys:
		parseSysComponent(result, rawStatus)
	case ComponentWiFi:
		parseWiFiComponent(result, rawStatus)
	}
}

// switchStatusData holds parsed switch status fields.
type switchStatusData struct {
	ID      int      `json:"id"`
	Output  bool     `json:"output"`
	Source  string   `json:"source,omitempty"`
	APower  *float64 `json:"apower,omitempty"`
	Voltage *float64 `json:"voltage,omitempty"`
	Current *float64 `json:"current,omitempty"`
	AEnergy *struct {
		Total float64 `json:"total"`
	} `json:"aenergy,omitempty"`
	Temperature *struct {
		TC float64 `json:"tC"`
	} `json:"temperature,omitempty"`
}

// parseSwitchComponent parses a switch component status.
func parseSwitchComponent(result *ParsedStatus, rawStatus json.RawMessage) {
	sw, err := unmarshalJSON[switchStatusData](rawStatus)
	if err != nil {
		return
	}

	result.Switches = append(result.Switches, SwitchState{
		ID:     sw.ID,
		On:     sw.Output,
		Source: sw.Source,
	})

	// Accumulate power metrics (switches without PM components report these directly)
	accumulateSwitchMetrics(result, sw)
}

// accumulateSwitchMetrics adds switch power metrics to the result.
func accumulateSwitchMetrics(result *ParsedStatus, sw *switchStatusData) {
	if sw.APower != nil {
		result.Power += *sw.APower
		if result.SwitchPowers != nil {
			result.SwitchPowers[sw.ID] = *sw.APower
		}
	}
	if sw.Voltage != nil && result.Voltage == 0 {
		result.Voltage = *sw.Voltage
	}
	if sw.Current != nil && result.Current == 0 {
		result.Current = *sw.Current
	}
	if sw.AEnergy != nil {
		result.TotalEnergy += sw.AEnergy.Total
	}
	if sw.Temperature != nil && result.Temperature == 0 {
		result.Temperature = sw.Temperature.TC
	}
}

// lightStatusData holds parsed light status fields.
type lightStatusData struct {
	ID      int      `json:"id"`
	Output  bool     `json:"output"`
	APower  *float64 `json:"apower,omitempty"`
	Voltage *float64 `json:"voltage,omitempty"`
	Current *float64 `json:"current,omitempty"`
	AEnergy *struct {
		Total float64 `json:"total"`
	} `json:"aenergy,omitempty"`
}

// parseLightComponent parses a light component status.
func parseLightComponent(result *ParsedStatus, rawStatus json.RawMessage) {
	lt, err := unmarshalJSON[lightStatusData](rawStatus)
	if err != nil {
		return
	}
	result.Lights = append(result.Lights, LightState{ID: lt.ID, On: lt.Output})
	accumulateLightMetrics(result, lt)
}

// accumulateLightMetrics adds light power metrics to the result.
func accumulateLightMetrics(result *ParsedStatus, lt *lightStatusData) {
	if lt.APower != nil {
		if result.LightPowers != nil {
			result.LightPowers[lt.ID] = *lt.APower
		}
		result.Power += *lt.APower
	}
	if lt.Voltage != nil && *lt.Voltage > 0 {
		result.Voltage = *lt.Voltage
	}
	if lt.Current != nil {
		result.Current += *lt.Current
	}
	if lt.AEnergy != nil {
		result.TotalEnergy += lt.AEnergy.Total
	}
}

// coverStatusData holds parsed cover status fields.
type coverStatusData struct {
	ID      int     `json:"id"`
	State   string  `json:"state"`
	APower  float64 `json:"apower,omitempty"`
	Voltage float64 `json:"voltage,omitempty"`
	Current float64 `json:"current,omitempty"`
}

// parseCoverComponent parses a cover component status.
func parseCoverComponent(result *ParsedStatus, rawStatus json.RawMessage) {
	cv, err := unmarshalJSON[coverStatusData](rawStatus)
	if err != nil {
		return
	}

	result.Covers = append(result.Covers, CoverState{ID: cv.ID, State: cv.State})
	accumulateCoverMetrics(result, cv)
}

// accumulateCoverMetrics adds cover power metrics to the result.
func accumulateCoverMetrics(result *ParsedStatus, cv *coverStatusData) {
	if cv.APower > 0 {
		result.Power += cv.APower
		if result.CoverPowers != nil {
			result.CoverPowers[cv.ID] = cv.APower
		}
	}
	if cv.Voltage > 0 && result.Voltage == 0 {
		result.Voltage = cv.Voltage
	}
	if cv.Current > 0 && result.Current == 0 {
		result.Current = cv.Current
	}
}

// inputStatusData holds parsed input status fields.
type inputStatusData struct {
	ID      int      `json:"id"`
	State   *bool    `json:"state,omitempty"`
	Percent *float64 `json:"percent,omitempty"`
}

// parseInputComponent parses an input component status.
func parseInputComponent(result *ParsedStatus, rawStatus json.RawMessage) {
	inp, err := unmarshalJSON[inputStatusData](rawStatus)
	if err != nil {
		return
	}

	state := false
	if inp.State != nil {
		state = *inp.State
	}

	// Type and Name are only available in config, not status
	result.Inputs = append(result.Inputs, InputState{
		ID:    inp.ID,
		State: state,
	})
}

// parsePMComponent parses a PM/PM1 (power meter) component status.
func parsePMComponent(result *ParsedStatus, rawStatus json.RawMessage) {
	pm, err := unmarshalJSON[model.PMStatus](rawStatus)
	if err != nil {
		return
	}
	result.PM = append(result.PM, *pm)
	accumulatePMMetrics(result, pm)
}

// accumulatePMMetrics adds PM power metrics to the result.
func accumulatePMMetrics(result *ParsedStatus, pm *model.PMStatus) {
	result.Power += pm.APower
	if result.PMPowers != nil {
		result.PMPowers[pm.ID] = pm.APower
	}
	if result.Voltage == 0 && pm.Voltage > 0 {
		result.Voltage = pm.Voltage
	}
	if result.Current == 0 && pm.Current > 0 {
		result.Current = pm.Current
	}
	if pm.AEnergy != nil {
		result.TotalEnergy += pm.AEnergy.Total
	}
}

// parseEMComponent parses an EM (3-phase energy meter) component status.
func parseEMComponent(result *ParsedStatus, rawStatus json.RawMessage) {
	em, err := unmarshalJSON[model.EMStatus](rawStatus)
	if err != nil {
		return
	}
	result.EM = append(result.EM, *em)
	result.Power += em.TotalActivePower
	if result.EMPowers != nil {
		result.EMPowers[em.ID] = em.TotalActivePower
	}
	result.Current += em.TotalCurrent
	if result.Voltage == 0 && em.AVoltage > 0 {
		result.Voltage = em.AVoltage
	}
}

// parseEM1Component parses an EM1 (single-phase energy meter) component status.
func parseEM1Component(result *ParsedStatus, rawStatus json.RawMessage) {
	em1, err := unmarshalJSON[model.EM1Status](rawStatus)
	if err != nil {
		return
	}
	result.EM1 = append(result.EM1, *em1)
	result.Power += em1.ActPower
	if result.EM1Powers != nil {
		result.EM1Powers[em1.ID] = em1.ActPower
	}
	if result.Voltage == 0 && em1.Voltage > 0 {
		result.Voltage = em1.Voltage
	}
	if result.Current == 0 && em1.Current > 0 {
		result.Current = em1.Current
	}
}

// temperatureData holds parsed temperature component fields.
type temperatureData struct {
	TC float64 `json:"tC"`
}

// parseTemperatureComponent parses a temperature component status.
func parseTemperatureComponent(result *ParsedStatus, rawStatus json.RawMessage) {
	temp, err := unmarshalJSON[temperatureData](rawStatus)
	if err == nil && result.Temperature == 0 {
		result.Temperature = temp.TC
	}
}

// sysDeviceTemp holds the device_temp field from sys status.
type sysDeviceTemp struct {
	DeviceTemp *temperatureData `json:"device_temp,omitempty"`
}

// parseSysComponent parses system status (uptime, RAM, FS, etc.).
func parseSysComponent(result *ParsedStatus, rawStatus json.RawMessage) {
	sys, err := unmarshalJSON[SysInfo](rawStatus)
	if err != nil {
		return
	}
	result.Sys = sys

	// Extract device temperature if present (some devices report it in sys)
	if sysTemp, err := unmarshalJSON[sysDeviceTemp](rawStatus); err == nil {
		if sysTemp.DeviceTemp != nil && result.Temperature == 0 {
			result.Temperature = sysTemp.DeviceTemp.TC
		}
	}

	if sys.MAC != "" {
		result.MAC = sys.MAC
	}
}

// parseWiFiComponent parses WiFi status (SSID, RSSI, IP).
func parseWiFiComponent(result *ParsedStatus, rawStatus json.RawMessage) {
	wifi, err := unmarshalJSON[WiFiInfo](rawStatus)
	if err != nil {
		return
	}
	result.WiFi = wifi
}

// ParseGen1Status parses a Gen1 /status response into ParsedStatus.
// Gen1 devices have a different JSON structure than Gen2+.
func ParseGen1Status(deviceID string, statusMap map[string]json.RawMessage) (*ParsedStatus, error) {
	result := &ParsedStatus{
		DeviceID:   deviceID,
		Generation: 1,
		Components: statusMap,
	}

	// Parse each Gen1-specific field
	if raw, ok := statusMap[Gen1Relays]; ok {
		parseGen1Relays(result, raw)
	}
	if raw, ok := statusMap[Gen1Meters]; ok {
		parseGen1Meters(result, raw)
	}
	if raw, ok := statusMap[Gen1EMeters]; ok {
		parseGen1EMeters(result, raw)
	}
	if raw, ok := statusMap[Gen1Lights]; ok {
		parseGen1Lights(result, raw)
	}
	if raw, ok := statusMap[Gen1Rollers]; ok {
		parseGen1Rollers(result, raw)
	}
	if raw, ok := statusMap[Gen1Inputs]; ok {
		parseGen1Inputs(result, raw)
	}
	if raw, ok := statusMap[Gen1Temperature]; ok {
		parseGen1Temperature(result, raw)
	}
	if raw, ok := statusMap[Gen1ExtTemperature]; ok {
		parseGen1ExtTemperature(result, raw)
	}
	if raw, ok := statusMap[Gen1WiFiSta]; ok {
		parseGen1WiFi(result, raw)
	}

	parseGen1SysInfo(result, statusMap)
	return result, nil
}

// gen1RelayStatus holds Gen1 relay fields.
type gen1RelayStatus struct {
	IsOn   bool   `json:"ison"`
	Source string `json:"source,omitempty"`
}

// parseGen1Relays parses Gen1 relays array.
func parseGen1Relays(result *ParsedStatus, raw json.RawMessage) {
	relays, err := unmarshalJSON[[]gen1RelayStatus](raw)
	if err != nil {
		return
	}
	for i, r := range *relays {
		result.Switches = append(result.Switches, SwitchState{ID: i, On: r.IsOn, Source: r.Source})
	}
}

// gen1MeterStatus holds Gen1 meter fields.
type gen1MeterStatus struct {
	Power float64 `json:"power"`
	Total float64 `json:"total"`
}

// parseGen1Meters parses Gen1 meters array.
func parseGen1Meters(result *ParsedStatus, raw json.RawMessage) {
	meters, err := unmarshalJSON[[]gen1MeterStatus](raw)
	if err != nil {
		return
	}
	for _, m := range *meters {
		result.Power += m.Power
		result.TotalEnergy += m.Total
	}
}

// gen1EMeterStatus holds Gen1 energy meter fields.
type gen1EMeterStatus struct {
	Power       float64 `json:"power"`
	Voltage     float64 `json:"voltage"`
	Current     float64 `json:"current"`
	TotalEnergy float64 `json:"total"`
}

// parseGen1EMeters parses Gen1 emeters array.
func parseGen1EMeters(result *ParsedStatus, raw json.RawMessage) {
	emeters, err := unmarshalJSON[[]gen1EMeterStatus](raw)
	if err != nil {
		return
	}
	for _, em := range *emeters {
		result.Power += em.Power
		result.Current += em.Current
		if result.Voltage == 0 && em.Voltage > 0 {
			result.Voltage = em.Voltage
		}
		result.TotalEnergy += em.TotalEnergy
	}
}

// gen1LightStatus holds Gen1 light fields.
type gen1LightStatus struct {
	IsOn bool `json:"ison"`
}

// parseGen1Lights parses Gen1 lights array.
func parseGen1Lights(result *ParsedStatus, raw json.RawMessage) {
	lights, err := unmarshalJSON[[]gen1LightStatus](raw)
	if err != nil {
		return
	}
	for i, lt := range *lights {
		result.Lights = append(result.Lights, LightState{ID: i, On: lt.IsOn})
	}
}

// gen1RollerStatus holds Gen1 roller (cover) fields.
type gen1RollerStatus struct {
	State string  `json:"state"`
	Power float64 `json:"power"`
}

// parseGen1Rollers parses Gen1 rollers array.
func parseGen1Rollers(result *ParsedStatus, raw json.RawMessage) {
	rollers, err := unmarshalJSON[[]gen1RollerStatus](raw)
	if err != nil {
		return
	}
	for i, r := range *rollers {
		// Normalize state: Gen1 uses "open"/"close"/"stop", Gen2 uses "open"/"closed"/"stopped"
		state := normalizeGen1CoverState(r.State)
		result.Covers = append(result.Covers, CoverState{ID: i, State: state})
		result.Power += r.Power
	}
}

// normalizeGen1CoverState converts Gen1 cover states to Gen2 format.
func normalizeGen1CoverState(state string) string {
	switch state {
	case "close":
		return CoverStateClosed
	case "stop":
		return CoverStateStopped
	default:
		return state
	}
}

// gen1InputStatus holds Gen1 input fields.
type gen1InputStatus struct {
	Input    int    `json:"input"`
	Event    string `json:"event,omitempty"`
	EventCnt int    `json:"event_cnt,omitempty"`
}

// parseGen1Inputs parses Gen1 inputs array.
func parseGen1Inputs(result *ParsedStatus, raw json.RawMessage) {
	inputs, err := unmarshalJSON[[]gen1InputStatus](raw)
	if err != nil {
		return
	}
	for i, inp := range *inputs {
		// Gen1 input field is 0/1 representing LOW/HIGH state
		result.Inputs = append(result.Inputs, InputState{
			ID:    i,
			State: inp.Input != 0,
		})
	}
}

// parseGen1Temperature parses Gen1 temperature field (single value).
func parseGen1Temperature(result *ParsedStatus, raw json.RawMessage) {
	var temp float64
	if err := json.Unmarshal(raw, &temp); err == nil && result.Temperature == 0 {
		result.Temperature = temp
	}
}

// gen1ExtTempSensor holds Gen1 external temperature sensor fields.
type gen1ExtTempSensor struct {
	TC      float64 `json:"tC"`
	IsValid bool    `json:"is_valid"`
}

// parseGen1ExtTemperature parses Gen1 ext_temperature (add-on sensors).
func parseGen1ExtTemperature(result *ParsedStatus, raw json.RawMessage) {
	sensors, err := unmarshalJSON[map[string]gen1ExtTempSensor](raw)
	if err != nil {
		return
	}
	for _, s := range *sensors {
		if s.IsValid && result.Temperature == 0 {
			result.Temperature = s.TC
			break
		}
	}
}

// gen1WiFiStatus holds Gen1 wifi_sta fields.
type gen1WiFiStatus struct {
	Connected bool   `json:"connected"`
	SSID      string `json:"ssid"`
	IP        string `json:"ip"`
	RSSI      int    `json:"rssi"`
}

// parseGen1WiFi parses Gen1 wifi_sta field.
func parseGen1WiFi(result *ParsedStatus, raw json.RawMessage) {
	wifi, err := unmarshalJSON[gen1WiFiStatus](raw)
	if err != nil {
		return
	}
	if wifi.Connected {
		result.WiFi = &WiFiInfo{
			StaIP:  wifi.IP,
			Status: "got ip",
			SSID:   wifi.SSID,
			RSSI:   wifi.RSSI,
		}
	}
}

// gen1TopLevel holds Gen1 top-level status fields for sys info.
type gen1TopLevel struct {
	Uptime   int    `json:"uptime"`
	RAMTotal int    `json:"ram_total"`
	RAMFree  int    `json:"ram_free"`
	FSSize   int    `json:"fs_size"`
	FSFree   int    `json:"fs_free"`
	MAC      string `json:"mac"`
	Time     string `json:"time,omitempty"`
}

// parseGen1SysInfo extracts system info from Gen1 status.
func parseGen1SysInfo(result *ParsedStatus, statusMap map[string]json.RawMessage) {
	// Reconstruct full JSON to parse top-level fields
	fullJSON, err := json.Marshal(statusMap)
	if err != nil {
		return
	}
	topLevel, err := unmarshalJSON[gen1TopLevel](fullJSON)
	if err != nil {
		return
	}

	result.Sys = &SysInfo{
		Uptime:  topLevel.Uptime,
		RAMFree: topLevel.RAMFree,
		RAMSize: topLevel.RAMTotal,
		FSFree:  topLevel.FSFree,
		FSSize:  topLevel.FSSize,
		MAC:     topLevel.MAC,
		Time:    topLevel.Time,
	}

	if topLevel.MAC != "" {
		result.MAC = topLevel.MAC
	}
}

// ApplyParsedStatus applies parsed status to DeviceData, preserving existing data when new is nil.
func ApplyParsedStatus(data *DeviceData, parsed *ParsedStatus) {
	if parsed == nil {
		return
	}

	applyParsedComponentStates(data, parsed)
	applyParsedPowerMaps(data, parsed)
	applyParsedMetrics(data, parsed)
	applyParsedSnapshot(data, parsed)
	applyParsedWiFi(data, parsed.WiFi)
	applyParsedSys(data, parsed.Sys)
	applyParsedDeviceID(data, parsed)
}

// applyParsedComponentStates updates component states (replace if present, keep existing if empty).
func applyParsedComponentStates(data *DeviceData, parsed *ParsedStatus) {
	if len(parsed.Switches) > 0 {
		data.Switches = parsed.Switches
	}
	if len(parsed.Lights) > 0 {
		data.Lights = parsed.Lights
	}
	if len(parsed.Covers) > 0 {
		data.Covers = parsed.Covers
	}
	if len(parsed.Inputs) > 0 {
		data.Inputs = parsed.Inputs
	}
}

// applyParsedPowerMaps updates per-component power maps for accurate WebSocket aggregation.
func applyParsedPowerMaps(data *DeviceData, parsed *ParsedStatus) {
	if len(parsed.SwitchPowers) > 0 {
		data.SwitchPowers = parsed.SwitchPowers
	}
	if len(parsed.LightPowers) > 0 {
		data.LightPowers = parsed.LightPowers
	}
	if len(parsed.CoverPowers) > 0 {
		data.CoverPowers = parsed.CoverPowers
	}
	if len(parsed.PMPowers) > 0 {
		data.PMPowers = parsed.PMPowers
	}
	if len(parsed.EMPowers) > 0 {
		data.EMPowers = parsed.EMPowers
	}
	if len(parsed.EM1Powers) > 0 {
		data.EM1Powers = parsed.EM1Powers
	}
}

// applyParsedMetrics updates power/energy metrics.
func applyParsedMetrics(data *DeviceData, parsed *ParsedStatus) {
	data.Power = parsed.Power
	data.Voltage = parsed.Voltage
	data.Current = parsed.Current
	data.TotalEnergy = parsed.TotalEnergy
	if parsed.Temperature > 0 {
		data.Temperature = parsed.Temperature
	}
}

// applyParsedSnapshot updates monitoring snapshot if we have PM/EM data.
func applyParsedSnapshot(data *DeviceData, parsed *ParsedStatus) {
	if len(parsed.PM) == 0 && len(parsed.EM) == 0 && len(parsed.EM1) == 0 {
		return
	}
	if data.Snapshot == nil {
		data.Snapshot = &model.MonitoringSnapshot{}
	}
	data.Snapshot.PM = parsed.PM
	data.Snapshot.EM = parsed.EM
	data.Snapshot.EM1 = parsed.EM1
	data.Snapshot.Online = true
}

// applyParsedWiFi applies WiFi info if present.
func applyParsedWiFi(data *DeviceData, wifi *WiFiInfo) {
	if wifi == nil {
		return
	}
	data.WiFi = &shelly.WiFiStatus{
		StaIP:  wifi.StaIP,
		Status: wifi.Status,
		SSID:   wifi.SSID,
		RSSI:   wifi.RSSI,
	}
}

// applyParsedSys applies system info if present.
func applyParsedSys(data *DeviceData, sys *SysInfo) {
	if sys == nil {
		return
	}
	data.Sys = &shelly.SysStatus{
		MAC:             sys.MAC,
		Uptime:          sys.Uptime,
		Time:            sys.Time,
		RAMFree:         sys.RAMFree,
		RAMSize:         sys.RAMSize,
		FSFree:          sys.FSFree,
		FSSize:          sys.FSSize,
		RestartRequired: sys.RestartRequired,
	}
}

// applyParsedDeviceID applies device identification if present.
func applyParsedDeviceID(data *DeviceData, parsed *ParsedStatus) {
	if parsed.Model != "" && data.Device.Model == "" {
		data.Device.Model = parsed.Model
	}
	if parsed.Generation > 0 && data.Device.Generation == 0 {
		data.Device.Generation = parsed.Generation
	}
	if parsed.MAC != "" && data.Device.MAC == "" {
		data.Device.MAC = parsed.MAC
	}
}

// ApplyIncrementalUpdate applies a single component status change to DeviceData.
// Used for WebSocket real-time updates that only contain one component's state.
func ApplyIncrementalUpdate(data *DeviceData, componentType string, componentID int, status json.RawMessage) {
	switch componentType {
	case ComponentSwitch:
		applyIncrementalSwitch(data, componentID, status)
	case ComponentLight:
		applyIncrementalLight(data, componentID, status)
	case ComponentCover:
		applyIncrementalCover(data, componentID, status)
	case ComponentInput:
		applyIncrementalInput(data, componentID, status)
	case ComponentPM, ComponentPM1:
		applyIncrementalPM(data, componentID, status)
	case ComponentEM:
		applyIncrementalEM(data, componentID, status)
	case ComponentEM1:
		applyIncrementalEM1(data, componentID, status)
	}
}

// incrementalSwitchStatus holds incremental switch update fields.
type incrementalSwitchStatus struct {
	Output  *bool    `json:"output,omitempty"`
	APower  *float64 `json:"apower,omitempty"`
	Voltage *float64 `json:"voltage,omitempty"`
	Current *float64 `json:"current,omitempty"`
	AEnergy *struct {
		Total float64 `json:"total"`
	} `json:"aenergy,omitempty"`
}

// applyIncrementalSwitch applies a switch status change.
func applyIncrementalSwitch(data *DeviceData, id int, raw json.RawMessage) {
	status, err := unmarshalJSON[incrementalSwitchStatus](raw)
	if err != nil {
		return
	}

	stateChanged := false
	if status.Output != nil {
		stateChanged = updateOrAppendSwitch(&data.Switches, id, *status.Output)
	}

	if status.APower != nil {
		// Update per-component power and recalculate total
		data.EnsurePowerMaps()
		data.SwitchPowers[id] = *status.APower
		data.Power = data.SumPowers()
	} else if stateChanged {
		// State changed but no power data - mark for HTTP refresh
		// This handles the case where WebSocket only sends output state without apower
		data.NeedsRefresh = true
	}

	if status.Voltage != nil {
		data.Voltage = *status.Voltage
	}
	if status.Current != nil {
		data.Current = *status.Current
	}
	if status.AEnergy != nil {
		data.TotalEnergy = status.AEnergy.Total
	}
}

// updateOrAppendSwitch updates an existing switch or appends a new one.
// Returns true if the state actually changed (or new switch was added).
func updateOrAppendSwitch(switches *[]SwitchState, id int, on bool) bool {
	for i := range *switches {
		if (*switches)[i].ID == id {
			if (*switches)[i].On == on {
				return false // No change
			}
			(*switches)[i].On = on
			return true // State changed
		}
	}
	*switches = append(*switches, SwitchState{ID: id, On: on})
	return true // New switch added, treat as change
}

// incrementalLightStatus holds incremental light update fields.
type incrementalLightStatus struct {
	Output  *bool    `json:"output,omitempty"`
	APower  *float64 `json:"apower,omitempty"`
	Voltage *float64 `json:"voltage,omitempty"`
	Current *float64 `json:"current,omitempty"`
	AEnergy *struct {
		Total float64 `json:"total"`
	} `json:"aenergy,omitempty"`
}

// applyIncrementalLight applies a light status change.
func applyIncrementalLight(data *DeviceData, id int, raw json.RawMessage) {
	status, err := unmarshalJSON[incrementalLightStatus](raw)
	if err != nil {
		return
	}

	stateChanged := false
	if status.Output != nil {
		stateChanged = updateOrAppendLight(&data.Lights, id, *status.Output)
	}

	if status.APower != nil {
		// Update per-component power and recalculate total
		data.EnsurePowerMaps()
		data.LightPowers[id] = *status.APower
		data.Power = data.SumPowers()
	} else if stateChanged {
		// State changed but no power data - mark for HTTP refresh
		// This handles the case where WebSocket only sends output state without apower
		data.NeedsRefresh = true
	}

	if status.Voltage != nil {
		data.Voltage = *status.Voltage
	}
	if status.Current != nil {
		data.Current = *status.Current
	}
	if status.AEnergy != nil {
		data.TotalEnergy = status.AEnergy.Total
	}
}

// updateOrAppendLight updates an existing light or appends a new one.
// Returns true if the state actually changed (or new light was added).
func updateOrAppendLight(lights *[]LightState, id int, on bool) bool {
	for i := range *lights {
		if (*lights)[i].ID == id {
			if (*lights)[i].On == on {
				return false // No change
			}
			(*lights)[i].On = on
			return true // State changed
		}
	}
	*lights = append(*lights, LightState{ID: id, On: on})
	return true // New light added, treat as change
}

// incrementalCoverStatus holds incremental cover update fields.
type incrementalCoverStatus struct {
	State  *string  `json:"state,omitempty"`
	APower *float64 `json:"apower,omitempty"`
}

// applyIncrementalCover applies a cover status change.
func applyIncrementalCover(data *DeviceData, id int, raw json.RawMessage) {
	status, err := unmarshalJSON[incrementalCoverStatus](raw)
	if err != nil {
		return
	}

	stateChanged := false
	if status.State != nil {
		stateChanged = updateOrAppendCover(&data.Covers, id, *status.State)
	}

	if status.APower != nil {
		// Update per-component power and recalculate total
		data.EnsurePowerMaps()
		data.CoverPowers[id] = *status.APower
		data.Power = data.SumPowers()
	} else if stateChanged {
		// State changed but no power data - mark for HTTP refresh
		data.NeedsRefresh = true
	}
}

// updateOrAppendCover updates an existing cover or appends a new one.
// Returns true if the state actually changed (or new cover was added).
func updateOrAppendCover(covers *[]CoverState, id int, state string) bool {
	for i := range *covers {
		if (*covers)[i].ID == id {
			if (*covers)[i].State == state {
				return false // No change
			}
			(*covers)[i].State = state
			return true // State changed
		}
	}
	*covers = append(*covers, CoverState{ID: id, State: state})
	return true // New cover added, treat as change
}

// incrementalInputStatus holds incremental input update fields.
type incrementalInputStatus struct {
	State *bool `json:"state,omitempty"`
}

// applyIncrementalInput applies an input status change.
func applyIncrementalInput(data *DeviceData, id int, raw json.RawMessage) {
	status, err := unmarshalJSON[incrementalInputStatus](raw)
	if err != nil || status.State == nil {
		return
	}
	updateOrAppendInput(&data.Inputs, id, *status.State)
}

// updateOrAppendInput updates an existing input or appends a new one.
func updateOrAppendInput(inputs *[]InputState, id int, state bool) {
	for i := range *inputs {
		if (*inputs)[i].ID == id {
			(*inputs)[i].State = state
			return
		}
	}
	*inputs = append(*inputs, InputState{ID: id, State: state})
}

// applyIncrementalPM applies a PM/PM1 status change.
func applyIncrementalPM(data *DeviceData, id int, raw json.RawMessage) {
	pm, err := unmarshalJSON[model.PMStatus](raw)
	if err != nil {
		return
	}

	// Update per-component power and recalculate total
	data.EnsurePowerMaps()
	data.PMPowers[id] = pm.APower
	data.Power = data.SumPowers()

	if pm.Voltage > 0 {
		data.Voltage = pm.Voltage
	}
	if pm.Current > 0 {
		data.Current = pm.Current
	}
	if pm.AEnergy != nil {
		data.TotalEnergy = pm.AEnergy.Total
	}
}

// applyIncrementalEM applies an EM status change.
func applyIncrementalEM(data *DeviceData, id int, raw json.RawMessage) {
	em, err := unmarshalJSON[model.EMStatus](raw)
	if err != nil {
		return
	}

	// Update per-component power and recalculate total
	data.EnsurePowerMaps()
	data.EMPowers[id] = em.TotalActivePower
	data.Power = data.SumPowers()

	data.Current = em.TotalCurrent
	if em.AVoltage > 0 {
		data.Voltage = em.AVoltage
	}
}

// applyIncrementalEM1 applies an EM1 status change.
func applyIncrementalEM1(data *DeviceData, id int, raw json.RawMessage) {
	em1, err := unmarshalJSON[model.EM1Status](raw)
	if err != nil {
		return
	}

	// Update per-component power and recalculate total
	data.EnsurePowerMaps()
	data.EM1Powers[id] = em1.ActPower
	data.Power = data.SumPowers()

	if em1.Voltage > 0 {
		data.Voltage = em1.Voltage
	}
	if em1.Current > 0 {
		data.Current = em1.Current
	}
}

// ParseComponentName extracts the component type and ID from a component name.
// For example, "switch:0" returns ("switch", 0) and "pm:1" returns ("pm", 1).
func ParseComponentName(name string) (componentType string, componentID int) {
	parts := strings.SplitN(name, ":", 2)
	if len(parts) != 2 {
		return name, 0
	}
	componentType = parts[0]
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return componentType, 0
	}
	return componentType, id
}

// componentRule defines how to detect and configure component capabilities.
type componentRule struct {
	patterns []string
	caps     ComponentCapabilities
	checkPM  bool // If true, also check for "PM" substring to set HasPM
}

// componentRules defines detection rules in priority order.
// First matching rule wins.
var componentRules = []componentRule{
	// Energy meters (check first to avoid false EM match from "1PM")
	{patterns: []string{"3EM", "SPEM"}, caps: ComponentCapabilities{HasEM: true}},
	// Plugs (SHPLG for Gen1, SNPL for Plus Plug)
	{patterns: []string{"SHPLG", "SNPL"}, caps: ComponentCapabilities{HasSwitches: true, HasPM: true, NumSwitches: 1}},
	// Bulbs (Gen1 Bulb Duo, Vintage, etc.)
	{patterns: []string{"SHBDUO", "SHBLB", "SHVIN"}, caps: ComponentCapabilities{HasLights: true, NumLights: 1}},
	// Dimmers and RGBW (lights) - includes Plus Wall Dimmer (SNDM)
	{patterns: []string{"RGBW", "SHDM", "SNDM"}, caps: ComponentCapabilities{HasLights: true, NumLights: 1}},
	// Shelly 2.5 can be cover or switch mode
	{patterns: []string{"SHSW-25", "2.5"}, caps: ComponentCapabilities{HasCovers: true, HasSwitches: true, NumCovers: 1, NumSwitches: 2}},
	// Cover-specific devices
	{patterns: []string{"COVER", "SHSPM"}, caps: ComponentCapabilities{HasCovers: true, NumCovers: 1}},
	// Pro 4PM
	{patterns: []string{"PRO4"}, caps: ComponentCapabilities{HasSwitches: true, NumSwitches: 4}, checkPM: true},
	// Plus/Pro 2PM or Shelly 2
	{patterns: []string{"PLUS2", "PRO2", "102"}, caps: ComponentCapabilities{HasSwitches: true, NumSwitches: 2}, checkPM: true},
	// Plus/Pro 1 or other switches (SHSW, SNSW)
	{patterns: []string{"PLUS1", "PRO1", "SHSW", "SNSW"}, caps: ComponentCapabilities{HasSwitches: true, NumSwitches: 1}, checkPM: true},
}

// DetectComponents returns component capabilities based on device model string.
// This is used for optimization - we only try to parse components the device has.
func DetectComponents(modelStr string) ComponentCapabilities {
	upper := strings.ToUpper(modelStr)

	for _, rule := range componentRules {
		if matchesAny(upper, rule.patterns) {
			caps := rule.caps
			if rule.checkPM && strings.Contains(upper, "PM") {
				caps.HasPM = true
			}
			return caps
		}
	}

	return ComponentCapabilities{}
}

// matchesAny returns true if s contains any of the patterns.
func matchesAny(s string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

// ComponentNames maps component IDs to their user-configured names.
type ComponentNames struct {
	Switches map[int]string
	Lights   map[int]string
	Covers   map[int]string
	Inputs   map[int]string
}

// componentConfigData holds common config fields we care about.
type componentConfigData struct {
	ID   int     `json:"id"`
	Name *string `json:"name,omitempty"`
	Type string  `json:"type,omitempty"` // For inputs
}

// ParseFullConfig parses a Gen2+ Shelly.GetConfig response to extract component names.
func ParseFullConfig(configMap map[string]json.RawMessage) *ComponentNames {
	names := &ComponentNames{
		Switches: make(map[int]string),
		Lights:   make(map[int]string),
		Covers:   make(map[int]string),
		Inputs:   make(map[int]string),
	}

	for key, raw := range configMap {
		prefix, _ := ParseComponentName(key)
		cfg, err := unmarshalJSON[componentConfigData](raw)
		if err != nil || cfg.Name == nil || *cfg.Name == "" {
			continue
		}

		switch prefix {
		case ComponentSwitch:
			names.Switches[cfg.ID] = *cfg.Name
		case ComponentLight:
			names.Lights[cfg.ID] = *cfg.Name
		case ComponentCover:
			names.Covers[cfg.ID] = *cfg.Name
		case ComponentInput:
			names.Inputs[cfg.ID] = *cfg.Name
		}
	}

	return names
}

// ParseGen1Config parses a Gen1 /settings response to extract component names.
func ParseGen1Config(configMap map[string]json.RawMessage) *ComponentNames {
	names := &ComponentNames{
		Switches: make(map[int]string),
		Lights:   make(map[int]string),
		Covers:   make(map[int]string),
		Inputs:   make(map[int]string),
	}

	// Gen1 relays have name field
	if raw, ok := configMap[Gen1Relays]; ok {
		var relays []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(raw, &relays); err == nil {
			for i, r := range relays {
				if r.Name != "" {
					names.Switches[i] = r.Name
				}
			}
		}
	}

	// Gen1 lights have name field
	if raw, ok := configMap[Gen1Lights]; ok {
		var lights []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(raw, &lights); err == nil {
			for i, l := range lights {
				if l.Name != "" {
					names.Lights[i] = l.Name
				}
			}
		}
	}

	// Gen1 rollers have name field
	if raw, ok := configMap[Gen1Rollers]; ok {
		var rollers []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(raw, &rollers); err == nil {
			for i, r := range rollers {
				if r.Name != "" {
					names.Covers[i] = r.Name
				}
			}
		}
	}

	return names
}

// ApplyConfigNames applies component names from config to DeviceData.
func ApplyConfigNames(data *DeviceData, names *ComponentNames) {
	if names == nil {
		return
	}

	for i := range data.Switches {
		if name, ok := names.Switches[data.Switches[i].ID]; ok {
			data.Switches[i].Name = name
		}
	}
	for i := range data.Lights {
		if name, ok := names.Lights[data.Lights[i].ID]; ok {
			data.Lights[i].Name = name
		}
	}
	for i := range data.Covers {
		if name, ok := names.Covers[data.Covers[i].ID]; ok {
			data.Covers[i].Name = name
		}
	}
	for i := range data.Inputs {
		if name, ok := names.Inputs[data.Inputs[i].ID]; ok {
			data.Inputs[i].Name = name
		}
	}
}
