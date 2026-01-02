package shelly

import (
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
)

func TestGetEMDataCSVURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		resolver  DeviceResolver
		device    string
		id        int
		startTS   *int64
		endTS     *int64
		addKeys   bool
		wantURL   string
		wantError bool
	}{
		{
			name: "basic URL without parameters",
			resolver: &mockResolver{
				device: model.Device{
					Name:       testDevice,
					Address:    "192.168.1.100",
					Generation: 2,
				},
			},
			device:  testDevice,
			id:      0,
			wantURL: "http://192.168.1.100/emdata/0/data.csv?",
		},
		{
			name: "URL with start timestamp",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  testDevice,
			id:      0,
			startTS: int64Ptr(1609459200),
			wantURL: "http://192.168.1.100/emdata/0/data.csv?ts=1609459200",
		},
		{
			name: "URL with both timestamps",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  testDevice,
			id:      1,
			startTS: int64Ptr(1609459200),
			endTS:   int64Ptr(1609545600),
			wantURL: "http://192.168.1.100/emdata/1/data.csv?ts=1609459200&end_ts=1609545600",
		},
		{
			name: "URL with add_keys",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  testDevice,
			id:      0,
			addKeys: true,
			wantURL: "http://192.168.1.100/emdata/0/data.csv?add_keys=true",
		},
		{
			name: "URL with all parameters",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  testDevice,
			id:      2,
			startTS: int64Ptr(1609459200),
			endTS:   int64Ptr(1609545600),
			addKeys: true,
			wantURL: "http://192.168.1.100/emdata/2/data.csv?ts=1609459200&end_ts=1609545600&add_keys=true",
		},
		{
			name: "unknown device",
			resolver: &mockResolver{
				err: model.ErrDeviceNotFound,
			},
			device:    "unknown",
			id:        0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &Service{resolver: tt.resolver}
			url, err := svc.GetEMDataCSVURL(tt.device, tt.id, tt.startTS, tt.endTS, tt.addKeys)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if url != tt.wantURL {
				t.Errorf("GetEMDataCSVURL() = %q, want %q", url, tt.wantURL)
			}
		})
	}
}

func TestGetEM1DataCSVURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		resolver  DeviceResolver
		device    string
		id        int
		startTS   *int64
		endTS     *int64
		addKeys   bool
		wantURL   string
		wantError bool
	}{
		{
			name: "basic URL without parameters",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  testDevice,
			id:      0,
			wantURL: "http://192.168.1.100/em1data/0/data.csv?",
		},
		{
			name: "URL with start timestamp",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  testDevice,
			id:      0,
			startTS: int64Ptr(1609459200),
			wantURL: "http://192.168.1.100/em1data/0/data.csv?ts=1609459200",
		},
		{
			name: "URL with both timestamps",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  testDevice,
			id:      1,
			startTS: int64Ptr(1609459200),
			endTS:   int64Ptr(1609545600),
			wantURL: "http://192.168.1.100/em1data/1/data.csv?ts=1609459200&end_ts=1609545600",
		},
		{
			name: "URL with add_keys",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  testDevice,
			id:      0,
			addKeys: true,
			wantURL: "http://192.168.1.100/em1data/0/data.csv?add_keys=true",
		},
		{
			name: "URL with all parameters",
			resolver: &mockResolver{
				device: model.Device{
					Address: "192.168.1.100",
				},
			},
			device:  testDevice,
			id:      2,
			startTS: int64Ptr(1609459200),
			endTS:   int64Ptr(1609545600),
			addKeys: true,
			wantURL: "http://192.168.1.100/em1data/2/data.csv?ts=1609459200&end_ts=1609545600&add_keys=true",
		},
		{
			name: "unknown device",
			resolver: &mockResolver{
				err: model.ErrDeviceNotFound,
			},
			device:    "unknown",
			id:        0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &Service{resolver: tt.resolver}
			url, err := svc.GetEM1DataCSVURL(tt.device, tt.id, tt.startTS, tt.endTS, tt.addKeys)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if url != tt.wantURL {
				t.Errorf("GetEM1DataCSVURL() = %q, want %q", url, tt.wantURL)
			}
		})
	}
}

