package monitoring

import (
	"github.com/tj-smith47/shelly-go/gen2/components"
)

// energyDataBlock represents a block of energy data with period and power values.
type energyDataBlock interface {
	GetPeriod() int
	GetPowerValues() []float64
}

type emDataBlockAdapter struct{ b components.EMDataBlock }

func (a emDataBlockAdapter) GetPeriod() int { return a.b.Period }
func (a emDataBlockAdapter) GetPowerValues() []float64 {
	powers := make([]float64, len(a.b.Values))
	for i, v := range a.b.Values {
		powers[i] = v.TotalActivePower
	}
	return powers
}

type em1DataBlockAdapter struct{ b components.EM1DataBlock }

func (a em1DataBlockAdapter) GetPeriod() int { return a.b.Period }
func (a em1DataBlockAdapter) GetPowerValues() []float64 {
	powers := make([]float64, len(a.b.Values))
	for i, v := range a.b.Values {
		powers[i] = v.ActivePower
	}
	return powers
}

// calculateMetrics is the generic implementation for energy metrics calculation.
func calculateMetrics(blocks []energyDataBlock) (energy, avgPower, peakPower float64, dataPoints int) {
	var totalPower float64
	for _, block := range blocks {
		for _, power := range block.GetPowerValues() {
			totalPower += power
			if power > peakPower {
				peakPower = power
			}
			energy += power * float64(block.GetPeriod()) / 3600.0
			dataPoints++
		}
	}
	if dataPoints > 0 {
		avgPower = totalPower / float64(dataPoints)
	}
	energy /= 1000.0 // Wh to kWh
	return
}

// CalculateEMMetrics calculates energy metrics from EM data history.
func CalculateEMMetrics(data *components.EMDataGetDataResult) (energy, avgPower, peakPower float64, dataPoints int) {
	blocks := make([]energyDataBlock, len(data.Data))
	for i, b := range data.Data {
		blocks[i] = emDataBlockAdapter{b}
	}
	return calculateMetrics(blocks)
}

// CalculateEM1Metrics calculates energy metrics from EM1 data history.
func CalculateEM1Metrics(data *components.EM1DataGetDataResult) (energy, avgPower, peakPower float64, dataPoints int) {
	blocks := make([]energyDataBlock, len(data.Data))
	for i, b := range data.Data {
		blocks[i] = em1DataBlockAdapter{b}
	}
	return calculateMetrics(blocks)
}
