// Package output provides output formatting utilities for the CLI.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/output/jsonfmt"
	"github.com/tj-smith47/shelly-cli/internal/output/tmplfmt"
	"github.com/tj-smith47/shelly-cli/internal/output/yamlfmt"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Format represents an output format.
type Format string

// Output format constants.
const (
	FormatJSON     Format = "json"
	FormatYAML     Format = "yaml"
	FormatTable    Format = "table"
	FormatText     Format = "text"
	FormatTemplate Format = "template"
)

// Formatter defines the interface for output formatters.
type Formatter interface {
	Format(w io.Writer, data any) error
}

// GetFormat returns the current output format from config.
func GetFormat() Format {
	format := viper.GetString("output")
	switch strings.ToLower(format) {
	case "json":
		return FormatJSON
	case "yaml", "yml":
		return FormatYAML
	case "table":
		return FormatTable
	case "text", "plain":
		return FormatText
	case "template", "go-template":
		return FormatTemplate
	default:
		return FormatTable
	}
}

// GetTemplate returns the current template string from config.
func GetTemplate() string {
	return viper.GetString("template")
}

// Print outputs data in the configured format.
func Print(data any) error {
	return PrintTo(os.Stdout, data)
}

// PrintTo outputs data to the specified writer in the configured format.
func PrintTo(w io.Writer, data any) error {
	formatter := NewFormatter(GetFormat())
	return formatter.Format(w, data)
}

// PrintJSON outputs data as JSON.
func PrintJSON(data any) error {
	return jsonfmt.New().Format(os.Stdout, data)
}

// PrintYAML outputs data as YAML.
func PrintYAML(data any) error {
	return yamlfmt.New().Format(os.Stdout, data)
}

// JSON outputs data as JSON to the specified writer.
func JSON(w io.Writer, data any) error {
	return jsonfmt.New().Format(w, data)
}

// YAML outputs data as YAML to the specified writer.
func YAML(w io.Writer, data any) error {
	return yamlfmt.New().Format(w, data)
}

// NewFormatter creates a formatter for the given format.
func NewFormatter(format Format) Formatter {
	switch format {
	case FormatJSON:
		return jsonfmt.New()
	case FormatYAML:
		return yamlfmt.New()
	case FormatTable:
		return NewTableFormatter()
	case FormatText:
		return NewTextFormatter()
	case FormatTemplate:
		return tmplfmt.New(GetTemplate())
	default:
		return NewTableFormatter()
	}
}

// TextFormatter formats output as plain text.
type TextFormatter struct{}

// NewTextFormatter creates a new text formatter.
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// Format outputs data as plain text.
func (f *TextFormatter) Format(w io.Writer, data any) error {
	// Handle different types
	switch v := data.(type) {
	case string:
		_, err := fmt.Fprintln(w, v)
		return err
	case []string:
		for _, s := range v {
			if _, err := fmt.Fprintln(w, s); err != nil {
				return err
			}
		}
		return nil
	case fmt.Stringer:
		_, err := fmt.Fprintln(w, v.String())
		return err
	default:
		_, err := fmt.Fprintf(w, "%+v\n", v)
		return err
	}
}

// TableFormatter formats output as a table using reflection.
// For slices of structs, it creates a table with struct field names as headers.
// For other types, it falls back to text format.
type TableFormatter struct{}

// NewTableFormatter creates a new table formatter.
func NewTableFormatter() *TableFormatter {
	return &TableFormatter{}
}

// Format outputs data as a table.
func (f *TableFormatter) Format(w io.Writer, data any) error {
	table := f.buildTable(data)
	if table == nil {
		return NewTextFormatter().Format(w, data)
	}
	return table.PrintTo(w)
}

// buildTable attempts to build a table from structured data.
// Returns nil if the data cannot be represented as a table.
func (f *TableFormatter) buildTable(data any) *Table {
	return buildTableFromData(data)
}