// int64Ptr returns a pointer to the given int64 value.
func int64Ptr(v int64) *int64 {
	return &v
}

func TestEMStatus_Fields(t *testing.T) {
	t.Parallel()

	pf := 0.95
	freq := 50.0
	ncurrent := 0.1

	status := model.EMStatus{
		ID:               0,
		AVoltage:         230.5,
		ACurrent:         10.0,
		AActivePower:     2300.0,
		AApparentPower:   2400.0,
		APowerFactor:     &pf,
		AFreq:            &freq,
		BVoltage:         231.0,
		BCurrent:         9.5,
		BActivePower:     2200.0,
		BApparentPower:   2300.0,
		CVoltage:         229.5,
		CCurrent:         10.5,
		CActivePower:     2400.0,
		CApparentPower:   2500.0,
		NCurrent:         &ncurrent,
		TotalCurrent:     30.0,
		TotalActivePower: 6900.0,
		TotalAprtPower:   7200.0,
		Errors:           []string{"overpower"},
	}

	if status.ID != 0 {
		t.Errorf("ID = %d, want 0", status.ID)
	}
	if status.AVoltage != 230.5 {
		t.Errorf("AVoltage = %f, want 230.5", status.AVoltage)
	}
	if status.TotalActivePower != 6900.0 {
		t.Errorf("TotalActivePower = %f, want 6900.0", status.TotalActivePower)
	}
	if len(status.Errors) != 1 {
		t.Errorf("Errors length = %d, want 1", len(status.Errors))
	}
}

func TestEM1Status_Fields(t *testing.T) {
	t.Parallel()

	pf := 0.98
	freq := 60.0

	status := model.EM1Status{
		ID:        0,
		Voltage:   120.5,
		Current:   5.0,
		ActPower:  600.0,
		AprtPower: 620.0,
		PF:        &pf,
		Freq:      &freq,
		Errors:    nil,
	}

	if status.ID != 0 {
		t.Errorf("ID = %d, want 0", status.ID)
	}
	if status.Voltage != 120.5 {
		t.Errorf("Voltage = %f, want 120.5", status.Voltage)
	}
	if status.ActPower != 600.0 {
		t.Errorf("ActPower = %f, want 600.0", status.ActPower)
	}
	if *status.PF != 0.98 {
		t.Errorf("PF = %f, want 0.98", *status.PF)
	}
}

func TestPMStatus_Fields(t *testing.T) {
	t.Parallel()

	freq := 50.0
	aenergy := &model.PMEnergyCounters{
		Total:    12345.67,
		ByMinute: []float64{10.0, 10.5, 11.0},
	}

	status := model.PMStatus{
		ID:      0,
		Voltage: 230.0,
		Current: 5.5,
		APower:  1250.0,
		Freq:    &freq,
		AEnergy: aenergy,
	}

	if status.Voltage != 230.0 {
		t.Errorf("Voltage = %f, want 230.0", status.Voltage)
	}
	if status.AEnergy == nil {
		t.Fatal("AEnergy is nil")
	}
	if status.AEnergy.Total != 12345.67 {
		t.Errorf("AEnergy.Total = %f, want 12345.67", status.AEnergy.Total)
	}
}

func TestPMEnergyCounters_Fields(t *testing.T) {
	t.Parallel()

	ts := int64(1234567890)
	ec := model.PMEnergyCounters{
		Total:    5000.0,
		ByMinute: []float64{10.0, 12.0, 8.0},
		MinuteTs: &ts,
	}

	if ec.Total != 5000.0 {
		t.Errorf("Total = %f, want 5000.0", ec.Total)
	}
	if len(ec.ByMinute) != 3 {
		t.Errorf("ByMinute length = %d, want 3", len(ec.ByMinute))
	}
	if *ec.MinuteTs != 1234567890 {
		t.Errorf("MinuteTs = %d, want 1234567890", *ec.MinuteTs)
	}
}

