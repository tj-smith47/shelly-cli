package diagram

import (
	"fmt"
	"strings"
)

// compactRenderer renders minimal terminal-focused box layout diagrams.
type compactRenderer struct{}

func (r *compactRenderer) Render(m DeviceModel) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(header(m))
	b.WriteString("\n\n")

	switch m.Topology {
	case SingleRelay:
		r.renderSingleRelay(&b)
	case DualRelay:
		r.renderDualRelay(&b)
	case QuadRelay:
		r.renderQuadRelay(&b)
	case Dimmer:
		r.renderDimmer(&b)
	case InputOnly:
		r.renderInputOnly(&b, m)
	case Plug:
		r.renderPlug(&b)
	case EnergyMonitor:
		r.renderEnergyMonitor(&b, m)
	case RGBW:
		r.renderRGBW(&b)
	}

	b.WriteString("\n")
	b.WriteString(specsFooter(m))
	b.WriteString("\n")
	return b.String()
}

func (r *compactRenderer) renderSingleRelay(b *strings.Builder) {
	b.WriteString("  ┌─────────────────┐\n")
	b.WriteString("  │  IN:  L  N  SW  │\n")
	b.WriteString("  │  OUT: O         │\n")
	b.WriteString("  └─────────────────┘\n")
	b.WriteString("  L,N = Power  SW = Switch  O = Output\n")
}

func (r *compactRenderer) renderDualRelay(b *strings.Builder) {
	b.WriteString("  ┌──────────────────────┐\n")
	b.WriteString("  │  IN:  L  N  S1  S2   │\n")
	b.WriteString("  │  OUT: O1  O2         │\n")
	b.WriteString("  └──────────────────────┘\n")
	b.WriteString("  L,N = Power  S1,S2 = Switches  O1,O2 = Outputs\n")
}

func (r *compactRenderer) renderQuadRelay(b *strings.Builder) {
	b.WriteString("  ┌──────────────────────────────┐\n")
	b.WriteString("  │  IN:  L  N  S1  S2  S3  S4  │\n")
	b.WriteString("  │  OUT: O1  O2  O3  O4        │\n")
	b.WriteString("  └──────────────────────────────┘\n")
	b.WriteString("  L,N = Power  S1-S4 = Switches  O1-O4 = Outputs\n")
}

func (r *compactRenderer) renderDimmer(b *strings.Builder) {
	b.WriteString("  ┌──────────────────────┐\n")
	b.WriteString("  │  IN:  L  N  SW1  SW2 │\n")
	b.WriteString("  │  OUT: O (dimmed)     │\n")
	b.WriteString("  └──────────────────────┘\n")
	b.WriteString("  L,N = Power  SW1,SW2 = Switches  O = Dimmed output\n")
}

func (r *compactRenderer) renderInputOnly(b *strings.Builder, m DeviceModel) {
	inputs := inputCount(m)
	label := fmt.Sprintf("SW1-SW%d", inputs)
	b.WriteString("  ┌─────────────────────┐\n")
	fmt.Fprintf(b, "  │  IN:  L  N  %s  │\n", label)
	b.WriteString("  │  OUT: (none)        │\n")
	b.WriteString("  └─────────────────────┘\n")
	fmt.Fprintf(b, "  L,N = Power  %s = %d digital inputs (no relay)\n", label, inputs)
}

func (r *compactRenderer) renderPlug(b *strings.Builder) {
	b.WriteString("  ┌─────────────────────┐\n")
	b.WriteString("  │  Plug-in device     │\n")
	b.WriteString("  │  No wiring needed   │\n")
	b.WriteString("  │  ◉ Relay + Meter    │\n")
	b.WriteString("  └─────────────────────┘\n")
	b.WriteString("  Plug directly into wall outlet\n")
}

func (r *compactRenderer) renderEnergyMonitor(b *strings.Builder, m DeviceModel) {
	channels := emChannelCount(m)
	label := fmt.Sprintf("CT1-CT%d", channels)
	b.WriteString("  ┌──────────────────────┐\n")
	fmt.Fprintf(b, "  │  IN:  L  N  %s   │\n", label)
	b.WriteString("  │  OUT: (monitoring)   │\n")
	b.WriteString("  └──────────────────────┘\n")
	fmt.Fprintf(b, "  L,N = Power  %s = %d CT clamp(s)\n", label, channels)
}

func (r *compactRenderer) renderRGBW(b *strings.Builder) {
	b.WriteString("  ┌──────────────────────────┐\n")
	b.WriteString("  │  IN:  V+  GND            │\n")
	b.WriteString("  │  OUT: R  G  B  W         │\n")
	b.WriteString("  └──────────────────────────┘\n")
	b.WriteString("  V+,GND = DC power  R,G,B,W = LED channels\n")
}
