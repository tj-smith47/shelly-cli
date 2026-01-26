package diagram

import (
	"fmt"
	"strings"
)

// detailedRenderer renders top-down installer-friendly layouts with annotations.
type detailedRenderer struct{}

func (r *detailedRenderer) Render(m DeviceModel) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(header(m))
	b.WriteString("\n\n")

	switch m.Topology {
	case SingleRelay:
		r.renderSingleRelay(&b, m)
	case DualRelay:
		r.renderDualRelay(&b, m)
	case QuadRelay:
		r.renderQuadRelay(&b, m)
	case Dimmer:
		r.renderDimmer(&b, m)
	case InputOnly:
		r.renderInputOnly(&b, m)
	case Plug:
		r.renderPlug(&b, m)
	case EnergyMonitor:
		r.renderEnergyMonitor(&b, m)
	case RGBW:
		r.renderRGBW(&b, m)
	}

	b.WriteString("\n")
	b.WriteString(specsFooter(m))
	b.WriteString("\n")
	return b.String()
}

func (r *detailedRenderer) renderSingleRelay(b *strings.Builder, m DeviceModel) {
	neutralNote := "optional"
	if m.Specs.NeutralRequired {
		neutralNote = "required"
	}
	b.WriteString("  POWER SUPPLY\n")
	b.WriteString("  ════════════\n")
	fmt.Fprintf(b, "  L (Line) ──────────────────┐  %s\n", m.Specs.Voltage)
	fmt.Fprintf(b, "  N (Neutral) ───────────────┤  (%s)\n", neutralNote)
	b.WriteString("                             │\n")
	b.WriteString("              ┌──────────────┴──────────────┐\n")
	fmt.Fprintf(b, "              │         %-20s │\n", m.Name)
	b.WriteString("              │                             │\n")
	fmt.Fprintf(b, "              │  Relay: %.0fA                 │\n", m.Specs.MaxAmps)
	b.WriteString("              └──────┬───────────────┬──────┘\n")
	b.WriteString("                     │               │\n")
	b.WriteString("  SWITCH INPUT       │    OUTPUT     │\n")
	b.WriteString("  ════════════       │    ══════     │\n")
	b.WriteString("  SW ────────────────┘    O ─────────┘──── Load\n")
}

func (r *detailedRenderer) renderDualRelay(b *strings.Builder, m DeviceModel) {
	b.WriteString("  POWER SUPPLY\n")
	b.WriteString("  ════════════\n")
	fmt.Fprintf(b, "  L (Line) ──────────────────┐  %s\n", m.Specs.Voltage)
	b.WriteString("  N (Neutral) ───────────────┤\n")
	b.WriteString("                             │\n")
	b.WriteString("              ┌──────────────┴──────────────┐\n")
	fmt.Fprintf(b, "              │         %-20s │\n", m.Name)
	b.WriteString("              │                             │\n")
	fmt.Fprintf(b, "              │  Relay 1: %.0fA               │\n", m.Specs.MaxAmps)
	fmt.Fprintf(b, "              │  Relay 2: %.0fA               │\n", m.Specs.MaxAmps)
	b.WriteString("              └──┬───┬──────────┬──────┬────┘\n")
	b.WriteString("                 │   │          │      │\n")
	b.WriteString("  SWITCH INPUTS  │   │  OUTPUTS │      │\n")
	b.WriteString("  ═════════════  │   │  ═══════ │      │\n")
	b.WriteString("  S1 ────────────┘   │  O1 ─────┘──── Load 1\n")
	b.WriteString("  S2 ───────────────┘  O2 ────────── Load 2\n")
}