func TestDeviceEvent_Fields(t *testing.T) {
	t.Parallel()

	event := model.DeviceEvent{
		Device:      testDevice,
		Event:       "switch.on",
		Component:   "switch",
		ComponentID: 0,
		Data:        map[string]any{"output": true},
	}

	if event.Device != testDevice {
		t.Errorf("Device = %q, want testDevice", event.Device)
	}
	if event.Event != "switch.on" {
		t.Errorf("Event = %q, want switch.on", event.Event)
	}
	if event.Component != "switch" {
		t.Errorf("Component = %q, want switch", event.Component)
	}
	if event.ComponentID != 0 {
		t.Errorf("ComponentID = %d, want 0", event.ComponentID)
	}
}

func TestMonitoringOptions_Fields(t *testing.T) {
	t.Parallel()

	opts := MonitoringOptions{
		Count:         10,
		IncludePower:  true,
		IncludeEnergy: true,
	}

	if opts.Count != 10 {
		t.Errorf("Count = %d, want 10", opts.Count)
	}
	if !opts.IncludePower {
		t.Error("IncludePower should be true")
	}
	if !opts.IncludeEnergy {
		t.Error("IncludeEnergy should be true")
	}
}

func TestMonitoringSnapshot_Fields(t *testing.T) {
	t.Parallel()

	snapshot := model.MonitoringSnapshot{
		Device: testDevice,
		EM:     []model.EMStatus{{ID: 0}},
		EM1:    []model.EM1Status{{ID: 0}},
		PM:     []model.PMStatus{{ID: 0}},
		Online: true,
		Error:  "",
	}

	if snapshot.Device != testDevice {
		t.Errorf("Device = %q, want testDevice", snapshot.Device)
	}
	if len(snapshot.EM) != 1 {
		t.Errorf("EM length = %d, want 1", len(snapshot.EM))
	}
	if !snapshot.Online {
		t.Error("Online should be true")
	}
}

func TestDeviceSnapshot_Fields(t *testing.T) {
	t.Parallel()

	snapshot := DeviceSnapshot{
		Device:  testDevice,
		Address: "192.168.1.100",
		Error:   nil,
	}

	if snapshot.Device != testDevice {
		t.Errorf("Device = %q, want testDevice", snapshot.Device)
	}
	if snapshot.Address != "192.168.1.100" {
		t.Errorf("Address = %q, want 192.168.1.100", snapshot.Address)
	}
}

func TestPMStatus_MeterReading(t *testing.T) {
	t.Parallel()

	energy := 1000.0
	freq := 50.0
	pm := &model.PMStatus{
		APower:  500.0,
		Voltage: 230.0,
		Current: 2.2,
		AEnergy: &model.PMEnergyCounters{Total: energy},
		Freq:    &freq,
	}

	if pm.GetPower() != 500.0 {
		t.Errorf("GetPower() = %f, want 500.0", pm.GetPower())
	}
	if pm.GetVoltage() != 230.0 {
		t.Errorf("GetVoltage() = %f, want 230.0", pm.GetVoltage())
	}
	if pm.GetCurrent() != 2.2 {
		t.Errorf("GetCurrent() = %f, want 2.2", pm.GetCurrent())
	}
	if e := pm.GetEnergy(); e == nil || *e != 1000.0 {
		t.Errorf("GetEnergy() = %v, want 1000.0", e)
	}
	if f := pm.GetFreq(); f == nil || *f != 50.0 {
		t.Errorf("GetFreq() = %v, want 50.0", f)
	}
}

func TestPMStatus_MeterReadingNilEnergy(t *testing.T) {
	t.Parallel()

	pm := &model.PMStatus{
		APower:  100.0,
		AEnergy: nil,
	}

	if pm.GetEnergy() != nil {
		t.Error("GetEnergy() should return nil when AEnergy is nil")
	}
}

