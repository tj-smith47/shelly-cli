package shelly

import (
	"context"
	"encoding/json"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// deviceInfoResult holds parsed device info from API.
type deviceInfoResult struct {
	ID  string `json:"id"`
	MAC string `json:"mac"`
	App string `json:"app"`
	Ver string `json:"ver"`
}

// parseDeviceInfo parses raw API result into deviceInfoResult.
func parseDeviceInfo(rawResult any) (*deviceInfoResult, bool) {
	jsonBytes, err := json.Marshal(rawResult)
	if err != nil {
		return nil, false
	}
	var info deviceInfoResult
	if err := json.Unmarshal(jsonBytes, &info); err != nil {
		return nil, false
	}
	return &info, true
}

// extractPower extracts active power from device status.
func extractPower(rawStatus any) (float64, bool) {
	statusBytes, err := json.Marshal(rawStatus)
	if err != nil {
		return 0, false
	}
	var status map[string]interface{}
	if err := json.Unmarshal(statusBytes, &status); err != nil {
		return 0, false
	}
	// Look for switch:0 or em:0 with power info
	for _, val := range status {
		if valMap, ok := val.(map[string]interface{}); ok {
			if power, ok := valMap["apower"].(float64); ok {
				return power, true
			}
		}
	}
	return 0, false
}

// extractAuthEnabled extracts auth_en from device info.
func extractAuthEnabled(rawInfo any) bool {
	infoBytes, err := json.Marshal(rawInfo)
	if err != nil {
		return false
	}
	var info struct {
		Auth bool `json:"auth_en"`
	}
	if err := json.Unmarshal(infoBytes, &info); err != nil {
		return false
	}
	return info.Auth
}

// extractCloudConnected extracts connected status from cloud status.
func extractCloudConnected(rawStatus any) bool {
	statusBytes, err := json.Marshal(rawStatus)
	if err != nil {
		return false
	}
	var status struct {
		Connected bool `json:"connected"`
	}
	if err := json.Unmarshal(statusBytes, &status); err != nil {
		return false
	}
	return status.Connected
}

// GenerateDevicesReport generates a device inventory report.
func (s *Service) GenerateDevicesReport(ctx context.Context, devices map[string]model.Device) model.DeviceReport {
	report := model.NewDeviceReport("devices")
	var online, offline int

	for name, deviceCfg := range devices {
		info := model.DeviceReportInfo{
			Name:   name,
			IP:     deviceCfg.Address,
			Online: false,
		}

		// Try to get device info
		conn, connErr := s.Connect(ctx, name)
		if connErr != nil {
			report.Devices = append(report.Devices, info)
			offline++
			continue
		}

		rawResult, callErr := conn.Call(ctx, "Shelly.GetDeviceInfo", nil)
		iostreams.CloseWithDebug("closing device report connection", conn)

		if callErr != nil {
			report.Devices = append(report.Devices, info)
			offline++
			continue
		}

		if deviceInfo, ok := parseDeviceInfo(rawResult); ok {
			info.Online = true
			info.Model = deviceInfo.App
			info.Firmware = deviceInfo.Ver
			info.MAC = deviceInfo.MAC
			online++
		}

		if !info.Online {
			offline++
		}

		report.Devices = append(report.Devices, info)
	}

	report.Summary["total"] = len(report.Devices)
	report.Summary["online"] = online
	report.Summary["offline"] = offline

	return report
}

// GenerateEnergyReport generates an energy consumption report.
func (s *Service) GenerateEnergyReport(ctx context.Context, devices map[string]model.Device) model.DeviceReport {
	report := model.NewDeviceReport("energy")
	var totalPower float64
	var devicesWithEnergy int

	for name := range devices {
		conn, connErr := s.Connect(ctx, name)
		if connErr != nil {
			continue
		}

		rawStatus, callErr := conn.Call(ctx, "Shelly.GetStatus", nil)
		iostreams.CloseWithDebug("closing energy report connection", conn)

		if callErr != nil {
			continue
		}

		if power, ok := extractPower(rawStatus); ok {
			totalPower += power
			devicesWithEnergy++
		}
	}

	report.Summary["total_power_w"] = totalPower
	report.Summary["devices_reporting"] = devicesWithEnergy

	return report
}

// GenerateAuditReport generates a security audit report.
func (s *Service) GenerateAuditReport(ctx context.Context, devices map[string]model.Device) model.DeviceReport {
	report := model.NewDeviceReport("audit")
	var authEnabled, cloudEnabled, outdated int

	for name := range devices {
		conn, connErr := s.Connect(ctx, name)
		if connErr != nil {
			continue
		}

		rawDeviceInfo, infoErr := conn.Call(ctx, "Shelly.GetDeviceInfo", nil)
		if infoErr == nil && extractAuthEnabled(rawDeviceInfo) {
			authEnabled++
		}

		rawCloudStatus, cloudErr := conn.Call(ctx, "Cloud.GetStatus", nil)
		if cloudErr == nil && extractCloudConnected(rawCloudStatus) {
			cloudEnabled++
		}

		iostreams.CloseWithDebug("closing audit report connection", conn)
	}

	report.Summary["devices_scanned"] = len(devices)
	report.Summary["auth_enabled"] = authEnabled
	report.Summary["auth_disabled"] = len(devices) - authEnabled
	report.Summary["cloud_connected"] = cloudEnabled
	report.Summary["outdated_firmware"] = outdated

	return report
}
