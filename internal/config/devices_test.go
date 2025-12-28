package config

import (
	"path/filepath"
	"testing"
)

func TestNormalizeDeviceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase only", "kitchen", "kitchen"},
		{"mixed case", "Kitchen", "kitchen"},
		{"all caps", "KITCHEN", "kitchen"},
		{"spaces to dashes", "living room", "living-room"},
		{"underscores to dashes", "living_room", "living-room"},
		{"mixed separators", "Living_Room Light", "living-room-light"},
		{"special chars removed", "device@home!", "devicehome"},
		{"multiple spaces", "my   device", "my-device"},
		{"leading trailing spaces", "  device  ", "device"},
		{"leading trailing dashes", "--device--", "device"},
		{"only special chars", "!!!@@@###", ""},
		{"empty string", "", ""},
		{"numbers preserved", "device123", "device123"},
		{"complex name", "Master Bathroom_Light (Main)", "master-bathroom-light-main"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NormalizeDeviceName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeDeviceName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateDeviceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "kitchen", false},
		{"valid with spaces", "Living Room", false},
		{"valid with underscores", "living_room", false},
		{"valid with dashes", "living-room", false},
		{"valid with numbers", "device123", false},
		{"valid complex", "Master Bathroom Light", false},
		{"empty", "", true},
		{"contains forward slash", "my/device", true},
		{"contains backslash", "my\\device", true},
		{"contains colon", "my:device", true},
		{"only special chars", "@@@", true},
		{"normalizes to empty", "!!!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateDeviceName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDeviceName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateGroupName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "lights", false},
		{"valid with spaces", "Living Room", false},
		{"valid with underscores", "all_lights", false},
		{"valid with dashes", "all-lights", false},
		{"valid with numbers", "group1", false},
		{"empty", "", true},
		{"contains forward slash", "my/group", true},
		{"contains backslash", "my\\group", true},
		{"contains colon", "my:group", true},
		{"only special chars", "@@@", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateGroupName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGroupName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

//nolint:gocyclo // test function with many assertions
func TestManager_GroupOperations(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Create a group
	if err := m.CreateGroup("lights"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}

	// Verify group exists
	grp, ok := m.GetGroup("lights")
	if !ok {
		t.Fatal("GetGroup() returned false")
	}
	if len(grp.Devices) != 0 {
		t.Errorf("expected empty devices, got %d", len(grp.Devices))
	}

	// List groups
	groups := m.ListGroups()
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}

	// Register a device and add to group
	if err := m.RegisterDevice("kitchen", "192.168.1.1", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.AddDeviceToGroup("lights", "kitchen"); err != nil {
		t.Fatalf("AddDeviceToGroup() error: %v", err)
	}

	// Verify device in group
	grp, _ = m.GetGroup("lights")
	if len(grp.Devices) != 1 || grp.Devices[0] != "kitchen" {
		t.Errorf("expected devices [kitchen], got %v", grp.Devices)
	}

	// Get group devices
	devices, err := m.GetGroupDevices("lights")
	if err != nil {
		t.Fatalf("GetGroupDevices() error: %v", err)
	}
	if len(devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(devices))
	}
	if devices[0].Address != "192.168.1.1" {
		t.Errorf("expected address 192.168.1.1, got %s", devices[0].Address)
	}

	// Remove device from group
	if err := m.RemoveDeviceFromGroup("lights", "kitchen"); err != nil {
		t.Fatalf("RemoveDeviceFromGroup() error: %v", err)
	}
	grp, _ = m.GetGroup("lights")
	if len(grp.Devices) != 0 {
		t.Errorf("expected empty devices after remove, got %d", len(grp.Devices))
	}

	// Delete group
	if err := m.DeleteGroup("lights"); err != nil {
		t.Fatalf("DeleteGroup() error: %v", err)
	}
	_, ok = m.GetGroup("lights")
	if ok {
		t.Error("group should not exist after DeleteGroup()")
	}
}

func TestManager_CreateGroup_Duplicate(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if err := m.CreateGroup("lights"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}

	// Try to create duplicate
	err := m.CreateGroup("lights")
	if err == nil {
		t.Error("expected error creating duplicate group")
	}
}

func TestManager_AddDeviceToGroup_Errors(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Add device to nonexistent group
	err := m.AddDeviceToGroup("nonexistent", "device")
	if err == nil {
		t.Error("expected error adding to nonexistent group")
	}

	// Create group and add a device
	if err := m.CreateGroup("lights"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}
	if err := m.RegisterDevice("kitchen", "192.168.1.1", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.AddDeviceToGroup("lights", "kitchen"); err != nil {
		t.Fatalf("AddDeviceToGroup() error: %v", err)
	}

	// Try to add same device again (should fail - duplicate)
	err = m.AddDeviceToGroup("lights", "kitchen")
	if err == nil {
		t.Error("expected error adding duplicate device to group")
	}
}

func TestManager_RenameDevice(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Register a device
	if err := m.RegisterDevice("old-name", "192.168.1.1", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Rename it
	if err := m.RenameDevice("old-name", "new-name"); err != nil {
		t.Fatalf("RenameDevice() error: %v", err)
	}

	// Old name should not exist
	_, ok := m.GetDevice("old-name")
	if ok {
		t.Error("old name should not exist after rename")
	}

	// New name should exist with same address
	dev, ok := m.GetDevice("new-name")
	if !ok {
		t.Fatal("new name should exist after rename")
	}
	if dev.Address != "192.168.1.1" {
		t.Errorf("expected address 192.168.1.1, got %s", dev.Address)
	}
}

func TestManager_RenameDevice_Errors(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Rename nonexistent device
	err := m.RenameDevice("nonexistent", "new-name")
	if err == nil {
		t.Error("expected error renaming nonexistent device")
	}

	// Register two devices
	if err := m.RegisterDevice("device1", "192.168.1.1", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice(device1) error: %v", err)
	}
	if err := m.RegisterDevice("device2", "192.168.1.2", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice(device2) error: %v", err)
	}

	// Rename to existing name
	err = m.RenameDevice("device1", "device2")
	if err == nil {
		t.Error("expected error renaming to existing name")
	}
}

func TestManager_UpdateDeviceAddress(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Register a device
	if err := m.RegisterDevice("kitchen", "192.168.1.1", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Update address
	if err := m.UpdateDeviceAddress("kitchen", "192.168.1.100"); err != nil {
		t.Fatalf("UpdateDeviceAddress() error: %v", err)
	}

	// Verify new address
	dev, _ := m.GetDevice("kitchen")
	if dev.Address != "192.168.1.100" {
		t.Errorf("expected address 192.168.1.100, got %s", dev.Address)
	}
}

func TestManager_UnregisterDevice(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Register a device
	if err := m.RegisterDevice("kitchen", "192.168.1.1", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Unregister it
	if err := m.UnregisterDevice("kitchen"); err != nil {
		t.Fatalf("UnregisterDevice() error: %v", err)
	}

	// Should not exist
	_, ok := m.GetDevice("kitchen")
	if ok {
		t.Error("device should not exist after UnregisterDevice()")
	}
}

func TestManager_UnregisterDevice_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	err := m.UnregisterDevice("nonexistent")
	if err == nil {
		t.Error("expected error unregistering nonexistent device")
	}
}
