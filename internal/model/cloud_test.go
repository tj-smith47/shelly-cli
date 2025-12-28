package model

import "testing"

func TestCloudEvent_GetDeviceID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		event  CloudEvent
		wantID string
	}{
		{
			name:   "prefer DeviceID over Device",
			event:  CloudEvent{DeviceID: "device-123", Device: "device-456"},
			wantID: "device-123",
		},
		{
			name:   "fall back to Device when DeviceID empty",
			event:  CloudEvent{DeviceID: "", Device: "device-456"},
			wantID: "device-456",
		},
		{
			name:   "DeviceID only",
			event:  CloudEvent{DeviceID: "device-123"},
			wantID: "device-123",
		},
		{
			name:   "Device only",
			event:  CloudEvent{Device: "device-456"},
			wantID: "device-456",
		},
		{
			name:   "both empty",
			event:  CloudEvent{},
			wantID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.event.GetDeviceID()
			if got != tt.wantID {
				t.Errorf("GetDeviceID() = %q, want %q", got, tt.wantID)
			}
		})
	}
}

func TestCloudEvent_Fields(t *testing.T) {
	t.Parallel()

	online := 1
	event := CloudEvent{
		Event:     "online",
		DeviceID:  "shelly-1pm-abc123",
		Device:    "shellyplug-s-def456",
		Online:    &online,
		Timestamp: 1699999999,
	}

	if event.Event != "online" {
		t.Errorf("Event = %q, want %q", event.Event, "online")
	}
	if event.DeviceID != "shelly-1pm-abc123" {
		t.Errorf("DeviceID = %q, want %q", event.DeviceID, "shelly-1pm-abc123")
	}
	if event.Device != "shellyplug-s-def456" {
		t.Errorf("Device = %q, want %q", event.Device, "shellyplug-s-def456")
	}
	if event.Online == nil || *event.Online != 1 {
		t.Errorf("Online = %v, want 1", event.Online)
	}
	if event.Timestamp != 1699999999 {
		t.Errorf("Timestamp = %d, want 1699999999", event.Timestamp)
	}
}
