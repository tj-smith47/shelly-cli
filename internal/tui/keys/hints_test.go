package keys

import "testing"

func TestFormatHints_AllFit(t *testing.T) {
	t.Parallel()
	hints := []Hint{
		{"e", "edit"},
		{"n", "new"},
		{"d", "del"},
	}
	got := FormatHints(hints, 100)
	want := "e:edit n:new d:del"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatHints_DropsFromEnd(t *testing.T) {
	t.Parallel()
	hints := []Hint{
		{"e", "edit"},
		{"n", "new"},
		{"d", "del"},
	}
	// "e:edit n:new" = 12 chars
	got := FormatHints(hints, 12)
	want := "e:edit n:new"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatHints_OnlyFirst(t *testing.T) {
	t.Parallel()
	hints := []Hint{
		{"e", "edit"},
		{"n", "new"},
	}
	got := FormatHints(hints, 6)
	want := "e:edit"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatHints_TooNarrowReturnsFirst(t *testing.T) {
	t.Parallel()
	hints := []Hint{
		{"e", "edit"},
	}
	got := FormatHints(hints, 3)
	want := "e:edit" // Returns anyway, renderer truncates
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatHints_Empty(t *testing.T) {
	t.Parallel()
	got := FormatHints(nil, 100)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestFormatHints_ZeroWidth(t *testing.T) {
	t.Parallel()
	hints := []Hint{{"e", "edit"}}
	got := FormatHints(hints, 0)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestFormatHints_GroupedKeys(t *testing.T) {
	t.Parallel()
	hints := []Hint{
		{"j/k", "nav"},
		{"e", "edit"},
	}
	got := FormatHints(hints, 100)
	want := "j/k:nav e:edit"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