func TestEMStatus_MeterReading(t *testing.T) {
	t.Parallel()

	freq := 50.0
	em := &model.EMStatus{
		TotalActivePower: 5000.0,
		AVoltage:         230.0,
		TotalCurrent:     22.0,
		AFreq:            &freq,
	}

	if em.GetPower() != 5000.0 {
		t.Errorf("GetPower() = %f, want 5000.0", em.GetPower())
	}
	if em.GetVoltage() != 230.0 {
		t.Errorf("GetVoltage() = %f, want 230.0", em.GetVoltage())
	}
	if em.GetCurrent() != 22.0 {
		t.Errorf("GetCurrent() = %f, want 22.0", em.GetCurrent())
	}
	if em.GetEnergy() != nil {
		t.Error("GetEnergy() should return nil for EMStatus")
	}
	if f := em.GetFreq(); f == nil || *f != 50.0 {
		t.Errorf("GetFreq() = %v, want 50.0", f)
	}
}

func TestEM1Status_MeterReading(t *testing.T) {
	t.Parallel()

	freq := 60.0
	em1 := &model.EM1Status{
		ActPower: 1000.0,
		Voltage:  120.0,
		Current:  8.5,
		Freq:     &freq,
	}

	if em1.GetPower() != 1000.0 {
		t.Errorf("GetPower() = %f, want 1000.0", em1.GetPower())
	}
	if em1.GetVoltage() != 120.0 {
		t.Errorf("GetVoltage() = %f, want 120.0", em1.GetVoltage())
	}
	if em1.GetCurrent() != 8.5 {
		t.Errorf("GetCurrent() = %f, want 8.5", em1.GetCurrent())
	}
	if em1.GetEnergy() != nil {
		t.Error("GetEnergy() should return nil for EM1Status")
	}
}

func TestPrometheusMetric_Fields(t *testing.T) {
	t.Parallel()

	metric := export.PrometheusMetric{
		Name:   "shelly_power_watts",
		Help:   "Current power in watts",
		Type:   "gauge",
		Labels: map[string]string{"device": "test"},
		Value:  100.5,
	}

	if metric.Name != "shelly_power_watts" {
		t.Errorf("Name = %q, want shelly_power_watts", metric.Name)
	}
	if metric.Type != "gauge" {
		t.Errorf("Type = %q, want gauge", metric.Type)
	}
	if metric.Value != 100.5 {
		t.Errorf("Value = %f, want 100.5", metric.Value)
	}
}

func TestPrometheusMetrics_Fields(t *testing.T) {
	t.Parallel()

	metrics := export.PrometheusMetrics{
		Metrics: []export.PrometheusMetric{
			{Name: "test1", Value: 1.0},
			{Name: "test2", Value: 2.0},
		},
	}

	if len(metrics.Metrics) != 2 {
		t.Errorf("Metrics length = %d, want 2", len(metrics.Metrics))
	}
}

func TestComponentReading_Fields(t *testing.T) {
	t.Parallel()

	energy := 5000.0
	freq := 50.0
	reading := model.ComponentReading{
		Device:  testDevice,
		Type:    "pm",
		ID:      0,
		Phase:   "a",
		Power:   1000.0,
		Voltage: 230.0,
		Current: 4.3,
		Energy:  &energy,
		Freq:    &freq,
	}

	if reading.Device != testDevice {
		t.Errorf("Device = %q, want testDevice", reading.Device)
	}
	if reading.Type != "pm" {
		t.Errorf("Type = %q, want pm", reading.Type)
	}
	if reading.Phase != "a" {
		t.Errorf("Phase = %q, want a", reading.Phase)
	}
}

func TestInfluxDBPoint_Fields(t *testing.T) {
	t.Parallel()

	point := export.InfluxDBPoint{
		Measurement: "shelly_power",
		Tags:        map[string]string{"device": "test"},
		Fields:      map[string]float64{"power": 100.0},
	}

	if point.Measurement != "shelly_power" {
		t.Errorf("Measurement = %q, want shelly_power", point.Measurement)
	}
	if point.Tags["device"] != "test" {
		t.Errorf("Tags[device] = %q, want test", point.Tags["device"])
	}
	if point.Fields["power"] != 100.0 {
		t.Errorf("Fields[power] = %f, want 100.0", point.Fields["power"])
	}
}

func TestBuildPowerPromMetrics(t *testing.T) {
	t.Parallel()

	labels := map[string]string{"device": "test"}
	metrics := export.BuildPowerPromMetrics(labels, 100.0, 230.0, 0.5)

	if len(metrics) != 3 {
		t.Fatalf("expected 3 metrics, got %d", len(metrics))
	}

	// Verify power metric
	found := false
	for _, m := range metrics {
		if m.Name == "shelly_power_watts" && m.Value == 100.0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("power metric not found or incorrect value")
	}
}

