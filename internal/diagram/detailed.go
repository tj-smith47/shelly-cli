package diagram

import (
	"fmt"
	"strings"
)

// detailedRenderer renders top-down installer-friendly layouts with annotations.
type detailedRenderer struct{}

const detailedIndent = "              "

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

// Single relay: junctions at interior offsets 6 (SW) and w-7 (O).
// Box left edge at column 14, interior starts at column 15.
// SW junction at column 21, O junction at column w+8.
func (r *detailedRenderer) renderSingleRelay(b *strings.Builder, m DeviceModel) {
	w := detailedWidth(m.Name)
	neutralNote := "optional"
	if m.Specs.NeutralRequired {
		neutralNote = "required"
	}
	b.WriteString("  POWER SUPPLY\n")
	b.WriteString("  ════════════\n")
	fmt.Fprintf(b, "  L (Line) ──────────────────┐  %s\n", m.Specs.Voltage)
	fmt.Fprintf(b, "  N (Neutral) ───────────────┤  (%s)\n", neutralNote)
	b.WriteString("                             │\n")
	b.WriteString(detailedTopBorder(detailedIndent, w))
	b.WriteString(detailedLine(detailedIndent, w, fmt.Sprintf("         %s", m.Name)))
	b.WriteString(detailedLine(detailedIndent, w, ""))
	b.WriteString(detailedLine(detailedIndent, w, fmt.Sprintf("  Relay: %s", formatAmps(m.Specs.MaxAmps))))
	// Bottom border: 6 dashes + ┬(SW) + gap + ┬(O) + 6 dashes
	fmt.Fprintf(b, "%s└%s┬%s┬%s┘\n", detailedIndent,
		strings.Repeat("─", 6),
		strings.Repeat("─", w-14),
		strings.Repeat("─", 6))
	// Connector verticals: SW at col 21, O at col w+8
	fmt.Fprintf(b, "%s       │%s│\n", detailedIndent, strings.Repeat(" ", w-14))
	fmt.Fprintf(b, "  SWITCH INPUT       │    OUTPUT%s│\n", strings.Repeat(" ", w-24))
	fmt.Fprintf(b, "  ════════════       │    ══════%s│\n", strings.Repeat(" ", w-24))
	fmt.Fprintf(b, "  SW ────────────────┘    O %s┘──── Load\n", strings.Repeat("─", w-20))
}

// Dual relay: junctions at interior offsets 2 (S1), 6 (S2), w-12 (O1), w-5 (O2).
// S1 at col 17, S2 at col 21, O1 at col w+3, O2 at col w+10.
func (r *detailedRenderer) renderDualRelay(b *strings.Builder, m DeviceModel) {
	w := detailedWidth(m.Name)
	b.WriteString("  POWER SUPPLY\n")
	b.WriteString("  ════════════\n")
	fmt.Fprintf(b, "  L (Line) ──────────────────┐  %s\n", m.Specs.Voltage)
	b.WriteString("  N (Neutral) ───────────────┤\n")
	b.WriteString("                             │\n")
	b.WriteString(detailedTopBorder(detailedIndent, w))
	b.WriteString(detailedLine(detailedIndent, w, fmt.Sprintf("         %s", m.Name)))
	b.WriteString(detailedLine(detailedIndent, w, ""))
	b.WriteString(detailedLine(detailedIndent, w, fmt.Sprintf("  Relay 1: %s", formatAmps(m.Specs.MaxAmps))))
	b.WriteString(detailedLine(detailedIndent, w, fmt.Sprintf("  Relay 2: %s", formatAmps(m.Specs.MaxAmps))))
	// Bottom: 2─ ┬(S1) 3─ ┬(S2) gap ┬(O1) 6─ ┬(O2) 4─
	fmt.Fprintf(b, "%s└%s┬%s┬%s┬%s┬%s┘\n", detailedIndent,
		strings.Repeat("─", 2),
		strings.Repeat("─", 3),
		strings.Repeat("─", w-19),
		strings.Repeat("─", 6),
		strings.Repeat("─", 4))
	// Connector verticals: S1(17) S2(21) O1(w+3) O2(w+10)
	fmt.Fprintf(b, "%s   │   │%s│      │\n", detailedIndent, strings.Repeat(" ", w-19))
	fmt.Fprintf(b, "  SWITCH INPUTS  │   │  OUTPUTS %s│      │\n", strings.Repeat(" ", w-29))
	fmt.Fprintf(b, "  ═════════════  │   │  ═══════ %s│      │\n", strings.Repeat(" ", w-29))
	// S1 at 17, O1 at w+3, O2 at w+10
	fmt.Fprintf(b, "  S1 ────────────┘   │  O1 %s┘%s│──── Load 1\n",
		strings.Repeat("─", w-24), strings.Repeat("─", 6))
	// S2 at 21, O2 at w+10
	fmt.Fprintf(b, "  S2 ────────────────┘  O2 %s┘──── Load 2\n",
		strings.Repeat("─", w-17))
}

