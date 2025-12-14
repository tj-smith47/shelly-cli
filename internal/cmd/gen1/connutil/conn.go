// Package connutil provides shared connection utilities for Gen1 device commands.
package connutil

import (
	"context"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// DefaultTimeout is the default connection timeout for Gen1 devices.
const DefaultTimeout = 10 * time.Second

// ConnectGen1 resolves device config and connects to a Gen1 device.
// It handles device resolution, authentication setup, and connection with progress indication.
func ConnectGen1(ctx context.Context, ios *iostreams.IOStreams, deviceName string) (*client.Gen1Client, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	devCfg, err := config.ResolveDevice(deviceName)
	if err != nil {
		return nil, err
	}

	device := model.Device{
		Name:    devCfg.Name,
		Address: devCfg.Address,
	}
	if devCfg.Auth != nil {
		device.Auth = &model.Auth{
			Username: devCfg.Auth.Username,
			Password: devCfg.Auth.Password,
		}
	}

	ios.StartProgress("Connecting to device...")
	gen1Client, err := client.ConnectGen1(ctx, device)
	ios.StopProgress()

	if err != nil {
		return nil, err
	}

	return gen1Client, nil
}
