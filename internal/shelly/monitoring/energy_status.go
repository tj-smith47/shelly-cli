package monitoring

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// Energy monitor component types reported in model.EnergyStatusEntry.Type.
const (
	EnergyTypeEM  = "em"
	EnergyTypeEM1 = "em1"
)

// CollectEnergyStatuses reads every EM and EM1 component on each device
// concurrently. Entries follow the order of devices. A Gen1 device, a device
// that cannot be reached or has no EM or EM1 component, and a component that
// fails to read are reported in skipped; the other devices are unaffected.
func (s *Service) CollectEnergyStatuses(ctx context.Context, devices []string) (entries []model.EnergyStatusEntry, skipped []model.EnergyStatusSkip) {
	perEntries := make([][]model.EnergyStatusEntry, len(devices))
	perSkipped := make([][]model.EnergyStatusSkip, len(devices))

	var g errgroup.Group
	g.SetLimit(config.GetGlobalMaxConcurrent())
	for i, device := range devices {
		g.Go(func() error {
			perEntries[i], perSkipped[i] = s.collectDeviceEnergyStatus(ctx, device)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, nil
	}

	entries = make([]model.EnergyStatusEntry, 0, len(devices))
	for i := range devices {
		entries = append(entries, perEntries[i]...)
		skipped = append(skipped, perSkipped[i]...)
	}
	return entries, skipped
}

func (s *Service) collectDeviceEnergyStatus(ctx context.Context, device string) ([]model.EnergyStatusEntry, []model.EnergyStatusSkip) {
	// EM and EM1 are Gen2+ RPC components; a Gen1 device answers their probe
	// with HTTP 404, which would otherwise read as "unreachable".
	if resolved, err := s.connector.ResolveWithGeneration(ctx, device); err == nil && resolved.Generation == 1 {
		return nil, []model.EnergyStatusSkip{{Device: device, Reason: "Gen1 device; energy status reads Gen2+ EM and EM1 components only"}}
	}
	emIDs, err := s.ListEMComponents(ctx, device)
	if err != nil {
		return nil, []model.EnergyStatusSkip{{Device: device, Reason: fmt.Sprintf("unreachable: %v", err)}}
	}
	em1IDs, err := s.ListEM1Components(ctx, device)
	if err != nil {
		return nil, []model.EnergyStatusSkip{{Device: device, Reason: fmt.Sprintf("unreachable: %v", err)}}
	}
	if len(emIDs) == 0 && len(em1IDs) == 0 {
		return nil, []model.EnergyStatusSkip{{Device: device, Reason: "no energy monitor (EM or EM1) component"}}
	}

	var entries []model.EnergyStatusEntry
	var skipped []model.EnergyStatusSkip
	for _, id := range emIDs {
		st, err := s.GetEMStatus(ctx, device, id)
		if err != nil {
			skipped = append(skipped, model.EnergyStatusSkip{Device: device, Reason: fmt.Sprintf("em:%d: %v", id, err)})
			continue
		}
		entries = append(entries, model.EnergyStatusEntry{Name: device, Type: EnergyTypeEM, ID: id, Power: st.TotalActivePower, EM: st})
	}
	for _, id := range em1IDs {
		st, err := s.GetEM1Status(ctx, device, id)
		if err != nil {
			skipped = append(skipped, model.EnergyStatusSkip{Device: device, Reason: fmt.Sprintf("em1:%d: %v", id, err)})
			continue
		}
		entries = append(entries, model.EnergyStatusEntry{Name: device, Type: EnergyTypeEM1, ID: id, Power: st.ActPower, EM1: st})
	}
	return entries, skipped
}
