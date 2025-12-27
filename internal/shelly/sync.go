package shelly

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// SyncResult holds the result of syncing a device config.
type SyncResult struct {
	Config map[string]any
	Err    error
}

// FetchDeviceConfig fetches config from a device and returns it as a map.
func (s *Service) FetchDeviceConfig(ctx context.Context, device string) SyncResult {
	conn, err := s.Connect(ctx, device)
	if err != nil {
		return SyncResult{Err: err}
	}

	rawResult, err := conn.Call(ctx, "Shelly.GetConfig", nil)
	iostreams.CloseWithDebug("closing sync connection", conn)
	if err != nil {
		return SyncResult{Err: err}
	}

	jsonBytes, err := json.Marshal(rawResult)
	if err != nil {
		return SyncResult{Err: fmt.Errorf("marshal: %w", err)}
	}

	var deviceConfig map[string]any
	if err := json.Unmarshal(jsonBytes, &deviceConfig); err != nil {
		return SyncResult{Err: fmt.Errorf("unmarshal: %w", err)}
	}

	return SyncResult{Config: deviceConfig}
}

// PushDeviceConfig pushes config to a device.
func (s *Service) PushDeviceConfig(ctx context.Context, device string, cfg map[string]any) error {
	conn, err := s.Connect(ctx, device)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	_, err = conn.Call(ctx, "Shelly.SetConfig", map[string]any{"config": cfg})
	iostreams.CloseWithDebug("closing sync push connection", conn)
	if err != nil {
		return err
	}
	return nil
}
