package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplayEMDataHistory(t *testing.T) {
	t.Parallel()

	t.Run("with data", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := &components.EMDataGetDataResult{
			Data: []components.EMDataBlock{
				{
					TS:     1700000000,
					Period: 60,
					Values: []components.EMDataValues{
						{
							TotalActivePower: 1500.5,
							AActivePower:     500.0,
							BActivePower:     500.0,
							CActivePower:     500.5,
							AVoltage:         230.0,
							BVoltage:         230.0,
							CVoltage:         230.0,
						},
					},
				},
			},
		}

		DisplayEMDataHistory(ios, data, 0, nil, nil, 0)

		output := out.String()
		if !strings.Contains(output, "Energy History") {
			t.Error("output should contain 'Energy History'")
		}
		if !strings.Contains(output, "1500.5") {
			t.Error("output should contain power value")
		}
	})

	t.Run("with time range", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := &components.EMDataGetDataResult{
			Data: []components.EMDataBlock{
				{
					TS:     1700000000,
					Period: 60,
					Values: []components.EMDataValues{
						{
							TotalActivePower: 1000.0,
							AVoltage:         230.0,
							BVoltage:         230.0,
							CVoltage:         230.0,
						},
					},
				},
			},
		}
		startTS := int64(1700000000)
		endTS := int64(1700003600)

		DisplayEMDataHistory(ios, data, 0, &startTS, &endTS, 0)

		output := out.String()
		if !strings.Contains(output, "From:") {
			t.Error("output should contain 'From:'")
		}
		if !strings.Contains(output, "To:") {
			t.Error("output should contain 'To:'")
		}
	})

	t.Run("empty data", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		data := &components.EMDataGetDataResult{
			Data: []components.EMDataBlock{},
		}

		DisplayEMDataHistory(ios, data, 0, nil, nil, 0)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No data available") {
			t.Errorf("output should contain warning, got %q", allOutput)
		}
	})

	t.Run("with limit", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := &components.EMDataGetDataResult{
			Data: []components.EMDataBlock{
				{
					TS:     1700000000,
					Period: 60,
					Values: []components.EMDataValues{
						{TotalActivePower: 100.0, AVoltage: 230.0, BVoltage: 230.0, CVoltage: 230.0},
						{TotalActivePower: 200.0, AVoltage: 230.0, BVoltage: 230.0, CVoltage: 230.0},
						{TotalActivePower: 300.0, AVoltage: 230.0, BVoltage: 230.0, CVoltage: 230.0},
					},
				},
			},
		}

		DisplayEMDataHistory(ios, data, 0, nil, nil, 2)

		output := out.String()
		if !strings.Contains(output, "showing first 2") {
			t.Error("output should contain limit message")
		}
	})
}

func TestDisplayEM1DataHistory(t *testing.T) {
	t.Parallel()

	t.Run("with data", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		pf := 0.95
		data := &components.EM1DataGetDataResult{
			Data: []components.EM1DataBlock{
				{
					TS:     1700000000,
					Period: 60,
					Values: []components.EM1DataValues{
						{
							ActivePower: 500.0,
							Voltage:     230.0,
							Current:     2.2,
							PowerFactor: &pf,
						},
					},
				},
			},
		}

		DisplayEM1DataHistory(ios, data, 0, nil, nil, 0)

		output := out.String()
		if !strings.Contains(output, "Energy History (EM1)") {
			t.Error("output should contain 'Energy History (EM1)'")
		}
		if !strings.Contains(output, "500.00W") {
			t.Error("output should contain power value")
		}
		if !strings.Contains(output, "PF:") {
			t.Error("output should contain power factor")
		}
	})

	t.Run("without power factor", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := &components.EM1DataGetDataResult{
			Data: []components.EM1DataBlock{
				{
					TS:     1700000000,
					Period: 60,
					Values: []components.EM1DataValues{
						{
							ActivePower: 500.0,
							Voltage:     230.0,
							Current:     2.2,
							PowerFactor: nil,
						},
					},
				},
			},
		}

		DisplayEM1DataHistory(ios, data, 0, nil, nil, 0)

		output := out.String()
		if strings.Contains(output, "PF:") {
			t.Error("output should not contain power factor when nil")
		}
	})

	t.Run("empty data", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		data := &components.EM1DataGetDataResult{
			Data: []components.EM1DataBlock{},
		}

		DisplayEM1DataHistory(ios, data, 0, nil, nil, 0)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No data available") {
			t.Errorf("output should contain warning, got %q", allOutput)
		}
	})

	t.Run("with limit", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := &components.EM1DataGetDataResult{
			Data: []components.EM1DataBlock{
				{
					TS:     1700000000,
					Period: 60,
					Values: []components.EM1DataValues{
						{ActivePower: 100.0, Voltage: 230.0, Current: 0.5},
						{ActivePower: 200.0, Voltage: 230.0, Current: 1.0},
						{ActivePower: 300.0, Voltage: 230.0, Current: 1.5},
					},
				},
			},
		}

		DisplayEM1DataHistory(ios, data, 0, nil, nil, 2)

		output := out.String()
		if !strings.Contains(output, "showing first 2") {
			t.Error("output should contain limit message")
		}
	})
}

