package connection

import (
	"github.com/tj-smith47/shelly-cli/internal/client"
)

// DeviceClient provides a unified interface for both Gen1 and Gen2 device connections.
// Use IsGen1() to determine which generation, then Gen1() or Gen2() to access
// generation-specific APIs.
type DeviceClient struct {
	gen1 *client.Gen1Client
	gen2 *client.Client
}

// NewGen1Client creates a DeviceClient wrapping a Gen1 connection.
func NewGen1Client(conn *client.Gen1Client) *DeviceClient {
	return &DeviceClient{gen1: conn}
}

// NewGen2Client creates a DeviceClient wrapping a Gen2+ connection.
func NewGen2Client(conn *client.Client) *DeviceClient {
	return &DeviceClient{gen2: conn}
}

// IsGen1 returns true if this is a Gen1 device connection.
func (c *DeviceClient) IsGen1() bool {
	return c.gen1 != nil
}

// IsGen2 returns true if this is a Gen2+ device connection.
func (c *DeviceClient) IsGen2() bool {
	return c.gen2 != nil
}

// Generation returns the device generation (1 for Gen1, 2+ for Gen2).
func (c *DeviceClient) Generation() int {
	if c.gen1 != nil {
		return c.gen1.Info().Generation
	}
	if c.gen2 != nil {
		return c.gen2.Info().Generation
	}
	return 0
}

// Gen1 returns the Gen1 client. Panics if this is not a Gen1 connection.
// Check IsGen1() first.
func (c *DeviceClient) Gen1() *client.Gen1Client {
	if c.gen1 == nil {
		panic("DeviceClient: not a Gen1 connection, check IsGen1() first")
	}
	return c.gen1
}

// Gen2 returns the Gen2+ client. Panics if this is not a Gen2 connection.
// Check IsGen2() first.
func (c *DeviceClient) Gen2() *client.Client {
	if c.gen2 == nil {
		panic("DeviceClient: not a Gen2 connection, check IsGen2() first")
	}
	return c.gen2
}

// Info returns the device information. Works for both generations.
func (c *DeviceClient) Info() *client.DeviceInfo {
	if c.gen1 != nil {
		return c.gen1.Info()
	}
	if c.gen2 != nil {
		return c.gen2.Info()
	}
	return nil
}

// Close closes the device connection. Works for both generations.
func (c *DeviceClient) Close() error {
	if c.gen1 != nil {
		return c.gen1.Close()
	}
	if c.gen2 != nil {
		return c.gen2.Close()
	}
	return nil
}
