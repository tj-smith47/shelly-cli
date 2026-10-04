package factories

import "testing"

func TestPluralize(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"alias":    "aliases",
		"watch":    "watches",
		"Light":    "lights",
		"scene":    "scenes",
		"template": "templates",
		"rgbw":     "rgbws",
	}
	for word, want := range tests {
		if got := pluralize(word); got != want {
			t.Errorf("pluralize(%q) = %q, want %q", word, got, want)
		}
	}
}

func TestWithArticle(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"alias":    "an alias",
		"input":    "an input",
		"scene":    "a scene",
		"template": "a template",
		"rgb":      "an rgb",
		"RGBW":     "an RGBW",
		"cover":    "a cover",
	}
	for word, want := range tests {
		if got := withArticle(word); got != want {
			t.Errorf("withArticle(%q) = %q, want %q", word, got, want)
		}
	}
}
