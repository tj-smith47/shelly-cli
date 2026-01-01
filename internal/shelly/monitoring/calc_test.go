package monitoring

import (
	"testing"

	"github.com/tj-smith47/shelly-go/gen2/components"
)

func TestEMDataBlockAdapter(t *testing.T) {
	t.Parallel()

	block := components.EMDataBlock{
		Period: 120,
		Values: []components.EMDataValues{
			{TotalActivePower: 50.0},
			{TotalActivePower: 100.0},
		},
	}

	adapter := emDataBlockAdapter{b: block}

	if adapter.GetPeriod() != 120 {
		t.Errorf("GetPeriod() = %d, want 120", adapter.GetPeriod())
	}

	powers := adapter.GetPowerValues()
	if len(powers) != 2 {
		t.Fatalf("GetPowerValues() len = %d, want 2", len(powers))
	}
	if powers[0] != 50.0 || powers[1] != 100.0 {
		t.Errorf("GetPowerValues() = %v, want [50.0, 100.0]", powers)
	}
}

func TestEM1DataBlockAdapter(t *testing.T) {
	t.Parallel()

	block := components.EM1DataBlock{
		Period: 180,
		Values: []components.EM1DataValues{
			{ActivePower: 25.0},
			{ActivePower: 75.0},
		},
	}

	adapter := em1DataBlockAdapter{b: block}

	if adapter.GetPeriod() != 180 {
		t.Errorf("GetPeriod() = %d, want 180", adapter.GetPeriod())
	}

	powers := adapter.GetPowerValues()
	if len(powers) != 2 {
		t.Fatalf("GetPowerValues() len = %d, want 2", len(powers))
	}
	if powers[0] != 25.0 || powers[1] != 75.0 {
		t.Errorf("GetPowerValues() = %v, want [25.0, 75.0]", powers)
	}
}

// mockEnergyBlock is a test helper that implements energyDataBlock.
type mockEnergyBlock struct {
	period int
	powers []float64
}

func (m mockEnergyBlock) GetPeriod() int            { return m.period }
func (m mockEnergyBlock) GetPowerValues() []float64 { return m.powers }

//nolint:gocyclo // Test function with multiple subtests is inherently complex.
func TestCalculateMetrics(t *testing.T) {
	t.Parallel()

	t.Run("calculates metrics from data", func(t *testing.T) {
		t.Parallel()

		blocks := []energyDataBlock{
			mockEnergyBlock{period: 60, powers: []float64{100.0, 200.0, 150.0}},
		}

		energy, avgPower, peakPower, dataPoints := calculateMetrics(blocks)

		if dataPoints != 3 {
			t.Errorf("dataPoints = %d, want 3", dataPoints)
		}
		if peakPower != 200.0 {
			t.Errorf("peakPower = %f, want 200.0", peakPower)
		}
		// avgPower = (100 + 200 + 150) / 3 = 150
		if avgPower != 150.0 {
			t.Errorf("avgPower = %f, want 150.0", avgPower)
		}
		// energy = sum(power * period/3600) / 1000 = (100*60/3600 + 200*60/3600 + 150*60/3600) / 1000
		// = (1.667 + 3.333 + 2.5) / 1000 = 0.0075 kWh
		if energy < 0.007 || energy > 0.008 {
			t.Errorf("energy = %f, expected ~0.0075 kWh", energy)
		}
	})

	t.Run("handles empty data", func(t *testing.T) {
		t.Parallel()

		blocks := []energyDataBlock{}

		energy, avgPower, peakPower, dataPoints := calculateMetrics(blocks)

		if dataPoints != 0 {
			t.Errorf("dataPoints = %d, want 0", dataPoints)
		}
		if energy != 0 || avgPower != 0 || peakPower != 0 {
			t.Error("expected all zeros for empty data")
		}
	})

	t.Run("handles multiple blocks with different periods", func(t *testing.T) {
		t.Parallel()

		blocks := []energyDataBlock{
			mockEnergyBlock{period: 60, powers: []float64{100.0}},
			mockEnergyBlock{period: 120, powers: []float64{200.0}},
		}

		_, _, peakPower, dataPoints := calculateMetrics(blocks)

		if dataPoints != 2 {
			t.Errorf("dataPoints = %d, want 2", dataPoints)
		}
		if peakPower != 200.0 {
			t.Errorf("peakPower = %f, want 200.0", peakPower)
		}
	})

	t.Run("handles empty powers in blocks", func(t *testing.T) {
		t.Parallel()

		blocks := []energyDataBlock{
			mockEnergyBlock{period: 60, powers: []float64{}},
		}

		energy, avgPower, peakPower, dataPoints := calculateMetrics(blocks)

		if dataPoints != 0 {
			t.Errorf("dataPoints = %d, want 0", dataPoints)
		}
		if energy != 0 || avgPower != 0 || peakPower != 0 {
			t.Error("expected all zeros for empty powers")
		}
	})

	t.Run("finds peak power correctly", func(t *testing.T) {
		t.Parallel()

		blocks := []energyDataBlock{
			mockEnergyBlock{period: 60, powers: []float64{10.0, 50.0, 30.0}},
			mockEnergyBlock{period: 60, powers: []float64{25.0, 75.0, 40.0}},
		}

		_, _, peakPower, _ := calculateMetrics(blocks)

		if peakPower != 75.0 {
			t.Errorf("peakPower = %f, want 75.0", peakPower)
		}
	})
}

