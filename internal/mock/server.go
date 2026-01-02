package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// DeviceServer mocks Shelly device HTTP endpoints.
type DeviceServer struct {
	*httptest.Server
	fixtures *Fixtures
	mu       sync.RWMutex
	state    map[string]DeviceState
	upgrader websocket.Upgrader
}

// NewDeviceServer creates a mock HTTP server for device requests.
func NewDeviceServer(fixtures *Fixtures) *DeviceServer {
	ds := &DeviceServer{
		fixtures: fixtures,
		state:    make(map[string]DeviceState),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

	for name, state := range fixtures.DeviceStates {
		ds.state[name] = copyState(state)
	}

	ds.Server = httptest.NewServer(http.HandlerFunc(ds.handleRequest))
	return ds
}

func copyState(src DeviceState) DeviceState {
	dst := make(DeviceState)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// DeviceURL returns the mock URL for a specific device.
func (ds *DeviceServer) DeviceURL(deviceName string) string {
	return ds.URL + "/devices/" + deviceName
}

func (ds *DeviceServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/devices/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 1 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	deviceName := parts[0]
	endpoint := ""
	if len(parts) > 1 {
		endpoint = "/" + parts[1]
	}

	ds.mu.RLock()
	state, hasState := ds.state[deviceName]
	ds.mu.RUnlock()

	device := ds.findDevice(deviceName)
	if device == nil {
		http.NotFound(w, r)
		return
	}

	if !hasState {
		state = make(DeviceState)
	}

	if device.Generation == 2 || device.Generation == 0 {
		ds.handleGen2(w, r, endpoint, state, device)
	} else {
		ds.handleGen1(w, r, endpoint, state, device)
	}
}

func (ds *DeviceServer) findDevice(name string) *DeviceFixture {
	for i := range ds.fixtures.Config.Devices {
		d := &ds.fixtures.Config.Devices[i]
		if d.Name == name || strings.EqualFold(d.Name, name) {
			return d
		}
	}
	return nil
}

// rpcRequest represents a JSON-RPC request from shelly-go.
type rpcRequest struct {
	ID     int            `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

// rpcResponse represents a JSON-RPC response for shelly-go.
type rpcResponse struct {
	ID     int    `json:"id"`
	Result any    `json:"result,omitempty"`
	Error  *error `json:"error,omitempty"`
}

func (ds *DeviceServer) handleGen2(w http.ResponseWriter, r *http.Request, endpoint string, state DeviceState, device *DeviceFixture) {
	// Handle WebSocket upgrade on /rpc endpoint
	if endpoint == "/rpc" && websocket.IsWebSocketUpgrade(r) {
		ds.handleWebSocketRPC(w, r, state, device)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Handle JSON-RPC endpoint
	if endpoint == "/rpc" {
		ds.handleGen2RPC(w, r, state, device)
		return
	}

	// Legacy direct endpoint access (for testing)
	switch {
	case endpoint == "/rpc/Shelly.GetDeviceInfo":
		ds.writeGen2DeviceInfo(w, device)

	case endpoint == "/rpc/Shelly.GetStatus":
		ds.writeJSON(w, state)

	case endpoint == "/rpc/Shelly.GetConfig":
		ds.writeJSON(w, map[string]any{
			"sys": map[string]any{"device": map[string]any{"name": device.Name}},
		})

	case endpoint == "/rpc/Shelly.GetComponents":
		ds.writeComponents(w, state)

	case strings.HasPrefix(endpoint, "/rpc/Switch."):
		ds.handleSwitchRPC(w, r, endpoint, state, device)

	case strings.HasPrefix(endpoint, "/rpc/Cover."):
		ds.handleCoverRPC(w, r, endpoint, state, device)

	case strings.HasPrefix(endpoint, "/rpc/Light."):
		ds.handleLightRPC(w, r, endpoint, state, device)

	default:
		http.NotFound(w, r)
	}
}

//nolint:gocyclo // RPC method router has inherent branching complexity
func (ds *DeviceServer) handleGen2RPC(w http.ResponseWriter, r *http.Request, state DeviceState, device *DeviceFixture) {
	var req rpcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ds.writeRPCError(w, 0, "invalid request")
		return
	}

	var result any
	switch req.Method {
	case "Shelly.GetDeviceInfo":
		result = ds.gen2DeviceInfo(device)

	case "Shelly.GetStatus":
		result = state

	case "Shelly.GetConfig":
		result = map[string]any{
			"sys": map[string]any{"device": map[string]any{"name": device.Name}},
		}

	case "Sys.GetConfig":
		result = ds.getSysConfig(state, device)

	case "Sys.GetStatus":
		result = ds.getSysStatus(state, device)

	case "Shelly.GetComponents":
		result = ds.getComponents(state)

	case "Switch.GetStatus":
		id := ds.getIDFromParams(req.Params)
		result = ds.getComponentState(state, fmt.Sprintf("switch:%d", id))

	case "Switch.Set":
		id := ds.getIDFromParams(req.Params)
		on, ok := req.Params["on"].(bool)
		if !ok {
			on = false
		}
		key := fmt.Sprintf("switch:%d", id)
		wasOn := ds.updateSwitchState(device.Name, key, on)
		result = map[string]any{"was_on": wasOn}

	case "Switch.Toggle":
		id := ds.getIDFromParams(req.Params)
		key := fmt.Sprintf("switch:%d", id)
		wasOn := ds.toggleSwitchState(device.Name, key)
		result = map[string]any{"was_on": wasOn}

	case "Cover.GetStatus":
		id := ds.getIDFromParams(req.Params)
		result = ds.getComponentState(state, fmt.Sprintf("cover:%d", id))

	case "Cover.Open", "Cover.Close", "Cover.Stop":
		result = map[string]any{}

	case "Input.GetStatus":
		id := ds.getIDFromParams(req.Params)
		result = ds.getComponentState(state, fmt.Sprintf("input:%d", id))

	case "Input.GetConfig":
		id := ds.getIDFromParams(req.Params)
		result = ds.getInputConfig(state, id)

	case "Light.GetStatus":
		id := ds.getIDFromParams(req.Params)
		result = ds.getComponentState(state, fmt.Sprintf("light:%d", id))

	case "Light.Set":
		id := ds.getIDFromParams(req.Params)
		on, ok := req.Params["on"].(bool)
		if !ok {
			on = false
		}
		brightness := 0
		if b, ok := req.Params["brightness"].(float64); ok {
			brightness = int(b)
		}
		key := fmt.Sprintf("light:%d", id)
		ds.updateLightState(device.Name, key, on, brightness)
		result = map[string]any{}

	case "Light.Toggle":
		id := ds.getIDFromParams(req.Params)
		key := fmt.Sprintf("light:%d", id)
		wasOn := ds.toggleSwitchState(device.Name, key)
		result = map[string]any{"was_on": wasOn}

	case "RGB.GetStatus":
		id := ds.getIDFromParams(req.Params)
		result = ds.getComponentState(state, fmt.Sprintf("rgb:%d", id))

	case "RGB.Set":
		id := ds.getIDFromParams(req.Params)
		key := fmt.Sprintf("rgb:%d", id)
		ds.updateRGBState(device.Name, key, req.Params)
		result = map[string]any{}

	case "RGB.Toggle":
		id := ds.getIDFromParams(req.Params)
		key := fmt.Sprintf("rgb:%d", id)
		wasOn := ds.toggleSwitchState(device.Name, key)
		result = map[string]any{"was_on": wasOn}

	case "RGBW.GetStatus":
		id := ds.getIDFromParams(req.Params)
		result = ds.getComponentState(state, fmt.Sprintf("rgbw:%d", id))

	case "RGBW.Set":
		id := ds.getIDFromParams(req.Params)
		key := fmt.Sprintf("rgbw:%d", id)
		ds.updateRGBWState(device.Name, key, req.Params)
		result = map[string]any{}

	case "RGBW.Toggle":
		id := ds.getIDFromParams(req.Params)
		key := fmt.Sprintf("rgbw:%d", id)
		wasOn := ds.toggleSwitchState(device.Name, key)
		result = map[string]any{"was_on": wasOn}

	case "Schedule.List":
		result = ds.getScheduleList(state, device.Name)

	case "Schedule.Create":
		result = ds.createSchedule(req.Params, device.Name)

	case "Schedule.Update":
		result = ds.updateSchedule(req.Params, device.Name)

	case "Schedule.Delete":
		result = ds.deleteSchedule(req.Params, device.Name)

	case "Schedule.DeleteAll":
		result = map[string]any{}

	case "Script.GetCode":
		id := ds.getIDFromParams(req.Params)
		result = ds.getScriptCode(state, id)

	case "Script.List":
		result = ds.getScriptList(state)

	case "Script.Start", "Script.Stop":
		result = map[string]any{"was_running": false}

	case "Script.Create":
		result = map[string]any{"id": 1}

	case "Script.PutCode":
		// Mock script code upload - return success
		result = map[string]any{"len": 0}

	case "Script.SetConfig":
		// Mock script config update - return success
		result = map[string]any{"restart_required": false}

	case "Script.Eval":
		// Mock script eval - return a result
		result = map[string]any{"result": 3}

	case "Script.GetStatus":
		// Mock script status
		id := ds.getIDFromParams(req.Params)
		result = map[string]any{
			"id":        id,
			"running":   true,
			"mem_usage": 1024,
			"mem_peak":  2048,
			"mem_free":  4096,
		}

	case "Script.Delete":
		// Mock script delete - return success
		result = map[string]any{}

	case "Shelly.SetAuth":
		// Mock auth enable/disable - return success
		result = map[string]any{}

	case "Zigbee.SetConfig":
		// Return success for Zigbee config operations
		result = map[string]any{"restart_required": false}

	case "Zigbee.GetConfig":
		// Return mock Zigbee config
		result = map[string]any{"enable": true}

	case "Zigbee.GetStatus":
		// Return mock Zigbee status
		result = map[string]any{
			"network_state": "joined",
			"eui64":         "0x00124B001234ABCD",
			"pan_id":        float64(12345),
			"channel":       float64(15),
		}

	case "Zigbee.StartNetworkSteering":
		// Return success for starting network steering
		result = map[string]any{}

	case "Wifi.GetStatus":
		// Return mock WiFi status
		result = map[string]any{
			"sta_ip":          "192.168.1.100",
			"status":          "got ip",
			"ssid":            "HomeNetwork",
			"rssi":            float64(-45),
			"ap_client_count": float64(0),
		}

	case "Wifi.GetConfig":
		// Return mock WiFi config
		result = map[string]any{
			"ap": map[string]any{
				"ssid":           "ShellyAP",
				"is_open":        true,
				"enable":         true,
				"range_extender": map[string]any{"enable": false},
			},
			"sta": map[string]any{
				"ssid":     "HomeNetwork",
				"is_open":  false,
				"enable":   true,
				"ipv4mode": "dhcp",
			},
			"sta1": map[string]any{
				"ssid":     "",
				"is_open":  false,
				"enable":   false,
				"ipv4mode": "dhcp",
			},
			"roam": map[string]any{
				"rssi_thr": -80,
				"interval": 60,
			},
		}

	case "Wifi.SetConfig":
		// Return success for WiFi config operations
		result = map[string]any{"restart_required": false}

	case "Wifi.Scan":
		// Return mock WiFi scan results
		result = map[string]any{
			"results": []map[string]any{
				{
					"ssid":    "HomeNetwork",
					"bssid":   "AA:BB:CC:DD:EE:FF",
					"auth":    3,
					"channel": 6,
					"rssi":    -45,
				},
				{
					"ssid":    "GuestNetwork",
					"bssid":   "11:22:33:44:55:66",
					"auth":    3,
					"channel": 11,
					"rssi":    -60,
				},
				{
					"ssid":    "NeighborWiFi",
					"bssid":   "AA:11:BB:22:CC:33",
					"auth":    3,
					"channel": 1,
					"rssi":    -75,
				},
			},
		}

	case "Wifi.ListAPClients":
		// Return mock AP clients list
		result = map[string]any{
			"ts": float64(1700000000),
			"ap_clients": []map[string]any{
				{
					"mac":       "AA:BB:CC:DD:EE:01",
					"ip":        "192.168.33.2",
					"ip_static": false,
					"mport":     0,
					"since":     float64(3600),
				},
				{
					"mac":       "AA:BB:CC:DD:EE:02",
					"ip":        "192.168.33.3",
					"ip_static": false,
					"mport":     0,
					"since":     float64(1800),
				},
			},
		}

	case "MQTT.GetStatus":
		// Return MQTT status from device state or default
		if mqttState, ok := state["mqtt"].(map[string]any); ok {
			result = mqttState
		} else {
			result = map[string]any{"connected": false}
		}

	case "MQTT.GetConfig":
		// Return MQTT config from device state or default
		if mqttConfig, ok := state["mqtt_config"].(map[string]any); ok {
			result = mqttConfig
		} else {
			result = map[string]any{
				"enable":       false,
				"server":       "",
				"client_id":    "",
				"topic_prefix": "",
				"rpc_ntf":      true,
				"status_ntf":   false,
			}
		}

	case "MQTT.SetConfig":
		// Return success for MQTT config operations
		result = map[string]any{"restart_required": false}

	case "Eth.GetStatus":
		// Return mock Ethernet status from device state or default
		if ethState, ok := state["eth"].(map[string]any); ok {
			result = ethState
		} else {
			result = map[string]any{
				"ip": "192.168.1.50",
			}
		}

	case "Eth.GetConfig":
		// Return mock Ethernet config
		result = map[string]any{
			"enable":   true,
			"ipv4mode": "dhcp",
		}

	case "Eth.SetConfig":
		// Return success for Ethernet config operations
		result = map[string]any{"restart_required": false}

	case "Cloud.GetStatus":
		// Return cloud status from device state or default
		if cloudState, ok := state["cloud"].(map[string]any); ok {
			result = cloudState
		} else {
			result = map[string]any{"connected": false}
		}

	case "Cloud.GetConfig":
		// Return cloud config from device state or default
		if cloudConfig, ok := state["cloud_config"].(map[string]any); ok {
			result = cloudConfig
		} else {
			result = map[string]any{
				"enable": false,
				"server": "shelly-13-eu.shelly.cloud:6022/jrpc",
			}
		}

	case "Ws.GetConfig":
		// Return WebSocket config from device state or default
		if wsConfig, ok := state["ws_config"].(map[string]any); ok {
			result = wsConfig
		} else {
			result = map[string]any{
				"enable": true,
				"server": "",
				"ssl_ca": "*",
			}
		}

	case "Ws.GetStatus":
		// Return WebSocket status from device state or default
		if wsStatus, ok := state["ws_status"].(map[string]any); ok {
			result = wsStatus
		} else {
			result = map[string]any{
				"connected": false,
			}
		}

	case "Modbus.GetStatus":
		// Return Modbus status from device state or default
		if modbusState, ok := state["modbus"].(map[string]any); ok {
			result = modbusState
		} else {
			result = map[string]any{
				"enabled": false,
			}
		}

	case "Modbus.GetConfig":
		// Return Modbus config from device state or default
		if modbusConfig, ok := state["modbus_config"].(map[string]any); ok {
			result = modbusConfig
		} else {
			result = map[string]any{
				"enable": false,
			}
		}

	case "Modbus.SetConfig":
		// Return success for Modbus config operations
		result = map[string]any{"restart_required": false}

	case "EM.GetStatus":
		// Return EM component status from device state
		id := ds.getIDFromParams(req.Params)
		result = ds.getEMStatus(state, id)

	case "EM.ResetCounters":
		// Reset EM counters - return success
		result = map[string]any{}

	case "EM1.GetStatus":
		// Return EM1 component status from device state
		id := ds.getIDFromParams(req.Params)
		result = ds.getEM1Status(state, id)

	case "EMData.GetRecords":
		// Return EMData records from device state or default
		id := ds.getIDFromParams(req.Params)
		result = ds.getEMDataRecords(state, id)

	case "EMData.GetData":
		// Return EMData history from device state or default
		id := ds.getIDFromParams(req.Params)
		result = ds.getEMDataHistory(state, id)

	case "EM1Data.GetRecords":
		// Return EM1Data records from device state or default
		id := ds.getIDFromParams(req.Params)
		result = ds.getEM1DataRecords(state, id)

	case "EM1Data.GetData":
		// Return EM1Data history from device state or default
		id := ds.getIDFromParams(req.Params)
		result = ds.getEM1DataHistory(state, id)

	case "BTHome.GetStatus":
		// Return BTHome component status from device state
		result = ds.getBTHomeStatus(state)

	case "BTHomeDevice.GetStatus":
		// Return BTHome device status from device state
		id := ds.getIDFromParams(req.Params)
		result = ds.getBTHomeDeviceStatus(state, id)

	case "BTHomeDevice.GetConfig":
		// Return BTHome device config from device state
		id := ds.getIDFromParams(req.Params)
		result = ds.getBTHomeDeviceConfig(state, id)

	case "BTHomeDevice.GetKnownObjects":
		// Return BTHome device known objects from device state
		id := ds.getIDFromParams(req.Params)
		result = ds.getBTHomeDeviceKnownObjects(state, id)

	case "BTHome.StartDeviceDiscovery":
		// Return success for BTHome discovery start
		result = map[string]any{}

	case "BTHome.AddDevice":
		// Return success for BTHome device add
		result = map[string]any{"key": "mock-key-12345"}

	case "BTHome.DeleteDevice":
		// Return success for BTHome device delete
		result = map[string]any{}

	case "Shelly.ListMethods":
		// Return mock list of available methods for debug/methods command
		result = map[string]any{
			"methods": []string{
				"Shelly.GetDeviceInfo",
				"Shelly.GetStatus",
				"Shelly.GetConfig",
				"Shelly.Reboot",
				"Shelly.ListMethods",
				"Switch.GetStatus",
				"Switch.Set",
				"Switch.Toggle",
				"Cover.GetStatus",
				"Cover.Open",
				"Cover.Close",
				"Light.GetStatus",
				"Light.Set",
				"Script.List",
				"Script.GetCode",
				"Script.Start",
				"Script.Stop",
				"Sys.GetConfig",
				"Wifi.GetStatus",
				"MQTT.GetStatus",
				"MQTT.GetConfig",
				"Input.GetStatus",
				"Input.GetConfig",
				"Input.SetConfig",
			},
		}

	case "LoRa.GetConfig":
		// Return LoRa config from device state or default
		id := ds.getIDFromParams(req.Params)
		result = ds.getLoRaConfig(state, id)

	case "LoRa.GetStatus":
		// Return LoRa status from device state or default
		id := ds.getIDFromParams(req.Params)
		result = ds.getLoRaStatus(state, id)

	case "LoRa.SetConfig":
		// Return success for LoRa config operations
		result = map[string]any{"restart_required": false}

	case "LoRa.SendBytes":
		// Return success for LoRa send operations
		result = map[string]any{}

	case "Shelly.PutUserCA":
		// Return success for CA certificate installation
		result = map[string]any{"len": 1024}

	case "Shelly.PutTLSClientCert":
		// Return success for client certificate installation
		result = map[string]any{"len": 2048}

	case "Shelly.SetConfig":
		// Return success for device configuration import
		result = map[string]any{"restart_required": false}

	case "Shelly.CheckForUpdate":
		// Return firmware update status from device state or default
		result = map[string]any{
			"stable": map[string]any{
				"version":  "1.5.0",
				"build_id": "20250101-120000/1.5.0",
			},
		}

	case "Shelly.Update":
		// Return success for firmware update
		result = map[string]any{}

	case "Shelly.Reboot":
		// Return success for reboot
		result = map[string]any{}

	case "Shelly.FactoryReset":
		// Return success for factory reset
		result = map[string]any{}

	case "Matter.GetStatus":
		// Return Matter status from device state
		result = ds.getMatterStatus(state)

	case "Matter.GetConfig":
		// Return Matter config from device state
		result = ds.getMatterConfig(state)

	case "Matter.SetConfig":
		// Return success for Matter config operations
		result = map[string]any{"restart_required": false}

	case "Matter.FactoryReset":
		// Return success for Matter factory reset
		result = map[string]any{}

	case "Matter.GetCommissioningCode":
		// Return commissioning code from device state
		result = ds.getMatterCommissioningCode(state)

	case "Thermostat.GetStatus":
		// Return thermostat status from device state
		id := ds.getIDFromParams(req.Params)
		result = ds.getThermostatStatus(state, id)

	case "Thermostat.GetConfig":
		// Return thermostat config from device state
		id := ds.getIDFromParams(req.Params)
		result = ds.getThermostatConfig(state, id)

	case "Thermostat.SetConfig":
		// Check if device state has error simulation enabled
		if _, hasError := state["thermostat_error"]; hasError {
			ds.writeRPCError(w, req.ID, "thermostat config failed")
			return
		}
		// Return success for thermostat config operations
		result = map[string]any{"restart_required": false}

	case "Thermostat.Override":
		// Return success for thermostat override operations
		result = map[string]any{}

	case "Thermostat.CancelOverride":
		// Return success for thermostat cancel override operations
		result = map[string]any{}

	case "Virtual.Add":
		// Return success for virtual component creation
		id := 200
		if idVal, ok := req.Params["id"].(float64); ok && idVal > 0 {
			id = int(idVal)
		}
		result = map[string]any{"id": id}

	case "Virtual.Delete":
		// Return success for virtual component deletion
		result = map[string]any{}

	case "Boolean.Set", "Boolean.Toggle":
		// Return success for virtual boolean operations
		result = map[string]any{}

	case "Boolean.GetStatus":
		// Return virtual boolean status
		id := ds.getIDFromParams(req.Params)
		result = ds.getVirtualBooleanStatus(state, id)

	case "Number.Set":
		// Return success for virtual number operations
		result = map[string]any{}

	case "Number.GetStatus":
		// Return virtual number status
		id := ds.getIDFromParams(req.Params)
		result = ds.getVirtualNumberStatus(state, id)

	case "Text.Set":
		// Return success for virtual text operations
		result = map[string]any{}

	case "Text.GetStatus":
		// Return virtual text status
		id := ds.getIDFromParams(req.Params)
		result = ds.getVirtualTextStatus(state, id)

	case "Enum.Set":
		// Return success for virtual enum operations
		result = map[string]any{}

	case "Enum.GetStatus":
		// Return virtual enum status
		id := ds.getIDFromParams(req.Params)
		result = ds.getVirtualEnumStatus(state, id)

	case "Button.Trigger":
		// Return success for virtual button trigger
		result = map[string]any{}

	default:
		ds.writeRPCError(w, req.ID, "method not found")
		return
	}

	ds.writeRPCResult(w, req.ID, result)
}

func (ds *DeviceServer) gen2DeviceInfo(device *DeviceFixture) map[string]any {
	mac := strings.ReplaceAll(device.MAC, ":", "")
	return map[string]any{
		"id":    "shelly" + strings.ToLower(strings.ReplaceAll(device.Model, " ", "")) + "-" + mac,
		"mac":   device.MAC,
		"model": device.Model,
		"gen":   device.Generation,
		"fw_id": "20241210-092317/1.4.4-g6d2a586",
		"ver":   "1.4.4",
		"app":   device.Type,
		"name":  device.Name,
	}
}

func (ds *DeviceServer) getComponents(state DeviceState) map[string]any {
	var comps []map[string]string
	for key := range state {
		if strings.Contains(key, ":") {
			comps = append(comps, map[string]string{"key": key})
		}
	}
	return map[string]any{"components": comps}
}

func (ds *DeviceServer) getComponentState(state DeviceState, key string) any {
	if comp, ok := state[key]; ok {
		return comp
	}
	return map[string]any{}
}

func (ds *DeviceServer) getIDFromParams(params map[string]any) int {
	if id, ok := params["id"].(float64); ok {
		return int(id)
	}
	return 0
}

func (ds *DeviceServer) writeRPCResult(w http.ResponseWriter, id int, result any) {
	resp := rpcResponse{ID: id, Result: result}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ds *DeviceServer) writeRPCError(w http.ResponseWriter, id int, message string) {
	w.WriteHeader(http.StatusNotFound)
	resp := map[string]any{
		"id":    id,
		"error": map[string]any{"code": -32601, "message": message},
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ds *DeviceServer) handleGen1(w http.ResponseWriter, r *http.Request, endpoint string, state DeviceState, device *DeviceFixture) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case endpoint == "/shelly":
		ds.writeJSON(w, map[string]any{
			"type": device.Type,
			"mac":  strings.ReplaceAll(device.MAC, ":", ""),
		})

	case endpoint == "/status":
		ds.writeJSON(w, state)

	case endpoint == "/settings":
		ds.writeJSON(w, map[string]any{
			"device": map[string]any{"type": device.Type},
			"name":   device.Name,
		})

	case strings.HasPrefix(endpoint, "/relay/"):
		ds.handleGen1Relay(w, r, endpoint, state, device)

	case strings.HasPrefix(endpoint, "/light/"):
		ds.handleGen1Light(w, r, endpoint, state, device)

	case strings.HasPrefix(endpoint, "/settings/actions"):
		ds.handleGen1Actions(w, r, device)

	default:
		http.NotFound(w, r)
	}
}

func (ds *DeviceServer) writeGen2DeviceInfo(w http.ResponseWriter, device *DeviceFixture) {
	mac := strings.ReplaceAll(device.MAC, ":", "")
	ds.writeJSON(w, map[string]any{
		"id":    "shelly" + strings.ToLower(strings.ReplaceAll(device.Model, " ", "")) + "-" + mac,
		"mac":   device.MAC,
		"model": device.Model,
		"gen":   device.Generation,
		"fw_id": "20241210-092317/1.4.4-g6d2a586",
		"ver":   "1.4.4",
		"app":   device.Type,
		"name":  device.Name,
	})
}

func (ds *DeviceServer) writeComponents(w http.ResponseWriter, state DeviceState) {
	var comps []map[string]string
	for key := range state {
		if strings.Contains(key, ":") {
			comps = append(comps, map[string]string{"key": key})
		}
	}
	ds.writeJSON(w, map[string]any{"components": comps})
}

func (ds *DeviceServer) handleSwitchRPC(w http.ResponseWriter, r *http.Request, endpoint string, state DeviceState, device *DeviceFixture) {
	switch endpoint {
	case "/rpc/Switch.GetStatus":
		id := ds.parseIDParam(r)
		ds.writeComponentState(w, state, fmt.Sprintf("switch:%d", id))

	case "/rpc/Switch.Set":
		var req struct {
			ID *int `json:"id"`
			On bool `json:"on"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			ds.writeJSON(w, map[string]any{"was_on": false})
			return
		}
		id := 0
		if req.ID != nil {
			id = *req.ID
		}
		key := fmt.Sprintf("switch:%d", id)
		wasOn := ds.updateSwitchState(device.Name, key, req.On)
		ds.writeJSON(w, map[string]any{"was_on": wasOn})

	case "/rpc/Switch.Toggle":
		id := ds.parseIDParam(r)
		key := fmt.Sprintf("switch:%d", id)
		wasOn := ds.toggleSwitchState(device.Name, key)
		ds.writeJSON(w, map[string]any{"was_on": wasOn})

	default:
		http.NotFound(w, r)
	}
}

