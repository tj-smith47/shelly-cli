package base

import "testing"

func TestSizableModel(t *testing.T) {
	t.Parallel()

	t.Run("initial values are zero", func(t *testing.T) {
		t.Parallel()
		var s SizableModel
		if s.Width() != 0 {
			t.Errorf("expected width 0, got %d", s.Width())
		}
		if s.Height() != 0 {
			t.Errorf("expected height 0, got %d", s.Height())
		}
	})

	t.Run("SetSize updates dimensions", func(t *testing.T) {
		t.Parallel()
		var s SizableModel
		s = s.SetSize(100, 50)
		if s.Width() != 100 {
			t.Errorf("expected width 100, got %d", s.Width())
		}
		if s.Height() != 50 {
			t.Errorf("expected height 50, got %d", s.Height())
		}
	})

	t.Run("SetSize is immutable", func(t *testing.T) {
		t.Parallel()
		original := SizableModel{}
		updated := original.SetSize(100, 50)
		if original.Width() != 0 {
			t.Errorf("original should be unchanged, got width %d", original.Width())
		}
		if updated.Width() != 100 {
			t.Errorf("updated should have width 100, got %d", updated.Width())
		}
	})
}

func TestFocusableModel(t *testing.T) {
	t.Parallel()

	t.Run("initial values", func(t *testing.T) {
		t.Parallel()
		var f FocusableModel
		if f.IsFocused() {
			t.Error("expected not focused initially")
		}
		if f.PanelIndex() != 0 {
			t.Errorf("expected panel index 0, got %d", f.PanelIndex())
		}
	})

	t.Run("SetFocused", func(t *testing.T) {
		t.Parallel()
		var f FocusableModel
		f = f.SetFocused(true)
		if !f.IsFocused() {
			t.Error("expected focused after SetFocused(true)")
		}
		f = f.SetFocused(false)
		if f.IsFocused() {
			t.Error("expected not focused after SetFocused(false)")
		}
	})

	t.Run("SetPanelIndex", func(t *testing.T) {
		t.Parallel()
		var f FocusableModel
		f = f.SetPanelIndex(3)
		if f.PanelIndex() != 3 {
			t.Errorf("expected panel index 3, got %d", f.PanelIndex())
		}
	})

	t.Run("SetFocused is immutable", func(t *testing.T) {
		t.Parallel()
		original := FocusableModel{}
		updated := original.SetFocused(true)
		if original.IsFocused() {
			t.Error("original should be unchanged")
		}
		if !updated.IsFocused() {
			t.Error("updated should be focused")
		}
	})
}

func TestPanelModel(t *testing.T) {
	t.Parallel()

	t.Run("combines SizableModel and FocusableModel", func(t *testing.T) {
		t.Parallel()
		var p PanelModel
		p = p.SetSize(100, 50)
		p = p.SetFocused(true)
		p = p.SetPanelIndex(2)

		if p.Width() != 100 {
			t.Errorf("expected width 100, got %d", p.Width())
		}
		if p.Height() != 50 {
			t.Errorf("expected height 50, got %d", p.Height())
		}
		if !p.IsFocused() {
			t.Error("expected focused")
		}
		if p.PanelIndex() != 2 {
			t.Errorf("expected panel index 2, got %d", p.PanelIndex())
		}
	})

	t.Run("PanelModel methods are immutable", func(t *testing.T) {
		t.Parallel()
		original := PanelModel{}
		sized := original.SetSize(100, 50)
		focused := sized.SetFocused(true)
		indexed := focused.SetPanelIndex(3)

		if original.Width() != 0 {
			t.Error("original should be unchanged")
		}
		if sized.IsFocused() {
			t.Error("sized should not be focused")
		}
		if focused.PanelIndex() != 0 {
			t.Error("focused should have panel index 0")
		}
		if indexed.Width() != 100 || !indexed.IsFocused() || indexed.PanelIndex() != 3 {
			t.Error("indexed should have all values set")
		}
	})
}