// Quad relay: S1-S4 at interior offsets 2,5,8,11 and O1-O4 at w-16,w-12,w-8,w-4.
// Uses wider indent (9 spaces) and minimum width 41.
func (r *detailedRenderer) renderQuadRelay(b *strings.Builder, m DeviceModel) {
	w := detailedWidth(m.Name)
	if w < 41 {
		w = 41
	}
	indent := "         "
	b.WriteString("  POWER SUPPLY\n")
	b.WriteString("  ════════════\n")
	fmt.Fprintf(b, "  L (Line) ──────────────────┐  %s\n", m.Specs.Voltage)
	b.WriteString("  N (Neutral) ───────────────┤\n")
	b.WriteString("                             │\n")
	b.WriteString(detailedTopBorder(indent, w))
	b.WriteString(detailedLine(indent, w, fmt.Sprintf("              %s", m.Name)))
	b.WriteString(detailedLine(indent, w, ""))
	b.WriteString(detailedLine(indent, w, fmt.Sprintf("  Relay 1-4: %s each", formatAmps(m.Specs.MaxAmps))))
	// Bottom: 2─ ┬(S1) 2─ ┬(S2) 2─ ┬(S3) 2─ ┬(S4) (w-28)─ ┬(O1) 3─ ┬(O2) 3─ ┬(O3) 3─ ┬(O4) 3─
	fmt.Fprintf(b, "%s└%s┬%s┬%s┬%s┬%s┬%s┬%s┬%s┬%s┘\n", indent,
		strings.Repeat("─", 2),
		strings.Repeat("─", 2),
		strings.Repeat("─", 2),
		strings.Repeat("─", 2),
		strings.Repeat("─", w-28),
		strings.Repeat("─", 3),
		strings.Repeat("─", 3),
		strings.Repeat("─", 3),
		strings.Repeat("─", 3))
	// S1-S4 at cols 12,15,18,21; O1-O4 at cols w-6,w-2,w+2,w+6
	fmt.Fprintf(b, "%s   │  │  │  │%s│   │   │   │\n", indent, strings.Repeat(" ", w-28))
	fmt.Fprintf(b, "  INPUTS    │  │  │  │  OUTS%s│   │   │   │\n", strings.Repeat(" ", w-34))
	fmt.Fprintf(b, "  ══════    │  │  │  │  ════%s│   │   │   │\n", strings.Repeat(" ", w-34))
	fmt.Fprintf(b, "  S1 %s┘  │  │  │  O1 %s┘   │   │   │──── Load 1\n",
		strings.Repeat("─", 7), strings.Repeat("─", w-33))
	fmt.Fprintf(b, "  S2 %s┘  │  │  O2 %s┘   │   │──── Load 2\n",
		strings.Repeat("─", 10), strings.Repeat("─", w-29))
	fmt.Fprintf(b, "  S3 %s┘  │  O3 %s┘   │──── Load 3\n",
		strings.Repeat("─", 13), strings.Repeat("─", w-25))
	fmt.Fprintf(b, "  S4 %s┘  O4 %s┘──── Load 4\n",
		strings.Repeat("─", 16), strings.Repeat("─", w-21))
}

