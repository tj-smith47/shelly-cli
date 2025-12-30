package connection

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ExecuteGen2 performs the actual Gen2+ connection and function execution.
// Includes automatic IP remapping: if connection fails and device has a MAC,
// attempts mDNS discovery to find the device's new IP address.
func (m *Manager) ExecuteGen2(ctx context.Context, dev model.Device, fn func(*client.Client) error) error {
	conn, err := client.Connect(ctx, dev)
	if err != nil {
		// Try IP remapping if connection failed and we have a MAC address
		conn, err = m.tryIPRemap(ctx, dev, err)
		if err != nil {
			return err
		}
	}
	defer iostreams.CloseWithDebug("closing device connection", conn)

	return fn(conn)
}

// ExecuteGen1 performs the actual Gen1 connection and function execution.
// Includes automatic IP remapping: if connection fails and device has a MAC,
// attempts mDNS discovery to find the device's new IP address.
func (m *Manager) ExecuteGen1(ctx context.Context, dev model.Device, fn func(*client.Gen1Client) error) error {
	conn, err := client.ConnectGen1(ctx, dev)
	if err != nil {
		// Try IP remapping if connection failed and we have a MAC address
		conn, err = m.tryGen1IPRemap(ctx, dev, err)
		if err != nil {
			return err
		}
	}
	defer iostreams.CloseWithDebug("closing gen1 device connection", conn)

	return fn(conn)
}

// tryIPRemap attempts to remap a device's IP address via mDNS discovery.
// Returns a new connection if remapping succeeds, or the original error if not.
func (m *Manager) tryIPRemap(ctx context.Context, dev model.Device, originalErr error) (*client.Client, error) {
	// Only attempt remap for connection errors with a known MAC
	if !isConnectionError(originalErr) || dev.MAC == "" || m.discoverer == nil {
		return nil, originalErr
	}

	iostreams.DebugCat(iostreams.CategoryDevice, "connection failed for %s, attempting MAC discovery...", dev.Name)

	// Quick mDNS discovery (~2 seconds)
	newIP, discoverErr := m.discoverer.DiscoverByMAC(ctx, dev.MAC)
	if discoverErr != nil {
		iostreams.DebugErr("MAC discovery failed", discoverErr)
		return nil, originalErr
	}
	if newIP == "" || newIP == dev.Address {
		// Not found or same IP - return original error
		return nil, originalErr
	}

	iostreams.DebugCat(iostreams.CategoryDevice, "found new IP %s for MAC %s, verifying...", newIP, dev.MAC)

	// Try connecting to new IP
	devCopy := dev
	devCopy.Address = newIP
	conn, retryErr := client.Connect(ctx, devCopy)
	if retryErr != nil {
		iostreams.DebugErr("connection to new IP failed", retryErr)
		return nil, originalErr
	}

	// Verify MAC matches (security check)
	info := conn.Info()
	if info == nil || model.NormalizeMAC(info.MAC) != dev.NormalizedMAC() {
		iostreams.DebugCat(iostreams.CategoryDevice, "MAC mismatch: expected %s, got %s", dev.NormalizedMAC(), model.NormalizeMAC(info.MAC))
		iostreams.CloseWithDebug("closing mismatched connection", conn)
		return nil, originalErr
	}

	// Success! Update config silently
	if err := config.UpdateDeviceAddress(dev.Name, newIP); err != nil {
		iostreams.DebugErr("failed to persist new IP", err)
		// Continue anyway - connection works, just won't persist
	} else {
		iostreams.DebugCat(iostreams.CategoryDevice, "remapped %s: %s -> %s", dev.Name, dev.Address, newIP)
	}

	return conn, nil
}

// tryGen1IPRemap attempts to remap a Gen1 device's IP address via mDNS discovery.
// Returns a new connection if remapping succeeds, or the original error if not.
func (m *Manager) tryGen1IPRemap(ctx context.Context, dev model.Device, originalErr error) (*client.Gen1Client, error) {
	// Only attempt remap for connection errors with a known MAC
	if !isConnectionError(originalErr) || dev.MAC == "" || m.discoverer == nil {
		return nil, originalErr
	}

	iostreams.DebugCat(iostreams.CategoryDevice, "Gen1 connection failed for %s, attempting MAC discovery...", dev.Name)

	// Quick mDNS discovery (~2 seconds)
	newIP, discoverErr := m.discoverer.DiscoverByMAC(ctx, dev.MAC)
	if discoverErr != nil {
		iostreams.DebugErr("MAC discovery failed", discoverErr)
		return nil, originalErr
	}
	if newIP == "" || newIP == dev.Address {
		// Not found or same IP - return original error
		return nil, originalErr
	}

	iostreams.DebugCat(iostreams.CategoryDevice, "found new IP %s for MAC %s, verifying...", newIP, dev.MAC)

	// Try connecting to new IP
	devCopy := dev
	devCopy.Address = newIP
	conn, retryErr := client.ConnectGen1(ctx, devCopy)
	if retryErr != nil {
		iostreams.DebugErr("Gen1 connection to new IP failed", retryErr)
		return nil, originalErr
	}

	// For Gen1, we can't easily verify MAC from connection info,
	// but the discovery already matched by MAC, so we trust it

	// Success! Update config silently
	if err := config.UpdateDeviceAddress(dev.Name, newIP); err != nil {
		iostreams.DebugErr("failed to persist new IP", err)
		// Continue anyway - connection works, just won't persist
	} else {
		iostreams.DebugCat(iostreams.CategoryDevice, "remapped Gen1 %s: %s -> %s", dev.Name, dev.Address, newIP)
	}

	return conn, nil
}
