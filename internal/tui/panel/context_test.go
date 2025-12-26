package panel

import "testing"

func TestContext_Chaining(t *testing.T) {
	t.Parallel()

	c := NewContext("Test Panel").
		SetBadge("3").
		SetFooter("j/k:nav enter:select").
		SetPanelIndex(2).
		SetFocused(true).
		SetSize(80, 24)

	if c.Title() != "Test Panel" {
		t.Errorf("Title() = %q, want %q", c.Title(), "Test Panel")
	}
	if c.Badge() != "3" {
		t.Errorf("Badge() = %q, want %q", c.Badge(), "3")
	}
	if c.Footer() != "j/k:nav enter:select" {
		t.Errorf("Footer() = %q, want %q", c.Footer(), "j/k:nav enter:select")
	}
	if c.PanelIndex() != 2 {
		t.Errorf("PanelIndex() = %d, want 2", c.PanelIndex())
	}
	if !c.Focused() {
		t.Error("Focused() should be true")
	}
	if c.Width() != 80 {
		t.Errorf("Width() = %d, want 80", c.Width())
	}
	if c.Height() != 24 {
		t.Errorf("Height() = %d, want 24", c.Height())
	}
}

func TestContext_ContentHeight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		height     int
		focused    bool
		footer     string
		wantHeight int
	}{
		{"unfocused no footer", 20, false, "", 17},
		{"focused with footer", 20, true, "keys", 16},
		{"focused no footer", 20, true, "", 17},
		{"small panel", 5, true, "keys", 1},
		{"tiny panel", 2, true, "keys", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := NewContext("Test").
				SetSize(80, tt.height).
				SetFocused(tt.focused).
				SetFooter(tt.footer)

			if got := c.ContentHeight(); got != tt.wantHeight {
				t.Errorf("ContentHeight() = %d, want %d", got, tt.wantHeight)
			}
		})
	}
}

func TestContext_ContentWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		width     int
		wantWidth int
	}{
		{"normal", 80, 76},
		{"small", 10, 6},
		{"tiny", 3, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := NewContext("Test").SetSize(tt.width, 20)

			if got := c.ContentWidth(); got != tt.wantWidth {
				t.Errorf("ContentWidth() = %d, want %d", got, tt.wantWidth)
			}
		})
	}
}

func TestContext_DefaultValues(t *testing.T) {
	t.Parallel()

	c := NewContext("Title Only")

	if c.Badge() != "" {
		t.Errorf("default Badge() = %q, want empty", c.Badge())
	}
	if c.Footer() != "" {
		t.Errorf("default Footer() = %q, want empty", c.Footer())
	}
	if c.PanelIndex() != 0 {
		t.Errorf("default PanelIndex() = %d, want 0", c.PanelIndex())
	}
	if c.Focused() {
		t.Error("default Focused() should be false")
	}
}
