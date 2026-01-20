// Package views provides view management for the TUI.
package views

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

// DashboardDeps holds dependencies for the Dashboard view.
type DashboardDeps struct {
	Ctx context.Context
}

// Validate ensures all required dependencies are set.
func (d DashboardDeps) Validate() error {
	if d.Ctx == nil {
		return errNilContext
	}
	return nil
}

// Dashboard is the main dashboard view showing devices, events, and status.
// This view displays a multi-panel layout with:
// - Events stream (left column)
// - Device list (center column)
// - Device info or JSON overlay (right column)
//
// The Dashboard is implemented as a thin wrapper that signals app.go
// to render the multi-panel layout. The actual rendering logic remains
// in app.go to minimize refactoring while enabling the tabbed interface.
// Panel focus is managed by focusState (single source of truth in app.go).
type Dashboard struct {
	ctx    context.Context
	id     ViewID
	width  int
	height int

	// selectedDevice tracks the currently selected device for propagation
	selectedDevice string
}

// NewDashboard creates a new Dashboard view.
func NewDashboard(deps DashboardDeps) *Dashboard {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("dashboard view init", err)
		panic("views/dashboard: invalid deps: " + err.Error())
	}

	return &Dashboard{
		ctx: deps.Ctx,
		id:  tabs.TabDashboard,
	}
}

// ID returns the view ID.
func (d *Dashboard) ID() ViewID {
	return d.id
}

// Init returns the initial command for the Dashboard.
func (d *Dashboard) Init() tea.Cmd {
	return nil
}

// Update handles messages for the Dashboard view.
// Most updates are handled by app.go; this just tracks state for the View interface.
func (d *Dashboard) Update(msg tea.Msg) (View, tea.Cmd) {
	if dsMsg, ok := msg.(DashboardDeviceSelectedMsg); ok {
		if d.selectedDevice != dsMsg.Device {
			d.selectedDevice = dsMsg.Device
			// Emit message for other views
			return d, func() tea.Msg {
				return DeviceSelectedMsg(dsMsg)
			}
		}
	}

	return d, nil
}

// View renders the Dashboard.
// Returns empty string because app.go handles the actual rendering
// when the Dashboard view is active. This is a delegation pattern.
func (d *Dashboard) View() string {
	// The actual rendering is done by app.go's renderMultiPanelLayout.
	// This method returns empty so the ViewManager knows to delegate.
	return ""
}

// SetSize sets the view dimensions.
func (d *Dashboard) SetSize(width, height int) View {
	d.width = width
	d.height = height
	return d
}

// SelectedDevice returns the currently selected device.
func (d *Dashboard) SelectedDevice() string {
	return d.selectedDevice
}

// DashboardDeviceSelectedMsg is sent when a device is selected in the Dashboard.
// This is an internal message that triggers DeviceSelectedMsg emission.
type DashboardDeviceSelectedMsg struct {
	Device  string
	Address string
}

// IsDashboardView returns true if this is the Dashboard view.
// Used by app.go to determine if it should render the multi-panel layout.
func (d *Dashboard) IsDashboardView() bool {
	return true
}
