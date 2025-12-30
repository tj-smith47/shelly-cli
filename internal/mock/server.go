package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// DeviceServer mocks Shelly device HTTP endpoints.
type DeviceServer struct {
	*httptest.Server
	fixtures *Fixtures
	mu       sync.RWMutex
	state    map[string]DeviceState
}

// NewDeviceServer creates a mock HTTP server for device requests.
func NewDeviceServer(fixtures *Fixtures) *DeviceServer {
	ds := &DeviceServer{
		fixtures: fixtures,
		state:    make(map[string]DeviceState),
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

	case "Wifi.GetStatus":
		// Return mock WiFi status
		result = map[string]any{
			"sta_ip":          "192.168.1.100",
			"status":          "got ip",
			"ssid":            "HomeNetwork",
			"rssi":            float64(-45),
			"ap_client_count": float64(0),
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
	actions := ds.state[device.Name]["actions"].(map[string]any)
	actions[actionKey] = map[string]any{
		"index":   index,
		"name":    name,
		"enabled": enabled == "true",
	}
	ds.mu.Unlock()

	// Return success response (Gen1 actions endpoint returns the action settings)
	ds.writeJSON(w, map[string]any{
		"actions": []map[string]any{
			{
				"index":   index,
				"name":    name,
				"enabled": enabled == "true",
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
