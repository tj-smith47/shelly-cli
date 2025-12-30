package mock

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestFixtures() *Fixtures {
	return &Fixtures{
		Config: ConfigFixture{
			Devices: []DeviceFixture{
				{Name: "Gen2 Switch", MAC: "AA:BB:CC:DD:EE:01", Model: "Shelly Plus 1PM", Type: "SNSW-001P16EU", Generation: 2},
				{Name: "Gen1 Relay", MAC: "AA:BB:CC:DD:EE:02", Model: "Shelly 1", Type: "SHSW-1", Generation: 1},
				{Name: "Gen2 Cover", MAC: "AA:BB:CC:DD:EE:03", Model: "Shelly Plus 2PM", Type: "SNSW-102P16EU", Generation: 2},
				{Name: "Gen2 Dimmer", MAC: "AA:BB:CC:DD:EE:04", Model: "Shelly Plus Dimmer", Type: "SNDM-0013US", Generation: 2},
				{Name: "Gen1 RGBW", MAC: "AA:BB:CC:DD:EE:05", Model: "Shelly RGBW2", Type: "SHRGBW2", Generation: 1},
			},
		},
		DeviceStates: map[string]DeviceState{
			"Gen2 Switch": {"switch:0": map[string]any{"output": true, "apower": 45.2}},
			"Gen1 Relay":  {"relay": map[string]any{"ison": false}},
			"Gen2 Cover":  {"cover:0": map[string]any{"state": "stopped", "current_pos": 100}},
			"Gen2 Dimmer": {"light:0": map[string]any{"output": true, "brightness": 75}},
			"Gen1 RGBW":   {"lights": []any{map[string]any{"ison": true, "brightness": 80}}},
		},
	}
}

func httpGet(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func httpPost(t *testing.T, url string, body []byte) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func closeBody(t *testing.T, resp *http.Response) {
	t.Helper()
	if err := resp.Body.Close(); err != nil {
		t.Logf("warning: close body: %v", err)
	}
}

func TestNewDeviceServer(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()
	require.NotNil(t, server)
	assert.NotEmpty(t, server.URL)
}

func TestDeviceServer_DeviceURL(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()
	assert.Contains(t, server.DeviceURL("Test"), server.URL+"/devices/Test")
}

func TestDeviceServer_Gen2_DeviceInfo(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("Gen2 Switch")+"/rpc/Shelly.GetDeviceInfo")
	defer closeBody(t, resp)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var info map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&info))
	assert.Equal(t, "Gen2 Switch", info["name"])
	assert.Equal(t, float64(2), info["gen"])
}

