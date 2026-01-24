// Package wireless provides wireless protocol operations for Shelly devices.
package wireless

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// BTHomeDeviceStatus represents BTHome device status from Shelly.GetStatus.
type BTHomeDeviceStatus struct {
	ID         int     `json:"id"`
	RSSI       *int    `json:"rssi"`
	Battery    *int    `json:"battery"`
	LastUpdate float64 `json:"last_updated_ts"`
}

// CollectBTHomeDevices collects BTHomeDevice components from device status.
// Returns a slice of BTHomeDeviceStatus for all bthomedevice:* keys in the status map.
func CollectBTHomeDevices(status map[string]json.RawMessage, ios *iostreams.IOStreams) []BTHomeDeviceStatus {
	devices := make([]BTHomeDeviceStatus, 0, len(status))

	for key, raw := range status {
		if !strings.HasPrefix(key, "bthomedevice:") {
			continue
		}

		var devStatus BTHomeDeviceStatus
		if err := json.Unmarshal(raw, &devStatus); err != nil {
			if ios != nil {
				ios.Debug("failed to unmarshal BTHome device %s: %v", key, err)
			}
			continue
		}

		devices = append(devices, devStatus)
	}

	return devices
}

// BTHomeSensorStatus represents BTHome sensor status from Shelly.GetStatus.
type BTHomeSensorStatus struct {
	ID           int     `json:"id"`
	Value        any     `json:"value"`
	LastUpdateTS float64 `json:"last_updated_ts"`
}

// CollectBTHomeSensors collects BTHomeSensor components from device status.
// Returns a slice of BTHomeSensorStatus for all bthomesensor:* keys in the status map.
func CollectBTHomeSensors(status map[string]json.RawMessage, ios *iostreams.IOStreams) []BTHomeSensorStatus {
	sensors := make([]BTHomeSensorStatus, 0, len(status))

	for key, raw := range status {
		if !strings.HasPrefix(key, "bthomesensor:") {
			continue
		}

		var sensorStatus BTHomeSensorStatus
		if err := json.Unmarshal(raw, &sensorStatus); err != nil {
			if ios != nil {
				ios.Debug("failed to unmarshal BTHome sensor %s: %v", key, err)
			}
			continue
		}

		sensors = append(sensors, sensorStatus)
	}

	return sensors
}

// BTHomeAddDeviceResult represents the result of adding a BTHome device.
type BTHomeAddDeviceResult struct {
	Key string `json:"key"`
}

// BTHomeAddDevice adds a BTHome device to a Shelly gateway.
func (s *Service) BTHomeAddDevice(ctx context.Context, identifier, addr, name string) (BTHomeAddDeviceResult, error) {
	var result BTHomeAddDeviceResult
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		config := map[string]any{
			"addr": addr,
		}
		if name != "" {
			config["name"] = name
		}
		params := map[string]any{
			"config": config,
		}

		resultAny, err := conn.Call(ctx, "BTHome.AddDevice", params)
		if err != nil {
			return fmt.Errorf("failed to add BTHome device: %w", err)
		}

		resultMap, ok := resultAny.(map[string]any)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}

		if key, ok := resultMap["key"].(string); ok {
			result.Key = key
		}
		return nil
	})
	return result, err
}

// BTHomeStartDiscovery starts BTHome device discovery on a gateway.
func (s *Service) BTHomeStartDiscovery(ctx context.Context, identifier string, duration int) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		params := map[string]any{
			"duration": duration,
		}
		_, err := conn.Call(ctx, "BTHome.StartDeviceDiscovery", params)
		if err != nil {
			return fmt.Errorf("failed to start BTHome discovery: %w", err)
		}
		return nil
	})
}

// BTHomeRemoveDevice removes a BTHome device from a gateway.
func (s *Service) BTHomeRemoveDevice(ctx context.Context, identifier string, deviceID int) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		params := map[string]any{
			"id": deviceID,
		}
		_, err := conn.Call(ctx, "BTHome.DeleteDevice", params)
		if err != nil {
			return fmt.Errorf("failed to remove BTHome device: %w", err)
		}
		return nil
	})
}

