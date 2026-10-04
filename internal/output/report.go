// Package output provides pure formatters (data → string).
package output

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// FormatDevicesReportText formats the device inventory report as plain text.
func FormatDevicesReportText(r model.DevicesReport) string {
	var b strings.Builder
	writeReportHeader(&b, "Device inventory", r.Timestamp)
	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	writeRow(tw, "NAME", "IP", "MODEL", "GEN", "FIRMWARE", "MAC", "STATUS")
	for _, d := range r.Devices {
		writeRow(tw, d.Name, d.IP, dash(d.Model), genLabel(d.Generation), dash(d.Firmware), dash(d.MAC), onlineLabel(d.Online))
	}
	flush(tw)
	fmt.Fprintf(&b, "\nSummary: %d devices, %d online, %d offline\n", r.Summary.Total, r.Summary.Online, r.Summary.Offline)
	return b.String()
}

// FormatEnergyReportText formats the energy report as plain text.
func FormatEnergyReportText(r model.EnergyReport) string {
	var b strings.Builder
	writeReportHeader(&b, "Energy", r.Timestamp)
	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	writeRow(tw, "NAME", "STATUS", "POWER")
	for _, d := range r.Devices {
		power := "no meter"
		switch {
		case !d.Online:
			power = "-"
		case d.Reporting:
			power = fmt.Sprintf("%.1f W", d.PowerW)
		}
		writeRow(tw, d.Name, onlineLabel(d.Online), power)
	}
	flush(tw)
	s := r.Summary
	fmt.Fprintf(&b, "\nTotal power: %.1f W from %d reporting devices (%d devices, %d online, %d offline)\n",
		s.TotalPowerW, s.DevicesReporting, s.Total, s.Online, s.Offline)
	return b.String()
}

// FormatAuditReportText formats the security audit report as plain text.
func FormatAuditReportText(r model.AuditReport) string {
	var b strings.Builder
	writeReportHeader(&b, "Security audit", r.Timestamp)
	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	writeRow(tw, "NAME", "REACHABLE", "AUTH", "CLOUD", "FIRMWARE", "UPDATE")
	for _, d := range r.Devices {
		update := "-"
		if d.FirmwareOutdated != nil {
			update = "up to date"
			if *d.FirmwareOutdated {
				update = d.FirmwareAvailable
			}
		}
		writeRow(tw, d.Name, yesNo(d.Reachable), boolLabel(d.AuthEnabled, "enabled", "disabled"),
			boolLabel(d.CloudConnected, "connected", "local"), dash(d.FirmwareCurrent), update)
	}
	flush(tw)

	var findings []string
	for _, d := range r.Devices {
		for _, issue := range d.Issues {
			findings = append(findings, fmt.Sprintf("  %s: issue: %s", d.Name, issue))
		}
		for _, warning := range d.Warnings {
			findings = append(findings, fmt.Sprintf("  %s: warning: %s", d.Name, warning))
		}
	}
	if len(findings) > 0 {
		b.WriteString("\nFindings:\n")
		b.WriteString(strings.Join(findings, "\n"))
		b.WriteString("\n")
	}

	s := r.Summary
	fmt.Fprintf(&b, "\nSummary: %d devices scanned, %d reachable, %d unreachable\n", s.DevicesScanned, s.Reachable, s.Unreachable)
	fmt.Fprintf(&b, "  Authentication: %d enabled, %d disabled\n", s.AuthEnabled, s.AuthDisabled)
	fmt.Fprintf(&b, "  Cloud connected: %d\n", s.CloudConnected)
	fmt.Fprintf(&b, "  Outdated firmware: %d\n", s.OutdatedFirmware)
	fmt.Fprintf(&b, "  Issues: %d, warnings: %d\n", s.Issues, s.Warnings)
	return b.String()
}

func writeReportHeader(b *strings.Builder, title string, ts time.Time) {
	fmt.Fprintf(b, "Shelly %s report\n", title)
	fmt.Fprintf(b, "Generated: %s\n\n", ts.Format(time.RFC3339))
}

// writeRow and flush cannot fail: the tabwriter writes to a strings.Builder.
func writeRow(tw *tabwriter.Writer, cells ...string) {
	if _, err := fmt.Fprintln(tw, strings.Join(cells, "\t")+"\t"); err != nil {
		panic(err)
	}
}

func flush(tw *tabwriter.Writer) {
	if err := tw.Flush(); err != nil {
		panic(err)
	}
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func genLabel(gen int) string {
	if gen == 0 {
		return "-"
	}
	return strconv.Itoa(gen)
}

func onlineLabel(online bool) string {
	if online {
		return LabelOnlineLower
	}
	return LabelOfflineLower
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func boolLabel(v *bool, yes, no string) string {
	switch {
	case v == nil:
		return "-"
	case *v:
		return yes
	default:
		return no
	}
}