func TestDeviceServer_Gen2_Status(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("Gen2 Switch")+"/rpc/Shelly.GetStatus")
	defer closeBody(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeviceServer_Gen2_Config(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("Gen2 Switch")+"/rpc/Shelly.GetConfig")
	defer closeBody(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeviceServer_Gen2_Components(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("Gen2 Switch")+"/rpc/Shelly.GetComponents")
	defer closeBody(t, resp)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Contains(t, result, "components")
}

//nolint:paralleltest // Subtests share server state
func TestDeviceServer_Gen2_Switch(t *testing.T) {
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	t.Run("GetStatus", func(t *testing.T) {
		resp := httpPost(t, server.DeviceURL("Gen2 Switch")+"/rpc/Switch.GetStatus", []byte(`{"id":0}`))
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Set", func(t *testing.T) {
		resp := httpPost(t, server.DeviceURL("Gen2 Switch")+"/rpc/Switch.Set", []byte(`{"id":0,"on":false}`))
		defer closeBody(t, resp)

		var result map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		assert.Equal(t, true, result["was_on"])
	})

	t.Run("Toggle", func(t *testing.T) {
		resp := httpPost(t, server.DeviceURL("Gen2 Switch")+"/rpc/Switch.Toggle", []byte(`{"id":0}`))
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Set invalid body", func(t *testing.T) {
		resp := httpPost(t, server.DeviceURL("Gen2 Switch")+"/rpc/Switch.Set", []byte(`invalid`))
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Unknown method", func(t *testing.T) {
		resp := httpGet(t, server.DeviceURL("Gen2 Switch")+"/rpc/Switch.Unknown")
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

//nolint:paralleltest // Subtests share server state
func TestDeviceServer_Gen2_Cover(t *testing.T) {
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	t.Run("GetStatus", func(t *testing.T) {
		resp := httpPost(t, server.DeviceURL("Gen2 Cover")+"/rpc/Cover.GetStatus", []byte(`{"id":0}`))
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Open", func(t *testing.T) {
		resp := httpGet(t, server.DeviceURL("Gen2 Cover")+"/rpc/Cover.Open")
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Close", func(t *testing.T) {
		resp := httpGet(t, server.DeviceURL("Gen2 Cover")+"/rpc/Cover.Close")
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Stop", func(t *testing.T) {
		resp := httpGet(t, server.DeviceURL("Gen2 Cover")+"/rpc/Cover.Stop")
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Unknown", func(t *testing.T) {
		resp := httpGet(t, server.DeviceURL("Gen2 Cover")+"/rpc/Cover.Unknown")
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

//nolint:paralleltest // Subtests share server state
func TestDeviceServer_Gen2_Light(t *testing.T) {
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	t.Run("GetStatus", func(t *testing.T) {
		resp := httpPost(t, server.DeviceURL("Gen2 Dimmer")+"/rpc/Light.GetStatus", []byte(`{"id":0}`))
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Set", func(t *testing.T) {
		resp := httpPost(t, server.DeviceURL("Gen2 Dimmer")+"/rpc/Light.Set", []byte(`{"id":0,"on":true,"brightness":50}`))
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Set invalid", func(t *testing.T) {
		resp := httpPost(t, server.DeviceURL("Gen2 Dimmer")+"/rpc/Light.Set", []byte(`invalid`))
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Unknown", func(t *testing.T) {
		resp := httpGet(t, server.DeviceURL("Gen2 Dimmer")+"/rpc/Light.Unknown")
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestDeviceServer_Gen1_Shelly(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("Gen1 Relay")+"/shelly")
	defer closeBody(t, resp)

	var info map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&info))
	assert.Equal(t, "SHSW-1", info["type"])
}

func TestDeviceServer_Gen1_Status(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("Gen1 Relay")+"/status")
	defer closeBody(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeviceServer_Gen1_Settings(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("Gen1 Relay")+"/settings")
	defer closeBody(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

//nolint:paralleltest // Subtests share server state
func TestDeviceServer_Gen1_Relay(t *testing.T) {
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	t.Run("get", func(t *testing.T) {
		resp := httpGet(t, server.DeviceURL("Gen1 Relay")+"/relay/0")
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("turn on", func(t *testing.T) {
		resp := httpGet(t, server.DeviceURL("Gen1 Relay")+"/relay/0?turn=on")
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("turn off", func(t *testing.T) {
		resp := httpGet(t, server.DeviceURL("Gen1 Relay")+"/relay/0?turn=off")
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestDeviceServer_Gen1_Light(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("Gen1 RGBW")+"/light/0")
	defer closeBody(t, resp)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, true, result["ison"])
}

//nolint:paralleltest // Subtests share server state
func TestDeviceServer_NotFound(t *testing.T) {
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	t.Run("unknown device", func(t *testing.T) {
		resp := httpGet(t, server.DeviceURL("Unknown")+"/rpc/Shelly.GetDeviceInfo")
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("invalid path", func(t *testing.T) {
		resp := httpGet(t, server.URL+"/invalid")
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("gen2 unknown", func(t *testing.T) {
		resp := httpGet(t, server.DeviceURL("Gen2 Switch")+"/rpc/Unknown.Method")
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("gen1 unknown", func(t *testing.T) {
		resp := httpGet(t, server.DeviceURL("Gen1 Relay")+"/unknown")
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestDeviceServer_GetState(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	state := server.GetState("Gen2 Switch")
	require.NotNil(t, state)
	assert.Contains(t, state, "switch:0")
}

func TestDeviceServer_StateModification(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	state := server.GetState("Gen2 Switch")
	sw, ok := state["switch:0"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, sw["output"])

	resp := httpPost(t, server.DeviceURL("Gen2 Switch")+"/rpc/Switch.Set", []byte(`{"id":0,"on":false}`))
	_, err := io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)
	closeBody(t, resp)

	state2 := server.GetState("Gen2 Switch")
	sw2, ok := state2["switch:0"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, false, sw2["output"])
}

func TestDeviceServer_NoState(t *testing.T) {
	t.Parallel()
	fixtures := &Fixtures{
		Config:       ConfigFixture{Devices: []DeviceFixture{{Name: "NoState", Generation: 2}}},
		DeviceStates: map[string]DeviceState{},
	}
	server := NewDeviceServer(fixtures)
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("NoState")+"/rpc/Shelly.GetStatus")
	defer closeBody(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeviceServer_Gen1_RelayNoState(t *testing.T) {
	t.Parallel()
	fixtures := &Fixtures{
		Config:       ConfigFixture{Devices: []DeviceFixture{{Name: "NoState", Generation: 1}}},
		DeviceStates: map[string]DeviceState{},
	}
	server := NewDeviceServer(fixtures)
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("NoState")+"/relay/0?turn=on")
	defer closeBody(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeviceServer_Gen1_LightNoState(t *testing.T) {
	t.Parallel()
	fixtures := &Fixtures{
		Config:       ConfigFixture{Devices: []DeviceFixture{{Name: "NoLight", Generation: 1}}},
		DeviceStates: map[string]DeviceState{"NoLight": {}},
	}
	server := NewDeviceServer(fixtures)
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("NoLight")+"/light/0")
	defer closeBody(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeviceServer_ComponentState_Empty(t *testing.T) {
	t.Parallel()
	fixtures := &Fixtures{
		Config:       ConfigFixture{Devices: []DeviceFixture{{Name: "Empty", Generation: 2}}},
		DeviceStates: map[string]DeviceState{"Empty": {}},
	}
	server := NewDeviceServer(fixtures)
	defer server.Close()

	resp := httpPost(t, server.DeviceURL("Empty")+"/rpc/Switch.GetStatus", []byte(`{"id":0}`))
	defer closeBody(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeviceServer_CaseInsensitiveDeviceName(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("gen2 switch")+"/rpc/Shelly.GetDeviceInfo")
	defer closeBody(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeviceServer_SwitchSetNewDevice(t *testing.T) {
	t.Parallel()
	fixtures := &Fixtures{
		Config:       ConfigFixture{Devices: []DeviceFixture{{Name: "New", Generation: 2}}},
		DeviceStates: map[string]DeviceState{},
	}
	server := NewDeviceServer(fixtures)
	defer server.Close()

	resp := httpPost(t, server.DeviceURL("New")+"/rpc/Switch.Set", []byte(`{"id":0,"on":true}`))
	defer closeBody(t, resp)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, false, result["was_on"])
}

func TestDeviceServer_ToggleNewDevice(t *testing.T) {
	t.Parallel()
	fixtures := &Fixtures{
		Config:       ConfigFixture{Devices: []DeviceFixture{{Name: "New", Generation: 2}}},
		DeviceStates: map[string]DeviceState{},
	}
	server := NewDeviceServer(fixtures)
	defer server.Close()

	resp := httpPost(t, server.DeviceURL("New")+"/rpc/Switch.Toggle", []byte(`{"id":0}`))
	defer closeBody(t, resp)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, false, result["was_on"])
}

func TestDeviceServer_LightSetNewDevice(t *testing.T) {
	t.Parallel()
	fixtures := &Fixtures{
		Config:       ConfigFixture{Devices: []DeviceFixture{{Name: "New", Generation: 2}}},
		DeviceStates: map[string]DeviceState{},
	}
	server := NewDeviceServer(fixtures)
	defer server.Close()

	resp := httpPost(t, server.DeviceURL("New")+"/rpc/Light.Set", []byte(`{"id":0,"on":true,"brightness":50}`))
	defer closeBody(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeviceServer_Gen1_RelayExistingState(t *testing.T) {
	t.Parallel()
	fixtures := &Fixtures{
		Config:       ConfigFixture{Devices: []DeviceFixture{{Name: "Relay", Generation: 1}}},
		DeviceStates: map[string]DeviceState{"Relay": {"relay": map[string]any{"ison": true}}},
	}
	server := NewDeviceServer(fixtures)
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("Relay")+"/relay/0?turn=off")
	defer closeBody(t, resp)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, false, result["ison"])
}

func TestDeviceServer_DeviceWithNoEndpoint(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	resp := httpGet(t, server.URL+"/devices/Gen2 Switch")
	defer closeBody(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeviceServer_ParseIDWithValidJSON(t *testing.T) {
	t.Parallel()
	server := NewDeviceServer(newTestFixtures())
	defer server.Close()

	resp := httpPost(t, server.DeviceURL("Gen2 Switch")+"/rpc/Switch.GetStatus", []byte(`{"id":1}`))
	defer closeBody(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeviceServer_Gen1_RelayGetExistingState(t *testing.T) {
	t.Parallel()
	fixtures := &Fixtures{
		Config:       ConfigFixture{Devices: []DeviceFixture{{Name: "Relay", Generation: 1}}},
		DeviceStates: map[string]DeviceState{"Relay": {"relay": map[string]any{"ison": true}}},
	}
	server := NewDeviceServer(fixtures)
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("Relay")+"/relay/0")
	defer closeBody(t, resp)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, true, result["ison"])
}

func TestDeviceServer_SpacesInName(t *testing.T) {
	t.Parallel()
	fixtures := &Fixtures{
		Config: ConfigFixture{
			Devices: []DeviceFixture{
				{Name: "Living Room Light", MAC: "AA:BB:CC:DD:EE:01", Model: "Shelly Plus 1PM", Type: "SNSW-001P16EU", Generation: 2},
			},
		},
		DeviceStates: map[string]DeviceState{
			"Living Room Light": {"switch:0": map[string]any{"output": true}},
		},
	}
	server := NewDeviceServer(fixtures)
	defer server.Close()

	resp := httpGet(t, server.DeviceURL("Living Room Light")+"/rpc/Shelly.GetDeviceInfo")
	defer closeBody(t, resp)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var info map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&info))
	assert.Equal(t, "Living Room Light", info["name"])
}

//nolint:paralleltest // Subtests share server state
func TestDeviceServer_JSONRPC(t *testing.T) {
	fixtures := &Fixtures{
		Config: ConfigFixture{
			Devices: []DeviceFixture{
				{Name: "Test Device", MAC: "AA:BB:CC:DD:EE:01", Model: "Shelly Plus 1PM", Type: "SNSW-001P16EU", Generation: 2},
			},
		},
		DeviceStates: map[string]DeviceState{
			"Test Device": {"switch:0": map[string]any{"output": true}, "light:0": map[string]any{"output": false, "brightness": 50}},
		},
	}
	server := NewDeviceServer(fixtures)
	defer server.Close()

	t.Run("GetDeviceInfo", func(t *testing.T) {
		body := []byte(`{"id":1,"method":"Shelly.GetDeviceInfo"}`)
		resp := httpPost(t, server.DeviceURL("Test Device")+"/rpc", body)
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var result map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		assert.Equal(t, float64(1), result["id"])
		info, ok := result["result"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "Test Device", info["name"])
	})

	t.Run("GetStatus", func(t *testing.T) {
		body := []byte(`{"id":2,"method":"Shelly.GetStatus"}`)
		resp := httpPost(t, server.DeviceURL("Test Device")+"/rpc", body)
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetConfig", func(t *testing.T) {
		body := []byte(`{"id":3,"method":"Shelly.GetConfig"}`)
		resp := httpPost(t, server.DeviceURL("Test Device")+"/rpc", body)
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetComponents", func(t *testing.T) {
		body := []byte(`{"id":4,"method":"Shelly.GetComponents"}`)
		resp := httpPost(t, server.DeviceURL("Test Device")+"/rpc", body)
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Switch.GetStatus", func(t *testing.T) {
		body := []byte(`{"id":5,"method":"Switch.GetStatus","params":{"id":0}}`)
		resp := httpPost(t, server.DeviceURL("Test Device")+"/rpc", body)
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Switch.Set", func(t *testing.T) {
		body := []byte(`{"id":6,"method":"Switch.Set","params":{"id":0,"on":false}}`)
		resp := httpPost(t, server.DeviceURL("Test Device")+"/rpc", body)
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Switch.Toggle", func(t *testing.T) {
		body := []byte(`{"id":7,"method":"Switch.Toggle","params":{"id":0}}`)
		resp := httpPost(t, server.DeviceURL("Test Device")+"/rpc", body)
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Cover.GetStatus", func(t *testing.T) {
		body := []byte(`{"id":8,"method":"Cover.GetStatus","params":{"id":0}}`)
		resp := httpPost(t, server.DeviceURL("Test Device")+"/rpc", body)
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Cover.Open", func(t *testing.T) {
		body := []byte(`{"id":9,"method":"Cover.Open","params":{"id":0}}`)
		resp := httpPost(t, server.DeviceURL("Test Device")+"/rpc", body)
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Light.GetStatus", func(t *testing.T) {
		body := []byte(`{"id":10,"method":"Light.GetStatus","params":{"id":0}}`)
		resp := httpPost(t, server.DeviceURL("Test Device")+"/rpc", body)
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Light.Set", func(t *testing.T) {
		body := []byte(`{"id":11,"method":"Light.Set","params":{"id":0,"on":true,"brightness":75}}`)
		resp := httpPost(t, server.DeviceURL("Test Device")+"/rpc", body)
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Unknown method", func(t *testing.T) {
		body := []byte(`{"id":12,"method":"Unknown.Method"}`)
		resp := httpPost(t, server.DeviceURL("Test Device")+"/rpc", body)
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		resp := httpPost(t, server.DeviceURL("Test Device")+"/rpc", []byte(`invalid`))
		defer closeBody(t, resp)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
