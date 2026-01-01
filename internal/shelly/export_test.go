package shelly

import "github.com/tj-smith47/shelly-go/integrator"

// SetNewIntegratorClient overrides the integrator client factory for testing.
// Returns a function that restores the original factory.
func SetNewIntegratorClient(fn func(tag, token string) *integrator.Client) func() {
	old := newIntegratorClient
	newIntegratorClient = fn
	return func() { newIntegratorClient = old }
}

// SetNewFleetManager overrides the fleet manager factory for testing.
// Returns a function that restores the original factory.
func SetNewFleetManager(fn func(client *integrator.Client) *integrator.FleetManager) func() {
	old := newFleetManager
	newFleetManager = fn
	return func() { newFleetManager = old }
}
