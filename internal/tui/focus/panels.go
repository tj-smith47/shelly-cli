// Package focus provides unified focus state management for the TUI.
package focus

import "github.com/tj-smith47/shelly-cli/internal/tui/tabs"

// GlobalPanelID identifies any focusable panel across all tabs/views.
// This replaces the fragmented per-view panel enums.
type GlobalPanelID int

// GlobalPanelID constants define all focusable panels across tabs.
const (
	// PanelNone indicates no panel is focused.
	PanelNone GlobalPanelID = iota
	// PanelDeviceList is the device list panel (shared across Dashboard, Automation, Config).
	PanelDeviceList
	// PanelDashboardInfo is the device info panel on the Dashboard tab.
	PanelDashboardInfo
	// PanelDashboardEvents is the events sidebar on the Dashboard tab.
	PanelDashboardEvents
	// PanelDashboardEnergyBars is the power consumption bars on the Dashboard tab.
	PanelDashboardEnergyBars
	// PanelDashboardEnergyHistory is the energy history sparklines on the Dashboard tab.
	PanelDashboardEnergyHistory
	// PanelDashboardJSON is the JSON viewer panel on the Dashboard tab.
	PanelDashboardJSON
	// PanelAutoScripts is the scripts list panel on the Automation tab.
	PanelAutoScripts
	// PanelAutoSchedules is the schedules list panel on the Automation tab.
	PanelAutoSchedules
	// PanelAutoWebhooks is the webhooks list panel on the Automation tab.
	PanelAutoWebhooks
	// PanelAutoVirtuals is the virtual components panel on the Automation tab.
	PanelAutoVirtuals
	// PanelAutoKVS is the key-value store panel on the Automation tab.
	PanelAutoKVS
	// PanelAutoAlerts is the alerts list panel on the Automation tab.
	PanelAutoAlerts
	// PanelConfigWiFi is the WiFi settings panel on the Config tab.
	PanelConfigWiFi
	// PanelConfigSystem is the system settings panel on the Config tab.
	PanelConfigSystem
	// PanelConfigCloud is the cloud settings panel on the Config tab.
	PanelConfigCloud
	// PanelConfigSecurity is the security settings panel on the Config tab.
	PanelConfigSecurity
	// PanelConfigBLE is the Bluetooth settings panel on the Config tab.
	PanelConfigBLE
	// PanelConfigInputs is the input configuration panel on the Config tab.
	PanelConfigInputs
	// PanelConfigProtocols is the protocol settings panel on the Config tab.
	PanelConfigProtocols
	// PanelConfigSmartHome is the smart home integrations panel on the Config tab.
	PanelConfigSmartHome
	// PanelManageDiscovery is the device discovery panel on the Manage tab.
	PanelManageDiscovery
	// PanelManageFirmware is the firmware management panel on the Manage tab.
	PanelManageFirmware
	// PanelManageBackup is the backup/restore panel on the Manage tab.
	PanelManageBackup
	// PanelManageScenes is the scene management panel on the Manage tab.
	PanelManageScenes
	// PanelManageTemplates is the config templates panel on the Manage tab.
	PanelManageTemplates
	// PanelManageBatch is the batch operations panel on the Manage tab.
	PanelManageBatch
	// PanelManageProvisioning is the provisioning wizard overlay on the Manage tab.
	PanelManageProvisioning
	// PanelManageMigration is the migration wizard overlay on the Manage tab.
	PanelManageMigration
	// PanelMonitorSummary is the summary bar on the Monitor tab (non-focusable).
	PanelMonitorSummary
	// PanelMonitorPowerRanking is the power ranking panel on the Monitor tab.
	PanelMonitorPowerRanking
	// PanelMonitorEnvironment is the environment sensors panel on the Monitor tab.
	PanelMonitorEnvironment
	// PanelMonitorAlerts is the alerts panel on the Monitor tab.
	PanelMonitorAlerts
	// PanelMonitorEventFeed is the event feed panel on the Monitor tab.
	PanelMonitorEventFeed
	// PanelFleetDevices is the cloud device list panel on the Fleet tab.
	PanelFleetDevices
	// PanelFleetGroups is the device groups panel on the Fleet tab.
	PanelFleetGroups
	// PanelFleetHealth is the health status panel on the Fleet tab.
	PanelFleetHealth
	// PanelFleetOperations is the batch operations panel on the Fleet tab.
	PanelFleetOperations
)

