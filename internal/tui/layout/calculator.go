// Package layout provides flexible panel sizing and rendering utilities for the TUI.
package layout

// PanelID uniquely identifies a panel within a view.
type PanelID int

// PanelConfig defines how a panel should be sized.
type PanelConfig struct {
	ID            PanelID
	MinHeight     int  // Minimum rows when collapsed (includes border)
	MaxHeight     int  // Maximum rows when expanded (0 = fill available)
	ExpandOnFocus bool // Grow to MaxHeight when focused
}

// Column represents a vertical column of panels.
type Column struct {
	Panels      []PanelConfig
	Width       int
	TotalHeight int
	FocusedID   PanelID
}

// CalculatePanelHeights distributes available height among panels in a column.
// When a panel is focused and has ExpandOnFocus=true, it gets maximum space.
// Other panels shrink to their MinHeight.
func (c *Column) CalculatePanelHeights() map[PanelID]int {
	return c.CalculatePanelHeightsWithMax(len(c.Panels))
}

// CalculatePanelHeightsWithMax distributes available height using a specified max panel count.
// This allows balancing heights across multiple columns with different panel counts.
func (c *Column) CalculatePanelHeightsWithMax(maxPanels int) map[PanelID]int {
	if len(c.Panels) == 0 {
		return make(map[PanelID]int)
	}

	// Find which panel should expand
	expandingPanel := c.findExpandingPanel()
	if expandingPanel == -1 {
		// No panel is expanding - distribute evenly based on max panels
		return c.distributeEvenlyWithMax(maxPanels)
	}

	return c.distributeWithExpansion(expandingPanel)
}

// findExpandingPanel returns the ID of the panel that should expand, or -1 if none.
func (c *Column) findExpandingPanel() PanelID {
	for _, p := range c.Panels {
		if p.ID == c.FocusedID && p.ExpandOnFocus {
			return p.ID
		}
	}
	return -1
}

// distributeWithExpansion allocates heights with one panel expanded.
func (c *Column) distributeWithExpansion(expandingID PanelID) map[PanelID]int {
	heights := make(map[PanelID]int)

	// Calculate minimum space needed by non-expanding panels
	minSpaceNeeded := 0
	var expandingMaxHeight int
	for _, p := range c.Panels {
		if p.ID != expandingID {
			minSpaceNeeded += p.MinHeight
		} else {
			expandingMaxHeight = p.MaxHeight
		}
	}

	// Calculate expanding panel height
	expandingHeight := max(0, c.TotalHeight-minSpaceNeeded)
	if expandingMaxHeight > 0 && expandingHeight > expandingMaxHeight {
		expandingHeight = expandingMaxHeight
	}

	// Assign heights
	for _, p := range c.Panels {
		if p.ID == expandingID {
			heights[p.ID] = expandingHeight
		} else {
			heights[p.ID] = p.MinHeight
		}
	}

	// Handle remaining space after max constraint
	c.distributeRemaining(heights, expandingID)

	return heights
}

// distributeRemaining gives any leftover space to the last non-expanding panel.
func (c *Column) distributeRemaining(heights map[PanelID]int, expandingID PanelID) {
	usedHeight := 0
	for _, h := range heights {
		usedHeight += h
	}
	remaining := c.TotalHeight - usedHeight
	if remaining <= 0 {
		return
	}
	// Give remaining to last non-expanding panel
	for i := len(c.Panels) - 1; i >= 0; i-- {
		p := c.Panels[i]
		if p.ID != expandingID {
			heights[p.ID] += remaining
			return
		}
	}
}

// distributeEvenlyWithMax gives each panel equal height based on a max panel count.
// This enables balanced heights across columns with different panel counts.
// Panels get height based on maxPanels, with remaining space distributed evenly.
func (c *Column) distributeEvenlyWithMax(maxPanels int) map[PanelID]int {
	heights := make(map[PanelID]int)
	if len(c.Panels) == 0 {
		return heights
	}

	// Use the larger of maxPanels or actual panel count
	if maxPanels < len(c.Panels) {
		maxPanels = len(c.Panels)
	}

	// Calculate base height per "slot" using max panels
	baseHeight := c.TotalHeight / maxPanels

	// Each panel gets the base height
	for _, p := range c.Panels {
		heights[p.ID] = baseHeight
	}

	// Distribute remaining space (unused slots + remainder) evenly among actual panels
	usedHeight := baseHeight * len(c.Panels)
	remaining := c.TotalHeight - usedHeight
	if remaining > 0 {
		perPanel := remaining / len(c.Panels)
		extraRemainder := remaining % len(c.Panels)

		for i, p := range c.Panels {
			heights[p.ID] += perPanel
			// Give extra pixels to last panels
			if i >= len(c.Panels)-extraRemainder {
				heights[p.ID]++
			}
		}
	}

	return heights
}

// TwoColumnLayout manages a two-column panel layout.
type TwoColumnLayout struct {
	LeftColumn  Column
	RightColumn Column
	TotalWidth  int
	TotalHeight int
	LeftRatio   float64 // Ratio of total width for left column (0.0-1.0)
	Gap         int     // Gap between columns
}

// NewTwoColumnLayout creates a new two-column layout.
func NewTwoColumnLayout(leftRatio float64, gap int) *TwoColumnLayout {
	if leftRatio <= 0 || leftRatio >= 1 {
		leftRatio = 0.5
	}
	if gap < 0 {
		gap = 1
	}
	return &TwoColumnLayout{
		LeftRatio: leftRatio,
		Gap:       gap,
	}
}