// ParseFormat parses a format string into a Format.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON, nil
	case "yaml", "yml":
		return FormatYAML, nil
	case "table":
		return FormatTable, nil
	case "text", "plain":
		return FormatText, nil
	case "template", "go-template":
		return FormatTemplate, nil
	default:
		return "", fmt.Errorf("unknown format: %s", s)
	}
}

// ValidFormats returns a list of valid format strings.
func ValidFormats() []string {
	return []string{"json", "yaml", "table", "text", "template"}
}

// Template outputs data using the specified template to the given writer.
func Template(w io.Writer, tmpl string, data any) error {
	return tmplfmt.New(tmpl).Format(w, data)
}

// PrintTemplate outputs data using the specified template to stdout.
func PrintTemplate(tmpl string, data any) error {
	return Template(os.Stdout, tmpl, data)
}

// IsQuiet returns true if quiet mode is enabled.
func IsQuiet() bool {
	return viper.GetBool("quiet")
}

// IsVerbose returns true if verbose mode is enabled.
func IsVerbose() bool {
	return viper.GetBool("verbose")
}

// WantsJSON returns true if the output format is JSON.
func WantsJSON() bool {
	return GetFormat() == FormatJSON
}

// WantsYAML returns true if the output format is YAML.
func WantsYAML() bool {
	return GetFormat() == FormatYAML
}

// WantsTable returns true if the output format is table.
func WantsTable() bool {
	return GetFormat() == FormatTable
}

// WantsStructured returns true if the output format needs raw data (JSON, YAML, or template).
func WantsStructured() bool {
	format := GetFormat()
	return format == FormatJSON || format == FormatYAML || format == FormatTemplate
}

// FormatOutput prints data in the configured output format.
// For table format, it uses the Table type from the output package.
// For other formats (json, yaml, template), it uses the standard formatters.
func FormatOutput(w io.Writer, data any) error {
	formatter := NewFormatter(GetFormat())
	return formatter.Format(w, data)
}

// FormatPlaceholder returns dimmed placeholder text.
func FormatPlaceholder(text string) string {
	return theme.Dim().Render(text)
}

// FormatActionCount returns themed action count string.
func FormatActionCount(count int) string {
	if count == 0 {
		return theme.StatusWarn().Render("0 (empty)")
	}
	if count == 1 {
		return theme.StatusOK().Render("1 action")
	}
	return theme.StatusOK().Render(fmt.Sprintf("%d actions", count))
}

// FormatDeviceCount returns a device count string with proper pluralization.
func FormatDeviceCount(count int) string {
	if count == 0 {
		return "0 (empty)"
	}
	if count == 1 {
		return "1 device"
	}
	return fmt.Sprintf("%d devices", count)
}

// FormatConfigValue converts any configuration value to a display string.
// It handles nil, bool, float64, string, maps, and slices appropriately.
func FormatConfigValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return "<not set>"
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		// Check if it's an integer
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case string:
		if val == "" {
			return "<empty>"
		}
		return val
	case map[string]interface{}, []interface{}:
		data, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(data)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// FormatConfigTable builds a table from a configuration map.
// It expects a map[string]interface{} where keys are setting names.
// Returns nil if config is not a map.
func FormatConfigTable(config interface{}) *Table {
	cfgMap, ok := config.(map[string]interface{})
	if !ok {
		return nil
	}

	table := NewTable("Setting", "Value")
	for key, value := range cfgMap {
		table.AddRow(key, FormatConfigValue(value))
	}
	return table
}

// FormatFloat formats a float64 value for CSV/data export.
// It uses automatic precision to avoid unnecessary trailing zeros.
func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// FormatFloatPtr formats a *float64 value for CSV/data export.
// It returns an empty string for nil values.
func FormatFloatPtr(f *float64) string {
	if f == nil {
		return ""
	}
	return strconv.FormatFloat(*f, 'f', -1, 64)
}

// FormatSize formats a byte count as a human-readable string.
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// ValueTruncateTable is the default truncation length for table display.
const ValueTruncateTable = 40