// Dimmer: junctions at interior offsets 6 (SW) and w-7 (O), same as SingleRelay.
func (r *detailedRenderer) renderDimmer(b *strings.Builder, m DeviceModel) {
	w := detailedWidth(m.Name)
	neutralNote := "optional"
	if m.Specs.NeutralRequired {
		neutralNote = "required"
	}
	b.WriteString("  POWER SUPPLY\n")
	b.WriteString("  ════════════\n")
	fmt.Fprintf(b, "  L (Line) ──────────────────┐  %s\n", m.Specs.Voltage)
	fmt.Fprintf(b, "  N (Neutral) ───────────────┤  (%s)\n", neutralNote)
	b.WriteString("                             │\n")
	b.WriteString(detailedTopBorder(detailedIndent, w))
	b.WriteString(detailedLine(detailedIndent, w, fmt.Sprintf("         %s", m.Name)))
	b.WriteString(detailedLine(detailedIndent, w, ""))
	b.WriteString(detailedLine(detailedIndent, w, "  Trailing-edge dimmer"))
	// Bottom border: 6─ ┬(SW) gap ┬(O) 6─
	fmt.Fprintf(b, "%s└%s┬%s┬%s┘\n", detailedIndent,
		strings.Repeat("─", 6),
		strings.Repeat("─", w-14),
		strings.Repeat("─", 6))
	// Connector verticals: SW at col 21, O at col w+8
	fmt.Fprintf(b, "%s       │%s│\n", detailedIndent, strings.Repeat(" ", w-14))
	fmt.Fprintf(b, "  SWITCH INPUTS      │    OUTPUT%s│\n", strings.Repeat(" ", w-24))
	fmt.Fprintf(b, "  ═════════════      │    ══════%s│\n", strings.Repeat(" ", w-24))
	fmt.Fprintf(b, "  SW1 (up) ──────────┤    O %s┘──── Light\n", strings.Repeat("─", w-20))
	b.WriteString("  SW2 (down) ────────┘\n")
}

