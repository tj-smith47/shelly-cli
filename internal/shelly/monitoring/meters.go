package monitoring

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// maxComponentID is the maximum component ID to probe when discovering components.
const maxComponentID = 10

// GetEMStatus returns the status of an Energy Monitor (EM) component.
func (s *Service) GetEMStatus(ctx context.Context, device string, id int) (*model.EMStatus, error) {
	var result *model.EMStatus
	err := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		em := components.NewEM(conn.RPCClient(), id)
		status, err := em.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = emStatusFromComponent(status)
		return nil
	})
	return result, err
}

// GetEM1Status returns the status of a single-phase Energy Monitor (EM1) component.
func (s *Service) GetEM1Status(ctx context.Context, device string, id int) (*model.EM1Status, error) {
	var result *model.EM1Status
	err := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		em1 := components.NewEM1(conn.RPCClient(), id)
		status, err := em1.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = em1StatusFromComponent(status)
		return nil
	})
	return result, err
}

// GetPMStatus returns the status of a Power Meter (PM) component.
func (s *Service) GetPMStatus(ctx context.Context, device string, id int) (*model.PMStatus, error) {
	var result *model.PMStatus
	err := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		pm := components.NewPM(conn.RPCClient(), id)
		status, err := pm.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = pmStatusFromComponent(status)
		return nil
	})
	return result, err
}

// GetPM1Status returns the status of a Power Meter (PM1) component.
func (s *Service) GetPM1Status(ctx context.Context, device string, id int) (*model.PMStatus, error) {
	var result *model.PMStatus
	err := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		pm1 := components.NewPM1(conn.RPCClient(), id)
		status, err := pm1.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = pm1StatusFromComponent(status)
		return nil
	})
	return result, err
}

// ResetEMCounters resets energy counters on an EM component.
func (s *Service) ResetEMCounters(ctx context.Context, device string, id int, counterTypes []string) error {
	return s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		em := components.NewEM(conn.RPCClient(), id)
		return em.ResetCounters(ctx, counterTypes)
	})
}

// ResetPMCounters resets energy counters on a PM component.
func (s *Service) ResetPMCounters(ctx context.Context, device string, id int, counterTypes []string) error {
	return s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		pm := components.NewPM(conn.RPCClient(), id)
		return pm.ResetCounters(ctx, counterTypes)
	})
}

// ResetPM1Counters resets energy counters on a PM1 component.
func (s *Service) ResetPM1Counters(ctx context.Context, device string, id int, counterTypes []string) error {
	return s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		pm1 := components.NewPM1(conn.RPCClient(), id)
		return pm1.ResetCounters(ctx, counterTypes)
	})
}

// ListEMComponents returns a list of EM component IDs on a device.
func (s *Service) ListEMComponents(ctx context.Context, device string) ([]int, error) {
	var ids []int
	err := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		ids = discoverEMComponents(ctx, conn)
		return nil
	})
	return ids, err
}

// ListEM1Components returns a list of EM1 component IDs on a device.
func (s *Service) ListEM1Components(ctx context.Context, device string) ([]int, error) {
	var ids []int
	err := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		ids = discoverEM1Components(ctx, conn)
		return nil
	})
	return ids, err
}

// ListPMComponents returns a list of PM component IDs on a device.
func (s *Service) ListPMComponents(ctx context.Context, device string) ([]int, error) {
	var ids []int
	err := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		ids = discoverPMComponents(ctx, conn)
		return nil
	})
	return ids, err
}

// ListPM1Components returns a list of PM1 component IDs on a device.
func (s *Service) ListPM1Components(ctx context.Context, device string) ([]int, error) {
	var ids []int
	err := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		ids = discoverPM1Components(ctx, conn)
		return nil
	})
	return ids, err
}

// Discovery helpers

func discoverEMComponents(ctx context.Context, conn *client.Client) []int {
	ids := make([]int, 0, maxComponentID)
	for id := range maxComponentID {
		em := components.NewEM(conn.RPCClient(), id)
		if _, err := em.GetStatus(ctx); err != nil {
			break
		}
		ids = append(ids, id)
	}
	return ids
}

func discoverEM1Components(ctx context.Context, conn *client.Client) []int {
	ids := make([]int, 0, maxComponentID)
	for id := range maxComponentID {
		em1 := components.NewEM1(conn.RPCClient(), id)
		if _, err := em1.GetStatus(ctx); err != nil {
			break
		}
		ids = append(ids, id)
	}
	return ids
}