func TestCalculateEMMetrics(t *testing.T) {
	t.Parallel()

	t.Run("calculates from EM data", func(t *testing.T) {
		t.Parallel()

		data := &components.EMDataGetDataResult{
			Data: []components.EMDataBlock{
				{
					Period: 60,
					Values: []components.EMDataValues{
						{TotalActivePower: 100.0},
						{TotalActivePower: 200.0},
						{TotalActivePower: 150.0},
					},
				},
			},
		}

		energy, avgPower, peakPower, dataPoints := CalculateEMMetrics(data)

		if dataPoints != 3 {
			t.Errorf("dataPoints = %d, want 3", dataPoints)
		}
		if peakPower != 200.0 {
			t.Errorf("peakPower = %f, want 200.0", peakPower)
		}
		if avgPower != 150.0 {
			t.Errorf("avgPower = %f, want 150.0", avgPower)
		}
		if energy < 0.007 || energy > 0.008 {
			t.Errorf("energy = %f, expected ~0.0075 kWh", energy)
		}
	})

	t.Run("handles empty data", func(t *testing.T) {
		t.Parallel()

		data := &components.EMDataGetDataResult{
			Data: []components.EMDataBlock{},
		}

		energy, avgPower, peakPower, dataPoints := CalculateEMMetrics(data)

		if dataPoints != 0 || energy != 0 || avgPower != 0 || peakPower != 0 {
			t.Error("expected all zeros for empty data")
		}
	})
}

func TestCalculateEM1Metrics(t *testing.T) {
	t.Parallel()

	t.Run("calculates from EM1 data", func(t *testing.T) {
		t.Parallel()

		data := &components.EM1DataGetDataResult{
			Data: []components.EM1DataBlock{
				{
					Period: 60,
					Values: []components.EM1DataValues{
						{ActivePower: 50.0},
						{ActivePower: 100.0},
						{ActivePower: 75.0},
					},
				},
			},
		}

		energy, avgPower, peakPower, dataPoints := CalculateEM1Metrics(data)

		if dataPoints != 3 {
			t.Errorf("dataPoints = %d, want 3", dataPoints)
		}
		if peakPower != 100.0 {
			t.Errorf("peakPower = %f, want 100.0", peakPower)
		}
		if avgPower != 75.0 {
			t.Errorf("avgPower = %f, want 75.0", avgPower)
		}
		// energy should be small positive value
		if energy <= 0 {
			t.Errorf("energy = %f, expected positive value", energy)
		}
	})

	t.Run("handles empty data", func(t *testing.T) {
		t.Parallel()

		data := &components.EM1DataGetDataResult{
			Data: []components.EM1DataBlock{},
		}

		energy, avgPower, peakPower, dataPoints := CalculateEM1Metrics(data)

		if dataPoints != 0 || energy != 0 || avgPower != 0 || peakPower != 0 {
			t.Error("expected all zeros for empty data")
		}
	})
}
