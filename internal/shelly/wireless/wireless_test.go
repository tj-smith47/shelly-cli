// Package wireless provides wireless protocol operations for Shelly devices.
package wireless

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// mockParent is a test double for Parent interface.
type mockParent struct {
	withConnectionFn func(ctx context.Context, identifier string, fn func(*client.Client) error) error
	rawRPCFn         func(ctx context.Context, identifier, method string, params map[string]any) (any, error)
}

func (m *mockParent) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	if m.withConnectionFn != nil {
		return m.withConnectionFn(ctx, identifier, fn)
	}
	return nil
}

// errRawRPCNotMocked is a sentinel error for unmocked RawRPC calls.
var errRawRPCNotMocked = errors.New("RawRPC not mocked")

func (m *mockParent) RawRPC(ctx context.Context, identifier, method string, params map[string]any) (any, error) {
	if m.rawRPCFn != nil {
		return m.rawRPCFn(ctx, identifier, method, params)
	}
	return nil, errRawRPCNotMocked
}

func TestNew(t *testing.T) {
	t.Parallel()

	parent := &mockParent{}
	svc := New(parent)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.parent != parent {
		t.Error("expected parent to be set")
	}
}

func TestIsBLENotSupportedError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "BLE not supported sentinel",
			err:  discovery.ErrBLENotSupported,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsBLENotSupportedError(tt.err)
			if got != tt.want {
				t.Errorf("IsBLENotSupportedError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBTHomeDeviceStatus_Fields(t *testing.T) {
	t.Parallel()

	rssi := -55
	battery := 75

	status := BTHomeDeviceStatus{
		ID:         1,
		RSSI:       &rssi,
		Battery:    &battery,
		LastUpdate: 1700000000.5,
	}

	if status.ID != 1 {
		t.Errorf("got ID=%d, want 1", status.ID)
	}
	if status.RSSI == nil || *status.RSSI != rssi {
		t.Errorf("got RSSI=%v, want %d", status.RSSI, rssi)
	}
	if status.Battery == nil || *status.Battery != battery {
		t.Errorf("got Battery=%v, want %d", status.Battery, battery)
	}
	if status.LastUpdate != 1700000000.5 {
		t.Errorf("got LastUpdate=%f, want 1700000000.5", status.LastUpdate)
	}
}

func TestBTHomeSensorStatus_Fields(t *testing.T) {
	t.Parallel()

	status := BTHomeSensorStatus{
		ID:           2,
		Value:        23.5,
		LastUpdateTS: 1700000001.0,
	}

	if status.ID != 2 {
		t.Errorf("got ID=%d, want 2", status.ID)
	}
	if status.Value != 23.5 {
		t.Errorf("got Value=%v, want 23.5", status.Value)
	}
	if status.LastUpdateTS != 1700000001.0 {
		t.Errorf("got LastUpdateTS=%f, want 1700000001.0", status.LastUpdateTS)
	}
}

func TestBTHomeAddDeviceResult_Fields(t *testing.T) {
	t.Parallel()

	result := BTHomeAddDeviceResult{
		Key: "bthomedevice:100",
	}

	if result.Key != "bthomedevice:100" {
		t.Errorf("got Key=%q, want %q", result.Key, "bthomedevice:100")
	}
}

func TestCollectBTHomeDevices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    map[string]json.RawMessage
		wantCount int
	}{
		{
			name:      "empty status",
			status:    map[string]json.RawMessage{},
			wantCount: 0,
		},
		{
			name: "no bthomedevice keys",
			status: map[string]json.RawMessage{
				"switch:0": json.RawMessage(`{"output": true}`),
				"input:0":  json.RawMessage(`{"state": false}`),
			},
			wantCount: 0,
		},
		{
			name: "with bthomedevice keys",
			status: map[string]json.RawMessage{
				"bthomedevice:0": json.RawMessage(`{"id": 0, "rssi": -55}`),
				"bthomedevice:1": json.RawMessage(`{"id": 1, "rssi": -60}`),
				"switch:0":       json.RawMessage(`{"output": true}`),
			},
			wantCount: 2,
		},
		{
			name: "invalid bthomedevice JSON",
			status: map[string]json.RawMessage{
				"bthomedevice:0": json.RawMessage(`{invalid json`),
				"bthomedevice:1": json.RawMessage(`{"id": 1}`),
			},
			wantCount: 1, // Only valid one should be collected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			devices := CollectBTHomeDevices(tt.status, nil)

			if len(devices) != tt.wantCount {
				t.Errorf("got %d devices, want %d", len(devices), tt.wantCount)
			}
		})
	}
}