// FetchBTHomeDevices fetches all BTHome devices from a gateway with their config.
func (s *Service) FetchBTHomeDevices(ctx context.Context, identifier string, ios *iostreams.IOStreams) ([]model.BTHomeDeviceInfo, error) {
	var devices []model.BTHomeDeviceInfo

	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		status, err := getBTHomeDeviceStatus(ctx, conn)
		if err != nil {
			return err
		}

		deviceStatuses := CollectBTHomeDevices(status, ios)
		devices = enrichBTHomeDevices(ctx, conn, deviceStatuses)
		return nil
	})

	return devices, err
}

func getBTHomeDeviceStatus(ctx context.Context, conn *client.Client) (map[string]json.RawMessage, error) {
	result, err := conn.Call(ctx, "Shelly.GetStatus", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var status map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	return status, nil
}

func enrichBTHomeDevices(ctx context.Context, conn *client.Client, deviceStatuses []BTHomeDeviceStatus) []model.BTHomeDeviceInfo {
	devices := make([]model.BTHomeDeviceInfo, 0, len(deviceStatuses))

	for _, devStatus := range deviceStatuses {
		name, addr := getBTHomeDeviceConfig(ctx, conn, devStatus.ID)
		devices = append(devices, model.BTHomeDeviceInfo{
			ID:         devStatus.ID,
			Name:       name,
			Addr:       addr,
			RSSI:       devStatus.RSSI,
			Battery:    devStatus.Battery,
			LastUpdate: devStatus.LastUpdate,
		})
	}

	return devices
}

func getBTHomeDeviceConfig(ctx context.Context, conn *client.Client, id int) (name, addr string) {
	configResult, err := conn.Call(ctx, "BTHomeDevice.GetConfig", map[string]any{"id": id})
	if err != nil {
		return "", ""
	}

	var cfg struct {
		Name *string `json:"name"`
		Addr string  `json:"addr"`
	}
	cfgBytes, err := json.Marshal(configResult)
	if err != nil {
		return "", ""
	}
	if json.Unmarshal(cfgBytes, &cfg) != nil {
		return "", ""
	}

	if cfg.Name != nil {
		name = *cfg.Name
	}
	return name, cfg.Addr
}

// FetchBTHomeComponentStatus fetches the BTHome component status.
func (s *Service) FetchBTHomeComponentStatus(ctx context.Context, identifier string) (model.BTHomeComponentStatus, error) {
	var status model.BTHomeComponentStatus

	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		result, err := conn.Call(ctx, "BTHome.GetStatus", nil)
		if err != nil {
			return fmt.Errorf("failed to get BTHome status: %w", err)
		}

		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &status); err != nil {
			return fmt.Errorf("failed to parse status: %w", err)
		}

		return nil
	})

	return status, err
}

// FetchBTHomeDeviceStatus fetches detailed status for a specific BTHome device.
func (s *Service) FetchBTHomeDeviceStatus(ctx context.Context, identifier string, id int) (model.BTHomeDeviceStatus, error) {
	var status model.BTHomeDeviceStatus
	params := map[string]any{"id": id}

	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		var err error
		status, err = getBTHomeDeviceStatusRPC(ctx, conn, params)
		if err != nil {
			return err
		}

		name, addr := getBTHomeDeviceConfig(ctx, conn, id)
		status.Name = name
		status.Addr = addr

		status.KnownObjects = getBTHomeKnownObjectsRPC(ctx, conn, params)
		return nil
	})

	return status, err
}

