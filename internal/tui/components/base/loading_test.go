package base

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
)

func TestLoadingState(t *testing.T) {
	t.Parallel()

	t.Run("NewLoadingState creates inactive state", func(t *testing.T) {
		t.Parallel()
		s := NewLoadingState(loading.WithMessage("Testing..."))
		if s.IsLoading() {
			t.Error("expected not loading initially")
		}
	})

	t.Run("SetLoading activates loading", func(t *testing.T) {
		t.Parallel()
		s := NewLoadingState()
		s = s.SetLoading(true)
		if !s.IsLoading() {
			t.Error("expected loading after SetLoading(true)")
		}
	})

	t.Run("StopLoading deactivates loading", func(t *testing.T) {
		t.Parallel()
		s := NewLoadingState()
		s = s.SetLoading(true)
		s = s.StopLoading()
		if s.IsLoading() {
			t.Error("expected not loading after StopLoading")
		}
	})

	t.Run("StartLoading activates and returns tick", func(t *testing.T) {
		t.Parallel()
		s := NewLoadingState()
		s, cmd := s.StartLoading()
		if !s.IsLoading() {
			t.Error("expected loading after StartLoading")
		}
		if cmd == nil {
			t.Error("expected tick command from StartLoading")
		}
	})

	t.Run("View returns empty when not loading", func(t *testing.T) {
		t.Parallel()
		s := NewLoadingState()
		if s.View() != "" {
			t.Error("expected empty view when not loading")
		}
	})

	t.Run("View returns loader view when loading", func(t *testing.T) {
		t.Parallel()
		s := NewLoadingState(loading.WithMessage("Testing..."))
		s = s.SetLoading(true)
		view := s.View()
		if view == "" {
			t.Error("expected non-empty view when loading")
		}
	})

	t.Run("SetSize updates loader size", func(t *testing.T) {
		t.Parallel()
		s := NewLoadingState()
		updated := s.SetSize(100, 50)
		// Size is applied to loader internally - just verify no panic
		// and that the method returns a new LoadingState
		_ = updated.loader.View() // access to verify no panic
	})

	t.Run("LoadingState is immutable", func(t *testing.T) {
		t.Parallel()
		original := NewLoadingState()
		updated := original.SetLoading(true)
		if original.IsLoading() {
			t.Error("original should be unchanged")
		}
		if !updated.IsLoading() {
			t.Error("updated should be loading")
		}
	})
}

func TestDualLoadingState(t *testing.T) {
	t.Parallel()

	t.Run("NewDualLoadingState creates inactive states", func(t *testing.T) {
		t.Parallel()
		d := NewDualLoadingState(
			[]loading.Option{loading.WithMessage("Primary")},
			[]loading.Option{loading.WithMessage("Secondary")},
		)
		if d.IsLoading() {
			t.Error("expected not loading initially")
		}
	})

	t.Run("IsLoading returns true when primary is loading", func(t *testing.T) {
		t.Parallel()
		d := NewDualLoadingState(nil, nil)
		d = d.SetPrimary(d.Primary().SetLoading(true))
		if !d.IsLoading() {
			t.Error("expected loading when primary is loading")
		}
	})

	t.Run("IsLoading returns true when secondary is loading", func(t *testing.T) {
		t.Parallel()
		d := NewDualLoadingState(nil, nil)
		d = d.SetSecondary(d.Secondary().SetLoading(true))
		if !d.IsLoading() {
			t.Error("expected loading when secondary is loading")
		}
	})

	t.Run("View returns primary view when both loading", func(t *testing.T) {
		t.Parallel()
		d := NewDualLoadingState(
			[]loading.Option{loading.WithMessage("Primary")},
			[]loading.Option{loading.WithMessage("Secondary")},
		)
		d = d.SetPrimary(d.Primary().SetLoading(true))
		d = d.SetSecondary(d.Secondary().SetLoading(true))
		view := d.View()
		if view == "" {
			t.Error("expected non-empty view")
		}
	})

	t.Run("View returns empty when not loading", func(t *testing.T) {
		t.Parallel()
		d := NewDualLoadingState(nil, nil)
		if d.View() != "" {
			t.Error("expected empty view when not loading")
		}
	})

	t.Run("SetSize updates both loaders", func(t *testing.T) {
		t.Parallel()
		d := NewDualLoadingState(nil, nil)
		updated := d.SetSize(100, 50)
		// Size is applied to loaders internally - just verify no panic
		// and that the method returns a new DualLoadingState
		_ = updated.primary.loader.View()   // access to verify no panic
		_ = updated.secondary.loader.View() // access to verify no panic
	})

	t.Run("DualLoadingState is immutable", func(t *testing.T) {
		t.Parallel()
		original := NewDualLoadingState(nil, nil)
		updated := original.SetPrimary(original.Primary().SetLoading(true))
		if original.IsLoading() {
			t.Error("original should be unchanged")
		}
		if !updated.IsLoading() {
			t.Error("updated should be loading")
		}
	})
}