func TestCollectBTHomeSensors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    map[string]json.RawMessage
		wantCount int
	}{
		{
			name:      "empty status",
			status:    map[string]json.RawMessage{},
			wantCount: 0,
		},
		{
			name: "no bthomesensor keys",
			status: map[string]json.RawMessage{
				"switch:0":       json.RawMessage(`{"output": true}`),
				"bthomedevice:0": json.RawMessage(`{"id": 0}`),
			},
			wantCount: 0,
		},
		{
			name: "with bthomesensor keys",
			status: map[string]json.RawMessage{
				"bthomesensor:0": json.RawMessage(`{"id": 0, "rssi": -55}`),
				"bthomesensor:1": json.RawMessage(`{"id": 1, "rssi": -60}`),
				"switch:0":       json.RawMessage(`{"output": true}`),
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sensors := CollectBTHomeSensors(tt.status, nil)

			if len(sensors) != tt.wantCount {
				t.Errorf("got %d sensors, want %d", len(sensors), tt.wantCount)
			}
		})
	}
}

func TestMatterConfig_Fields(t *testing.T) {
	t.Parallel()

	cfg := MatterConfig{Enable: true}

	if !cfg.Enable {
		t.Error("expected Enable to be true")
	}
}

func TestParseLoRaConfig(t *testing.T) {
	t.Parallel()

	cfgMap := map[string]any{
		"id":   float64(0),
		"freq": float64(868000000),
		"bw":   float64(125),
		"dr":   float64(7),
		"txp":  float64(14),
	}

	cfg := parseLoRaConfig(cfgMap)

	if cfg.ID != 0 {
		t.Errorf("got ID=%d, want 0", cfg.ID)
	}
	if cfg.Freq != 868000000 {
		t.Errorf("got Freq=%d, want 868000000", cfg.Freq)
	}
	if cfg.BW != 125 {
		t.Errorf("got BW=%d, want 125", cfg.BW)
	}
	if cfg.DR != 7 {
		t.Errorf("got DR=%d, want 7", cfg.DR)
	}
	if cfg.TxP != 14 {
		t.Errorf("got TxP=%d, want 14", cfg.TxP)
	}
}

func TestParseLoRaStatus(t *testing.T) {
	t.Parallel()

	statusMap := map[string]any{
		"id":   float64(0),
		"rssi": float64(-80),
		"snr":  float64(7.5),
	}

	status := parseLoRaStatus(statusMap)

	if status.ID != 0 {
		t.Errorf("got ID=%d, want 0", status.ID)
	}
	if status.RSSI != -80 {
		t.Errorf("got RSSI=%d, want -80", status.RSSI)
	}
	if status.SNR != 7.5 {
		t.Errorf("got SNR=%f, want 7.5", status.SNR)
	}
}

func TestZigbeeDevice_Fields(t *testing.T) {
	t.Parallel()

	device := model.ZigbeeDevice{
		Name:         "Living Room Light",
		Address:      "192.168.1.101",
		Model:        "SNSW-002P16EU",
		Enabled:      true,
		NetworkState: "router",
		EUI64:        "00:11:22:33:44:55:66:77",
	}

	if device.Name != "Living Room Light" {
		t.Errorf("got Name=%q, want %q", device.Name, "Living Room Light")
	}
	if device.Address != "192.168.1.101" {
		t.Errorf("got Address=%q, want %q", device.Address, "192.168.1.101")
	}
	if !device.Enabled {
		t.Error("expected Enabled to be true")
	}
	if device.NetworkState != "router" {
		t.Errorf("got NetworkState=%q, want %q", device.NetworkState, "router")
	}
	if device.EUI64 != "00:11:22:33:44:55:66:77" {
		t.Errorf("got EUI64=%q, want %q", device.EUI64, "00:11:22:33:44:55:66:77")
	}
}

