package base

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
)

// LoadingState manages a loading spinner and its active state.
// It encapsulates the common pattern of having a loading flag and loader model.
//
// Example usage:
//
//	type Model struct {
//	    base.LoadingState
//	    // other fields
//	}
//
//	func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
//	    if cmd := m.LoadingState.ForwardTick(msg); cmd != nil {
//	        return m, cmd
//	    }
//	    // handle other messages
//	}
type LoadingState struct {
	loading bool
	loader  loading.Model
}

// NewLoadingState creates a LoadingState with a configured loader.
func NewLoadingState(opts ...loading.Option) LoadingState {
	return LoadingState{
		loading: false,
		loader:  loading.New(opts...),
	}
}

// IsLoading returns whether the loading state is active.
func (s LoadingState) IsLoading() bool {
	return s.loading
}

// Loader returns the underlying loader model.
func (s LoadingState) Loader() loading.Model {
	return s.loader
}

// SetLoading sets the loading state and returns the updated LoadingState.
func (s LoadingState) SetLoading(active bool) LoadingState {
	s.loading = active
	return s
}

// SetLoader sets the loader model and returns the updated LoadingState.
func (s LoadingState) SetLoader(loader loading.Model) LoadingState {
	s.loader = loader
	return s
}

// StartLoading sets loading to true and returns the tick command to start animation.
func (s LoadingState) StartLoading() (LoadingState, tea.Cmd) {
	s.loading = true
	return s, s.loader.Tick()
}

// StopLoading sets loading to false.
func (s LoadingState) StopLoading() LoadingState {
	s.loading = false
	return s
}

// Update forwards tick messages to the loader when loading is active.
// Returns a command if the tick was handled, nil otherwise.
// Use this in your Update method to handle loader animation.
//
// Example:
//
//	if m.IsLoading() {
//	    m.LoadingState, cmd = m.LoadingState.Update(msg)
//	    if cmd != nil {
//	        return m, cmd
//	    }
//	}
func (s LoadingState) Update(msg tea.Msg) (LoadingState, tea.Cmd) {
	if !s.loading {
		return s, nil
	}
	var cmd tea.Cmd
	s.loader, cmd = s.loader.Update(msg)
	return s, cmd
}

// View returns the loader's view if loading, empty string otherwise.
func (s LoadingState) View() string {
	if !s.loading {
		return ""
	}
	return s.loader.View()
}

// SetSize updates the loader's size.
func (s LoadingState) SetSize(width, height int) LoadingState {
	s.loader = s.loader.SetSize(width, height)
	return s
}

// DualLoadingState manages two loading states (e.g., export and import).
// Use this when a component has multiple independent operations.
type DualLoadingState struct {
	primary   LoadingState
	secondary LoadingState
}

// NewDualLoadingState creates a DualLoadingState with two configured loaders.
func NewDualLoadingState(primaryOpts, secondaryOpts []loading.Option) DualLoadingState {
	return DualLoadingState{
		primary:   NewLoadingState(primaryOpts...),
		secondary: NewLoadingState(secondaryOpts...),
	}
}

// Primary returns the primary loading state.
func (d DualLoadingState) Primary() LoadingState {
	return d.primary
}

// Secondary returns the secondary loading state.
func (d DualLoadingState) Secondary() LoadingState {
	return d.secondary
}

// SetPrimary sets the primary loading state.
func (d DualLoadingState) SetPrimary(s LoadingState) DualLoadingState {
	d.primary = s
	return d
}

// SetSecondary sets the secondary loading state.
func (d DualLoadingState) SetSecondary(s LoadingState) DualLoadingState {
	d.secondary = s
	return d
}

// IsLoading returns true if either state is loading.
func (d DualLoadingState) IsLoading() bool {
	return d.primary.IsLoading() || d.secondary.IsLoading()
}

// Update forwards tick messages to the active loader(s).
func (d DualLoadingState) Update(msg tea.Msg) (DualLoadingState, tea.Cmd) {
	var cmds []tea.Cmd

	if d.primary.IsLoading() {
		var cmd tea.Cmd
		d.primary, cmd = d.primary.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if d.secondary.IsLoading() {
		var cmd tea.Cmd
		d.secondary, cmd = d.secondary.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return d, tea.Batch(cmds...)
}

// View returns the active loader's view (primary takes precedence).
func (d DualLoadingState) View() string {
	if d.primary.IsLoading() {
		return d.primary.View()
	}
	if d.secondary.IsLoading() {
		return d.secondary.View()
	}
	return ""
}

// SetSize updates both loaders' sizes.
func (d DualLoadingState) SetSize(width, height int) DualLoadingState {
	d.primary = d.primary.SetSize(width, height)
	d.secondary = d.secondary.SetSize(width, height)
	return d
}
