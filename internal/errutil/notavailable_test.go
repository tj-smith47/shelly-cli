package errutil

import (
	"errors"
	"fmt"
	"testing"

	"github.com/tj-smith47/shelly-go/rpc"
)

func TestNotAvailable(t *testing.T) {
	t.Parallel()

	noHandler := &rpc.ErrorObject{Code: 404, Message: "No handler for Thermostat.GetStatus"}
	other := errors.New("connection refused")

	tests := []struct {
		name    string
		feature string
		err     error
		want    string
	}{
		{"named from the device answer", "", noHandler,
			"Thermostat not available on this device: RPC error 404: No handler for Thermostat.GetStatus"},
		{"wrapped with %w", "", fmt.Errorf("failed to get status: %w", noHandler),
			"Thermostat not available on this device: failed to get status: RPC error 404: No handler for Thermostat.GetStatus"},
		{"explicit feature", "LoRa", &rpc.ErrorObject{Code: -106, Message: "method not found"},
			"LoRa not available on this device: RPC error -106: method not found"},
		{"JSON-RPC code, no method name", "", &rpc.ErrorObject{Code: -32601, Message: "Method not found"},
			"This feature not available on this device: RPC error -32601: Method not found"},
		{"other RPC error unchanged", "", &rpc.ErrorObject{Code: -103, Message: "bad id"}, "RPC error -103: bad id"},
		{"non-RPC error unchanged", "", other, "connection refused"},
		{"wrapped with %v is lost", "", errors.New("failed: " + noHandler.Error()),
			"failed: RPC error 404: No handler for Thermostat.GetStatus"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NotAvailable(tt.feature, tt.err)
			if got.Error() != tt.want {
				t.Errorf("NotAvailable() = %q, want %q", got.Error(), tt.want)
			}
			if !errors.Is(got, tt.err) {
				t.Error("NotAvailable() dropped the original error from the chain")
			}
		})
	}
}

func TestNotAvailable_DoesNotWrapTwice(t *testing.T) {
	t.Parallel()

	once := NotAvailable("", &rpc.ErrorObject{Code: 404, Message: "No handler for Zigbee.GetStatus"})
	if twice := NotAvailable("Zigbee", once); twice.Error() != once.Error() {
		t.Errorf("second NotAvailable = %q, want the first unchanged", twice)
	}
}
