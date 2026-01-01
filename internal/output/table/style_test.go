package table

import (
	"testing"

	"github.com/spf13/viper"
)

// mockModeChecker implements ModeChecker for testing.
type mockModeChecker struct {
	plain        bool
	colorEnabled bool
}

func (m *mockModeChecker) IsPlainMode() bool {
	return m.plain
}

func (m *mockModeChecker) ColorEnabled() bool {
	return m.colorEnabled
}

func TestPlainStyle(t *testing.T) {
	t.Parallel()

	style := PlainStyle()
	// Plain mode uses no borders for tab-separated output
	if style.BorderStyle != BorderNone {
		t.Error("PlainStyle should use no borders")
	}
	if style.ShowBorder {
		t.Error("PlainStyle should have ShowBorder disabled")
	}
	if !style.PlainMode {
		t.Error("PlainStyle should have PlainMode enabled")
	}
}

func TestNoColorStyle(t *testing.T) {
	t.Parallel()

	style := NoColorStyle()
	// No-color mode uses ASCII borders
	if style.BorderStyle != BorderASCII {
		t.Error("NoColorStyle should use ASCII borders")
	}
	if !style.ShowBorder {
		t.Error("NoColorStyle should have ShowBorder enabled")
	}
	if style.PlainMode {
		t.Error("NoColorStyle should not have PlainMode enabled")
	}
}

func TestDefaultStyle(t *testing.T) {
	t.Parallel()

	style := DefaultStyle()
	if style.BorderStyle != BorderRounded {
		t.Error("DefaultStyle should use rounded borders")
	}
	if !style.ShowBorder {
		t.Error("DefaultStyle should have ShowBorder enabled")
	}
	if style.PlainMode {
		t.Error("DefaultStyle should not have PlainMode enabled")
	}
}

func TestGetStyle(t *testing.T) {
	t.Parallel()

	t.Run("nil returns default", func(t *testing.T) {
		t.Parallel()
		style := GetStyle(nil)
		if style.BorderStyle != BorderRounded {
			t.Error("nil checker should return default style with rounded borders")
		}
	})

	t.Run("plain mode returns plain style with no borders", func(t *testing.T) {
		t.Parallel()
		checker := &mockModeChecker{plain: true, colorEnabled: true}
		style := GetStyle(checker)
		if style.BorderStyle != BorderNone {
			t.Error("plain mode should return plain style with no borders")
		}
		if !style.PlainMode {
			t.Error("plain mode should have PlainMode=true for tab-separated output")
		}
	})

	t.Run("no-color mode returns ASCII borders", func(t *testing.T) {
		t.Parallel()
		checker := &mockModeChecker{plain: false, colorEnabled: false}
		style := GetStyle(checker)
		if style.BorderStyle != BorderASCII {
			t.Error("no-color mode should return ASCII borders")
		}
	})

	t.Run("color enabled returns default style", func(t *testing.T) {
		t.Parallel()
		checker := &mockModeChecker{plain: false, colorEnabled: true}
		style := GetStyle(checker)
		if style.BorderStyle != BorderRounded {
			t.Error("color mode should return default style with rounded borders")
		}
	})
}

// Tests that use viper cannot be parallel due to global state.

//nolint:paralleltest // Uses viper global state
func TestShouldHideHeaders_Default(t *testing.T) {
	// Reset viper for clean test
	original := viper.GetBool("no-headers")
	viper.Set("no-headers", false)
	t.Cleanup(func() {
		viper.Set("no-headers", original)
	})

	hide := ShouldHideHeaders()
	if hide {
		t.Error("ShouldHideHeaders() default should be false")
	}
}

//nolint:paralleltest // Uses viper global state
func TestShouldHideHeaders_Set(t *testing.T) {
	original := viper.GetBool("no-headers")
	viper.Set("no-headers", true)
	t.Cleanup(func() {
		viper.Set("no-headers", original)
	})

	hide := ShouldHideHeaders()
	if !hide {
		t.Error("ShouldHideHeaders() should return true when set")
	}
}