func TestZigbeeStatus_Fields(t *testing.T) {
	t.Parallel()

	status := model.ZigbeeStatus{
		Enabled:          true,
		NetworkState:     "router",
		EUI64:            "00:11:22:33:44:55:66:77",
		PANID:            0x1234,
		Channel:          15,
		CoordinatorEUI64: "00:11:22:33:44:55:66:88",
	}

	if !status.Enabled {
		t.Error("expected Enabled to be true")
	}
	if status.NetworkState != "router" {
		t.Errorf("got NetworkState=%q, want %q", status.NetworkState, "router")
	}
	if status.PANID != 0x1234 {
		t.Errorf("got PANID=%d, want %d", status.PANID, 0x1234)
	}
	if status.Channel != 15 {
		t.Errorf("got Channel=%d, want 15", status.Channel)
	}
}

func TestMatterStatus_Fields(t *testing.T) {
	t.Parallel()

	status := model.MatterStatus{
		Enabled:        true,
		Commissionable: true,
		FabricsCount:   2,
	}

	if !status.Enabled {
		t.Error("expected Enabled to be true")
	}
	if !status.Commissionable {
		t.Error("expected Commissionable to be true")
	}
	if status.FabricsCount != 2 {
		t.Errorf("got FabricsCount=%d, want 2", status.FabricsCount)
	}
}

func TestCommissioningInfo_Fields(t *testing.T) {
	t.Parallel()

	info := model.CommissioningInfo{
		ManualCode:    "12345-67890",
		QRCode:        "MT:ABCDEFG",
		Discriminator: 1234,
		SetupPINCode:  123456,
		Available:     true,
	}

	if info.ManualCode != "12345-67890" {
		t.Errorf("got ManualCode=%q, want %q", info.ManualCode, "12345-67890")
	}
	if info.QRCode != "MT:ABCDEFG" {
		t.Errorf("got QRCode=%q, want %q", info.QRCode, "MT:ABCDEFG")
	}
	if info.Discriminator != 1234 {
		t.Errorf("got Discriminator=%d, want 1234", info.Discriminator)
	}
	if info.SetupPINCode != 123456 {
		t.Errorf("got SetupPINCode=%d, want 123456", info.SetupPINCode)
	}
	if !info.Available {
		t.Error("expected Available to be true")
	}
}

func TestLoRaFullStatus_Fields(t *testing.T) {
	t.Parallel()

	full := model.LoRaFullStatus{
		Config: &model.LoRaConfig{
			ID:   0,
			Freq: 868000000,
			BW:   125,
			DR:   7,
			TxP:  14,
		},
		Status: &model.LoRaStatus{
			ID:   0,
			RSSI: -80,
			SNR:  7.5,
		},
	}

	if full.Config == nil {
		t.Fatal("expected non-nil Config")
	}
	if full.Config.Freq != 868000000 {
		t.Errorf("got Config.Freq=%d, want 868000000", full.Config.Freq)
	}
	if full.Status == nil {
		t.Fatal("expected non-nil Status")
	}
	if full.Status.RSSI != -80 {
		t.Errorf("got Status.RSSI=%d, want -80", full.Status.RSSI)
	}
}

func TestBTHomeDeviceInfo_Fields(t *testing.T) {
	t.Parallel()

	rssi := -55
	battery := 80

	info := model.BTHomeDeviceInfo{
		ID:         1,
		Name:       "Temperature Sensor",
		Addr:       "AA:BB:CC:DD:EE:FF",
		RSSI:       &rssi,
		Battery:    &battery,
		LastUpdate: 1700000000.5,
	}

	if info.ID != 1 {
		t.Errorf("got ID=%d, want 1", info.ID)
	}
	if info.Name != "Temperature Sensor" {
		t.Errorf("got Name=%q, want %q", info.Name, "Temperature Sensor")
	}
	if info.Addr != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("got Addr=%q, want %q", info.Addr, "AA:BB:CC:DD:EE:FF")
	}
	if info.RSSI == nil || *info.RSSI != rssi {
		t.Errorf("got RSSI=%v, want %d", info.RSSI, rssi)
	}
	if info.Battery == nil || *info.Battery != battery {
		t.Errorf("got Battery=%v, want %d", info.Battery, battery)
	}
}

