// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// Thermostat operating modes.
const (
	thermostatModeHeat = "heat"
	thermostatModeCool = "cool"
)

// Thermostat schedule RPC literals.
const (
	rpcThermostatSetConfig = "Thermostat.SetConfig"
)

// Shared RPC config field names.
const (
	fieldConfig = "config"
	fieldEnable = "enable"
	fieldServer = "server"
	fieldSTA    = "sta"
)

// ValidThermostatModes contains the valid thermostat operating modes.
var ValidThermostatModes = map[string]bool{
	thermostatModeHeat: true,
	thermostatModeCool: true,
	ComponentTypeAuto:  true,
}

// ValidateThermostatMode validates that a thermostat mode is one of: heat, cool, auto.
// If allowEmpty is true, an empty string is also valid (used when mode is optional).
func ValidateThermostatMode(mode string, allowEmpty bool) error {
	if mode == "" {
		if allowEmpty {
			return nil
		}
		return fmt.Errorf("mode is required, must be one of: heat, cool, auto")
	}

	if !ValidThermostatModes[mode] {
		return fmt.Errorf("invalid mode %q, must be one of: heat, cool, auto", mode)
	}

	return nil
}

// ThermostatSchedule represents a schedule targeting a thermostat.
type ThermostatSchedule struct {
	ID           int      `json:"id"`
	Enabled      bool     `json:"enabled"`
	Timespec     string   `json:"timespec"`
	ThermostatID int      `json:"thermostat_id,omitempty"`
	TargetC      *float64 `json:"target_c,omitempty"`
	Mode         string   `json:"mode,omitempty"`
	Enable       *bool    `json:"enable,omitempty"`
}

// FilterThermostatSchedules filters schedule jobs to only those targeting thermostats.
// If showAll is true, returns all schedules without filtering.
// If thermostatID > 0, only returns schedules targeting that specific thermostat.
func FilterThermostatSchedules(jobs []components.ScheduleJob, thermostatID int, showAll bool) []ThermostatSchedule {
	var schedules []ThermostatSchedule

	for _, job := range jobs {
		if showAll {
			schedules = append(schedules, ThermostatSchedule{
				ID:       job.ID,
				Enabled:  job.Enable,
				Timespec: job.Timespec,
			})
			continue
		}

		if sched, ok := extractThermostatSchedule(job, thermostatID); ok {
			schedules = append(schedules, sched)
		}
	}

	return schedules
}

// extractThermostatSchedule checks if a schedule job targets a thermostat.
// Returns the extracted schedule and true if the job contains a Thermostat.* call.
// If filterID > 0, only matches schedules targeting that thermostat ID.
func extractThermostatSchedule(job components.ScheduleJob, filterID int) (ThermostatSchedule, bool) {
	for _, call := range job.Calls {
		if !strings.HasPrefix(call.Method, "Thermostat.") {
			continue
		}

		sched := ThermostatSchedule{
			ID:       job.ID,
			Enabled:  job.Enable,
			Timespec: job.Timespec,
		}

		params, ok := call.Params.(map[string]any)
		if !ok {
			return sched, true
		}

		extractThermostatParams(&sched, params)

		if filterID > 0 && sched.ThermostatID != filterID {
			continue
		}

		return sched, true
	}
	return ThermostatSchedule{}, false
}

// extractThermostatParams extracts thermostat-related parameters from a schedule call.
func extractThermostatParams(sched *ThermostatSchedule, params map[string]any) {
	if id, ok := params["id"].(float64); ok {
		sched.ThermostatID = int(id)
	}

	if config, ok := params[fieldConfig].(map[string]any); ok {
		extractConfigParams(sched, config)
	}

	if targetC, ok := params["target_C"].(float64); ok {
		sched.TargetC = &targetC
	}
}

// extractConfigParams extracts target temperature, mode, and enable state from config object.
func extractConfigParams(sched *ThermostatSchedule, config map[string]any) {
	if targetC, ok := config["target_C"].(float64); ok {
		sched.TargetC = &targetC
	}
	if mode, ok := config["thermostat_mode"].(string); ok {
		sched.Mode = mode
	}
	if enable, ok := config[fieldEnable].(bool); ok {
		sched.Enable = &enable
	}
}