func (r *detailedRenderer) renderQuadRelay(b *strings.Builder, m DeviceModel) {
	b.WriteString("  POWER SUPPLY\n")
	b.WriteString("  ════════════\n")
	fmt.Fprintf(b, "  L (Line) ──────────────────┐  %s\n", m.Specs.Voltage)
	b.WriteString("  N (Neutral) ───────────────┤\n")
	b.WriteString("                             │\n")
	b.WriteString("         ┌────────────────────┴────────────────────┐\n")
	fmt.Fprintf(b, "         │              %-26s │\n", m.Name)
	b.WriteString("         │                                        │\n")
	fmt.Fprintf(b, "         │  Relay 1-4: %.0fA each                   │\n", m.Specs.MaxAmps)
	b.WriteString("         └──┬──┬──┬──┬────────┬──┬──┬──┬──────────┘\n")
	b.WriteString("            │  │  │  │        │  │  │  │\n")
	b.WriteString("  INPUTS    │  │  │  │  OUTS  │  │  │  │\n")
	b.WriteString("  ══════    │  │  │  │  ════  │  │  │  │\n")
	b.WriteString("  S1 ───────┘  │  │  │  O1 ───┘  │  │  │──── Load 1\n")
	b.WriteString("  S2 ──────────┘  │  │  O2 ──────┘  │  │──── Load 2\n")
	b.WriteString("  S3 ─────────────┘  │  O3 ─────────┘  │──── Load 3\n")
	b.WriteString("  S4 ────────────────┘  O4 ────────────┘──── Load 4\n")
}

func (r *detailedRenderer) renderDimmer(b *strings.Builder, m DeviceModel) {
	neutralNote := "optional"
	if m.Specs.NeutralRequired {
		neutralNote = "required"
	}
	b.WriteString("  POWER SUPPLY\n")
	b.WriteString("  ════════════\n")
	fmt.Fprintf(b, "  L (Line) ──────────────────┐  %s\n", m.Specs.Voltage)
	fmt.Fprintf(b, "  N (Neutral) ───────────────┤  (%s)\n", neutralNote)
	b.WriteString("                             │\n")
	b.WriteString("              ┌──────────────┴──────────────┐\n")
	fmt.Fprintf(b, "              │         %-20s │\n", m.Name)
	b.WriteString("              │                             │\n")
	b.WriteString("              │  Trailing-edge dimmer       │\n")
	b.WriteString("              └──────┬───────────────┬──────┘\n")
	b.WriteString("                     │               │\n")
	b.WriteString("  SWITCH INPUTS      │    OUTPUT     │\n")
	b.WriteString("  ═════════════      │    ══════     │\n")
	b.WriteString("  SW1 (up) ──────────┤    O ─────────┘──── Light\n")
	b.WriteString("  SW2 (down) ────────┘\n")
}

func (r *detailedRenderer) renderInputOnly(b *strings.Builder, m DeviceModel) {
	inputs := inputCount(m)
	b.WriteString("  POWER SUPPLY\n")
	b.WriteString("  ════════════\n")
	fmt.Fprintf(b, "  L (Line) ──────────────────┐  %s\n", m.Specs.Voltage)
	b.WriteString("  N (Neutral) ───────────────┤\n")
	b.WriteString("                             │\n")
	b.WriteString("              ┌──────────────┴──────────────┐\n")
	fmt.Fprintf(b, "              │         %-20s │\n", m.Name)
	b.WriteString("              │                             │\n")
	fmt.Fprintf(b, "              │  %d digital inputs, no relay  │\n", inputs)
	b.WriteString("              └───────┬─────────────────────┘\n")
	b.WriteString("                      │\n")
	b.WriteString("  DIGITAL INPUTS      │\n")
	b.WriteString("  ══════════════      │\n")
	for i := 1; i <= inputs; i++ {
		fmt.Fprintf(b, "  SW%d ────────────────", i)
		if i < inputs {
			b.WriteString("┤\n")
		} else {
			b.WriteString("┘\n")
		}
	}
	b.WriteString("\n")
	b.WriteString("  No output terminals — input-only device\n")
}