func TestBTHomeComponentStatus_Fields(t *testing.T) {
	t.Parallel()

	status := model.BTHomeComponentStatus{
		Discovery: &model.BTHomeDiscoveryStatus{
			StartedAt: 1700000000,
			Duration:  30,
		},
	}

	if status.Discovery == nil {
		t.Fatal("expected non-nil Discovery")
	}
	if status.Discovery.StartedAt != 1700000000 {
		t.Errorf("got Discovery.StartedAt=%f, want 1700000000", status.Discovery.StartedAt)
	}
	if status.Discovery.Duration != 30 {
		t.Errorf("got Discovery.Duration=%d, want 30", status.Discovery.Duration)
	}
}

// ----- Connection Error Tests -----

func TestZigbeeEnable_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	err := svc.ZigbeeEnable(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestZigbeeDisable_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	err := svc.ZigbeeDisable(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestZigbeeStartNetworkSteering_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	err := svc.ZigbeeStartNetworkSteering(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestZigbeeGetStatus_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	result, err := svc.ZigbeeGetStatus(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestZigbeeGetConfig_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	result, err := svc.ZigbeeGetConfig(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestBTHomeAddDevice_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	result, err := svc.BTHomeAddDevice(context.Background(), "test-device", "AA:BB:CC:DD:EE:FF", "Test Sensor")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result.Key != "" {
		t.Errorf("expected empty Key, got %q", result.Key)
	}
}

func TestBTHomeStartDiscovery_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	err := svc.BTHomeStartDiscovery(context.Background(), "test-device", 30)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestBTHomeRemoveDevice_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	err := svc.BTHomeRemoveDevice(context.Background(), "test-device", 1)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestFetchBTHomeDevices_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	result, err := svc.FetchBTHomeDevices(context.Background(), "test-device", nil)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestFetchBTHomeComponentStatus_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	_, err := svc.FetchBTHomeComponentStatus(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestFetchBTHomeDeviceStatus_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	_, err := svc.FetchBTHomeDeviceStatus(context.Background(), "test-device", 1)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestLoRaSendBytes_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	err := svc.LoRaSendBytes(context.Background(), "test-device", 0, "test-data")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestLoRaSetConfig_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	err := svc.LoRaSetConfig(context.Background(), "test-device", 0, map[string]any{"freq": 868000000})

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestLoRaGetConfig_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	result, err := svc.LoRaGetConfig(context.Background(), "test-device", 0)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestLoRaGetStatus_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	result, err := svc.LoRaGetStatus(context.Background(), "test-device", 0)

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestMatterEnable_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	err := svc.MatterEnable(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestMatterDisable_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	err := svc.MatterDisable(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestMatterReset_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	err := svc.MatterReset(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestMatterGetSetupCode_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	result, err := svc.MatterGetSetupCode(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}

func TestMatterGetStatus_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	result, err := svc.MatterGetStatus(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestMatterGetConfig_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	result, err := svc.MatterGetConfig(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result.Enable {
		t.Error("expected Enable to be false")
	}
}

func TestMatterGetCommissioningCode_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	result, err := svc.MatterGetCommissioningCode(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result.Available {
		t.Error("expected Available to be false")
	}
}

func TestMatterIsCommissionable_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := New(parent)
	result, err := svc.MatterIsCommissionable(context.Background(), "test-device")

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if result {
		t.Error("expected result to be false")
	}
}

// ----- Identifier Passthrough Tests -----

func TestZigbeeEnable_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := New(parent)
	//nolint:errcheck // Intentionally ignoring error to test identifier passthrough
	_ = svc.ZigbeeEnable(context.Background(), "zigbee-device")

	if capturedIdentifier != "zigbee-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "zigbee-device")
	}
}

func TestZigbeeDisable_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := New(parent)
	//nolint:errcheck // Intentionally ignoring error to test identifier passthrough
	_ = svc.ZigbeeDisable(context.Background(), "zigbee-device-2")

	if capturedIdentifier != "zigbee-device-2" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "zigbee-device-2")
	}
}

func TestBTHomeAddDevice_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := New(parent)
	//nolint:errcheck // Intentionally ignoring error to test identifier passthrough
	_, _ = svc.BTHomeAddDevice(context.Background(), "bthome-gateway", "AA:BB:CC:DD:EE:FF", "Sensor")

	if capturedIdentifier != "bthome-gateway" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "bthome-gateway")
	}
}