// valueNull is the standard null representation.
const valueNull = "null"

// FormatJSONValue formats a value for JSON-like display.
// Strings are quoted, numbers are formatted cleanly, nil becomes "null".
func FormatJSONValue(v any) string {
	if v == nil {
		return valueNull
	}
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%.0f", val)
		}
		return fmt.Sprintf("%g", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// FormatDisplayValue formats a value for table display with truncation.
// Strings are quoted and truncated, maps/arrays are summarized.
func FormatDisplayValue(v any) string {
	if v == nil {
		return valueNull
	}
	switch val := v.(type) {
	case string:
		if len(val) > ValueTruncateTable {
			return fmt.Sprintf("%q...", val[:ValueTruncateTable-3])
		}
		return fmt.Sprintf("%q", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%.0f", val)
		}
		return fmt.Sprintf("%g", val)
	case map[string]any:
		return fmt.Sprintf("{%d fields}", len(val))
	case []any:
		return fmt.Sprintf("[%d items]", len(val))
	default:
		return fmt.Sprintf("%v", val)
	}
}

// FormatComponentStatus formats a component status map into a human-readable string.
// This extracts key fields based on the component type (switch, cover, light, etc.)
// and returns a meaningful summary instead of just "{N fields}".
func FormatComponentStatus(componentName string, status map[string]any) string {
	if len(status) == 0 {
		return "-"
	}

	// Determine component type from name (e.g., "switch:0" -> "switch")
	compType := extractComponentType(componentName)

	switch compType {
	case "switch":
		return formatSwitchStatus(status)
	case "light":
		return formatLightStatus(status)
	case "cover":
		return formatCoverStatus(status)
	case "input":
		return formatInputStatus(status)
	case "pm1", "pm":
		return formatPowerMeterStatus(status)
	case "temperature", "humidity", "devicepower", "illuminance":
		return formatSensorStatus(compType, status)
	case "sys":
		return formatSysStatus(status)
	case "wifi", "cloud", "mqtt", "ble", "eth":
		return formatNetworkStatus(compType, status)
	case "ws", "ota", "ui", "sntp":
		return formatSystemComponentStatus(compType, status)
	default:
		// For unknown components, show key-value pairs inline
		return formatGenericStatus(status)
	}
}

// extractComponentType extracts the type from a component name like "switch:0" -> "switch".
func extractComponentType(name string) string {
	for i, c := range name {
		if c == ':' {
			return name[:i]
		}
	}
	return name
}

// formatSwitchStatus formats switch component status.
func formatSwitchStatus(status map[string]any) string {
	var parts []string

	// Output state
	if output, ok := status["output"].(bool); ok {
		if output {
			parts = append(parts, theme.StatusOK().Render("ON"))
		} else {
			parts = append(parts, theme.Dim().Render("off"))
		}
	}

	// Active power
	if apower, ok := status["apower"].(float64); ok {
		parts = append(parts, FormatPowerCompact(apower))
	}

	// Voltage
	if voltage, ok := status["voltage"].(float64); ok {
		parts = append(parts, fmt.Sprintf("%.1fV", voltage))
	}

	// Current
	if current, ok := status["current"].(float64); ok && current > 0 {
		parts = append(parts, fmt.Sprintf("%.2fA", current))
	}

	// Temperature (overtemp protection)
	if tc, ok := status["temperature"].(map[string]any); ok {
		if temp, ok := tc["tC"].(float64); ok {
			parts = append(parts, fmt.Sprintf("%.1f°C", temp))
		}
	}

	if len(parts) == 0 {
		return formatGenericStatus(status)
	}
	return strings.Join(parts, ", ")
}

// formatLightStatus formats light component status.
func formatLightStatus(status map[string]any) string {
	var parts []string

	// Output state
	if output, ok := status["output"].(bool); ok {
		if output {
			parts = append(parts, theme.StatusOK().Render("ON"))
		} else {
			parts = append(parts, theme.Dim().Render("off"))
		}
	}

	// Brightness
	if brightness, ok := status["brightness"].(float64); ok {
		parts = append(parts, fmt.Sprintf("%d%%", int(brightness)))
	}

	// RGB color
	if rgb, ok := status["rgb"].([]any); ok && len(rgb) == 3 {
		parts = append(parts, fmt.Sprintf("RGB(%v,%v,%v)", rgb[0], rgb[1], rgb[2]))
	}

	if len(parts) == 0 {
		return formatGenericStatus(status)
	}
	return strings.Join(parts, ", ")
}

// formatCoverStatus formats cover component status.
func formatCoverStatus(status map[string]any) string {
	var parts []string

	// State (open, closed, opening, closing, stopped)
	if state, ok := status["state"].(string); ok {
		switch state {
		case "open":
			parts = append(parts, theme.StatusOK().Render("open"))
		case "closed":
			parts = append(parts, theme.Dim().Render("closed"))
		case "opening":
			parts = append(parts, theme.StatusWarn().Render("opening"))
		case "closing":
			parts = append(parts, theme.StatusWarn().Render("closing"))
		default:
			parts = append(parts, state)
		}
	}

	// Current position
	if pos, ok := status["current_pos"].(float64); ok {
		parts = append(parts, fmt.Sprintf("%d%%", int(pos)))
	}

	// Power consumption during movement
	if apower, ok := status["apower"].(float64); ok && apower > 0 {
		parts = append(parts, FormatPowerCompact(apower))
	}

	if len(parts) == 0 {
		return formatGenericStatus(status)
	}
	return strings.Join(parts, ", ")
}

// formatInputStatus formats input component status.
func formatInputStatus(status map[string]any) string {
	// State (true = triggered)
	if state, ok := status["state"].(bool); ok {
		if state {
			return theme.StatusWarn().Render("triggered")
		}
		return theme.Dim().Render("idle")
	}
	return formatGenericStatus(status)
}

// formatPowerMeterStatus formats power meter (pm1) component status.
func formatPowerMeterStatus(status map[string]any) string {
	var parts []string

	if apower, ok := status["apower"].(float64); ok {
		parts = append(parts, FormatPowerCompact(apower))
	}
	if voltage, ok := status["voltage"].(float64); ok {
		parts = append(parts, fmt.Sprintf("%.1fV", voltage))
	}
	if current, ok := status["current"].(float64); ok {
		parts = append(parts, fmt.Sprintf("%.2fA", current))
	}
	if freq, ok := status["freq"].(float64); ok {
		parts = append(parts, fmt.Sprintf("%.1fHz", freq))
	}

	if len(parts) == 0 {
		return formatGenericStatus(status)
	}
	return strings.Join(parts, ", ")
}

// formatSensorStatus formats sensor component status.
func formatSensorStatus(compType string, status map[string]any) string {
	switch compType {
	case "temperature":
		if tC, ok := status["tC"].(float64); ok {
			return fmt.Sprintf("%.1f°C", tC)
		}
	case "humidity":
		if rh, ok := status["rh"].(float64); ok {
			return fmt.Sprintf("%.1f%%", rh)
		}
	case "illuminance":
		if lux, ok := status["lux"].(float64); ok {
			return fmt.Sprintf("%.0f lux", lux)
		}
	case "devicepower":
		var parts []string
		if battery, ok := status["battery"].(map[string]any); ok {
			if percent, ok := battery["percent"].(float64); ok {
				parts = append(parts, fmt.Sprintf("%.0f%%", percent))
			}
		}
		if external, ok := status["external"].(map[string]any); ok {
			if present, ok := external["present"].(bool); ok && present {
				parts = append(parts, "external power")
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, ", ")
		}
	}
	return formatGenericStatus(status)
}

// formatSysStatus formats sys component status.
func formatSysStatus(status map[string]any) string {
	var parts []string

	// Check for available updates
	if updates, ok := status["available_updates"].(map[string]any); ok {
		if stable, ok := updates["stable"].(map[string]any); ok {
			if version, ok := stable["version"].(string); ok {
				parts = append(parts, theme.StatusOK().Render(fmt.Sprintf("update: %s", version)))
			}
		}
	}

	// Restart required
	if restart, ok := status["restart_required"].(bool); ok && restart {
		parts = append(parts, theme.StatusWarn().Render("restart required"))
	}

	// Uptime
	if uptime, ok := status["uptime"].(float64); ok {
		parts = append(parts, fmt.Sprintf("up %s", FormatDuration(time.Duration(uptime)*time.Second)))
	}

	// RAM free
	if ramFree, ok := status["ram_free"].(float64); ok {
		parts = append(parts, fmt.Sprintf("%s free", FormatSize(int64(ramFree))))
	}

	if len(parts) == 0 {
		return theme.Dim().Render("ok")
	}
	return strings.Join(parts, ", ")
}

// formatNetworkStatus formats network component status (wifi, cloud, mqtt, etc.).
//
//nolint:gocyclo // Component-specific formatting requires many cases
func formatNetworkStatus(compType string, status map[string]any) string {
	switch compType {
	case "wifi":
		if ssid, ok := status["ssid"].(string); ok {
			result := ssid
			if rssi, ok := status["rssi"].(float64); ok {
				result += fmt.Sprintf(" (%ddBm)", int(rssi))
			}
			return result
		}
		if sta, ok := status["sta_ip"].(string); ok {
			return sta
		}
	case "cloud":
		if connected, ok := status["connected"].(bool); ok {
			if connected {
				return theme.StatusOK().Render("connected")
			}
			return theme.Dim().Render("disconnected")
		}
	case "mqtt":
		if connected, ok := status["connected"].(bool); ok {
			if connected {
				return theme.StatusOK().Render("connected")
			}
			return theme.Dim().Render("disconnected")
		}
	case "ble":
		if enabled, ok := status["enabled"].(bool); ok {
			if enabled {
				return theme.StatusOK().Render("enabled")
			}
			return theme.Dim().Render("disabled")
		}
	case "eth":
		if ip, ok := status["ip"].(string); ok && ip != "" {
			return ip
		}
	}
	return formatGenericStatus(status)
}

// formatSystemComponentStatus formats other system components (ws, ota, ui, sntp).
func formatSystemComponentStatus(compType string, status map[string]any) string {
	switch compType {
	case "ws":
		if connected, ok := status["connected"].(bool); ok {
			if connected {
				return theme.StatusOK().Render("connected")
			}
			return theme.Dim().Render("disconnected")
		}
	case "ui":
		// UI usually doesn't have meaningful status
		return theme.Dim().Render("ok")
	}
	return formatGenericStatus(status)
}

// formatGenericStatus formats an unknown component status by showing key fields.
func formatGenericStatus(status map[string]any) string {
	if len(status) == 0 {
		return "-"
	}

	// Show up to 3 key fields inline
	parts := make([]string, 0, 4)
	count := 0
	for key, value := range status {
		if count >= 3 {
			remaining := len(status) - count
			if remaining > 0 {
				parts = append(parts, fmt.Sprintf("+%d more", remaining))
			}
			break
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, formatSimpleValue(value)))
		count++
	}
	return strings.Join(parts, ", ")
}

