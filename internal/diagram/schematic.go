package diagram

import (
	"fmt"
	"strings"
)

// schematicRenderer renders left-to-right circuit-style wiring diagrams.
type schematicRenderer struct{}

func (r *schematicRenderer) Render(m DeviceModel) string {
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

func (r *schematicRenderer) renderSingleRelay(b *strings.Builder, _ DeviceModel) {
	b.WriteString("               ┌────────────────┐\n")
	b.WriteString("    L ─────────┤ L            O ├──────── Load\n")
	b.WriteString("    N ─────────┤ N              │\n")
	b.WriteString("   SW ─────────┤ SW             │\n")
	b.WriteString("               └────────────────┘\n")
}

func (r *schematicRenderer) renderDualRelay(b *strings.Builder, _ DeviceModel) {
	b.WriteString("               ┌────────────────┐\n")
	b.WriteString("    L ─────────┤ L           O1 ├──────── Load 1\n")
	b.WriteString("    N ─────────┤ N           O2 ├──────── Load 2\n")
	b.WriteString("   S1 ─────────┤ S1             │\n")
	b.WriteString("   S2 ─────────┤ S2             │\n")
	b.WriteString("               └────────────────┘\n")
}

func (r *schematicRenderer) renderQuadRelay(b *strings.Builder, _ DeviceModel) {
	b.WriteString("               ┌────────────────┐\n")
	b.WriteString("    L ─────────┤ L           O1 ├──────── Load 1\n")
	b.WriteString("    N ─────────┤ N           O2 ├──────── Load 2\n")
	b.WriteString("   S1 ─────────┤ S1          O3 ├──────── Load 3\n")
	b.WriteString("   S2 ─────────┤ S2          O4 ├──────── Load 4\n")
	b.WriteString("   S3 ─────────┤ S3             │\n")
	b.WriteString("   S4 ─────────┤ S4             │\n")
	b.WriteString("               └────────────────┘\n")
}

func (r *schematicRenderer) renderDimmer(b *strings.Builder, _ DeviceModel) {
	b.WriteString("               ┌────────────────┐\n")
	b.WriteString("    L ─────────┤ L            O ├──────── Light\n")
	b.WriteString("    N ─────────┤ N              │\n")
	b.WriteString("  SW1 ─────────┤ SW1            │\n")
	b.WriteString("  SW2 ─────────┤ SW2            │\n")
	b.WriteString("               └────────────────┘\n")
}

func (r *schematicRenderer) renderInputOnly(b *strings.Builder, m DeviceModel) {
	inputs := inputCount(m)
	b.WriteString("               ┌────────────────┐\n")
	b.WriteString("    L ─────────┤ L              │\n")
	b.WriteString("    N ─────────┤ N              │\n")
	for i := 1; i <= inputs; i++ {
		fmt.Fprintf(b, "  SW%d ─────────┤ SW%d            │\n", i, i)
	}
	b.WriteString("               └────────────────┘\n")
}

func (r *schematicRenderer) renderPlug(b *strings.Builder, _ DeviceModel) {
	b.WriteString("    ╔════════════════════════╗\n")
	b.WriteString("    ║                        ║\n")
	b.WriteString("    ║    ┌──────────────┐    ║\n")
	b.WriteString("    ║    │   ◉  Relay    │    ║\n")
	b.WriteString("    ║    │   Power Meter │    ║\n")
	b.WriteString("    ║    └──────────────┘    ║\n")
	b.WriteString("    ║                        ║\n")
	b.WriteString("    ╚════════════════════════╝\n")
	b.WriteString("      Plug into wall outlet\n")
}

func (r *schematicRenderer) renderEnergyMonitor(b *strings.Builder, m DeviceModel) {
	channels := emChannelCount(m)
	b.WriteString("               ┌────────────────┐\n")
	b.WriteString("    L ─────────┤ L              │\n")
	b.WriteString("    N ─────────┤ N              │\n")
	b.WriteString("               │                │\n")
	for i := 1; i <= channels; i++ {
		label := string(rune('0' + i))
		b.WriteString("  CT")
		b.WriteString(label)
		b.WriteString(" ─ ─ ─ ─┤ CT")
		b.WriteString(label)
		b.WriteString("            │\n")
	}
	b.WriteString("               └────────────────┘\n")
	b.WriteString("    CT = Current Transformer clamp\n")
}

func (r *schematicRenderer) renderRGBW(b *strings.Builder, _ DeviceModel) {
	b.WriteString("               ┌────────────────┐\n")
	b.WriteString("   V+ ─────────┤ V+           R ├──────── Red\n")
	b.WriteString("  GND ─────────┤ GND          G ├──────── Green\n")
	b.WriteString("               │              B ├──────── Blue\n")
	b.WriteString("               │              W ├──────── White\n")
	b.WriteString("               └────────────────┘\n")
}