// SetSize updates the total dimensions available.
func (l *TwoColumnLayout) SetSize(width, height int) {
	l.TotalWidth = width
	l.TotalHeight = height

	// Calculate column widths
	leftWidth := int(float64(width) * l.LeftRatio)
	rightWidth := width - leftWidth - l.Gap

	l.LeftColumn.Width = leftWidth
	l.LeftColumn.TotalHeight = height

	l.RightColumn.Width = rightWidth
	l.RightColumn.TotalHeight = height
}

// SetFocus updates which panel is focused.
// Pass -1 to clear focus from all panels (equal distribution).
func (l *TwoColumnLayout) SetFocus(panelID PanelID) {
	// Clear focus when -1 is passed
	if panelID == -1 {
		l.LeftColumn.FocusedID = -1
		l.RightColumn.FocusedID = -1
		return
	}

	// Check if panel is in left or right column
	for _, p := range l.LeftColumn.Panels {
		if p.ID == panelID {
			l.LeftColumn.FocusedID = panelID
			l.RightColumn.FocusedID = -1
			return
		}
	}
	for _, p := range l.RightColumn.Panels {
		if p.ID == panelID {
			l.RightColumn.FocusedID = panelID
			l.LeftColumn.FocusedID = -1
			return
		}
	}
}

// Calculate returns the width and height for each panel.
// Heights are balanced across columns using the max panel count from either column.
func (l *TwoColumnLayout) Calculate() map[PanelID]PanelDimensions {
	result := make(map[PanelID]PanelDimensions)

	// Use max panel count for balanced heights across columns
	maxPanels := max(len(l.LeftColumn.Panels), len(l.RightColumn.Panels))

	leftHeights := l.LeftColumn.CalculatePanelHeightsWithMax(maxPanels)
	rightHeights := l.RightColumn.CalculatePanelHeightsWithMax(maxPanels)

	for id, h := range leftHeights {
		result[id] = PanelDimensions{
			Width:  l.LeftColumn.Width,
			Height: h,
		}
	}

	for id, h := range rightHeights {
		result[id] = PanelDimensions{
			Width:  l.RightColumn.Width,
			Height: h,
		}
	}

	return result
}

// PanelDimensions holds the calculated width and height for a panel.
type PanelDimensions struct {
	Width  int
	Height int
}

// ContentDimensions returns the usable content area (subtracting borders).
func (d PanelDimensions) ContentDimensions(borderWidth int) (width, height int) {
	width = d.Width - (borderWidth * 2)
	height = d.Height - 2 // Top and bottom border
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	return width, height
}

// ThreeColumnLayout manages a three-column panel layout (for JSON preview).
type ThreeColumnLayout struct {
	LeftColumn   Column
	MiddleColumn Column
	RightColumn  Column
	TotalWidth   int
	TotalHeight  int
	LeftRatio    float64 // Ratio for left column
	MiddleRatio  float64 // Ratio for middle column
	// Right column gets remaining space
	Gap            int
	RightCollapsed bool // When true, right column has 0 width
}

// NewThreeColumnLayout creates a new three-column layout.
func NewThreeColumnLayout(leftRatio, middleRatio float64, gap int) *ThreeColumnLayout {
	return &ThreeColumnLayout{
		LeftRatio:   leftRatio,
		MiddleRatio: middleRatio,
		Gap:         gap,
	}
}

// SetSize updates the total dimensions.
func (l *ThreeColumnLayout) SetSize(width, height int) {
	l.TotalWidth = width
	l.TotalHeight = height
	l.recalculateWidths()
}

// SetRightCollapsed shows/hides the right column (for slide-in effect).
func (l *ThreeColumnLayout) SetRightCollapsed(collapsed bool) {
	l.RightCollapsed = collapsed
	l.recalculateWidths()
}

func (l *ThreeColumnLayout) recalculateWidths() {
	if l.RightCollapsed {
		// Two-column mode
		leftWidth := int(float64(l.TotalWidth) * l.LeftRatio / (l.LeftRatio + l.MiddleRatio))
		middleWidth := l.TotalWidth - leftWidth - l.Gap

		l.LeftColumn.Width = leftWidth
		l.MiddleColumn.Width = middleWidth
		l.RightColumn.Width = 0
	} else {
		// Three-column mode
		leftWidth := int(float64(l.TotalWidth) * l.LeftRatio)
		middleWidth := int(float64(l.TotalWidth) * l.MiddleRatio)
		rightWidth := l.TotalWidth - leftWidth - middleWidth - (l.Gap * 2)

		l.LeftColumn.Width = leftWidth
		l.MiddleColumn.Width = middleWidth
		l.RightColumn.Width = rightWidth
	}

	l.LeftColumn.TotalHeight = l.TotalHeight
	l.MiddleColumn.TotalHeight = l.TotalHeight
	l.RightColumn.TotalHeight = l.TotalHeight
}

// Calculate returns dimensions for all panels.
func (l *ThreeColumnLayout) Calculate() map[PanelID]PanelDimensions {
	result := make(map[PanelID]PanelDimensions)

	leftHeights := l.LeftColumn.CalculatePanelHeights()
	middleHeights := l.MiddleColumn.CalculatePanelHeights()
	rightHeights := l.RightColumn.CalculatePanelHeights()

	for id, h := range leftHeights {
		result[id] = PanelDimensions{Width: l.LeftColumn.Width, Height: h}
	}
	for id, h := range middleHeights {
		result[id] = PanelDimensions{Width: l.MiddleColumn.Width, Height: h}
	}
	for id, h := range rightHeights {
		result[id] = PanelDimensions{Width: l.RightColumn.Width, Height: h}
	}

	return result
}
