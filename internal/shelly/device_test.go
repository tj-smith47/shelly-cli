package shelly

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
)

const (
	testModel = "SNSW-001P16EU"
)

func TestDeviceInfo_Fields(t *testing.T) {
	t.Parallel()

	info := DeviceInfo{
		ID:         "shellypro1pm-123456",
		MAC:        testMAC,
		Model:      testModel,
		Generation: 2,
		Firmware:   "1.0.0",
		App:        "Pro1PM",
		AuthEn:     true,
	}

	if info.ID != "shellypro1pm-123456" {
		t.Errorf("ID = %q, want shellypro1pm-123456", info.ID)
	}
	if info.MAC != testMAC {
		t.Errorf("MAC = %q, want %s", info.MAC, testMAC)
	}
	if info.Model != testModel {
		t.Errorf("Model = %q, want %s", info.Model, testModel)
	}
	if info.Generation != 2 {
		t.Errorf("Generation = %d, want 2", info.Generation)
	}
	if info.Firmware != "1.0.0" {
		t.Errorf("Firmware = %q, want 1.0.0", info.Firmware)
	}
	if info.App != "Pro1PM" {
		t.Errorf("App = %q, want Pro1PM", info.App)
	}
	if !info.AuthEn {
		t.Error("AuthEn = false, want true")
	}
}

func TestDeviceStatus_Fields(t *testing.T) {
	t.Parallel()

	status := DeviceStatus{
		Info: &DeviceInfo{
			ID:         "shellyplus1-123456",
			Generation: 2,
		},
		Status: map[string]any{
			"sys": map[string]any{
				"uptime": 12345,
			},
		},
	}

	if status.Info == nil {
		t.Fatal("Info is nil")
	}
	if status.Info.ID != "shellyplus1-123456" {
		t.Errorf("Info.ID = %q, want shellyplus1-123456", status.Info.ID)
	}
	if status.Status == nil {
		t.Fatal("Status is nil")
	}
	if _, ok := status.Status["sys"]; !ok {
		t.Error("Status missing 'sys' key")
	}
}

func TestDeviceInfo_ZeroValues(t *testing.T) {
	t.Parallel()

	info := DeviceInfo{}

	if info.ID != "" {
		t.Errorf("ID = %q, want empty", info.ID)
	}
	if info.Generation != 0 {
		t.Errorf("Generation = %d, want 0", info.Generation)
	}
	if info.AuthEn {
		t.Error("AuthEn = true, want false")
	}
}

func TestDeviceStatus_NilInfo(t *testing.T) {
	t.Parallel()

	status := DeviceStatus{
		Info:   nil,
		Status: map[string]any{},
	}

	if status.Info != nil {
		t.Error("Info should be nil")
	}
}

// TestWithGenAwareRestart_ResolveError covers the gen-detection guard: a device that
// cannot even be resolved was never restarted, so the error must surface honestly
// rather than be swallowed as a dropped-connection "success".
func TestWithGenAwareRestart_ResolveError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("device not found")
	svc := New(&generationAwareResolver{err: wantErr})
	err := svc.DeviceReboot(context.Background(), "ghost", 0)
	if !errors.Is(err, wantErr) {
		t.Fatalf("DeviceReboot err = %v, want resolve error %v", err, wantErr)
	}
}

// TestWithGenAwareRestart_ConnectivityDropIsSuccess covers the success signal: with a
// live context, a connectivity error from the restart action means the device tore
// the connection down as it restarted, which is reported as success (nil). The
// resolver supplies the generation, so gen-detection passes without a device round
// trip; the action then dials a refusing address and gets "connection refused".
//
// One generation suffices: the gen2 branch's WithConnection wiring is exercised by
// TestWithGenAwareRestart_NonConnectivityErrorIsFailure/gen2. This path lets the SDK
// run its full connect-retry budget (it must, so the context stays live and the
// connectivity error reaches the success classifier), so it is intentionally kept to
// a single connection rather than multiplied across generations and operations.
func TestWithGenAwareRestart_ConnectivityDropIsSuccess(t *testing.T) {
	t.Parallel()

	addr := refusingAddr(t)
	svc := New(&generationAwareResolver{device: deviceAt(addr, 1)},
		WithRateLimiter(ratelimit.New()))

	if err := svc.DeviceReboot(context.Background(), "dev", 0); err != nil {
		t.Errorf("DeviceReboot: connectivity drop should read as success, got %v", err)
	}
}

