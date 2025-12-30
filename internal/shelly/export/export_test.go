// Package export provides export format builders for device data.
package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

const (
	testAddress = "192.168.1.101"
)

func TestBuildAnsibleInventory(t *testing.T) {
	t.Parallel()

	devices := []model.DeviceData{
		{
			Name:       "device1",
			Address:    testAddress,
			Model:      "SNSW-002P16EU",
			Generation: 2,
			App:        "Plus2PM",
		},
		{
			Name:       "device2",
			Address:    "192.168.1.102",
			Model:      "SNSW-002P16EU",
			Generation: 2,
			App:        "Plus2PM",
		},
		{
			Name:       "device3",
			Address:    "192.168.1.103",
			Model:      "SHSW-1",
			Generation: 1,
			App:        "",
		},
	}

	inventory, data, err := BuildAnsibleInventory(devices, "shelly")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inventory == nil {
		t.Fatal("expected non-nil inventory")
	}
	if data == nil {
		t.Fatal("expected non-nil data")
	}

	// Check structure
	if inventory.All.Children == nil {
		t.Fatal("expected non-nil Children")
	}

	shellyGroup, ok := inventory.All.Children["shelly"]
	if !ok {
		t.Fatal("expected 'shelly' group")
	}

	if len(shellyGroup.Hosts) != 3 {
		t.Errorf("got %d hosts, want 3", len(shellyGroup.Hosts))
	}

	// Check that subgroups were created by model
	if len(shellyGroup.Children) < 2 {
		t.Errorf("expected at least 2 model subgroups, got %d", len(shellyGroup.Children))
	}
}

