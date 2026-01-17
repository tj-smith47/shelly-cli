// Package export provides business logic for exporting Shelly device data.
package export

import (
	"fmt"
	"strings"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// PrometheusMetrics represents metrics in Prometheus exposition format.
type PrometheusMetrics struct {
	Metrics []PrometheusMetric `json:"metrics"`
}

// PrometheusMetric represents a single Prometheus metric.
type PrometheusMetric struct {
	Name   string            `json:"name"`
	Help   string            `json:"help"`
	Type   string            `json:"type"` // gauge, counter
	Labels map[string]string `json:"labels"`
	Value  float64           `json:"value"`
}

// JSONMetricsDevice represents metrics for a single device in JSON output.
type JSONMetricsDevice struct {
	Device     string                   `json:"device"`
	Online     bool                     `json:"online"`
	Components []model.ComponentReading `json:"components,omitempty"`
}

// JSONMetricsOutput represents the JSON metrics output format.
type JSONMetricsOutput struct {
	Timestamp time.Time           `json:"timestamp"`
	Devices   []JSONMetricsDevice `json:"devices"`
}

// BuildPowerPromMetrics creates power, voltage, and current Prometheus metrics.
func BuildPowerPromMetrics(labels map[string]string, power, voltage, current float64) []PrometheusMetric {
	return []PrometheusMetric{
		{Name: "shelly_power_watts", Help: "Current power consumption in watts", Type: "gauge", Labels: labels, Value: power},
		{Name: "shelly_voltage_volts", Help: "Current voltage in volts", Type: "gauge", Labels: labels, Value: voltage},
		{Name: "shelly_current_amps", Help: "Current in amps", Type: "gauge", Labels: labels, Value: current},
	}
}

// ReadingsToPrometheusMetrics converts ComponentReadings to Prometheus metrics.
func ReadingsToPrometheusMetrics(readings []model.ComponentReading) []PrometheusMetric {
	metrics := make([]PrometheusMetric, 0, len(readings)*5)
	for _, r := range readings {
		labels := map[string]string{"device": r.Device, "component": r.Type, "component_id": fmt.Sprintf("%d", r.ID)}
		if r.Phase != "" {
			labels["phase"] = r.Phase
		}
		metrics = append(metrics, BuildPowerPromMetrics(labels, r.Power, r.Voltage, r.Current)...)
		if r.Energy != nil {
			metrics = append(metrics, PrometheusMetric{
				Name: "shelly_energy_wh_total", Help: "Total energy consumption in watt-hours",
				Type: "counter", Labels: labels, Value: *r.Energy,
			})
		}
		if r.Freq != nil {
			metrics = append(metrics, PrometheusMetric{
				Name: "shelly_frequency_hz", Help: "AC frequency in hertz",
				Type: "gauge", Labels: labels, Value: *r.Freq,
			})
		}
	}
	return metrics
}

// FormatPrometheusMetrics formats metrics as Prometheus exposition format.
func FormatPrometheusMetrics(metrics *PrometheusMetrics) string {
	var result string
	seen := make(map[string]bool)

	for _, m := range metrics.Metrics {
		// Print HELP and TYPE only once per metric name
		if !seen[m.Name] {
			result += fmt.Sprintf("# HELP %s %s\n", m.Name, m.Help)
			result += fmt.Sprintf("# TYPE %s %s\n", m.Name, m.Type)
			seen[m.Name] = true
		}

		// Format labels
		labels := ""
		if len(m.Labels) > 0 {
			first := true
			labels = "{"
			for k, v := range m.Labels {
				if !first {
					labels += ","
				}
				labels += fmt.Sprintf("%s=%q", k, v)
				first = false
			}
			labels += "}"
		}

		result += fmt.Sprintf("%s%s %g\n", m.Name, labels, m.Value)
	}

	return result
}

// ExtractWifiMetrics extracts WiFi RSSI metrics from device status.
func ExtractWifiMetrics(labels map[string]string, status map[string]any) []PrometheusMetric {
	wifi, ok := status["wifi"].(map[string]any)
	if !ok {
		return nil
	}
	rssi, ok := wifi["rssi"].(float64)
	if !ok {
		return nil
	}
	return []PrometheusMetric{{
		Name: "shelly_wifi_rssi", Help: "WiFi signal strength in dBm",
		Type: "gauge", Labels: labels, Value: rssi,
	}}
}

// ExtractSysPrometheusMetrics extracts system metrics from device status.
func ExtractSysPrometheusMetrics(labels map[string]string, status map[string]any) []PrometheusMetric {
	sys, ok := status["sys"].(map[string]any)
	if !ok {
		return nil
	}

	var metrics []PrometheusMetric
	if uptime, ok := sys["uptime"].(float64); ok {
		metrics = append(metrics, PrometheusMetric{
			Name: "shelly_uptime_seconds", Help: "Device uptime in seconds",
			Type: "counter", Labels: labels, Value: uptime,
		})
	}
	if ramFree, ok := sys["ram_free"].(float64); ok {
		metrics = append(metrics, PrometheusMetric{
			Name: "shelly_ram_free_bytes", Help: "Free RAM in bytes",
			Type: "gauge", Labels: labels, Value: ramFree,
		})
	}
	if ramTotal, ok := sys["ram_size"].(float64); ok {
		metrics = append(metrics, PrometheusMetric{
			Name: "shelly_ram_total_bytes", Help: "Total RAM in bytes",
			Type: "gauge", Labels: labels, Value: ramTotal,
		})
	}
	metrics = append(metrics, ExtractTempMetric(labels, sys)...)
	return metrics
}

// ExtractTempMetric extracts temperature metrics from sys status.
func ExtractTempMetric(labels map[string]string, sys map[string]any) []PrometheusMetric {
	temp, ok := sys["temperature"].(map[string]any)
	if !ok {
		return nil
	}
	tC, ok := temp["tC"].(float64)
	if !ok {
		return nil
	}
	return []PrometheusMetric{{
		Name: "shelly_temperature_celsius", Help: "Device temperature in Celsius",
		Type: "gauge", Labels: labels, Value: tC,
	}}
}

// ExtractSwitchPrometheusMetrics extracts switch state metrics from device status.
func ExtractSwitchPrometheusMetrics(device string, status map[string]any) []PrometheusMetric {
	// Count switches for pre-allocation
	count := 0
	for key := range status {
		if strings.HasPrefix(key, "switch:") {
			count++
		}
	}
	metrics := make([]PrometheusMetric, 0, count)
	for key, val := range status {
		if !strings.HasPrefix(key, "switch:") {
			continue
		}
		sw, ok := val.(map[string]any)
		if !ok {
			continue
		}
		output, ok := sw["output"].(bool)
		if !ok {
			continue
		}
		outputVal := 0.0
		if output {
			outputVal = 1.0
		}
		metrics = append(metrics, PrometheusMetric{
			Name: "shelly_switch_on", Help: "Switch state (1=on, 0=off)",
			Type: "gauge", Labels: map[string]string{"device": device, "component": key}, Value: outputVal,
		})
	}
	return metrics
}

// CollectSystemPrometheusMetrics collects system-level Prometheus metrics from device status.
func CollectSystemPrometheusMetrics(device string, status map[string]any) []PrometheusMetric {
	metrics := make([]PrometheusMetric, 0, 16)
	deviceLabels := map[string]string{"device": device}

	// WiFi RSSI
	metrics = append(metrics, ExtractWifiMetrics(deviceLabels, status)...)

	// System metrics (uptime, ram, temp)
	metrics = append(metrics, ExtractSysPrometheusMetrics(deviceLabels, status)...)

	// Switch states
	metrics = append(metrics, ExtractSwitchPrometheusMetrics(device, status)...)

	return metrics
}

// CollectMeterPrometheusMetrics is a generic collector for any meter type.
func CollectMeterPrometheusMetrics[T model.MeterReading](
	device, compType string,
	ids []int,
	getFunc func(id int) (T, error),
) []PrometheusMetric {
	metrics := make([]PrometheusMetric, 0, len(ids)*5)
	for _, id := range ids {
		status, err := getFunc(id)
		if err != nil {
			continue
		}
		labels := map[string]string{"device": device, "component": compType, "component_id": fmt.Sprintf("%d", id)}
		metrics = append(metrics, BuildPowerPromMetrics(labels, status.GetPower(), status.GetVoltage(), status.GetCurrent())...)
		if energy := status.GetEnergy(); energy != nil {
			metrics = append(metrics, PrometheusMetric{
				Name: "shelly_energy_wh_total", Help: "Total energy consumption in watt-hours",
				Type: "counter", Labels: labels, Value: *energy,
			})
		}
		if freq := status.GetFreq(); freq != nil {
			metrics = append(metrics, PrometheusMetric{
				Name: "shelly_frequency_hz", Help: "AC frequency in hertz",
				Type: "gauge", Labels: labels, Value: *freq,
			})
		}
	}
	return metrics
}