func TestBTHomeRemoveDevice_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := New(parent)
	//nolint:errcheck // Intentionally ignoring error to test identifier passthrough
	_ = svc.BTHomeRemoveDevice(context.Background(), "bthome-remove-device", 1)

	if capturedIdentifier != "bthome-remove-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "bthome-remove-device")
	}
}

func TestLoRaSendBytes_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := New(parent)
	//nolint:errcheck // Intentionally ignoring error to test identifier passthrough
	_ = svc.LoRaSendBytes(context.Background(), "lora-device", 0, "test-data")

	if capturedIdentifier != "lora-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "lora-device")
	}
}

func TestMatterEnable_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := New(parent)
	//nolint:errcheck // Intentionally ignoring error to test identifier passthrough
	_ = svc.MatterEnable(context.Background(), "matter-device")

	if capturedIdentifier != "matter-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "matter-device")
	}
}

func TestMatterReset_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	parent := &mockParent{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := New(parent)
	//nolint:errcheck // Intentionally ignoring error to test identifier passthrough
	_ = svc.MatterReset(context.Background(), "matter-reset-device")

	if capturedIdentifier != "matter-reset-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "matter-reset-device")
	}
}

// ----- Zero Value Tests -----

func TestBTHomeDeviceStatus_ZeroValue(t *testing.T) {
	t.Parallel()

	var status BTHomeDeviceStatus

	if status.ID != 0 {
		t.Errorf("got ID=%d, want 0", status.ID)
	}
	if status.RSSI != nil {
		t.Error("expected RSSI to be nil")
	}
	if status.Battery != nil {
		t.Error("expected Battery to be nil")
	}
	if status.LastUpdate != 0 {
		t.Errorf("got LastUpdate=%f, want 0", status.LastUpdate)
	}
}

func TestBTHomeSensorStatus_ZeroValue(t *testing.T) {
	t.Parallel()

	var status BTHomeSensorStatus

	if status.ID != 0 {
		t.Errorf("got ID=%d, want 0", status.ID)
	}
	if status.Value != nil {
		t.Error("expected Value to be nil")
	}
	if status.LastUpdateTS != 0 {
		t.Errorf("got LastUpdateTS=%f, want 0", status.LastUpdateTS)
	}
}

func TestBTHomeAddDeviceResult_ZeroValue(t *testing.T) {
	t.Parallel()

	var result BTHomeAddDeviceResult

	if result.Key != "" {
		t.Errorf("got Key=%q, want empty", result.Key)
	}
}

func TestMatterConfig_ZeroValue(t *testing.T) {
	t.Parallel()

	var cfg MatterConfig

	if cfg.Enable {
		t.Error("expected Enable to be false")
	}
}

func TestZigbeeDevice_ZeroValue(t *testing.T) {
	t.Parallel()

	var device model.ZigbeeDevice

	if device.Name != "" {
		t.Errorf("got Name=%q, want empty", device.Name)
	}
	if device.Address != "" {
		t.Errorf("got Address=%q, want empty", device.Address)
	}
	if device.Enabled {
		t.Error("expected Enabled to be false")
	}
	if device.NetworkState != "" {
		t.Errorf("got NetworkState=%q, want empty", device.NetworkState)
	}
	if device.EUI64 != "" {
		t.Errorf("got EUI64=%q, want empty", device.EUI64)
	}
}

func TestZigbeeStatus_ZeroValue(t *testing.T) {
	t.Parallel()

	var status model.ZigbeeStatus

	if status.Enabled {
		t.Error("expected Enabled to be false")
	}
	if status.NetworkState != "" {
		t.Errorf("got NetworkState=%q, want empty", status.NetworkState)
	}
	if status.PANID != 0 {
		t.Errorf("got PANID=%d, want 0", status.PANID)
	}
	if status.Channel != 0 {
		t.Errorf("got Channel=%d, want 0", status.Channel)
	}
}