// TestWithGenAwareRestart_NonConnectivityErrorIsFailure covers the real-failure
// branch: a reachable device that refuses the action (HTTP 500 / RPC error, not a
// dropped connection) returns a non-connectivity error, which must be propagated as
// a genuine failure rather than masked as a restart.
func TestWithGenAwareRestart_NonConnectivityErrorIsFailure(t *testing.T) {
	t.Parallel()

	t.Run("gen1", func(t *testing.T) {
		t.Parallel()
		d := newAPDevServer(t, 1)
		d.gen1.rebootErr = true
		svc := apdevService(d, 1)
		if err := svc.DeviceReboot(context.Background(), "dev", 0); err == nil {
			t.Fatal("DeviceReboot: a refused reboot must surface as an error, not a masked success")
		}
		if atomic.LoadInt32(&d.gen1.rebootHits) == 0 {
			t.Error("the device's /reboot endpoint was never hit")
		}
	})

	t.Run("gen1-factory-reset", func(t *testing.T) {
		t.Parallel()
		d := newAPDevServer(t, 1)
		d.gen1.resetErr = true
		svc := apdevService(d, 1)
		if err := svc.DeviceFactoryReset(context.Background(), "dev"); err == nil {
			t.Fatal("DeviceFactoryReset: a refused reset must surface as an error")
		}
		if atomic.LoadInt32(&d.gen1.resetHits) == 0 {
			t.Error("the device's /settings?reset=true endpoint was never hit")
		}
	})

	t.Run("gen2", func(t *testing.T) {
		t.Parallel()
		d := newAPDevServer(t, 2)
		d.gen2.rebootErr = true
		svc := apdevService(d, 2)
		if err := svc.DeviceReboot(context.Background(), "dev", 250); err == nil {
			t.Fatal("DeviceReboot gen2: an RPC-error reboot must surface as an error")
		}
		if atomic.LoadInt32(&d.gen2.rebootHits) == 0 {
			t.Error("the device's Shelly.Reboot RPC was never hit")
		}
	})
}

// TestDeviceFactoryReset_Gen2Success covers the Gen2 factory-reset closure: a reachable
// Gen2 device that accepts the Shelly.FactoryReset RPC returns success (nil). This is
// the generation branch the connectivity-drop and non-connectivity Gen1 cases do not
// reach, since they never run the Gen2 action against a live device.
func TestDeviceFactoryReset_Gen2Success(t *testing.T) {
	t.Parallel()
	d := newAPDevServer(t, 2)
	svc := apdevService(d, 2)

	if err := svc.DeviceFactoryReset(context.Background(), "dev"); err != nil {
		t.Fatalf("DeviceFactoryReset gen2 (reachable, accepts reset): %v", err)
	}
}

// TestWithGenAwareRestart_CancelledContextErrors covers the precedence guard: when the
// caller's own context is cancelled, the operation was aborted and the error must be
// surfaced — never swallowed as a successful restart — even though a cancellation
// reads as a connectivity failure. The cancelled context short-circuits gen-detection,
// so the same honest error is returned for both reboot and factory-reset.
func TestWithGenAwareRestart_CancelledContextErrors(t *testing.T) {
	t.Parallel()

	addr := refusingAddr(t)
	svc := New(&generationAwareResolver{device: deviceAt(addr, 1)},
		WithRateLimiter(ratelimit.New()))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := svc.DeviceReboot(ctx, "dev", 0); err == nil {
		t.Error("DeviceReboot with a cancelled context must error, not report a phantom restart")
	}
	if err := svc.DeviceFactoryReset(ctx, "dev"); err == nil {
		t.Error("DeviceFactoryReset with a cancelled context must error")
	}
}

// deviceAt builds a resolved device pinned to addr and generation, so gen-detection
// reads the generation from the resolver without a device round trip.
func deviceAt(addr string, gen int) model.Device {
	return model.Device{Name: "dev", Address: addr, Generation: gen}
}