func (ds *DeviceServer) handleCoverRPC(w http.ResponseWriter, r *http.Request, endpoint string, state DeviceState, _ *DeviceFixture) {
	switch endpoint {
	case "/rpc/Cover.GetStatus":
		id := ds.parseIDParam(r)
		ds.writeComponentState(w, state, fmt.Sprintf("cover:%d", id))

	case "/rpc/Cover.Open", "/rpc/Cover.Close", "/rpc/Cover.Stop":
		ds.writeJSON(w, map[string]any{})

	default:
		http.NotFound(w, r)
	}
}

func (ds *DeviceServer) handleLightRPC(w http.ResponseWriter, r *http.Request, endpoint string, state DeviceState, device *DeviceFixture) {
	switch endpoint {
	case "/rpc/Light.GetStatus":
		id := ds.parseIDParam(r)
		ds.writeComponentState(w, state, fmt.Sprintf("light:%d", id))

	case "/rpc/Light.Set":
		var req struct {
			ID         *int `json:"id"`
			On         bool `json:"on"`
			Brightness int  `json:"brightness"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			ds.writeJSON(w, map[string]any{})
			return
		}
		id := 0
		if req.ID != nil {
			id = *req.ID
		}
		key := fmt.Sprintf("light:%d", id)
		ds.updateLightState(device.Name, key, req.On, req.Brightness)
		ds.writeJSON(w, map[string]any{})

	default:
		http.NotFound(w, r)
	}
}

func (ds *DeviceServer) handleGen1Relay(w http.ResponseWriter, r *http.Request, _ string, _ DeviceState, device *DeviceFixture) {
	turn := r.URL.Query().Get("turn")
	if turn != "" {
		ds.mu.Lock()
		if ds.state[device.Name] == nil {
			ds.state[device.Name] = make(DeviceState)
		}
		if relay, ok := ds.state[device.Name]["relay"].(map[string]any); ok {
			relay["ison"] = turn == "on"
		} else {
			ds.state[device.Name]["relay"] = map[string]any{"ison": turn == "on"}
		}
		ds.mu.Unlock()
	}

	ds.mu.RLock()
	if relay, ok := ds.state[device.Name]["relay"]; ok {
		ds.mu.RUnlock()
		ds.writeJSON(w, relay)
	} else {
		ds.mu.RUnlock()
		ds.writeJSON(w, map[string]any{"ison": false})
	}
}

func (ds *DeviceServer) handleGen1Light(w http.ResponseWriter, _ *http.Request, _ string, state DeviceState, _ *DeviceFixture) {
	if lights, ok := state["lights"].([]any); ok && len(lights) > 0 {
		ds.writeJSON(w, lights[0])
	} else {
		ds.writeJSON(w, map[string]any{"ison": false})
	}
}

// handleGen1Actions handles Gen1 /settings/actions endpoint for action URL management.
func (ds *DeviceServer) handleGen1Actions(w http.ResponseWriter, r *http.Request, device *DeviceFixture) {
	// Parse query parameters
	index := r.URL.Query().Get("index")
	name := r.URL.Query().Get("name")
	enabled := r.URL.Query().Get("enabled")

	ds.mu.Lock()
	if ds.state[device.Name] == nil {
		ds.state[device.Name] = make(DeviceState)
	}
	if ds.state[device.Name]["actions"] == nil {
		ds.state[device.Name]["actions"] = make(map[string]any)
	}

	// Store the action state
	actionKey := index + "_" + name
	actions, ok := ds.state[device.Name]["actions"].(map[string]any)
	if !ok {
		actions = make(map[string]any)
	}
	actions[actionKey] = map[string]any{
		"index":   index,
		"name":    name,
		"enabled": enabled == strTrue,
	}
	ds.mu.Unlock()

	// Return success response (Gen1 actions endpoint returns the action settings)
	ds.writeJSON(w, map[string]any{
		"actions": []map[string]any{
			{
				"index":   index,
				"name":    name,
				"enabled": enabled == strTrue,
				"urls":    []string{},
			},
		},
	})
}

func (ds *DeviceServer) parseIDParam(r *http.Request) int {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return 0
	}
	return req.ID
}

func (ds *DeviceServer) writeComponentState(w http.ResponseWriter, state DeviceState, key string) {
	if comp, ok := state[key]; ok {
		ds.writeJSON(w, comp)
	} else {
		ds.writeJSON(w, map[string]any{})
	}
}

func (ds *DeviceServer) updateSwitchState(deviceName, key string, on bool) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.state[deviceName] == nil {
		ds.state[deviceName] = make(DeviceState)
	}

	wasOn := false
	if comp, ok := ds.state[deviceName][key].(map[string]any); ok {
		if v, ok := comp["output"].(bool); ok {
			wasOn = v
		}
		comp["output"] = on
	} else {
		ds.state[deviceName][key] = map[string]any{"output": on}
	}
	return wasOn
}

func (ds *DeviceServer) toggleSwitchState(deviceName, key string) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.state[deviceName] == nil {
		ds.state[deviceName] = make(DeviceState)
	}

	wasOn := false
	if comp, ok := ds.state[deviceName][key].(map[string]any); ok {
		if v, ok := comp["output"].(bool); ok {
			wasOn = v
		}
		comp["output"] = !wasOn
	} else {
		ds.state[deviceName][key] = map[string]any{"output": true}
	}
	return wasOn
}

func (ds *DeviceServer) updateLightState(deviceName, key string, on bool, brightness int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.state[deviceName] == nil {
		ds.state[deviceName] = make(DeviceState)
	}

	if comp, ok := ds.state[deviceName][key].(map[string]any); ok {
		comp["output"] = on
		if brightness > 0 {
			comp["brightness"] = brightness
		}
	} else {
		ds.state[deviceName][key] = map[string]any{"output": on, "brightness": brightness}
	}
}

func (ds *DeviceServer) updateRGBState(deviceName, key string, params map[string]any) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.state[deviceName] == nil {
		ds.state[deviceName] = make(DeviceState)
	}

	comp, ok := ds.state[deviceName][key].(map[string]any)
	if !ok {
		comp = map[string]any{"output": false}
		ds.state[deviceName][key] = comp
	}

	if on, ok := params["on"].(bool); ok {
		comp["output"] = on
	}
	if r, ok := params["red"].(float64); ok {
		comp["red"] = int(r)
	}
	if g, ok := params["green"].(float64); ok {
		comp["green"] = int(g)
	}
	if b, ok := params["blue"].(float64); ok {
		comp["blue"] = int(b)
	}
	if brightness, ok := params["brightness"].(float64); ok {
		comp["brightness"] = int(brightness)
	}
}

func (ds *DeviceServer) updateRGBWState(deviceName, key string, params map[string]any) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.state[deviceName] == nil {
		ds.state[deviceName] = make(DeviceState)
	}

	comp, ok := ds.state[deviceName][key].(map[string]any)
	if !ok {
		comp = map[string]any{"output": false}
		ds.state[deviceName][key] = comp
	}

	if on, ok := params["on"].(bool); ok {
		comp["output"] = on
	}
	if r, ok := params["red"].(float64); ok {
		comp["red"] = int(r)
	}
	if g, ok := params["green"].(float64); ok {
		comp["green"] = int(g)
	}
	if b, ok := params["blue"].(float64); ok {
		comp["blue"] = int(b)
	}
	if w, ok := params["white"].(float64); ok {
		comp["white"] = int(w)
	}
	if brightness, ok := params["brightness"].(float64); ok {
		comp["brightness"] = int(brightness)
	}
}

func (ds *DeviceServer) writeJSON(w http.ResponseWriter, v any) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetState returns the current state for a device (for testing).
func (ds *DeviceServer) GetState(deviceName string) DeviceState {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.state[deviceName]
}

// getInputConfig returns mock input configuration.
func (ds *DeviceServer) getInputConfig(_ DeviceState, id int) map[string]any {
	return map[string]any{
		"id":            id,
		"name":          fmt.Sprintf("Input %d", id),
		"type":          "switch",
		"enable":        true,
		"invert":        false,
		"factory_reset": true,
	}
}

// getScriptCode returns mock script code.
func (ds *DeviceServer) getScriptCode(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("script:%d", id)
	if script, ok := state[key].(map[string]any); ok {
		if code, ok := script["code"].(string); ok {
			return map[string]any{"data": code}
		}
	}
	// Return empty string for non-existent scripts (matches real behavior)
	return map[string]any{"data": ""}
}

// getScriptList returns mock script list.
func (ds *DeviceServer) getScriptList(state DeviceState) map[string]any {
	var scripts []map[string]any
	for key := range state {
		if strings.HasPrefix(key, "script:") {
			parts := strings.SplitN(key, ":", 2)
			if len(parts) == 2 {
				id := 0
				if _, err := fmt.Sscanf(parts[1], "%d", &id); err == nil {
					scripts = append(scripts, map[string]any{
						"id":      id,
						"name":    fmt.Sprintf("Script %d", id),
						"enable":  true,
						"running": false,
					})
				}
			}
		}
	}
	if len(scripts) == 0 {
		scripts = []map[string]any{}
	}
	return map[string]any{"scripts": scripts}
}

// getSysConfig returns mock Sys.GetConfig response for CoIoT and other system config.
func (ds *DeviceServer) getSysConfig(state DeviceState, device *DeviceFixture) map[string]any {
	result := map[string]any{
		"device": map[string]any{
			"name":         device.Name,
			"mac":          device.MAC,
			"model":        device.Model,
			"fw_id":        "20241210-092317/1.4.4-g6d2a586",
			"discoverable": true,
			"eco_mode":     false,
		},
		"location": map[string]any{
			"tz":  "UTC",
			"lat": 0.0,
			"lon": 0.0,
		},
		"debug": map[string]any{
			"level": 2,
		},
		"ui_data": map[string]any{},
		"rpc_udp": map[string]any{
			"dst_addr":    "",
			"listen_port": nil,
		},
		"sntp": map[string]any{
			"server": "time.google.com",
		},
		"cfg_rev": 0,
	}

	// Add CoIoT section if present in state
	if coiot, ok := state["coiot"].(map[string]any); ok {
		result["coiot"] = coiot
	}

	// Add sys section if present in state (overrides default)
	if sys, ok := state["sys"].(map[string]any); ok {
		result["sys"] = sys
	} else {
		result["sys"] = map[string]any{
			"device": map[string]any{
				"name": device.Name,
			},
		}
	}

	return result
}

// getSysStatus returns mock Sys.GetStatus response for firmware.
func (ds *DeviceServer) getSysStatus(_ DeviceState, device *DeviceFixture) map[string]any {
	return map[string]any{
		"mac":              device.MAC,
		"restart_required": false,
		"time":             "12:00",
		"unixtime":         1700000000,
		"uptime":           3600,
		"ram_size":         262144,
		"ram_free":         131072,
		"fs_size":          1048576,
		"fs_free":          524288,
		"cfg_rev":          10,
		"kvs_rev":          0,
		"schedule_rev":     0,
		"webhook_rev":      0,
		"available_updates": map[string]any{
			"stable": map[string]any{
				"version":  "1.4.4",
				"build_id": "20241210-092317/1.4.4-g6d2a586",
			},
		},
		"reset_reason": 1,
	}
}

// getEMStatus returns mock EM component status from device state.
func (ds *DeviceServer) getEMStatus(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("em:%d", id)
	if em, ok := state[key].(map[string]any); ok {
		return em
	}
	// Return default EM status if not in state
	return map[string]any{
		"id":               id,
		"a_current":        1.5,
		"a_voltage":        230.0,
		"a_act_power":      345.0,
		"a_aprt_power":     350.0,
		"a_pf":             0.98,
		"a_freq":           50.0,
		"b_current":        1.4,
		"b_voltage":        231.0,
		"b_act_power":      323.0,
		"b_aprt_power":     330.0,
		"b_pf":             0.97,
		"b_freq":           50.0,
		"c_current":        1.6,
		"c_voltage":        229.0,
		"c_act_power":      367.0,
		"c_aprt_power":     375.0,
		"c_pf":             0.97,
		"c_freq":           50.0,
		"n_current":        nil,
		"total_current":    4.5,
		"total_act_power":  1035.0,
		"total_aprt_power": 1055.0,
	}
}

// getEM1Status returns mock EM1 component status from device state.
func (ds *DeviceServer) getEM1Status(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("em1:%d", id)
	if em1, ok := state[key].(map[string]any); ok {
		return em1
	}
	// Return default EM1 status if not in state
	return map[string]any{
		"id":         id,
		"current":    2.5,
		"voltage":    230.0,
		"act_power":  575.0,
		"aprt_power": 580.0,
		"pf":         0.99,
		"freq":       50.0,
	}
}

// getEMDataRecords returns mock EMData records.
func (ds *DeviceServer) getEMDataRecords(_ DeviceState, _ int) map[string]any {
	return map[string]any{
		"data_blocks": []map[string]any{},
	}
}

// getEMDataHistory returns mock EMData history with valid sample data.
func (ds *DeviceServer) getEMDataHistory(_ DeviceState, _ int) map[string]any {
	return map[string]any{
		"data": []map[string]any{
			{
				"ts":     1700000000,
				"period": 60,
				"values": []map[string]any{
					{
						"total_act_power": 1500.0,
						"a_act_power":     500.0,
						"b_act_power":     500.0,
						"c_act_power":     500.0,
						"a_voltage":       230.0,
						"b_voltage":       230.0,
						"c_voltage":       230.0,
					},
				},
			},
		},
	}
}

// getEM1DataRecords returns mock EM1Data records.
func (ds *DeviceServer) getEM1DataRecords(_ DeviceState, _ int) map[string]any {
	return map[string]any{
		"data_blocks": []map[string]any{},
	}
}

// getEM1DataHistory returns mock EM1Data history with valid sample data.
func (ds *DeviceServer) getEM1DataHistory(_ DeviceState, _ int) map[string]any {
	return map[string]any{
		"data": []map[string]any{
			{
				"ts":     1700000000,
				"period": 60,
				"values": []map[string]any{
					{
						"act_power": 575.0,
						"voltage":   230.0,
						"current":   2.5,
						"pf":        0.99,
						"freq":      50.0,
					},
				},
			},
		},
	}
}

// getBTHomeStatus returns BTHome component status from device state.
func (ds *DeviceServer) getBTHomeStatus(state DeviceState) map[string]any {
	if bthome, ok := state["bthome"].(map[string]any); ok {
		return bthome
	}
	return map[string]any{
		"errors": []any{},
	}
}

// getBTHomeDeviceStatus returns BTHome device status from device state.
func (ds *DeviceServer) getBTHomeDeviceStatus(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("bthomedevice:%d", id)
	if dev, ok := state[key].(map[string]any); ok {
		return dev
	}
	return map[string]any{
		"id": id,
	}
}

// getBTHomeDeviceConfig returns BTHome device config from device state.
func (ds *DeviceServer) getBTHomeDeviceConfig(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("bthomedevice:%d_config", id)
	if cfg, ok := state[key].(map[string]any); ok {
		return cfg
	}
	return map[string]any{
		"id":   id,
		"addr": "",
		"name": nil,
	}
}

// getBTHomeDeviceKnownObjects returns BTHome device known objects from device state.
func (ds *DeviceServer) getBTHomeDeviceKnownObjects(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("bthomedevice:%d_known_objects", id)
	if objs, ok := state[key].(map[string]any); ok {
		return objs
	}
	return map[string]any{
		"objects": []any{},
	}
}

// getMatterStatus returns mock Matter status.
func (ds *DeviceServer) getMatterStatus(state DeviceState) map[string]any {
	if matter, ok := state["matter"].(map[string]any); ok {
		return matter
	}
	return map[string]any{
		"commissioning_in_progress": false,
		"operational":               false,
	}
}

// getMatterConfig returns mock Matter config.
func (ds *DeviceServer) getMatterConfig(state DeviceState) map[string]any {
	if matterCfg, ok := state["matter_config"].(map[string]any); ok {
		return matterCfg
	}
	return map[string]any{
		"enable": false,
	}
}

// getMatterCommissioningCode returns Matter commissioning code from device state.
func (ds *DeviceServer) getMatterCommissioningCode(state DeviceState) map[string]any {
	if code, ok := state["matter_code"].(map[string]any); ok {
		return code
	}
	// Return default commissioning code if not in state
	return map[string]any{
		"manual_code":    "34970112332",
		"qr_code":        "MT:Y3.13WAF00KA0648G00",
		"discriminator":  float64(3840),
		"setup_pin_code": float64(20202021),
	}
}

// getScheduleList returns mock schedule list.
func (ds *DeviceServer) getScheduleList(state DeviceState, deviceName string) map[string]any {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	// Check device state for schedules
	deviceState := ds.state[deviceName]
	if deviceState == nil {
		deviceState = state
	}

	var jobs []map[string]any
	for key := range deviceState {
		if len(key) > 9 && key[:9] == "schedule:" {
			if sched, ok := deviceState[key].(map[string]any); ok {
				jobs = append(jobs, sched)
			}
		}
	}
	if len(jobs) == 0 {
		jobs = []map[string]any{}
	}
	return map[string]any{"jobs": jobs}
}

// createSchedule creates a mock schedule and returns its ID.
func (ds *DeviceServer) createSchedule(params map[string]any, deviceName string) map[string]any {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.state[deviceName] == nil {
		ds.state[deviceName] = make(DeviceState)
	}

	// Find next available ID
	maxID := 0
	for key := range ds.state[deviceName] {
		if len(key) > 9 && key[:9] == "schedule:" {
			var id int
			if _, err := fmt.Sscanf(key, "schedule:%d", &id); err == nil && id > maxID {
				maxID = id
			}
		}
	}
	newID := maxID + 1

	// Create the schedule
	enable := true
	if e, ok := params["enable"].(bool); ok {
		enable = e
	}
	timespec := ""
	if ts, ok := params["timespec"].(string); ok {
		timespec = ts
	}
	var calls []any
	if c, ok := params["calls"].([]any); ok {
		calls = c
	}

	ds.state[deviceName][fmt.Sprintf("schedule:%d", newID)] = map[string]any{
		"id":       newID,
		"enable":   enable,
		"timespec": timespec,
		"calls":    calls,
	}

	return map[string]any{"id": newID}
}

// updateSchedule updates a mock schedule.
func (ds *DeviceServer) updateSchedule(params map[string]any, deviceName string) map[string]any {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	id := 0
	if idVal, ok := params["id"].(float64); ok {
		id = int(idVal)
	}

	key := fmt.Sprintf("schedule:%d", id)
	if ds.state[deviceName] == nil {
		ds.state[deviceName] = make(DeviceState)
	}

	sched, ok := ds.state[deviceName][key].(map[string]any)
	if !ok {
		sched = map[string]any{"id": id}
	}

	if e, ok := params["enable"].(bool); ok {
		sched["enable"] = e
	}
	if ts, ok := params["timespec"].(string); ok {
		sched["timespec"] = ts
	}
	if c, ok := params["calls"].([]any); ok {
		sched["calls"] = c
	}

	ds.state[deviceName][key] = sched
	return map[string]any{"rev": 1}
}

// deleteSchedule deletes a mock schedule.
func (ds *DeviceServer) deleteSchedule(params map[string]any, deviceName string) map[string]any {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	id := 0
	if idVal, ok := params["id"].(float64); ok {
		id = int(idVal)
	}

	key := fmt.Sprintf("schedule:%d", id)
	if ds.state[deviceName] != nil {
		delete(ds.state[deviceName], key)
	}

	return map[string]any{}
}

// handleWebSocketRPC handles WebSocket RPC connections for Gen2 devices.
func (ds *DeviceServer) handleWebSocketRPC(w http.ResponseWriter, r *http.Request, _ DeviceState, _ *DeviceFixture) {
	conn, err := ds.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			// Close errors expected when client disconnects
			return
		}
	}()

	// Simple echo-style WebSocket handler for testing
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Echo the message back (basic stub for WebSocket RPC)
		if err := conn.WriteMessage(messageType, message); err != nil {
			break
		}
	}
}

// getLoRaConfig returns mock LoRa config from device state.
func (ds *DeviceServer) getLoRaConfig(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("lora:%d_config", id)
	if cfg, ok := state[key].(map[string]any); ok {
		return cfg
	}
	// Return default LoRa config matching model.LoRaConfig structure
	return map[string]any{
		"id":   float64(id),
		"freq": float64(868000000),
		"bw":   float64(7),
		"dr":   float64(7),
		"txp":  float64(14),
	}
}

// getLoRaStatus returns mock LoRa status from device state.
func (ds *DeviceServer) getLoRaStatus(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("lora:%d", id)
	if status, ok := state[key].(map[string]any); ok {
		return status
	}
	// Return default LoRa status matching model.LoRaStatus structure
	return map[string]any{
		"id":   float64(id),
		"rssi": float64(-65),
		"snr":  float64(8.5),
	}
}

// getThermostatStatus returns mock thermostat status from device state.
func (ds *DeviceServer) getThermostatStatus(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("thermostat:%d", id)
	if status, ok := state[key].(map[string]any); ok {
		return status
	}
	// Return default thermostat status
	return map[string]any{
		"id":        float64(id),
		"enable":    true,
		"target_C":  float64(21.0),
		"current_C": float64(20.5),
		"output":    false,
	}
}

// getThermostatConfig returns mock thermostat config from device state.
func (ds *DeviceServer) getThermostatConfig(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("thermostat:%d_config", id)
	if cfg, ok := state[key].(map[string]any); ok {
		return cfg
	}
	// Return default thermostat config
	return map[string]any{
		"id":              float64(id),
		"type":            "heating",
		"enable":          true,
		"target_C":        float64(21.0),
		"thermostat_mode": "auto",
	}
}

// getVirtualBooleanStatus returns mock virtual boolean status from device state.
func (ds *DeviceServer) getVirtualBooleanStatus(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("boolean:%d", id)
	if status, ok := state[key].(map[string]any); ok {
		return status
	}
	return map[string]any{
		"id":    float64(id),
		"value": false,
	}
}

// getVirtualNumberStatus returns mock virtual number status from device state.
func (ds *DeviceServer) getVirtualNumberStatus(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("number:%d", id)
	if status, ok := state[key].(map[string]any); ok {
		return status
	}
	return map[string]any{
		"id":    float64(id),
		"value": float64(0),
	}
}

// getVirtualTextStatus returns mock virtual text status from device state.
func (ds *DeviceServer) getVirtualTextStatus(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("text:%d", id)
	if status, ok := state[key].(map[string]any); ok {
		return status
	}
	return map[string]any{
		"id":    float64(id),
		"value": "",
	}
}

// getVirtualEnumStatus returns mock virtual enum status from device state.
func (ds *DeviceServer) getVirtualEnumStatus(state DeviceState, id int) map[string]any {
	key := fmt.Sprintf("enum:%d", id)
	if status, ok := state[key].(map[string]any); ok {
		return status
	}
	return map[string]any{
		"id":    float64(id),
		"value": "",
	}
}