// formatSimpleValue formats a single value for inline display.
func formatSimpleValue(v any) string {
	switch val := v.(type) {
	case bool:
		return strconv.FormatBool(val)
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%.0f", val)
		}
		return fmt.Sprintf("%.2f", val)
	case string:
		if len(val) > 20 {
			return val[:17] + "..."
		}
		return val
	case map[string]any, []any:
		return "..."
	default:
		s := fmt.Sprintf("%v", val)
		if len(s) > 20 {
			return s[:17] + "..."
		}
		return s
	}
}

// FormatPowerCompact formats power in a compact form (e.g., "45.2W").
func FormatPowerCompact(watts float64) string {
	if watts >= 1000 {
		return fmt.Sprintf("%.1fkW", watts/1000)
	}
	if watts == 0 {
		return "0W"
	}
	return fmt.Sprintf("%.1fW", watts)
}

// ValueType returns the type name of a value for display.
func ValueType(v any) string {
	if v == nil {
		return valueNull
	}
	switch v.(type) {
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	default:
		return "unknown"
	}
}

// FormatDuration formats a duration for human-readable display.
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%dd", days)
}

// FormatParamsInline formats parameters as an inline string.
// Example: key1=value1, key2=value2.
func FormatParamsInline(params map[string]any) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, 0, len(params))
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(parts, ", ")
}