// InputOnly: single junction at interior offset 7.
// Junction column = 22.
func (r *detailedRenderer) renderInputOnly(b *strings.Builder, m DeviceModel) {
	w := detailedWidth(m.Name)
	inputs := inputCount(m)
	b.WriteString("  POWER SUPPLY\n")
	b.WriteString("  ════════════\n")
	fmt.Fprintf(b, "  L (Line) ──────────────────┐  %s\n", m.Specs.Voltage)
	b.WriteString("  N (Neutral) ───────────────┤\n")
	b.WriteString("                             │\n")
	b.WriteString(detailedTopBorder(detailedIndent, w))
	b.WriteString(detailedLine(detailedIndent, w, fmt.Sprintf("         %s", m.Name)))
	b.WriteString(detailedLine(detailedIndent, w, ""))
	b.WriteString(detailedLine(detailedIndent, w, fmt.Sprintf("  %d digital inputs, no relay", inputs)))
	// Bottom: 7─ ┬ rest─
	fmt.Fprintf(b, "%s└%s┬%s┘\n", detailedIndent,
		strings.Repeat("─", 7),
		strings.Repeat("─", w-8))
	// Junction at col 22
	fmt.Fprintf(b, "%s        │\n", detailedIndent)
	b.WriteString("  DIGITAL INPUTS      │\n")
	b.WriteString("  ══════════════      │\n")
	for i := 1; i <= inputs; i++ {
		if i < inputs {
			fmt.Fprintf(b, "  SW%d ────────────────┤\n", i)
		} else {
			fmt.Fprintf(b, "  SW%d ────────────────┘\n", i)
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

// EnergyMonitor: single junction at interior offset 7, same as InputOnly.
func (r *detailedRenderer) renderEnergyMonitor(b *strings.Builder, m DeviceModel) {
	w := detailedWidth(m.Name)
	channels := emChannelCount(m)
	b.WriteString("  POWER SUPPLY\n")
	b.WriteString("  ════════════\n")
	fmt.Fprintf(b, "  L (Line) ──────────────────┐  %s\n", m.Specs.Voltage)
	b.WriteString("  N (Neutral) ───────────────┤\n")
	b.WriteString("                             │\n")
	b.WriteString(detailedTopBorder(detailedIndent, w))
	b.WriteString(detailedLine(detailedIndent, w, fmt.Sprintf("         %s", m.Name)))
	b.WriteString(detailedLine(detailedIndent, w, ""))
	b.WriteString(detailedLine(detailedIndent, w, fmt.Sprintf("  %d-channel energy meter", channels)))
	// Bottom: 7─ ┬ rest─
	fmt.Fprintf(b, "%s└%s┬%s┘\n", detailedIndent,
		strings.Repeat("─", 7),
		strings.Repeat("─", w-8))
	// Junction at col 22
	fmt.Fprintf(b, "%s        │\n", detailedIndent)
	b.WriteString("  CT CLAMP INPUTS     │\n")
	b.WriteString("  ═══════════════     │\n")
	for i := 1; i <= channels; i++ {
		if i < channels {
			fmt.Fprintf(b, "  CT%d ─ ─ ─ ─ ─ ─ ─ ─┤  ← Clamp around conductor\n", i)
		} else {
			fmt.Fprintf(b, "  CT%d ─ ─ ─ ─ ─ ─ ─ ─┘  ← Clamp around conductor\n", i)
		}
	}
	b.WriteString("\n")
	b.WriteString("  CT clamps are non-invasive — no wire cutting needed.\n")
	b.WriteString("  Clamp each CT around the LIVE conductor to monitor.\n")
}

// RGBW: 4 junctions at interior offsets 2, 7, 12, 17 (4-dash gap between each).
// indent = 13 spaces. Junction columns: 16, 21, 26, 31.
func (r *detailedRenderer) renderRGBW(b *strings.Builder, m DeviceModel) {
	w := detailedWidth(m.Name)
	indent := "             "
	b.WriteString("  DC POWER SUPPLY\n")
	b.WriteString("  ═══════════════\n")
	fmt.Fprintf(b, "  V+ ────────────────────────┐  %s\n", m.Specs.Voltage)
	b.WriteString("  GND ───────────────────────┤\n")
	b.WriteString("                             │\n")
	b.WriteString(detailedTopBorder(indent, w))
	b.WriteString(detailedLine(indent, w, fmt.Sprintf("         %s", m.Name)))
	b.WriteString(detailedLine(indent, w, ""))
	b.WriteString(detailedLine(indent, w, "  4-channel PWM controller"))
	// Bottom: 2─ ┬(R) 4─ ┬(G) 4─ ┬(B) 4─ ┬(W) (w-18)─
	// Junctions at interior offsets 2(R), 7(G), 12(B), 17(W) → cols 16, 21, 26, 31
	fmt.Fprintf(b, "%s└%s┬%s┬%s┬%s┬%s┘\n", indent,
		strings.Repeat("─", 2),
		strings.Repeat("─", 4),
		strings.Repeat("─", 4),
		strings.Repeat("─", 4),
		strings.Repeat("─", w-18))
	fmt.Fprintf(b, "%s   │    │    │    │\n", indent)
	b.WriteString("  LED OUTPUTS   │    │    │    │\n")
	b.WriteString("  ═══════════   │    │    │    │\n")
	b.WriteString("  R (Red) ──────┘    │    │    │\n")
	b.WriteString("  G (Green) ─────────┘    │    │\n")
	b.WriteString("  B (Blue) ───────────────┘    │\n")
	b.WriteString("  W (White) ───────────────────┘\n")
	b.WriteString("\n")
	b.WriteString("  Connect common-anode LED strips to output channels.\n")
	b.WriteString("  V+ goes to LED strip V+, R/G/B/W to strip channels.\n")
}
