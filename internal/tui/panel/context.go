package panel

import "github.com/tj-smith47/shelly-cli/internal/tui/rendering"

// Context holds common state for panel rendering.
// It provides a consistent way to manage title, badge, footer, focus, and dimensions.
type Context struct {
	title      string
	badge      string
	footer     string
	panelIndex int
	focused    bool
	width      int
	height     int
}

// NewContext creates a new panel context with the given title.
func NewContext(title string) *Context {
	return &Context{title: title}
}

// SetTitle sets the panel title.
func (c *Context) SetTitle(title string) *Context {
	c.title = title
	return c
}

// SetBadge sets the panel badge (displayed next to title).
func (c *Context) SetBadge(badge string) *Context {
	c.badge = badge
	return c
}

// SetFooter sets the panel footer (keybindings hint).
func (c *Context) SetFooter(footer string) *Context {
	c.footer = footer
	return c
}

// SetPanelIndex sets the panel index for Shift+N navigation hint.
func (c *Context) SetPanelIndex(index int) *Context {
	c.panelIndex = index
	return c
}

// SetFocused sets whether the panel is focused.
func (c *Context) SetFocused(focused bool) *Context {
	c.focused = focused
	return c
}

// SetSize sets the panel dimensions.
func (c *Context) SetSize(width, height int) *Context {
	c.width = width
	c.height = height
	return c
}

// Title returns the panel title.
func (c *Context) Title() string {
	return c.title
}

// Badge returns the panel badge.
func (c *Context) Badge() string {
	return c.badge
}

// Footer returns the panel footer.
func (c *Context) Footer() string {
	return c.footer
}

// PanelIndex returns the panel index.
func (c *Context) PanelIndex() int {
	return c.panelIndex
}

// Focused returns whether the panel is focused.
func (c *Context) Focused() bool {
	return c.focused
}

// Width returns the panel width.
func (c *Context) Width() int {
	return c.width
}

// Height returns the panel height.
func (c *Context) Height() int {
	return c.height
}

// ApplyToRenderer applies the context settings to a rendering.Renderer.
func (c *Context) ApplyToRenderer(r *rendering.Renderer) *rendering.Renderer {
	r.SetTitle(c.title).
		SetFocused(c.focused).
		SetPanelIndex(c.panelIndex)

	if c.badge != "" {
		r.SetBadge(c.badge)
	}
	if c.footer != "" && c.focused {
		r.SetFooter(c.footer)
	}
	return r
}

// NewRenderer creates a new renderer with this context applied.
func (c *Context) NewRenderer() *rendering.Renderer {
	r := rendering.New(c.width, c.height)
	return c.ApplyToRenderer(r)
}

// ContentHeight returns the height available for content after accounting
// for borders, title, and footer.
func (c *Context) ContentHeight() int {
	// 2 for top/bottom borders, 1 for title line, 1 for footer (when focused)
	overhead := 4
	if !c.focused || c.footer == "" {
		overhead = 3
	}
	h := c.height - overhead
	if h < 1 {
		return 1
	}
	return h
}

// ContentWidth returns the width available for content after accounting
// for borders and padding.
func (c *Context) ContentWidth() int {
	// 2 for left/right borders, 2 for padding
	w := c.width - 4
	if w < 1 {
		return 1
	}
	return w
}

// FooterProvider is implemented by models that provide custom footer text.
type FooterProvider interface {
	FooterText() string
}

// BadgeProvider is implemented by models that provide custom badge text.
type BadgeProvider interface {
	BadgeText() string
}