// FormatParamsTable formats parameters as multi-line for table display.
// Example:
//
//	key1: value1
//	key2: value2
func FormatParamsTable(params map[string]any) string {
	if len(params) == 0 {
		return "-"
	}
	lines := make([]string, 0, len(params))
	for k, v := range params {
		lines = append(lines, fmt.Sprintf("%s: %v", k, v))
	}
	return strings.Join(lines, "\n")
}

// Truncate truncates a string to maxLen characters, adding "..." if truncated.
func Truncate(s string, maxLen int) string {
	// Use ansi.StringWidth for proper visual width calculation
	width := ansi.StringWidth(s)
	if width <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return ansi.Truncate(s, maxLen, "")
	}
	return ansi.Truncate(s, maxLen, "...")
}

// PadRight pads a string with spaces to reach the specified visual width.
// Uses ansi.StringWidth for proper calculation with ANSI escape codes.
func PadRight(s string, width int) string {
	sWidth := ansi.StringWidth(s)
	if sWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-sWidth)
}

// RenderProgressBar renders a text-based progress bar.
// value is the current value, maxVal is the maximum value.
// The bar width is fixed at 20 characters.
func RenderProgressBar(value, maxVal int) string {
	const barWidth = 20
	filled := (value * barWidth) / maxVal
	if filled > barWidth {
		filled = barWidth
	}
	bar := ""
	for range barWidth {
		if filled > 0 {
			bar += "█"
			filled--
		} else {
			bar += "░"
		}
	}
	return theme.Dim().Render("[") + bar + theme.Dim().Render("]")
}

