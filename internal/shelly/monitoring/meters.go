package monitoring

import (
	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

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