func getBTHomeDeviceStatusRPC(ctx context.Context, conn *client.Client, params map[string]any) (model.BTHomeDeviceStatus, error) {
	var status model.BTHomeDeviceStatus

	result, err := conn.Call(ctx, "BTHomeDevice.GetStatus", params)
	if err != nil {
		return status, fmt.Errorf("failed to get device status: %w", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return status, fmt.Errorf("failed to marshal result: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, &status); err != nil {
		return status, fmt.Errorf("failed to parse status: %w", err)
	}

	return status, nil
}

func getBTHomeKnownObjectsRPC(ctx context.Context, conn *client.Client, params map[string]any) []model.BTHomeKnownObj {
	objResult, err := conn.Call(ctx, "BTHomeDevice.GetKnownObjects", params)
	if err != nil {
		return nil
	}

	var objResp struct {
		Objects []model.BTHomeKnownObj `json:"objects"`
	}
	objBytes, err := json.Marshal(objResult)
	if err != nil {
		return nil
	}
	if json.Unmarshal(objBytes, &objResp) != nil {
		return nil
	}

	return objResp.Objects
}

// FetchBTHomeSensors fetches all BTHome sensors from a gateway with their values.
func (s *Service) FetchBTHomeSensors(ctx context.Context, identifier string, ios *iostreams.IOStreams) ([]model.BTHomeSensorInfo, error) {
	var sensors []model.BTHomeSensorInfo

	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		status, err := getBTHomeDeviceStatus(ctx, conn)
		if err != nil {
			return err
		}

		sensorStatuses := CollectBTHomeSensors(status, ios)
		sensors = enrichBTHomeSensors(ctx, conn, sensorStatuses)
		return nil
	})

	return sensors, err
}

func enrichBTHomeSensors(ctx context.Context, conn *client.Client, sensorStatuses []BTHomeSensorStatus) []model.BTHomeSensorInfo {
	sensors := make([]model.BTHomeSensorInfo, 0, len(sensorStatuses))

	for _, sensorStatus := range sensorStatuses {
		cfg := getBTHomeSensorConfig(ctx, conn, sensorStatus.ID)
		sensors = append(sensors, model.BTHomeSensorInfo{
			ID:           sensorStatus.ID,
			Name:         cfg.Name,
			Addr:         cfg.Addr,
			ObjID:        cfg.ObjID,
			Idx:          cfg.Idx,
			Value:        sensorStatus.Value,
			LastUpdateTS: sensorStatus.LastUpdateTS,
		})
	}

	return sensors
}

type bthomeSensorConfig struct {
	Name  string
	Addr  string
	ObjID int
	Idx   int
}

func getBTHomeSensorConfig(ctx context.Context, conn *client.Client, id int) bthomeSensorConfig {
	var cfg bthomeSensorConfig

	configResult, err := conn.Call(ctx, "BTHomeSensor.GetConfig", map[string]any{"id": id})
	if err != nil {
		return cfg
	}

	var rawCfg struct {
		Name  *string `json:"name"`
		Addr  string  `json:"addr"`
		ObjID int     `json:"obj_id"`
		Idx   int     `json:"idx"`
	}
	cfgBytes, err := json.Marshal(configResult)
	if err != nil {
		return cfg
	}
	if json.Unmarshal(cfgBytes, &rawCfg) != nil {
		return cfg
	}

	if rawCfg.Name != nil {
		cfg.Name = *rawCfg.Name
	}
	cfg.Addr = rawCfg.Addr
	cfg.ObjID = rawCfg.ObjID
	cfg.Idx = rawCfg.Idx

	return cfg
}

// FetchBTHomeObjectInfos fetches information about BTHome object types.
func (s *Service) FetchBTHomeObjectInfos(ctx context.Context, identifier string, objIDs []int) ([]model.BTHomeObjectInfo, error) {
	var infos []model.BTHomeObjectInfo

	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		params := map[string]any{
			"obj_ids": objIDs,
		}

		result, err := conn.Call(ctx, "BTHome.GetObjectInfos", params)
		if err != nil {
			return fmt.Errorf("failed to get object infos: %w", err)
		}

		var resp struct {
			Infos []model.BTHomeObjectInfo `json:"infos"`
		}
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &resp); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		infos = resp.Infos
		return nil
	})

	return infos, err
}
