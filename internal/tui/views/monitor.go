package views

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/monitor"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

// MonitorDeps holds dependencies for the monitor view.
type MonitorDeps struct {
	Ctx         context.Context
	Svc         *shelly.Service
	IOS         *iostreams.IOStreams
	EventStream *automation.EventStream // Shared event stream
}

// Validate ensures all required dependencies are set.
func (d MonitorDeps) Validate() error {
	if d.Ctx == nil {
		return errNilContext
	}
	if d.Svc == nil {
		return errNilService
	}
	if d.IOS == nil {
		return errNilIOStreams
	}
	if d.EventStream == nil {
		return errNilEventStream
	}
	return nil
}

// Monitor is the real-time monitoring view.
type Monitor struct {
	ctx context.Context
	id  ViewID

	monitor monitor.Model

	width  int
	height int
}

// NewMonitor creates a new monitor view.
func NewMonitor(deps MonitorDeps) *Monitor {
	if err := deps.Validate(); err != nil {
		panic("monitor: " + err.Error())
	}

	return &Monitor{
		ctx: deps.Ctx,
		id:  tabs.TabMonitor,
		monitor: monitor.New(monitor.Deps{
			Ctx:         deps.Ctx,
			Svc:         deps.Svc,
			IOS:         deps.IOS,
			EventStream: deps.EventStream,
		}),
	}
}

// ID returns the view ID.
func (m *Monitor) ID() ViewID {
	return m.id
}

// Init returns the initial command for the monitor view.
func (m *Monitor) Init() tea.Cmd {
	return m.monitor.Init()
}

// Update handles messages for the monitor view.
func (m *Monitor) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmd tea.Cmd
	m.monitor, cmd = m.monitor.Update(msg)
	return m, cmd
}

// View renders the monitor view.
func (m *Monitor) View() string {
	return m.monitor.View()
}

// SetSize sets the view dimensions.
func (m *Monitor) SetSize(width, height int) View {
	m.width = width
	m.height = height
	m.monitor = m.monitor.SetSize(width, height)
	return m
}
