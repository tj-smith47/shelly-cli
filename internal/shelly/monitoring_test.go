package shelly

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestGetEMDataCSVURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		resolver  DeviceResolver
		device    string
		id        int
		startTS   *int64
		endTS     *int64
		addKeys   bool
		wantURL   string
		wantError bool
	}{
		{
			name: "basic URL without parameters",
			resolver: &mockResolver{
				device: model.Device{
					Name:       "test-device",
					Address:    "192.168.1.100",
					Generation: 2,
				},
			},
			device:  "test-device",
			id:      0,
			wantURL: "http://192.168.1.100/emdata/0/data.csv?",
		},
		{
			name: "URL with start timestamp",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  "test-device",
			id:      0,
			startTS: int64Ptr(1609459200),
			wantURL: "http://192.168.1.100/emdata/0/data.csv?ts=1609459200",
		},
		{
			name: "URL with both timestamps",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  "test-device",
			id:      1,
			startTS: int64Ptr(1609459200),
			endTS:   int64Ptr(1609545600),
			wantURL: "http://192.168.1.100/emdata/1/data.csv?ts=1609459200&end_ts=1609545600",
		},
		{
			name: "URL with add_keys",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  "test-device",
			id:      0,
			addKeys: true,
			wantURL: "http://192.168.1.100/emdata/0/data.csv?add_keys=true",
		},
		{
			name: "URL with all parameters",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  "test-device",
			id:      2,
			startTS: int64Ptr(1609459200),
			endTS:   int64Ptr(1609545600),
			addKeys: true,
			wantURL: "http://192.168.1.100/emdata/2/data.csv?ts=1609459200&end_ts=1609545600&add_keys=true",
		},
		{
			name: "unknown device",
			resolver: &mockResolver{
				err: model.ErrDeviceNotFound,
			},
			device:    "unknown",
			id:        0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &Service{resolver: tt.resolver}
			url, err := svc.GetEMDataCSVURL(tt.device, tt.id, tt.startTS, tt.endTS, tt.addKeys)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if url != tt.wantURL {
				t.Errorf("GetEMDataCSVURL() = %q, want %q", url, tt.wantURL)
			}
		})
	}
}

func TestGetEM1DataCSVURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		resolver  DeviceResolver
		device    string
		id        int
		startTS   *int64
		endTS     *int64
		addKeys   bool
		wantURL   string
		wantError bool
	}{
		{
			name: "basic URL without parameters",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  "test-device",
			id:      0,
			wantURL: "http://192.168.1.100/em1data/0/data.csv?",
		},
		{
			name: "URL with start timestamp",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  "test-device",
			id:      0,
			startTS: int64Ptr(1609459200),
			wantURL: "http://192.168.1.100/em1data/0/data.csv?ts=1609459200",
		},
		{
			name: "URL with both timestamps",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  "test-device",
			id:      1,
			startTS: int64Ptr(1609459200),
			endTS:   int64Ptr(1609545600),
			wantURL: "http://192.168.1.100/em1data/1/data.csv?ts=1609459200&end_ts=1609545600",
		},
		{
			name: "URL with add_keys",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  "test-device",
			id:      0,
			addKeys: true,
			wantURL: "http://192.168.1.100/em1data/0/data.csv?add_keys=true",
		},
		{
			name: "URL with all parameters",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  "test-device",
			id:      2,
			startTS: int64Ptr(1609459200),
			endTS:   int64Ptr(1609545600),
			addKeys: true,
			wantURL: "http://192.168.1.100/em1data/2/data.csv?ts=1609459200&end_ts=1609545600&add_keys=true",
		},
		{
			name: "unknown device",
			resolver: &mockResolver{
				err: model.ErrDeviceNotFound,
			},
			device:    "unknown",
			id:        0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &Service{resolver: tt.resolver}
			url, err := svc.GetEM1DataCSVURL(tt.device, tt.id, tt.startTS, tt.endTS, tt.addKeys)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if url != tt.wantURL {
				t.Errorf("GetEM1DataCSVURL() = %q, want %q", url, tt.wantURL)
			}
		})
	}
}

// int64Ptr returns a pointer to the given int64 value.
func int64Ptr(v int64) *int64 {
	return &v
}
