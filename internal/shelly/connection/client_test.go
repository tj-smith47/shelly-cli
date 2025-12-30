package connection

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

func TestDeviceClient_Gen1(t *testing.T) {
	t.Parallel()

	t.Run("IsGen1 with gen1 client", func(t *testing.T) {
		t.Parallel()

		// Create a mock Gen1Client (nil pointer for testing purposes, but we're testing the wrapper)
		dc := NewGen1Client(&client.Gen1Client{})

		if !dc.IsGen1() {
			t.Error("expected IsGen1() to be true")
		}
		if dc.IsGen2() {
			t.Error("expected IsGen2() to be false")
		}
	})

	t.Run("IsGen2 with gen2 client", func(t *testing.T) {
		t.Parallel()

		dc := NewGen2Client(&client.Client{})

		if dc.IsGen1() {
			t.Error("expected IsGen1() to be false")
		}
		if !dc.IsGen2() {
			t.Error("expected IsGen2() to be true")
		}
	})

	t.Run("neither gen1 nor gen2", func(t *testing.T) {
		t.Parallel()

		dc := &DeviceClient{}

		if dc.IsGen1() {
			t.Error("expected IsGen1() to be false")
		}
		if dc.IsGen2() {
			t.Error("expected IsGen2() to be false")
		}
	})
}

func TestDeviceClient_Close(t *testing.T) {
	t.Parallel()

	t.Run("close empty client", func(t *testing.T) {
		t.Parallel()

		dc := &DeviceClient{}
		err := dc.Close()

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestDeviceClient_Info(t *testing.T) {
	t.Parallel()

	t.Run("info with empty client", func(t *testing.T) {
		t.Parallel()

		dc := &DeviceClient{}
		info := dc.Info()

		if info != nil {
			t.Error("expected nil info for empty client")
		}
	})
}

func TestDeviceClient_Generation(t *testing.T) {
	t.Parallel()

	t.Run("generation with empty client", func(t *testing.T) {
		t.Parallel()

		dc := &DeviceClient{}
		gen := dc.Generation()

		if gen != 0 {
			t.Errorf("expected generation 0 for empty client, got %d", gen)
		}
	})
}

func TestDeviceClient_Gen1Panic(t *testing.T) {
	t.Parallel()

	t.Run("Gen1 panics when not gen1 connection", func(t *testing.T) {
		t.Parallel()

		dc := NewGen2Client(&client.Client{})

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when calling Gen1() on Gen2 connection")
			}
		}()

		dc.Gen1()
	})
}

func TestDeviceClient_Gen2Panic(t *testing.T) {
	t.Parallel()

	t.Run("Gen2 panics when not gen2 connection", func(t *testing.T) {
		t.Parallel()

		dc := NewGen1Client(&client.Gen1Client{})

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when calling Gen2() on Gen1 connection")
			}
		}()

		dc.Gen2()
	})
}

func TestDeviceClient_Gen1NoClientPanic(t *testing.T) {
	t.Parallel()

	t.Run("Gen1 panics with empty client", func(t *testing.T) {
		t.Parallel()

		dc := &DeviceClient{}

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when calling Gen1() on empty connection")
			}
		}()

		dc.Gen1()
	})
}

func TestDeviceClient_Gen2NoClientPanic(t *testing.T) {
	t.Parallel()

	t.Run("Gen2 panics with empty client", func(t *testing.T) {
		t.Parallel()

		dc := &DeviceClient{}

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when calling Gen2() on empty connection")
			}
		}()

		dc.Gen2()
	})
}