// ThermostatScheduleParams contains parameters for creating a thermostat schedule.
type ThermostatScheduleParams struct {
	ThermostatID int
	Timespec     string
	Enabled      bool
	TargetC      *float64
	Mode         string
	EnableState  *bool // nil=no change, true=enable, false=disable
}

// BuildThermostatScheduleCall builds a Schedule.Create parameter map for a thermostat schedule.
func BuildThermostatScheduleCall(params ThermostatScheduleParams) map[string]any {
	config := buildThermostatConfig(params)

	call := map[string]any{
		"method": rpcThermostatSetConfig,
		"params": map[string]any{
			"id":        params.ThermostatID,
			fieldConfig: config,
		},
	}

	return map[string]any{
		fieldEnable: params.Enabled,
		"timespec":  params.Timespec,
		"calls":     []any{call},
	}
}

// buildThermostatConfig builds the thermostat config portion of a schedule call.
func buildThermostatConfig(params ThermostatScheduleParams) map[string]any {
	config := make(map[string]any)
	if params.TargetC != nil {
		config["target_C"] = *params.TargetC
	}
	if params.Mode != "" {
		config["thermostat_mode"] = params.Mode
	}
	if params.EnableState != nil {
		config[fieldEnable] = *params.EnableState
	}
	return config
}

// ScheduleCreateResult contains the result of a Schedule.Create call.
type ScheduleCreateResult struct {
	ID  int `json:"id"`
	Rev int `json:"rev"`
}

// ParseScheduleCreateResponse parses a Schedule.Create response.
func ParseScheduleCreateResponse(result any) (ScheduleCreateResult, error) {
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return ScheduleCreateResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var resp ScheduleCreateResult
	if err := json.Unmarshal(jsonBytes, &resp); err != nil {
		return ScheduleCreateResult{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return resp, nil
}

// CollectThermostats extracts thermostat information from a device's status and
// config maps (the results of Shelly.GetStatus and Shelly.GetConfig). The
// enable flag is a configuration property; the status "output" field reports
// only whether the valve is currently calling for heat, so the configured
// Enabled flag and the live Heating state are surfaced as distinct fields.
func CollectThermostats(status, config map[string]json.RawMessage) []model.ThermostatInfo {
	var thermostats []model.ThermostatInfo

	for key, raw := range status {
		if !strings.HasPrefix(key, "thermostat:") {
			continue
		}

		var st struct {
			ID       int      `json:"id"`
			Output   *bool    `json:"output"`
			TargetC  *float64 `json:"target_C"`
			CurrentC *float64 `json:"current_C"`
		}
		if err := json.Unmarshal(raw, &st); err != nil {
			continue
		}

		info := model.ThermostatInfo{
			ID:      st.ID,
			Heating: st.Output != nil && *st.Output,
		}
		if st.TargetC != nil {
			info.TargetC = *st.TargetC
		}

		// Enabled and mode live in config, not status.
		applyThermostatConfig(&info, config[key])

		thermostats = append(thermostats, info)
	}

	return thermostats
}

// applyThermostatConfig overlays the configuration-derived fields (enable flag,
// mode, and a target_C fallback) onto a ThermostatInfo built from status. A nil
// or unparsable config leaves the info unchanged.
func applyThermostatConfig(info *model.ThermostatInfo, cfgRaw json.RawMessage) {
	if len(cfgRaw) == 0 {
		return
	}
	var cfg struct {
		Enable  *bool    `json:"enable"`
		Mode    *string  `json:"thermostat_mode"`
		TargetC *float64 `json:"target_C"`
	}
	if err := json.Unmarshal(cfgRaw, &cfg); err != nil {
		return
	}
	info.Enabled = cfg.Enable != nil && *cfg.Enable
	if cfg.Mode != nil {
		info.Mode = *cfg.Mode
	}
	if info.TargetC == 0 && cfg.TargetC != nil {
		info.TargetC = *cfg.TargetC
	}
}