func TestDisplayEMStatus(t *testing.T) {
	t.Parallel()

	t.Run("basic status", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		status := &model.EMStatus{
			ID:               0,
			AVoltage:         230.0,
			BVoltage:         231.0,
			CVoltage:         229.0,
			ACurrent:         5.0,
			BCurrent:         5.1,
			CCurrent:         4.9,
			TotalCurrent:     15.0,
			AActivePower:     1150.0,
			BActivePower:     1175.0,
			CActivePower:     1125.0,
			TotalActivePower: 3450.0,
			AApparentPower:   1200.0,
			BApparentPower:   1225.0,
			CApparentPower:   1175.0,
			TotalAprtPower:   3600.0,
		}

		DisplayEMStatus(ios, status)

		output := out.String()
		if !strings.Contains(output, "Energy Monitor") {
			t.Error("output should contain 'Energy Monitor'")
		}
		if !strings.Contains(output, "230.00 V") {
			t.Error("output should contain voltage")
		}
		if !strings.Contains(output, "3450.00 W") {
			t.Error("output should contain total power")
		}
	})

	t.Run("with power factor and frequency", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		pf := 0.95
		freq := 50.0
		status := &model.EMStatus{
			ID:           0,
			AVoltage:     230.0,
			BVoltage:     231.0,
			CVoltage:     229.0,
			APowerFactor: &pf,
			BPowerFactor: &pf,
			CPowerFactor: &pf,
			AFreq:        &freq,
			BFreq:        &freq,
			CFreq:        &freq,
		}

		DisplayEMStatus(ios, status)

		output := out.String()
		if !strings.Contains(output, "0.95") {
			t.Error("output should contain power factor")
		}
		if !strings.Contains(output, "50.00 Hz") {
			t.Error("output should contain frequency")
		}
	})

	t.Run("with neutral current", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		nc := 0.5
		status := &model.EMStatus{
			ID:       0,
			NCurrent: &nc,
		}

		DisplayEMStatus(ios, status)

		output := out.String()
		if !strings.Contains(output, "Neutral Current") {
			t.Error("output should contain neutral current")
		}
	})

	t.Run("with errors", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		status := &model.EMStatus{
			ID:     0,
			Errors: []string{"phase_error"},
		}

		DisplayEMStatus(ios, status)

		output := out.String()
		if !strings.Contains(output, "Errors") {
			t.Error("output should contain errors")
		}
	})
}

func TestDisplayEM1Status(t *testing.T) {
	t.Parallel()

	t.Run("basic status", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		status := &model.EM1Status{
			ID:        0,
			Voltage:   230.0,
			Current:   5.0,
			ActPower:  1100.0,
			AprtPower: 1150.0,
		}

		DisplayEM1Status(ios, status)

		output := out.String()
		if !strings.Contains(output, "Energy Monitor (EM1)") {
			t.Error("output should contain 'Energy Monitor (EM1)'")
		}
		if !strings.Contains(output, "230.00 V") {
			t.Error("output should contain voltage")
		}
		if !strings.Contains(output, "1100.00 W") {
			t.Error("output should contain power")
		}
	})

	t.Run("with power factor and frequency", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		pf := 0.96
		freq := 60.0
		status := &model.EM1Status{
			ID:      0,
			Voltage: 120.0,
			Current: 10.0,
			PF:      &pf,
			Freq:    &freq,
		}

		DisplayEM1Status(ios, status)

		output := out.String()
		if !strings.Contains(output, "0.96") {
			t.Error("output should contain power factor")
		}
		if !strings.Contains(output, "60.00 Hz") {
			t.Error("output should contain frequency")
		}
	})

	t.Run("with errors", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		status := &model.EM1Status{
			ID:     0,
			Errors: []string{"overload"},
		}

		DisplayEM1Status(ios, status)

		output := out.String()
		if !strings.Contains(output, "Errors") {
			t.Error("output should contain errors")
		}
	})
}
