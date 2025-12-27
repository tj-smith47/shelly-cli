// Package backup provides backup and restore operations for Shelly devices.
package backup

import (
	"encoding/json"
	"fmt"

	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// ConvertBackupScripts converts shelly-go backup scripts to internal Script type.
func ConvertBackupScripts(scripts []*shellybackup.Script) []Script {
	result := make([]Script, len(scripts))
	for i, s := range scripts {
		result[i] = Script{
			ID:     s.ID,
			Name:   s.Name,
			Enable: s.Enable,
			Code:   s.Code,
		}
	}
	return result
}

// ConvertBackupSchedules converts raw schedule data to internal Schedule type.
func ConvertBackupSchedules(data json.RawMessage) []Schedule {
	var schedData struct {
		Jobs []struct {
			Enable   bool           `json:"enable"`
			Timespec string         `json:"timespec"`
			Calls    []ScheduleCall `json:"calls"`
		} `json:"jobs"`
	}
	if err := json.Unmarshal(data, &schedData); err != nil {
		return nil
	}

	result := make([]Schedule, len(schedData.Jobs))
	for i, j := range schedData.Jobs {
		result[i] = Schedule{
			Enable:   j.Enable,
			Timespec: j.Timespec,
			Calls:    j.Calls,
		}
	}
	return result
}

// WebhookInfoBackup represents webhook info from a backup file.
type WebhookInfoBackup struct {
	ID     int      `json:"id"`
	Cid    int      `json:"cid"`
	Enable bool     `json:"enable"`
	Event  string   `json:"event"`
	Name   string   `json:"name"`
	URLs   []string `json:"urls"`
}

// ConvertBackupWebhooks converts raw webhook data to internal WebhookInfoBackup type.
func ConvertBackupWebhooks(data json.RawMessage) []WebhookInfoBackup {
	var whData struct {
		Hooks []WebhookInfoBackup `json:"hooks"`
	}
	if err := json.Unmarshal(data, &whData); err != nil {
		return nil
	}
	return whData.Hooks
}

// CompareConfigs compares current and backup configurations.
func CompareConfigs(current, bkup map[string]any) []model.ConfigDiff {
	var diffs []model.ConfigDiff

	// Check for keys in backup that differ from current
	for key, backupVal := range bkup {
		currentVal, exists := current[key]
		if !exists {
			diffs = append(diffs, model.ConfigDiff{
				Path:     key,
				NewValue: backupVal,
				DiffType: model.DiffAdded,
			})
		} else if !utils.DeepEqualJSON(currentVal, backupVal) {
			diffs = append(diffs, model.ConfigDiff{
				Path:     key,
				OldValue: currentVal,
				NewValue: backupVal,
				DiffType: model.DiffChanged,
			})
		}
	}

	// Check for keys in current that are not in backup
	for key, currentVal := range current {
		if _, exists := bkup[key]; !exists {
			diffs = append(diffs, model.ConfigDiff{
				Path:     key,
				OldValue: currentVal,
				DiffType: model.DiffRemoved,
			})
		}
	}

	return diffs
}

// compareScripts compares current and backup scripts.
func compareScripts(current []ScriptInfoResult, bkup []Script) []model.ScriptDiff {
	var diffs []model.ScriptDiff

	currentMap := make(map[string]ScriptInfoResult)
	for _, s := range current {
		currentMap[s.Name] = s
	}

	backupMap := make(map[string]Script)
	for _, s := range bkup {
		backupMap[s.Name] = s
	}

	// Check backup scripts
	for name, backupScript := range backupMap {
		if _, exists := currentMap[name]; !exists {
			diffs = append(diffs, model.ScriptDiff{
				Name:     name,
				DiffType: model.DiffAdded,
				Details:  "script will be created",
			})
		} else {
			diffs = append(diffs, model.ScriptDiff{
				Name:     name,
				DiffType: model.DiffChanged,
				Details:  fmt.Sprintf("enable: %v", backupScript.Enable),
			})
		}
	}

	// Check for scripts not in backup
	for name := range currentMap {
		if _, exists := backupMap[name]; !exists {
			diffs = append(diffs, model.ScriptDiff{
				Name:     name,
				DiffType: model.DiffRemoved,
				Details:  "script not in backup (will not be deleted)",
			})
		}
	}

	return diffs
}

// compareSchedules compares current and backup schedules.
func compareSchedules(current []ScheduleJobResult, bkup []Schedule) []model.ScheduleDiff {
	var diffs []model.ScheduleDiff

	// Simple comparison by timespec
	currentTimespecs := make(map[string]bool)
	for _, s := range current {
		currentTimespecs[s.Timespec] = true
	}

	backupTimespecs := make(map[string]bool)
	for _, s := range bkup {
		backupTimespecs[s.Timespec] = true
	}

	for _, s := range bkup {
		if !currentTimespecs[s.Timespec] {
			diffs = append(diffs, model.ScheduleDiff{
				Timespec: s.Timespec,
				DiffType: model.DiffAdded,
				Details:  fmt.Sprintf("enable: %v", s.Enable),
			})
		}
	}

	for _, s := range current {
		if !backupTimespecs[s.Timespec] {
			diffs = append(diffs, model.ScheduleDiff{
				Timespec: s.Timespec,
				DiffType: model.DiffRemoved,
				Details:  "schedule not in backup",
			})
		}
	}

	return diffs
}

// compareWebhooks compares current and backup webhooks.
func compareWebhooks(current []WebhookInfoResult, bkup []WebhookInfoBackup) []model.WebhookDiff {
	var diffs []model.WebhookDiff

	currentMap := make(map[string]WebhookInfoResult)
	for _, w := range current {
		key := fmt.Sprintf("%s:%s", w.Event, w.Name)
		currentMap[key] = w
	}

	backupMap := make(map[string]WebhookInfoBackup)
	for _, w := range bkup {
		key := fmt.Sprintf("%s:%s", w.Event, w.Name)
		backupMap[key] = w
	}

	for key, backupWh := range backupMap {
		if _, exists := currentMap[key]; !exists {
			diffs = append(diffs, model.WebhookDiff{
				Event:    backupWh.Event,
				Name:     backupWh.Name,
				DiffType: model.DiffAdded,
				Details:  fmt.Sprintf("urls: %v", backupWh.URLs),
			})
		}
	}

	for key, currentWh := range currentMap {
		if _, exists := backupMap[key]; !exists {
			diffs = append(diffs, model.WebhookDiff{
				Event:    currentWh.Event,
				Name:     currentWh.Name,
				DiffType: model.DiffRemoved,
				Details:  "webhook not in backup",
			})
		}
	}

	return diffs
}

// DiffBackups compares two backups and returns differences.
func DiffBackups(backup1, backup2 *DeviceBackup) (*model.BackupDiff, error) {
	diff := &model.BackupDiff{}

	// Parse configs
	var config1, config2 map[string]any
	if err := json.Unmarshal(backup1.Config, &config1); err != nil {
		return nil, fmt.Errorf("failed to parse backup1 config: %w", err)
	}
	if err := json.Unmarshal(backup2.Config, &config2); err != nil {
		return nil, fmt.Errorf("failed to parse backup2 config: %w", err)
	}

	// Compare configurations
	diff.ConfigDiffs = CompareConfigs(config1, config2)

	// Compare scripts
	if backup1.Scripts != nil && backup2.Scripts != nil {
		scripts1 := ConvertBackupScripts(backup1.Scripts)
		scripts2 := ConvertBackupScripts(backup2.Scripts)
		diff.ScriptDiffs = diffScripts(scripts1, scripts2)
	}

	// Compare schedules
	if backup1.Schedules != nil && backup2.Schedules != nil {
		schedules1 := ConvertBackupSchedules(backup1.Schedules)
		schedules2 := ConvertBackupSchedules(backup2.Schedules)
		diff.ScheduleDiffs = diffSchedules(schedules1, schedules2)
	}

	// Compare webhooks
	if backup1.Webhooks != nil && backup2.Webhooks != nil {
		webhooks1 := ConvertBackupWebhooks(backup1.Webhooks)
		webhooks2 := ConvertBackupWebhooks(backup2.Webhooks)
		diff.WebhookDiffs = diffWebhooks(webhooks1, webhooks2)
	}

	return diff, nil
}

// diffScripts compares two sets of backup scripts.
func diffScripts(scripts1, scripts2 []Script) []model.ScriptDiff {
	var diffs []model.ScriptDiff

	map1 := make(map[string]Script)
	for _, s := range scripts1 {
		map1[s.Name] = s
	}

	map2 := make(map[string]Script)
	for _, s := range scripts2 {
		map2[s.Name] = s
	}

	// Check scripts in backup2
	for name, script2 := range map2 {
		if _, exists := map1[name]; !exists {
			diffs = append(diffs, model.ScriptDiff{
				Name:     name,
				DiffType: model.DiffAdded,
				Details:  "script in backup2 only",
			})
		} else {
			diffs = append(diffs, model.ScriptDiff{
				Name:     name,
				DiffType: model.DiffChanged,
				Details:  fmt.Sprintf("enable: %v", script2.Enable),
			})
		}
	}

	// Check for scripts in backup1 not in backup2
	for name := range map1 {
		if _, exists := map2[name]; !exists {
			diffs = append(diffs, model.ScriptDiff{
				Name:     name,
				DiffType: model.DiffRemoved,
				Details:  "script in backup1 only",
			})
		}
	}

	return diffs
}

// diffSchedules compares two sets of backup schedules.
func diffSchedules(schedules1, schedules2 []Schedule) []model.ScheduleDiff {
	var diffs []model.ScheduleDiff

	timespecs1 := make(map[string]bool)
	for _, s := range schedules1 {
		timespecs1[s.Timespec] = true
	}

	timespecs2 := make(map[string]bool)
	for _, s := range schedules2 {
		timespecs2[s.Timespec] = true
	}

	for _, s := range schedules2 {
		if !timespecs1[s.Timespec] {
			diffs = append(diffs, model.ScheduleDiff{
				Timespec: s.Timespec,
				DiffType: model.DiffAdded,
				Details:  "schedule in backup2 only",
			})
		}
	}

	for _, s := range schedules1 {
		if !timespecs2[s.Timespec] {
			diffs = append(diffs, model.ScheduleDiff{
				Timespec: s.Timespec,
				DiffType: model.DiffRemoved,
				Details:  "schedule in backup1 only",
			})
		}
	}

	return diffs
}

// diffWebhooks compares two sets of backup webhooks.
func diffWebhooks(webhooks1, webhooks2 []WebhookInfoBackup) []model.WebhookDiff {
	var diffs []model.WebhookDiff

	map1 := make(map[string]WebhookInfoBackup)
	for _, w := range webhooks1 {
		key := fmt.Sprintf("%s:%s", w.Event, w.Name)
		map1[key] = w
	}

	map2 := make(map[string]WebhookInfoBackup)
	for _, w := range webhooks2 {
		key := fmt.Sprintf("%s:%s", w.Event, w.Name)
		map2[key] = w
	}

	for key, wh2 := range map2 {
		if _, exists := map1[key]; !exists {
			diffs = append(diffs, model.WebhookDiff{
				Event:    wh2.Event,
				Name:     wh2.Name,
				DiffType: model.DiffAdded,
				Details:  "webhook in backup2 only",
			})
		}
	}

	for key, wh1 := range map1 {
		if _, exists := map2[key]; !exists {
			diffs = append(diffs, model.WebhookDiff{
				Event:    wh1.Event,
				Name:     wh1.Name,
				DiffType: model.DiffRemoved,
				Details:  "webhook in backup1 only",
			})
		}
	}

	return diffs
}