// EscapeWiFiQR escapes special characters in WiFi QR content.
// Escapes: backslash, semicolon, comma, colon.
func EscapeWiFiQR(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, ":", "\\:")
	return s
}

// FormatWiFiSignalStrength converts RSSI value to human-readable signal strength.
func FormatWiFiSignalStrength(rssi int) string {
	switch {
	case rssi >= -50:
		return "excellent"
	case rssi >= -60:
		return "good"
	case rssi >= -70:
		return "fair"
	default:
		return "weak"
	}
}

// FormatReleaseNotes formats release notes for display.
// Truncates to maxLen characters and indents each line.
func FormatReleaseNotes(body string) string {
	const maxLen = 500
	if len(body) > maxLen {
		body = body[:maxLen] + "..."
	}

	lines := strings.Split(body, "\n")
	for i, line := range lines {
		lines[i] = "  " + line
	}

	return strings.Join(lines, "\n")
}

// FormatDeviceGeneration returns a formatted generation string.
func FormatDeviceGeneration(gen int) string {
	return fmt.Sprintf("Gen%d", gen)
}

// ExtractMapSection extracts a map section from an RPC result.
// Useful for extracting subsections like "ws" from Sys.GetConfig results.
// Returns nil if the section doesn't exist or result isn't a map.
func ExtractMapSection(result any, key string) map[string]any {
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil
	}

	var m map[string]any
	if err := json.Unmarshal(jsonBytes, &m); err != nil {
		return nil
	}

	section, ok := m[key].(map[string]any)
	if !ok {
		return nil
	}

	return section
}