func TestNormalizeGroupName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"SNSW-002P16EU", "snsw_002p16eu"},
		{"SHSW-1", "shsw_1"},
		{"Plus 2PM", "plus_2pm"},
		{"model.name", "model_name"},
		{"simple", "simple"},
		{"Mixed-Case Model", "mixed_case_model"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got := NormalizeGroupName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeGroupName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildTerraformConfig(t *testing.T) {
	t.Parallel()

	devices := []model.DeviceData{
		{
			Name:       "living-room",
			Address:    testAddress,
			Model:      "SNSW-002P16EU",
			Generation: 2,
			App:        "Plus2PM",
		},
		{
			Name:       "kitchen",
			Address:    "192.168.1.102",
			Model:      "SHSW-1",
			Generation: 1,
			App:        "",
		},
	}

	config, err := BuildTerraformConfig(devices, "shelly_devices")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config == "" {
		t.Fatal("expected non-empty config")
	}

	// Check expected content
	if !strings.Contains(config, "shelly_devices") {
		t.Error("expected config to contain resource name")
	}
	if !strings.Contains(config, "living_room") {
		t.Error("expected config to contain normalized device name")
	}
	if !strings.Contains(config, testAddress) {
		t.Error("expected config to contain device address")
	}
	if !strings.Contains(config, "Plus2PM") {
		t.Error("expected config to contain app name")
	}
}

func TestNormalizeResourceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"Living Room", "living_room"},
		{"kitchen-light", "kitchen_light"},
		{"device.name", "device_name"},
		{"simple", "simple"},
		{"Mixed-Case Name", "mixed_case_name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got := NormalizeResourceName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeResourceName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatEMDataCSV(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    *components.EMDataGetDataResult
		wantErr bool
	}{
		{
			name:    "nil data",
			data:    nil,
			wantErr: true,
		},
		{
			name: "empty data",
			data: &components.EMDataGetDataResult{
				Data: []components.EMDataBlock{},
			},
			wantErr: true,
		},
		{
			name: "valid data",
			data: &components.EMDataGetDataResult{
				Data: []components.EMDataBlock{
					{
						TS:     1700000000,
						Period: 60,
						Values: []components.EMDataValues{
							{
								AVoltage:     230.5,
								ACurrent:     1.5,
								AActivePower: 345.75,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := FormatEMDataCSV(tt.data)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if data == nil {
				t.Error("expected non-nil data")
			}
		})
	}
}

func TestFormatEM1DataCSV(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    *components.EM1DataGetDataResult
		wantErr bool
	}{
		{
			name:    "nil data",
			data:    nil,
			wantErr: true,
		},
		{
			name: "empty data",
			data: &components.EM1DataGetDataResult{
				Data: []components.EM1DataBlock{},
			},
			wantErr: true,
		},
		{
			name: "valid data",
			data: &components.EM1DataGetDataResult{
				Data: []components.EM1DataBlock{
					{
						TS:     1700000000,
						Period: 60,
						Values: []components.EM1DataValues{
							{
								Voltage:     230.5,
								Current:     1.5,
								ActivePower: 345.75,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := FormatEM1DataCSV(tt.data)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if data == nil {
				t.Error("expected non-nil data")
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"with space", "with_space"},
		{"with/slash", "with_slash"},
		{"with\\backslash", "with_backslash"},
		{"with:colon", "with_colon"},
		{"with*star", "with_star"},
		{"with?question", "with_question"},
		{"with\"quote", "with_quote"},
		{"with<less", "with_less"},
		{"with>greater", "with_greater"},
		{"with|pipe", "with_pipe"},
		{"complex/with:many*bad?chars", "complex_with_many_bad_chars"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got := SanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestWriteBackupFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	bkp := &backup.DeviceBackup{}

	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"json format", "json", false},
		{"yaml format", "yaml", false},
		{"yml format", "yml", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(tmpDir, "backup-"+tt.format+"."+tt.format)
			err := WriteBackupFile(bkp, filePath, tt.format)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify file was created
			if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
				t.Error("expected file to be created")
			}
		})
	}
}

func TestIsBackupFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{"backup.json", true},
		{"backup.yaml", true},
		{"backup.yml", true},
		{"backup.txt", false},
		{"backup", false},
		{"file.JSON", false}, // case sensitive
		{"config.json.bak", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsBackupFile(tt.name)
			if got != tt.want {
				t.Errorf("IsBackupFile(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestMarshalBackup(t *testing.T) {
	t.Parallel()

	bkp := &backup.DeviceBackup{}

	// Test JSON format
	jsonData, err := MarshalBackup(bkp, "json")
	if err != nil {
		t.Errorf("unexpected error for JSON: %v", err)
	}
	if jsonData == nil {
		t.Error("expected non-nil JSON data")
	}

	// Test YAML format
	yamlData, err := MarshalBackup(bkp, "yaml")
	if err != nil {
		t.Errorf("unexpected error for YAML: %v", err)
	}
	if yamlData == nil {
		t.Error("expected non-nil YAML data")
	}

	// Test YML format
	ymlData, err := MarshalBackup(bkp, "yml")
	if err != nil {
		t.Errorf("unexpected error for YML: %v", err)
	}
	if ymlData == nil {
		t.Error("expected non-nil YML data")
	}
}

func TestAnsibleHost_Fields(t *testing.T) {
	t.Parallel()

	host := AnsibleHost{
		AnsibleHost: testAddress,
		ShellyModel: "SNSW-002P16EU",
		ShellyGen:   2,
		ShellyApp:   "Plus2PM",
	}

	if host.AnsibleHost != testAddress {
		t.Errorf("got AnsibleHost=%q, want %q", host.AnsibleHost, testAddress)
	}
	if host.ShellyModel != "SNSW-002P16EU" {
		t.Errorf("got ShellyModel=%q, want %q", host.ShellyModel, "SNSW-002P16EU")
	}
	if host.ShellyGen != 2 {
		t.Errorf("got ShellyGen=%d, want 2", host.ShellyGen)
	}
	if host.ShellyApp != "Plus2PM" {
		t.Errorf("got ShellyApp=%q, want %q", host.ShellyApp, "Plus2PM")
	}
}

func TestTerraformDevice_Fields(t *testing.T) {
	t.Parallel()

	device := TerraformDevice{
		Name:       "living_room",
		Address:    testAddress,
		Model:      "SNSW-002P16EU",
		Generation: 2,
		App:        "Plus2PM",
	}

	if device.Name != "living_room" {
		t.Errorf("got Name=%q, want %q", device.Name, "living_room")
	}
	if device.Address != testAddress {
		t.Errorf("got Address=%q, want %q", device.Address, testAddress)
	}
	if device.Model != "SNSW-002P16EU" {
		t.Errorf("got Model=%q, want %q", device.Model, "SNSW-002P16EU")
	}
	if device.Generation != 2 {
		t.Errorf("got Generation=%d, want 2", device.Generation)
	}
	if device.App != "Plus2PM" {
		t.Errorf("got App=%q, want %q", device.App, "Plus2PM")
	}
}

func TestFormatConstants(t *testing.T) {
	t.Parallel()

	if FormatCSV != "csv" {
		t.Errorf("got FormatCSV=%q, want %q", FormatCSV, "csv")
	}
	if FormatJSON != "json" {
		t.Errorf("got FormatJSON=%q, want %q", FormatJSON, "json")
	}
	if FormatYAML != "yaml" { //nolint:goconst // Testing constant value, FormatYAML constant exists
		t.Errorf("got FormatYAML=%q, want %q", FormatYAML, "yaml")
	}
}
