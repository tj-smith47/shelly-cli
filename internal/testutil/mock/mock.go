// Package mock provides mock device utilities for testing.
package mock

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"

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
// Uses the package-level filesystem from config.
func Dir() (string, error) {
	return DirWithFs(nil)
}

// DirWithFs returns the mock devices directory path, creating it if needed.
// If fs is nil, uses the package-level filesystem from config.
func DirWithFs(fs afero.Fs) (string, error) {
	configDir, err := config.Dir()
	if err != nil {
		return "", err
	}
	mockDir := filepath.Join(configDir, "mock")

	// Use provided fs or fall back to a new manager's fs (which uses package default)
	if fs == nil {
		mgr := config.NewManager("")
		fs = mgr.Fs()
	}

	if err := fs.MkdirAll(mockDir, 0o700); err != nil {
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
