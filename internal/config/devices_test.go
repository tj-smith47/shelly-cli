package config

import (
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

const (
	testDeviceIP   = "192.168.1.1"
	testAuthAdmin  = "admin"
	testDeviceName = "kitchen"
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

//nolint:gocyclo,paralleltest // test function with many assertions; modifies global state via SetFs
func TestManager_GroupOperations(t *testing.T) {
	m := setupManagerTest(t)

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
	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.AddDeviceToGroup("lights", testDeviceName); err != nil {
		t.Fatalf("AddDeviceToGroup() error: %v", err)
	}

	// Verify device in group
	grp, _ = m.GetGroup("lights")
	if len(grp.Devices) != 1 || grp.Devices[0] != testDeviceName {
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
	if devices[0].Address != testDeviceIP {
		t.Errorf("expected address testDeviceIP, got %s", devices[0].Address)
	}

	// Remove device from group
	if err := m.RemoveDeviceFromGroup("lights", testDeviceName); err != nil {
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

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_CreateGroup_Duplicate(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.CreateGroup("lights"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}

	// Try to create duplicate
	err := m.CreateGroup("lights")
	if err == nil {
		t.Error("expected error creating duplicate group")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_AddDeviceToGroup_Errors(t *testing.T) {
	m := setupManagerTest(t)

	// Add device to nonexistent group
	err := m.AddDeviceToGroup("nonexistent", "device")
	if err == nil {
		t.Error("expected error adding to nonexistent group")
	}

	// Create group and add a device
	if err := m.CreateGroup("lights"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}
	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.AddDeviceToGroup("lights", testDeviceName); err != nil {
		t.Fatalf("AddDeviceToGroup() error: %v", err)
	}

	// Try to add same device again (should fail - duplicate)
	err = m.AddDeviceToGroup("lights", testDeviceName)
	if err == nil {
		t.Error("expected error adding duplicate device to group")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RenameDevice(t *testing.T) {
	m := setupManagerTest(t)

	// Register a device
	if err := m.RegisterDevice("old-name", testDeviceIP, 2, "", "", nil); err != nil {
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
	if dev.Address != testDeviceIP {
		t.Errorf("expected address testDeviceIP, got %s", dev.Address)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RenameDevice_Errors(t *testing.T) {
	m := setupManagerTest(t)

	// Rename nonexistent device
	err := m.RenameDevice("nonexistent", "new-name")
	if err == nil {
		t.Error("expected error renaming nonexistent device")
	}

	// Register two devices
	if err := m.RegisterDevice("device1", testDeviceIP, 2, "", "", nil); err != nil {
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

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UpdateDeviceAddress(t *testing.T) {
	m := setupManagerTest(t)

	// Register a device
	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Update address
	if err := m.UpdateDeviceAddress(testDeviceName, "testDeviceIP00"); err != nil {
		t.Fatalf("UpdateDeviceAddress() error: %v", err)
	}

	// Verify new address
	dev, _ := m.GetDevice(testDeviceName)
	if dev.Address != "testDeviceIP00" {
		t.Errorf("expected address testDeviceIP00, got %s", dev.Address)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UnregisterDevice(t *testing.T) {
	m := setupManagerTest(t)

	// Register a device
	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Unregister it
	if err := m.UnregisterDevice(testDeviceName); err != nil {
		t.Fatalf("UnregisterDevice() error: %v", err)
	}

	// Should not exist
	_, ok := m.GetDevice(testDeviceName)
	if ok {
		t.Error("device should not exist after UnregisterDevice()")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UnregisterDevice_NotFound(t *testing.T) {
	m := setupManagerTest(t)

	err := m.UnregisterDevice("nonexistent")
	if err == nil {
		t.Error("expected error unregistering nonexistent device")
	}
}

// setupDevicesTest sets up an isolated environment for device package-level function tests.
func setupDevicesTest(t *testing.T) {
	t.Helper()
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })
	ResetDefaultManagerForTesting()
}

// setupManagerTest sets up an isolated Manager for testing.
// It uses an in-memory filesystem to avoid touching real files.
func setupManagerTest(t *testing.T) *Manager {
	t.Helper()
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })
	m := NewManager("/test/config/config.yaml")
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	return m
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_RegisterDeviceWithPlatform(t *testing.T) {
	setupDevicesTest(t)

	// Register with platform
	err := RegisterDeviceWithPlatform("tasmota-device", "192.168.1.50", 0, "smart-plug", "Sonoff Basic", "tasmota", nil)
	if err != nil {
		t.Errorf("RegisterDeviceWithPlatform() error = %v", err)
	}

	// Verify device was registered with platform
	dev, ok := GetDevice("tasmota-device")
	if !ok {
		t.Fatal("GetDevice() should find registered device")
	}
	if dev.Platform != "tasmota" {
		t.Errorf("dev.Platform = %q, want %q", dev.Platform, "tasmota")
	}
	if dev.Address != "192.168.1.50" {
		t.Errorf("dev.Address = %q, want %q", dev.Address, "192.168.1.50")
	}
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_UpdateDeviceInfo(t *testing.T) {
	setupDevicesTest(t)

	// Register a device
	if err := RegisterDevice("info-test", "192.168.1.60", 0, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error = %v", err)
	}

	// Update device info
	err := UpdateDeviceInfo("info-test", DeviceUpdates{
		Type:       "SPSW-001PE16EU",
		Model:      "Shelly Plus 1PM",
		Generation: 2,
		MAC:        "aa:bb:cc:dd:ee:ff",
	})
	if err != nil {
		t.Errorf("UpdateDeviceInfo() error = %v", err)
	}

	// Verify updates
	dev, ok := GetDevice("info-test")
	if !ok {
		t.Fatal("GetDevice() should find device")
	}
	if dev.Type != "SPSW-001PE16EU" {
		t.Errorf("dev.Type = %q, want %q", dev.Type, "SPSW-001PE16EU")
	}
	if dev.Model != "Shelly Plus 1PM" {
		t.Errorf("dev.Model = %q, want %q", dev.Model, "Shelly Plus 1PM")
	}
	if dev.Generation != 2 {
		t.Errorf("dev.Generation = %d, want %d", dev.Generation, 2)
	}
	if dev.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("dev.MAC = %q, want %q", dev.MAC, "AA:BB:CC:DD:EE:FF")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RegisterDeviceWithPlatform_Direct(t *testing.T) {
	m := setupManagerTest(t)

	// Register with platform
	auth := &model.Auth{Username: "testAuthAdmin", Password: "pass"}
	err := m.RegisterDeviceWithPlatform("esphome-device", "192.168.1.70", 0, "light", "ESPHome Light", "esphome", auth)
	if err != nil {
		t.Errorf("RegisterDeviceWithPlatform() error = %v", err)
	}

	// Verify
	dev, ok := m.GetDevice("esphome-device")
	if !ok {
		t.Fatal("GetDevice() should find device")
	}
	if dev.Platform != "esphome" {
		t.Errorf("dev.Platform = %q, want %q", dev.Platform, "esphome")
	}
	if dev.Auth == nil || dev.Auth.Username != "testAuthAdmin" {
		t.Error("dev.Auth should be set")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RegisterDeviceWithPlatform_InvalidName(t *testing.T) {
	m := setupManagerTest(t)

	// Register with invalid name (empty after normalization)
	err := m.RegisterDeviceWithPlatform("!!!", testDeviceIP, 0, "", "", "", nil)
	if err == nil {
		t.Error("expected error registering device with invalid name")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UnregisterDevice_WithDisplayName(t *testing.T) {
	m := setupManagerTest(t)

	// Register with display name
	if err := m.RegisterDevice("Master Bathroom", testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Unregister with display name
	if err := m.UnregisterDevice("Master Bathroom"); err != nil {
		t.Fatalf("UnregisterDevice() with display name error: %v", err)
	}

	// Verify it's gone
	_, ok := m.GetDevice("master-bathroom")
	if ok {
		t.Error("device should not exist after unregister")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UnregisterDevice_CleanupAliasesAndGroups(t *testing.T) {
	m := setupManagerTest(t)

	// Register device with alias and add to group
	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.AddDeviceAlias(testDeviceName, "kl"); err != nil {
		t.Fatalf("AddDeviceAlias() error: %v", err)
	}
	if err := m.CreateGroup("lights"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}
	if err := m.AddDeviceToGroup("lights", testDeviceName); err != nil {
		t.Fatalf("AddDeviceToGroup() error: %v", err)
	}

	// Unregister
	if err := m.UnregisterDevice(testDeviceName); err != nil {
		t.Fatalf("UnregisterDevice() error: %v", err)
	}

	// Verify device is gone
	_, ok := m.GetDevice(testDeviceName)
	if ok {
		t.Error("device should not exist after unregister")
	}

	// Group should still exist but device should be removed
	grp, ok := m.GetGroup("lights")
	if !ok {
		t.Fatal("group should still exist")
	}
	for _, d := range grp.Devices {
		if d == testDeviceName {
			t.Error("device should be removed from group")
		}
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UpdateDeviceAddress_SameAddress(t *testing.T) {
	m := setupManagerTest(t)

	// Register device
	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Update with same address (should be no-op)
	if err := m.UpdateDeviceAddress(testDeviceName, testDeviceIP); err != nil {
		t.Fatalf("UpdateDeviceAddress() same address error: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UpdateDeviceAddress_NotFound(t *testing.T) {
	m := setupManagerTest(t)

	err := m.UpdateDeviceAddress("nonexistent", "testDeviceIP00")
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RenameDevice_InvalidNewName(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.RegisterDevice("device1", testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Rename to invalid name
	err := m.RenameDevice("device1", "!!!")
	if err == nil {
		t.Error("expected error renaming to invalid name")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_GetDeviceAliases_NotFound(t *testing.T) {
	m := setupManagerTest(t)

	_, err := m.GetDeviceAliases("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_AddDeviceAlias_NotFound(t *testing.T) {
	m := setupManagerTest(t)

	err := m.AddDeviceAlias("nonexistent", "alias1")
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_AddDeviceAlias_InvalidAlias(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.RegisterDevice("device1", testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Try to add invalid alias
	err := m.AddDeviceAlias("device1", "-invalid")
	if err == nil {
		t.Error("expected error for invalid alias")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RemoveDeviceAlias_NotFound(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.RegisterDevice("device1", testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Try to remove nonexistent alias
	err := m.RemoveDeviceAlias("device1", "nonexistent")
	if err == nil {
		t.Error("expected error removing nonexistent alias")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_ResolveDevice_ByMAC(t *testing.T) {
	m := setupManagerTest(t)

	// Register device with MAC
	if err := m.RegisterDevice("device1", testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.UpdateDeviceInfo("device1", DeviceUpdates{MAC: "11:22:33:44:55:66"}); err != nil {
		t.Fatalf("UpdateDeviceInfo() error: %v", err)
	}

	// Resolve by MAC
	dev, err := m.ResolveDevice("11:22:33:44:55:66")
	if err != nil {
		t.Fatalf("ResolveDevice() by MAC error: %v", err)
	}
	if dev.Address != testDeviceIP {
		t.Errorf("dev.Address = %q, want %q", dev.Address, testDeviceIP)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_CheckAliasConflict_NoConflict(t *testing.T) {
	m := setupManagerTest(t)

	// No devices, no conflict
	err := m.CheckAliasConflict("new-alias", "")
	if err != nil {
		t.Errorf("CheckAliasConflict() should not return error: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_CheckAliasConflict_WithExclude(t *testing.T) {
	m := setupManagerTest(t)

	// Register device with alias
	if err := m.RegisterDevice("device1", testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.AddDeviceAlias("device1", "alias1"); err != nil {
		t.Fatalf("AddDeviceAlias() error: %v", err)
	}

	// Check conflict with exclude (same device)
	err := m.CheckAliasConflict("alias1", "device1")
	if err != nil {
		t.Errorf("CheckAliasConflict() with exclude should not return error: %v", err)
	}

	// Check conflict without exclude (different device)
	err = m.CheckAliasConflict("alias1", "")
	if err == nil {
		t.Error("CheckAliasConflict() without exclude should return error")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_GetGroupDevices_NotFound(t *testing.T) {
	m := setupManagerTest(t)

	_, err := m.GetGroupDevices("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent group")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RemoveDeviceFromGroup_NotFound(t *testing.T) {
	m := setupManagerTest(t)

	err := m.RemoveDeviceFromGroup("nonexistent", "device1")
	if err == nil {
		t.Error("expected error for nonexistent group")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RemoveDeviceFromGroup_DeviceNotInGroup(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.CreateGroup("lights"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}

	err := m.RemoveDeviceFromGroup("lights", "device-not-in-group")
	if err == nil {
		t.Error("expected error removing device not in group")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_DeleteGroup_NotFound(t *testing.T) {
	m := setupManagerTest(t)

	err := m.DeleteGroup("nonexistent")
	if err == nil {
		t.Error("expected error deleting nonexistent group")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RenameDevice_SameName(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Rename to same normalized name (different display name, same key)
	if err := m.RenameDevice(testDeviceName, "Kitchen"); err != nil {
		t.Errorf("RenameDevice() to same normalized key should succeed: %v", err)
	}

	// Check that device was renamed (display name updated)
	dev, ok := m.GetDevice(testDeviceName)
	if !ok {
		t.Error("device should still exist")
	}
	if dev.Name != "Kitchen" {
		t.Errorf("device name = %q, want %q", dev.Name, "Kitchen")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RenameDevice_AlreadyExists(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.RegisterDevice("living-room", "192.168.1.2", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Try to rename to existing device name
	err := m.RenameDevice(testDeviceName, "living-room")
	if err == nil {
		t.Error("expected error renaming to existing device name")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RemoveDeviceAlias_AliasNotFound(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Try to remove alias that doesn't exist
	err := m.RemoveDeviceAlias(testDeviceName, "nonexistent-alias")
	if err == nil {
		t.Error("expected error removing nonexistent alias")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_GetGroupDevices_FallbackToAddress(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.CreateGroup("test-group"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}

	// Add a device that doesn't exist (will fallback to using the identifier as address)
	m.mu.Lock()
	group := m.config.Groups["test-group"]
	group.Devices = append(group.Devices, "testDeviceIP00")
	m.config.Groups["test-group"] = group
	m.mu.Unlock()

	// GetGroupDevices should succeed, returning a device with the identifier as address
	devices, err := m.GetGroupDevices("test-group")
	if err != nil {
		t.Fatalf("GetGroupDevices() error: %v", err)
	}
	if len(devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(devices))
	}
	if devices[0].Address != "testDeviceIP00" {
		t.Errorf("device address = %q, want %q", devices[0].Address, "testDeviceIP00")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_CreateGroup_ValidationError(t *testing.T) {
	m := setupManagerTest(t)

	// Try to create group with invalid name
	err := m.CreateGroup("")
	if err == nil {
		t.Error("expected error creating group with empty name")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_ResolveDevice_ByAlias(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.RegisterDevice("kitchen-light", testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	if err := m.AddDeviceAlias("kitchen-light", "kl"); err != nil {
		t.Fatalf("AddDeviceAlias() error: %v", err)
	}

	// Resolve by alias
	dev, err := m.ResolveDevice("kl")
	if err != nil {
		t.Fatalf("ResolveDevice() error: %v", err)
	}
	if dev.Address != testDeviceIP {
		t.Errorf("resolved device address = %q, want %q", dev.Address, testDeviceIP)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_CheckAliasConflict_WithExcludedDevice(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.AddDeviceAlias(testDeviceName, "kl"); err != nil {
		t.Fatalf("AddDeviceAlias() error: %v", err)
	}

	// Check conflict with exclusion should not return error
	err := m.CheckAliasConflict("kl", testDeviceName)
	if err != nil {
		t.Errorf("CheckAliasConflict() should not error when excluding owning device: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UnregisterDevice_CleanupFromGroup(t *testing.T) {
	m := setupManagerTest(t)

	// Register device and add to group
	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.CreateGroup("lights"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}
	if err := m.AddDeviceToGroup("lights", testDeviceName); err != nil {
		t.Fatalf("AddDeviceToGroup() error: %v", err)
	}

	// Unregister device
	if err := m.UnregisterDevice(testDeviceName); err != nil {
		t.Fatalf("UnregisterDevice() error: %v", err)
	}

	// Device should be removed from group
	group, ok := m.GetGroup("lights")
	if !ok {
		t.Fatal("group should still exist")
	}
	for _, d := range group.Devices {
		if d == testDeviceName {
			t.Error("device should be removed from group")
		}
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RenameDevice_UpdatesGroupReferences(t *testing.T) {
	m := setupManagerTest(t)

	// Register device and add to group
	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.CreateGroup("lights"); err != nil {
		t.Fatalf("CreateGroup() error: %v", err)
	}
	if err := m.AddDeviceToGroup("lights", testDeviceName); err != nil {
		t.Fatalf("AddDeviceToGroup() error: %v", err)
	}

	// Rename device
	if err := m.RenameDevice(testDeviceName, "living-room"); err != nil {
		t.Fatalf("RenameDevice() error: %v", err)
	}

	// Group should now have the new device key
	group, ok := m.GetGroup("lights")
	if !ok {
		t.Fatal("group should still exist")
	}
	foundNewName := false
	for _, d := range group.Devices {
		if d == testDeviceName {
			t.Error("old device name should not be in group")
		}
		if d == "living-room" {
			foundNewName = true
		}
	}
	if !foundNewName {
		t.Error("new device name should be in group")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_AddDeviceAlias_ToDeviceWithExistingAliases(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Add first alias
	if err := m.AddDeviceAlias(testDeviceName, "k1"); err != nil {
		t.Fatalf("AddDeviceAlias() error: %v", err)
	}

	// Add second alias
	if err := m.AddDeviceAlias(testDeviceName, "k2"); err != nil {
		t.Fatalf("AddDeviceAlias() second error: %v", err)
	}

	// Verify both aliases exist
	aliases, err := m.GetDeviceAliases(testDeviceName)
	if err != nil {
		t.Fatalf("GetDeviceAliases() error: %v", err)
	}
	if len(aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(aliases))
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RemoveDeviceAlias_SingleAlias(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Add two aliases
	if err := m.AddDeviceAlias(testDeviceName, "k1"); err != nil {
		t.Fatalf("AddDeviceAlias() error: %v", err)
	}
	if err := m.AddDeviceAlias(testDeviceName, "k2"); err != nil {
		t.Fatalf("AddDeviceAlias() error: %v", err)
	}

	// Remove one alias
	if err := m.RemoveDeviceAlias(testDeviceName, "k1"); err != nil {
		t.Fatalf("RemoveDeviceAlias() error: %v", err)
	}

	// Verify only one alias remains
	aliases, err := m.GetDeviceAliases(testDeviceName)
	if err != nil {
		t.Fatalf("GetDeviceAliases() error: %v", err)
	}
	if len(aliases) != 1 {
		t.Errorf("expected 1 alias, got %d", len(aliases))
	}
	if aliases[0] != "k2" {
		t.Errorf("remaining alias = %q, want k2", aliases[0])
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_CheckAliasConflict_ConflictsWithName(t *testing.T) {
	m := setupManagerTest(t)

	// Register device with display name "Kitchen Light"
	if err := m.RegisterDevice("Kitchen Light", testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Try to use the display name as an alias for another device
	err := m.CheckAliasConflict("Kitchen Light", "")
	if err == nil {
		t.Error("expected conflict with device name")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_CheckAliasConflict_ConflictsWithOtherAlias(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.AddDeviceAlias(testDeviceName, "kl"); err != nil {
		t.Fatalf("AddDeviceAlias() error: %v", err)
	}

	// Register another device
	if err := m.RegisterDevice("living", "192.168.1.2", 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Try to use "kl" as alias for living room (conflicts with kitchen's alias)
	err := m.CheckAliasConflict("kl", "living")
	if err == nil {
		t.Error("expected conflict with other device's alias")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_AddDeviceAlias_DuplicateOnSameDevice(t *testing.T) {
	m := setupManagerTest(t)

	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Add alias
	if err := m.AddDeviceAlias(testDeviceName, "kl"); err != nil {
		t.Fatalf("AddDeviceAlias() error: %v", err)
	}

	// Try to add same alias again (case-insensitive)
	err := m.AddDeviceAlias(testDeviceName, "KL")
	if err == nil {
		t.Error("expected error adding duplicate alias")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_AddDeviceAlias_ByNormalizedName(t *testing.T) {
	m := setupManagerTest(t)

	// Register with display name
	if err := m.RegisterDevice("Kitchen Light", testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Add alias using normalized key
	if err := m.AddDeviceAlias("kitchen-light", "kl"); err != nil {
		t.Fatalf("AddDeviceAlias() by normalized key error: %v", err)
	}

	// Verify alias was added
	aliases, err := m.GetDeviceAliases("kitchen-light")
	if err != nil {
		t.Fatalf("GetDeviceAliases() error: %v", err)
	}
	if len(aliases) != 1 {
		t.Errorf("expected 1 alias, got %d", len(aliases))
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RemoveDeviceAlias_ByNormalizedName(t *testing.T) {
	m := setupManagerTest(t)

	// Register with display name
	if err := m.RegisterDevice("Kitchen Light", testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Add alias using display name
	if err := m.AddDeviceAlias("Kitchen Light", "kl"); err != nil {
		t.Fatalf("AddDeviceAlias() error: %v", err)
	}

	// Remove alias using normalized key
	if err := m.RemoveDeviceAlias("kitchen-light", "kl"); err != nil {
		t.Fatalf("RemoveDeviceAlias() by normalized key error: %v", err)
	}

	// Verify alias was removed
	aliases, err := m.GetDeviceAliases("kitchen-light")
	if err != nil {
		t.Fatalf("GetDeviceAliases() error: %v", err)
	}
	if len(aliases) != 0 {
		t.Errorf("expected 0 aliases, got %d", len(aliases))
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_RemoveDeviceAlias_ByDisplayName(t *testing.T) {
	m := setupManagerTest(t)

	// Register with display name (stored under normalized key)
	if err := m.RegisterDevice("Kitchen Light", testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Add alias
	if err := m.AddDeviceAlias("Kitchen Light", "kl"); err != nil {
		t.Fatalf("AddDeviceAlias() error: %v", err)
	}

	// Remove alias using display name (triggers normalized key lookup)
	if err := m.RemoveDeviceAlias("Kitchen Light", "kl"); err != nil {
		t.Fatalf("RemoveDeviceAlias() by display name error: %v", err)
	}

	// Verify alias was removed
	aliases, err := m.GetDeviceAliases("kitchen-light")
	if err != nil {
		t.Fatalf("GetDeviceAliases() error: %v", err)
	}
	if len(aliases) != 0 {
		t.Errorf("expected 0 aliases, got %d", len(aliases))
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UnregisterDevice_ByDisplayName(t *testing.T) {
	m := setupManagerTest(t)

	// Register with display name
	if err := m.RegisterDevice("Kitchen Light", testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Unregister using display name (triggers normalized key lookup)
	if err := m.UnregisterDevice("Kitchen Light"); err != nil {
		t.Fatalf("UnregisterDevice() by display name error: %v", err)
	}

	// Verify device was removed
	_, ok := m.GetDevice("kitchen-light")
	if ok {
		t.Error("device should not exist after unregister")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_ResolveDevice_ByDisplayName(t *testing.T) {
	m := setupManagerTest(t)

	// Register with specific display name
	if err := m.RegisterDevice("Kitchen Light", testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}

	// Resolve using a different case but same display name
	// This triggers the case-insensitive display name match path
	dev, err := m.ResolveDevice("KITCHEN LIGHT")
	if err != nil {
		t.Fatalf("ResolveDevice() error: %v", err)
	}
	if dev.Address != testDeviceIP {
		t.Errorf("resolved device address = %q, want %q", dev.Address, testDeviceIP)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_ResolveDevice_ByExactMAC(t *testing.T) {
	m := setupManagerTest(t)

	// Register device with MAC
	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.UpdateDeviceInfo(testDeviceName, DeviceUpdates{MAC: "AA:BB:CC:DD:EE:FF"}); err != nil {
		t.Fatalf("UpdateDeviceInfo() error: %v", err)
	}

	// Resolve using exact MAC (case-insensitive match)
	dev, err := m.ResolveDevice("aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("ResolveDevice() by exact MAC error: %v", err)
	}
	if dev.Address != testDeviceIP {
		t.Errorf("resolved device address = %q, want %q", dev.Address, testDeviceIP)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestManager_UnregisterDevice_CleanupFromMultipleGroups(t *testing.T) {
	m := setupManagerTest(t)

	// Register device and add to multiple groups
	if err := m.RegisterDevice(testDeviceName, testDeviceIP, 2, "", "", nil); err != nil {
		t.Fatalf("RegisterDevice() error: %v", err)
	}
	if err := m.CreateGroup("lights"); err != nil {
		t.Fatalf("CreateGroup() lights error: %v", err)
	}
	if err := m.CreateGroup("kitchen-devices"); err != nil {
		t.Fatalf("CreateGroup() kitchen-devices error: %v", err)
	}
	if err := m.AddDeviceToGroup("lights", testDeviceName); err != nil {
		t.Fatalf("AddDeviceToGroup() lights error: %v", err)
	}
	if err := m.AddDeviceToGroup("kitchen-devices", testDeviceName); err != nil {
		t.Fatalf("AddDeviceToGroup() kitchen-devices error: %v", err)
	}

	// Unregister device
	if err := m.UnregisterDevice(testDeviceName); err != nil {
		t.Fatalf("UnregisterDevice() error: %v", err)
	}

	// Verify device was removed from both groups
	group1, _ := m.GetGroup("lights")
	for _, d := range group1.Devices {
		if d == testDeviceName {
			t.Error("device should be removed from lights group")
		}
	}
	group2, _ := m.GetGroup("kitchen-devices")
	for _, d := range group2.Devices {
		if d == testDeviceName {
			t.Error("device should be removed from kitchen-devices group")
		}
	}
}