// panelNames maps panel IDs to their string names for debugging.
var panelNames = map[GlobalPanelID]string{
	PanelNone:                   "none",
	PanelDeviceList:             "deviceList",
	PanelDashboardInfo:          "dashboard.info",
	PanelDashboardEvents:        "dashboard.events",
	PanelDashboardEnergyBars:    "dashboard.energyBars",
	PanelDashboardEnergyHistory: "dashboard.energyHistory",
	PanelDashboardJSON:          "dashboard.json",
	PanelAutoScripts:            "auto.scripts",
	PanelAutoSchedules:          "auto.schedules",
	PanelAutoWebhooks:           "auto.webhooks",
	PanelAutoVirtuals:           "auto.virtuals",
	PanelAutoKVS:                "auto.kvs",
	PanelAutoAlerts:             "auto.alerts",
	PanelConfigWiFi:             "config.wifi",
	PanelConfigSystem:           "config.system",
	PanelConfigCloud:            "config.cloud",
	PanelConfigSecurity:         "config.security",
	PanelConfigBLE:              "config.ble",
	PanelConfigInputs:           "config.inputs",
	PanelConfigProtocols:        "config.protocols",
	PanelConfigSmartHome:        "config.smarthome",
	PanelManageDiscovery:        "manage.discovery",
	PanelManageFirmware:         "manage.firmware",
	PanelManageBackup:           "manage.backup",
	PanelManageScenes:           "manage.scenes",
	PanelManageTemplates:        "manage.templates",
	PanelManageBatch:            "manage.batch",
	PanelManageProvisioning:     "manage.provisioning",
	PanelManageMigration:        "manage.migration",
	PanelMonitorSummary:         "monitor.summary",
	PanelMonitorPowerRanking:    "monitor.powerRanking",
	PanelMonitorEnvironment:     "monitor.environment",
	PanelMonitorAlerts:          "monitor.alerts",
	PanelMonitorEventFeed:       "monitor.eventFeed",
	PanelFleetDevices:           "fleet.devices",
	PanelFleetGroups:            "fleet.groups",
	PanelFleetHealth:            "fleet.health",
	PanelFleetOperations:        "fleet.operations",
}

// panelIndexes maps panel IDs to their 1-based Shift+N indices.
var panelIndexes = map[GlobalPanelID]int{
	// Dashboard (DeviceList handled specially as index 1)
	PanelDashboardInfo:          2,
	PanelDashboardEvents:        3,
	PanelDashboardEnergyBars:    4,
	PanelDashboardEnergyHistory: 5,
	PanelDashboardJSON:          6,
	// Automation
	PanelAutoScripts:   2,
	PanelAutoSchedules: 3,
	PanelAutoWebhooks:  4,
	PanelAutoVirtuals:  5,
	PanelAutoKVS:       6,
	PanelAutoAlerts:    7,
	// Config
	PanelConfigWiFi:      2,
	PanelConfigSystem:    3,
	PanelConfigCloud:     4,
	PanelConfigSecurity:  5,
	PanelConfigBLE:       6,
	PanelConfigInputs:    7,
	PanelConfigProtocols: 8,
	PanelConfigSmartHome: 9,
	// Manage (no device list, so starts at 1)
	PanelManageDiscovery: 1,
	PanelManageFirmware:  2,
	PanelManageBackup:    3,
	PanelManageScenes:    4,
	PanelManageTemplates: 5,
	PanelManageBatch:     6,
	// Fleet (no device list, so starts at 1)
	PanelFleetDevices:    1,
	PanelFleetGroups:     2,
	PanelFleetHealth:     3,
	PanelFleetOperations: 4,
	// Monitor (Summary is non-focusable, so starts at 1 with PowerRanking)
	PanelMonitorPowerRanking: 1,
	PanelMonitorEnvironment:  2,
	PanelMonitorAlerts:       3,
	PanelMonitorEventFeed:    4,
}

// String returns the panel name for debugging.
func (p GlobalPanelID) String() string {
	if name, ok := panelNames[p]; ok {
		return name
	}
	return "unknown"
}

// TabFor returns which tab this panel belongs to.
func (p GlobalPanelID) TabFor() tabs.TabID {
	switch p {
	case PanelNone:
		return tabs.TabDashboard
	case PanelDeviceList, PanelDashboardInfo, PanelDashboardEvents,
		PanelDashboardEnergyBars, PanelDashboardEnergyHistory, PanelDashboardJSON:
		return tabs.TabDashboard
	case PanelAutoScripts, PanelAutoSchedules, PanelAutoWebhooks,
		PanelAutoVirtuals, PanelAutoKVS, PanelAutoAlerts:
		return tabs.TabAutomation
	case PanelConfigWiFi, PanelConfigSystem, PanelConfigCloud,
		PanelConfigSecurity, PanelConfigBLE, PanelConfigInputs,
		PanelConfigProtocols, PanelConfigSmartHome:
		return tabs.TabConfig
	case PanelManageDiscovery, PanelManageFirmware, PanelManageBackup,
		PanelManageScenes, PanelManageTemplates, PanelManageBatch,
		PanelManageProvisioning, PanelManageMigration:
		return tabs.TabManage
	case PanelMonitorSummary, PanelMonitorPowerRanking, PanelMonitorEnvironment,
		PanelMonitorAlerts, PanelMonitorEventFeed:
		return tabs.TabMonitor
	case PanelFleetDevices, PanelFleetGroups, PanelFleetHealth,
		PanelFleetOperations:
		return tabs.TabFleet
	}
	return tabs.TabDashboard
}

// PanelIndex returns the 1-based index for Shift+N hints (0 = not applicable).
func (p GlobalPanelID) PanelIndex() int {
	// Device list is always index 1
	if p == PanelDeviceList {
		return 1
	}
	if idx, ok := panelIndexes[p]; ok {
		return idx
	}
	return 0
}