func TestReadingsToPrometheusMetrics(t *testing.T) {
	t.Parallel()

	energy := 5000.0
	freq := 50.0
	readings := []model.ComponentReading{
		{
			Device:  "test",
			Type:    "pm",
			ID:      0,
			Power:   100.0,
			Voltage: 230.0,
			Current: 0.5,
			Energy:  &energy,
			Freq:    &freq,
		},
	}

	metrics := export.ReadingsToPrometheusMetrics(readings)

	// Should have power, voltage, current, energy, frequency = 5 metrics
	if len(metrics) < 5 {
		t.Errorf("expected at least 5 metrics, got %d", len(metrics))
	}
}

func TestReadingsToInfluxDBPoints(t *testing.T) {
	t.Parallel()

	energy := 5000.0
	readings := []model.ComponentReading{
		{
			Device:  "test",
			Type:    "pm",
			ID:      0,
			Power:   100.0,
			Voltage: 230.0,
			Current: 0.5,
			Energy:  &energy,
		},
	}

	// Use time.Now() for conversion
	points := export.ReadingsToInfluxDBPoints(readings, time.Now())
	if len(points) != 1 {
		t.Errorf("expected 1 point, got %d", len(points))
	}
	if points[0].Measurement != "shelly" {
		t.Errorf("Measurement = %q, want shelly", points[0].Measurement)
	}
}

func TestFormatPrometheusMetrics(t *testing.T) {
	t.Parallel()

	metrics := &export.PrometheusMetrics{
		Metrics: []export.PrometheusMetric{
			{
				Name:   "shelly_power_watts",
				Help:   "Power in watts",
				Type:   "gauge",
				Labels: map[string]string{"device": "test"},
				Value:  100.0,
			},
		},
	}

	result := export.FormatPrometheusMetrics(metrics)

	if result == "" {
		t.Error("expected non-empty result")
	}
	if len(result) < 10 {
		t.Errorf("result too short: %q", result)
	}
}

func TestFormatInfluxDBLineProtocol(t *testing.T) {
	t.Parallel()

	points := []export.InfluxDBPoint{
		{
			Measurement: "shelly",
			Tags:        map[string]string{"device": "test"},
			Fields:      map[string]float64{"power": 100.0},
		},
	}

	result := export.FormatInfluxDBLineProtocol(points)

	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestFormatInfluxDBPoint(t *testing.T) {
	t.Parallel()

	point := export.InfluxDBPoint{
		Measurement: "shelly",
		Tags:        map[string]string{"device": "test"},
		Fields:      map[string]float64{"power": 100.0},
	}

	result := export.FormatInfluxDBPoint(point)

	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestEscapeInfluxTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no escaping needed", "test", "test"},
		{"space", "test value", "test\\ value"},
		{"comma", "test,value", "test\\,value"},
		{"equals", "test=value", "test\\=value"},
		{"multiple", "a b,c=d", "a\\ b\\,c\\=d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := export.EscapeInfluxTag(tt.input)
			if got != tt.want {
				t.Errorf("escapeInfluxTag(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestJSONMetricsDevice_Fields(t *testing.T) {
	t.Parallel()

	device := export.JSONMetricsDevice{
		Device:     testDevice,
		Online:     true,
		Components: []model.ComponentReading{{Device: "test", Type: "pm", ID: 0}},
	}

	if device.Device != testDevice {
		t.Errorf("Device = %q, want testDevice", device.Device)
	}
	if !device.Online {
		t.Error("Online should be true")
	}
}

func TestJSONMetricsOutput_Fields(t *testing.T) {
	t.Parallel()

	output := export.JSONMetricsOutput{
		Devices: []export.JSONMetricsDevice{
			{Device: "test1", Online: true},
			{Device: "test2", Online: false},
		},
	}

	if len(output.Devices) != 2 {
		t.Errorf("Devices length = %d, want 2", len(output.Devices))
	}
}