func (r *detailedRenderer) renderPlug(b *strings.Builder, m DeviceModel) {
	b.WriteString("  INSTALLATION\n")
	b.WriteString("  ════════════\n")
	b.WriteString("  No wiring required — plug-in device\n")
	b.WriteString("\n")
	b.WriteString("    ╔══════════════════════════════╗\n")
	b.WriteString("    ║                              ║\n")
	b.WriteString("    ║    ┌────────────────────┐    ║\n")
	fmt.Fprintf(b, "    ║    │  %-18s │    ║\n", m.Name)
	b.WriteString("    ║    │                    │    ║\n")
	b.WriteString("    ║    │  ◉ Relay           │    ║\n")
	b.WriteString("    ║    │  ⚡ Power meter     │    ║\n")
	b.WriteString("    ║    │  📶 Wi-Fi + BLE     │    ║\n")
	b.WriteString("    ║    └────────────────────┘    ║\n")
	b.WriteString("    ║                              ║\n")
	b.WriteString("    ╚══════════════════════════════╝\n")
	b.WriteString("\n")
	b.WriteString("  1. Plug into wall outlet\n")
	b.WriteString("  2. Connect load to front socket\n")
	b.WriteString("  3. Configure via app or CLI\n")
}

func (r *detailedRenderer) renderEnergyMonitor(b *strings.Builder, m DeviceModel) {
	channels := emChannelCount(m)
	b.WriteString("  POWER SUPPLY\n")
	b.WriteString("  ════════════\n")
	fmt.Fprintf(b, "  L (Line) ──────────────────┐  %s\n", m.Specs.Voltage)
	b.WriteString("  N (Neutral) ───────────────┤\n")
	b.WriteString("                             │\n")
	b.WriteString("              ┌──────────────┴──────────────┐\n")
	fmt.Fprintf(b, "              │         %-20s │\n", m.Name)
	b.WriteString("              │                             │\n")
	fmt.Fprintf(b, "              │  %d-channel energy meter      │\n", channels)
	b.WriteString("              └───────┬─────────────────────┘\n")
	b.WriteString("                      │\n")
	b.WriteString("  CT CLAMP INPUTS     │\n")
	b.WriteString("  ═══════════════     │\n")
	for i := 1; i <= channels; i++ {
		fmt.Fprintf(b, "  CT%d ─ ─ ─ ─ ─ ─ ─ ", i)
		if i < channels {
			b.WriteString("┤  ← Clamp around conductor\n")
		} else {
			b.WriteString("┘  ← Clamp around conductor\n")
		}
	}
	b.WriteString("\n")
	b.WriteString("  CT clamps are non-invasive — no wire cutting needed.\n")
	b.WriteString("  Clamp each CT around the LIVE conductor to monitor.\n")
}

func (r *detailedRenderer) renderRGBW(b *strings.Builder, m DeviceModel) {
	b.WriteString("  DC POWER SUPPLY\n")
	b.WriteString("  ═══════════════\n")
	fmt.Fprintf(b, "  V+ ───────────────────────┐  %s\n", m.Specs.Voltage)
	b.WriteString("  GND ──────────────────────┤\n")
	b.WriteString("                            │\n")
	b.WriteString("             ┌──────────────┴──────────────┐\n")
	fmt.Fprintf(b, "             │         %-20s │\n", m.Name)
	b.WriteString("             │                             │\n")
	b.WriteString("             │  4-channel PWM controller   │\n")
	b.WriteString("             └──┬─────┬─────┬─────┬────────┘\n")
	b.WriteString("                │     │     │     │\n")
	b.WriteString("  LED OUTPUTS   │     │     │     │\n")
	b.WriteString("  ═══════════   │     │     │     │\n")
	b.WriteString("  R (Red) ──────┘     │     │     │\n")
	b.WriteString("  G (Green) ──────────┘     │     │\n")
	b.WriteString("  B (Blue) ─────────────────┘     │\n")
	b.WriteString("  W (White) ──────────────────────┘\n")
	b.WriteString("\n")
	b.WriteString("  Connect common-anode LED strips to output channels.\n")
	b.WriteString("  V+ goes to LED strip V+, R/G/B/W to strip channels.\n")
}
