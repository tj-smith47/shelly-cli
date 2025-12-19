// Package mock provides mock device utilities for testing.
package mock

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Device represents a mock device configuration.
type Device struct {
	Name     string                 `json:"name"`
	Model    string                 `json:"model"`
	Firmware string                 `json:"firmware"`
	MAC      string                 `json:"mac"`
	State    map[string]interface{} `json:"state"`
}

// Dir returns the mock devices directory path, creating it if needed.
func Dir() (string, error) {
	configDir, err := config.Dir()
	if err != nil {
		return "", err
	}
	mockDir := filepath.Join(configDir, "mock")
	if err := os.MkdirAll(mockDir, 0o700); err != nil {
		return "", err
	}
	return mockDir, nil
}

// GenerateMAC generates a deterministic MAC address based on the device name.
func GenerateMAC(name string) string {
	hash := 0
	for _, c := range name {
		hash = hash*31 + int(c)
	}
	return fmt.Sprintf("AA:BB:CC:%02X:%02X:%02X",
		(hash>>16)&0xFF,
		(hash>>8)&0xFF,
		hash&0xFF,
	)
}