func discoverPMComponents(ctx context.Context, conn *client.Client) []int {
	ids := make([]int, 0, maxComponentID)
	for id := range maxComponentID {
		pm := components.NewPM(conn.RPCClient(), id)
		if _, err := pm.GetStatus(ctx); err != nil {
			break
		}
		ids = append(ids, id)
	}
	return ids
}

func discoverPM1Components(ctx context.Context, conn *client.Client) []int {
	ids := make([]int, 0, maxComponentID)
	for id := range maxComponentID {
		pm1 := components.NewPM1(conn.RPCClient(), id)
		if _, err := pm1.GetStatus(ctx); err != nil {
			break
		}
		ids = append(ids, id)
	}
	return ids
}

// Component conversion helpers

// emStatusFromComponent converts shelly-go EMStatus to model.EMStatus.
func emStatusFromComponent(c *components.EMStatus) *model.EMStatus {
	return &model.EMStatus{
		ID:               c.ID,
		AVoltage:         c.AVoltage,
		ACurrent:         c.ACurrent,
		AActivePower:     c.AActivePower,
		AApparentPower:   c.AApparentPower,
		APowerFactor:     c.APowerFactor,
		AFreq:            c.AFreq,
		BVoltage:         c.BVoltage,
		BCurrent:         c.BCurrent,
		BActivePower:     c.BActivePower,
		BApparentPower:   c.BApparentPower,
		BPowerFactor:     c.BPowerFactor,
		BFreq:            c.BFreq,
		CVoltage:         c.CVoltage,
		CCurrent:         c.CCurrent,
		CActivePower:     c.CActivePower,
		CApparentPower:   c.CApparentPower,
		CPowerFactor:     c.CPowerFactor,
		CFreq:            c.CFreq,
		NCurrent:         c.NCurrent,
		TotalCurrent:     c.TotalCurrent,
		TotalActivePower: c.TotalActivePower,
		TotalAprtPower:   c.TotalApparentPower,
		Errors:           c.Errors,
	}
}

// em1StatusFromComponent converts shelly-go EM1Status to model.EM1Status.
func em1StatusFromComponent(c *components.EM1Status) *model.EM1Status {
	return &model.EM1Status{
		ID:        c.ID,
		Voltage:   c.Voltage,
		Current:   c.Current,
		ActPower:  c.ActPower,
		AprtPower: c.AprtPower,
		PF:        c.PF,
		Freq:      c.Freq,
		Errors:    c.Errors,
	}
}

// pmStatusFromComponent converts shelly-go PMStatus to model.PMStatus.
func pmStatusFromComponent(c *components.PMStatus) *model.PMStatus {
	result := &model.PMStatus{
		ID:      c.ID,
		Voltage: c.Voltage,
		Current: c.Current,
		APower:  c.APower,
		Freq:    c.Freq,
		Errors:  c.Errors,
	}
	if c.AEnergy != nil {
		result.AEnergy = &model.PMEnergyCounters{
			Total:    c.AEnergy.Total,
			ByMinute: c.AEnergy.ByMinute,
			MinuteTs: c.AEnergy.MinuteTs,
		}
	}
	if c.RetAEnergy != nil {
		result.RetAEnergy = &model.PMEnergyCounters{
			Total:    c.RetAEnergy.Total,
			ByMinute: c.RetAEnergy.ByMinute,
			MinuteTs: c.RetAEnergy.MinuteTs,
		}
	}
	return result
}

// pm1StatusFromComponent converts shelly-go PM1Status to model.PMStatus.
func pm1StatusFromComponent(c *components.PM1Status) *model.PMStatus {
	result := &model.PMStatus{
		ID:      c.ID,
		Voltage: c.Voltage,
		Current: c.Current,
		APower:  c.APower,
		Freq:    c.Freq,
		Errors:  c.Errors,
	}
	if c.AEnergy != nil {
		result.AEnergy = &model.PMEnergyCounters{
			Total:    c.AEnergy.Total,
			ByMinute: c.AEnergy.ByMinute,
			MinuteTs: c.AEnergy.MinuteTs,
		}
	}
	if c.RetAEnergy != nil {
		result.RetAEnergy = &model.PMEnergyCounters{
			Total:    c.RetAEnergy.Total,
			ByMinute: c.RetAEnergy.ByMinute,
			MinuteTs: c.RetAEnergy.MinuteTs,
		}
	}
	return result
}
