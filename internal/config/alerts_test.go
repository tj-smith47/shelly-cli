package config

import (
	"testing"
	"time"

	"github.com/spf13/afero"
)

func TestAlert_IsSnoozed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		snoozedTime string
		want        bool
	}{
		{
			name:        "empty snooze time",
			snoozedTime: "",
			want:        false,
		},
		{
			name:        "future snooze time",
			snoozedTime: time.Now().Add(time.Hour).Format(time.RFC3339),
			want:        true,
		},
		{
			name:        "past snooze time",
			snoozedTime: time.Now().Add(-time.Hour).Format(time.RFC3339),
			want:        false,
		},
		{
			name:        "invalid time format",
			snoozedTime: "invalid-time",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := Alert{
				Name:         "test",
				SnoozedUntil: tt.snoozedTime,
			}
			if got := a.IsSnoozed(); got != tt.want {
				t.Errorf("IsSnoozed() = %v, want %v", got, tt.want)
			}
		})
	}
}

// setupAlertsTest sets up an isolated environment for alert package-level function tests.
func setupAlertsTest(t *testing.T) {
	t.Helper()
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })
	ResetDefaultManagerForTesting()
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_CreateAlert(t *testing.T) {
	setupAlertsTest(t)

	err := CreateAlert("test-alert", "Test description", "device1", "offline", "notify", true)
	if err != nil {
		t.Errorf("CreateAlert() error = %v", err)
	}

	// Verify alert was created
	alert, ok := GetAlert("test-alert")
	if !ok {
		t.Fatal("GetAlert() should find created alert")
	}
	if alert.Name != "test-alert" {
		t.Errorf("alert.Name = %q, want %q", alert.Name, "test-alert")
	}
	if alert.Device != "device1" {
		t.Errorf("alert.Device = %q, want %q", alert.Device, "device1")
	}
	if alert.Condition != "offline" {
		t.Errorf("alert.Condition = %q, want %q", alert.Condition, "offline")
	}
	if !alert.Enabled {
		t.Error("alert.Enabled should be true")
	}
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_ListAlerts(t *testing.T) {
	setupAlertsTest(t)

	// Initially empty
	alerts := ListAlerts()
	if len(alerts) != 0 {
		t.Errorf("ListAlerts() should be empty, got %d", len(alerts))
	}

	// Create some alerts
	if err := CreateAlert("alert1", "", "dev1", "offline", "notify", true); err != nil {
		t.Fatalf("CreateAlert() error = %v", err)
	}
	if err := CreateAlert("alert2", "", "dev2", "power>100", "webhook:http://example.com", false); err != nil {
		t.Fatalf("CreateAlert() error = %v", err)
	}

	alerts = ListAlerts()
	if len(alerts) != 2 {
		t.Errorf("ListAlerts() should return 2 alerts, got %d", len(alerts))
	}
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_DeleteAlert(t *testing.T) {
	setupAlertsTest(t)

	// Create an alert
	if err := CreateAlert("to-delete", "", "dev1", "offline", "notify", true); err != nil {
		t.Fatalf("CreateAlert() error = %v", err)
	}

	// Delete it
	if err := DeleteAlert("to-delete"); err != nil {
		t.Errorf("DeleteAlert() error = %v", err)
	}

	// Verify it's gone
	_, ok := GetAlert("to-delete")
	if ok {
		t.Error("GetAlert() should not find deleted alert")
	}

	// Delete non-existent should error
	if err := DeleteAlert("nonexistent"); err == nil {
		t.Error("DeleteAlert() should error for non-existent alert")
	}
}

//nolint:paralleltest // Tests modify global state
func TestPackageLevel_UpdateAlert(t *testing.T) {
	setupAlertsTest(t)

	// Create an alert
	if err := CreateAlert("to-update", "", "dev1", "offline", "notify", true); err != nil {
		t.Fatalf("CreateAlert() error = %v", err)
	}

	// Update enabled to false
	enabled := false
	if err := UpdateAlert("to-update", &enabled, ""); err != nil {
		t.Errorf("UpdateAlert() error = %v", err)
	}

	alert, ok := GetAlert("to-update")
	if !ok {
		t.Fatal("GetAlert() should find alert")
	}
	if alert.Enabled {
		t.Error("alert.Enabled should be false after update")
	}

	// Update snooze time
	snoozeTime := time.Now().Add(time.Hour).Format(time.RFC3339)
	if err := UpdateAlert("to-update", nil, snoozeTime); err != nil {
		t.Errorf("UpdateAlert() error = %v", err)
	}

	alert, _ = GetAlert("to-update")
	if alert.SnoozedUntil != snoozeTime {
		t.Errorf("alert.SnoozedUntil = %q, want %q", alert.SnoozedUntil, snoozeTime)
	}

	// Update non-existent should error
	if err := UpdateAlert("nonexistent", &enabled, ""); err == nil {
		t.Error("UpdateAlert() should error for non-existent alert")
	}
}
