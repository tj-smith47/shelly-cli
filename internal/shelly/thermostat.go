// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-go/gen2/components"
)

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

	if config, ok := params["config"].(map[string]any); ok {
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
	if enable, ok := config["enable"].(bool); ok {
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
		"method": "Thermostat.SetConfig",
		"params": map[string]any{
			"id":     params.ThermostatID,
			"config": config,
		},
	}

	return map[string]any{
		"enable":   params.Enabled,
		"timespec": params.Timespec,
		"calls":    []any{call},
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
		config["enable"] = *params.EnableState
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
