package testutil

import (
	"context"
	"sync/atomic"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// Resolver is a generation-aware device resolver for tests. An identifier
// listed in Devices resolves to that device; any other resolves to Device.
// Every call returns Err when it is set.
type Resolver struct {
	Device  model.Device
	Devices map[string]model.Device
	Err     error
	calls   atomic.Int32
}

// Gen2At returns a Resolver mapping every identifier to the Gen2 device at addr.
func Gen2At(addr string) *Resolver {
	return &Resolver{Device: model.Device{Address: addr, Generation: 2}}
}

// Resolve returns the device for id.
func (r *Resolver) Resolve(id string) (model.Device, error) {
	r.calls.Add(1)
	if d, ok := r.Devices[id]; ok {
		return d, r.Err
	}
	return r.Device, r.Err
}

// ResolveWithGeneration is Resolve; the generation is already set on the device.
func (r *Resolver) ResolveWithGeneration(_ context.Context, id string) (model.Device, error) {
	return r.Resolve(id)
}

// Calls returns how many times the resolver was asked to resolve.
func (r *Resolver) Calls() int {
	return int(r.calls.Load())
}