func TestMatterStatus_ZeroValue(t *testing.T) {
	t.Parallel()

	var status model.MatterStatus

	if status.Enabled {
		t.Error("expected Enabled to be false")
	}
	if status.Commissionable {
		t.Error("expected Commissionable to be false")
	}
	if status.FabricsCount != 0 {
		t.Errorf("got FabricsCount=%d, want 0", status.FabricsCount)
	}
}

func TestCommissioningInfo_ZeroValue(t *testing.T) {
	t.Parallel()

	var info model.CommissioningInfo

	if info.ManualCode != "" {
		t.Errorf("got ManualCode=%q, want empty", info.ManualCode)
	}
	if info.QRCode != "" {
		t.Errorf("got QRCode=%q, want empty", info.QRCode)
	}
	if info.Discriminator != 0 {
		t.Errorf("got Discriminator=%d, want 0", info.Discriminator)
	}
	if info.SetupPINCode != 0 {
		t.Errorf("got SetupPINCode=%d, want 0", info.SetupPINCode)
	}
	if info.Available {
		t.Error("expected Available to be false")
	}
}

func TestLoRaFullStatus_ZeroValue(t *testing.T) {
	t.Parallel()

	var full model.LoRaFullStatus

	if full.Config != nil {
		t.Error("expected Config to be nil")
	}
	if full.Status != nil {
		t.Error("expected Status to be nil")
	}
}

func TestBTHomeDeviceInfo_ZeroValue(t *testing.T) {
	t.Parallel()

	var info model.BTHomeDeviceInfo

	if info.ID != 0 {
		t.Errorf("got ID=%d, want 0", info.ID)
	}
	if info.Name != "" {
		t.Errorf("got Name=%q, want empty", info.Name)
	}
	if info.Addr != "" {
		t.Errorf("got Addr=%q, want empty", info.Addr)
	}
	if info.RSSI != nil {
		t.Error("expected RSSI to be nil")
	}
	if info.Battery != nil {
		t.Error("expected Battery to be nil")
	}
	if info.LastUpdate != 0 {
		t.Errorf("got LastUpdate=%f, want 0", info.LastUpdate)
	}
}

// ----- Nil Parent Tests -----

func TestNewWithNilParent(t *testing.T) {
	t.Parallel()

	svc := New(nil)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.parent != nil {
		t.Error("expected parent to be nil")
	}
}

// ----- Parse Function Edge Cases -----

func TestParseLoRaConfig_EmptyMap(t *testing.T) {
	t.Parallel()

	cfgMap := map[string]any{}
	cfg := parseLoRaConfig(cfgMap)

	if cfg.ID != 0 {
		t.Errorf("got ID=%d, want 0", cfg.ID)
	}
	if cfg.Freq != 0 {
		t.Errorf("got Freq=%d, want 0", cfg.Freq)
	}
	if cfg.BW != 0 {
		t.Errorf("got BW=%d, want 0", cfg.BW)
	}
	if cfg.DR != 0 {
		t.Errorf("got DR=%d, want 0", cfg.DR)
	}
	if cfg.TxP != 0 {
		t.Errorf("got TxP=%d, want 0", cfg.TxP)
	}
}

func TestParseLoRaConfig_WrongTypes(t *testing.T) {
	t.Parallel()

	cfgMap := map[string]any{
		"id":   "not a number",
		"freq": "not a number",
		"bw":   "not a number",
		"dr":   "not a number",
		"txp":  "not a number",
	}
	cfg := parseLoRaConfig(cfgMap)

	if cfg.ID != 0 {
		t.Errorf("got ID=%d, want 0", cfg.ID)
	}
	if cfg.Freq != 0 {
		t.Errorf("got Freq=%d, want 0", cfg.Freq)
	}
}

func TestParseLoRaStatus_EmptyMap(t *testing.T) {
	t.Parallel()

	statusMap := map[string]any{}
	status := parseLoRaStatus(statusMap)

	if status.ID != 0 {
		t.Errorf("got ID=%d, want 0", status.ID)
	}
	if status.RSSI != 0 {
		t.Errorf("got RSSI=%d, want 0", status.RSSI)
	}
	if status.SNR != 0 {
		t.Errorf("got SNR=%f, want 0", status.SNR)
	}
}

func TestParseLoRaStatus_WrongTypes(t *testing.T) {
	t.Parallel()

	statusMap := map[string]any{
		"id":   "not a number",
		"rssi": "not a number",
		"snr":  "not a number",
	}
	status := parseLoRaStatus(statusMap)

	if status.ID != 0 {
		t.Errorf("got ID=%d, want 0", status.ID)
	}
	if status.RSSI != 0 {
		t.Errorf("got RSSI=%d, want 0", status.RSSI)
	}
	if status.SNR != 0 {
		t.Errorf("got SNR=%f, want 0", status.SNR)
	}
}

// ----- JSON Serialization Tests -----

func TestBTHomeDeviceStatus_JSONSerialization(t *testing.T) {
	t.Parallel()

	rssi := -55
	battery := 75
	status := BTHomeDeviceStatus{
		ID:         1,
		RSSI:       &rssi,
		Battery:    &battery,
		LastUpdate: 1700000000.5,
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded BTHomeDeviceStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != status.ID {
		t.Errorf("got ID=%d, want %d", decoded.ID, status.ID)
	}
	if decoded.RSSI == nil || *decoded.RSSI != rssi {
		t.Errorf("got RSSI=%v, want %d", decoded.RSSI, rssi)
	}
	if decoded.Battery == nil || *decoded.Battery != battery {
		t.Errorf("got Battery=%v, want %d", decoded.Battery, battery)
	}
}

func TestBTHomeSensorStatus_JSONSerialization(t *testing.T) {
	t.Parallel()

	status := BTHomeSensorStatus{
		ID:           2,
		Value:        23.5,
		LastUpdateTS: 1700000001.0,
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded BTHomeSensorStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != status.ID {
		t.Errorf("got ID=%d, want %d", decoded.ID, status.ID)
	}
	if decoded.Value != status.Value {
		t.Errorf("got Value=%v, want %v", decoded.Value, status.Value)
	}
	if decoded.LastUpdateTS != status.LastUpdateTS {
		t.Errorf("got LastUpdateTS=%f, want %f", decoded.LastUpdateTS, status.LastUpdateTS)
	}
}

func TestBTHomeAddDeviceResult_JSONSerialization(t *testing.T) {
	t.Parallel()

	result := BTHomeAddDeviceResult{Key: "bthomedevice:100"}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded BTHomeAddDeviceResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Key != result.Key {
		t.Errorf("got Key=%q, want %q", decoded.Key, result.Key)
	}
}

func TestMatterConfig_JSONSerialization(t *testing.T) {
	t.Parallel()

	cfg := MatterConfig{Enable: true}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded MatterConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Enable != cfg.Enable {
		t.Errorf("got Enable=%v, want %v", decoded.Enable, cfg.Enable)
	}
}

// ----- Interface Compliance Test -----

func TestParent_Interface(t *testing.T) {
	t.Parallel()

	// This test verifies that mockParent satisfies the Parent interface
	var _ Parent = (*mockParent)(nil)
}

// ----- Collect Functions Edge Cases -----

func TestCollectBTHomeDevices_NilStatus(t *testing.T) {
	t.Parallel()

	devices := CollectBTHomeDevices(nil, nil)

	if len(devices) != 0 {
		t.Errorf("got %d devices, want 0", len(devices))
	}
}

func TestCollectBTHomeSensors_NilStatus(t *testing.T) {
	t.Parallel()

	sensors := CollectBTHomeSensors(nil, nil)

	if len(sensors) != 0 {
		t.Errorf("got %d sensors, want 0", len(sensors))
	}
}

func TestCollectBTHomeSensors_InvalidJSON(t *testing.T) {
	t.Parallel()

	status := map[string]json.RawMessage{
		"bthomesensor:0": json.RawMessage(`{invalid json`),
		"bthomesensor:1": json.RawMessage(`{"id": 1}`),
	}

	sensors := CollectBTHomeSensors(status, nil)

	if len(sensors) != 1 {
		t.Errorf("got %d sensors, want 1", len(sensors))
	}
}
