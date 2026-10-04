package factories

import "strings"

// pluralize returns the lower-case plural form of a resource or component
// name (switch -> switches, alias -> aliases, light -> lights).
func pluralize(word string) string {
	word = strings.ToLower(word)
	switch {
	case strings.HasSuffix(word, "ch") || strings.HasSuffix(word, "sh") ||
		strings.HasSuffix(word, "x") || strings.HasSuffix(word, "s"):
		return word + "es"
	default:
		return word + "s"
	}
}

// withArticle returns the word preceded by "a" or "an" (a scene, an alias,
// an rgb). Every generated help string that names one resource goes through
// it so the article always matches the word.
func withArticle(word string) string {
	lower := strings.ToLower(word)
	// "rgb" and "rgbw" are spelled out when read aloud, starting with "ar".
	if strings.HasPrefix(lower, "rgb") || (lower != "" && strings.ContainsRune("aeiou", rune(lower[0]))) {
		return "an " + word
	}
	return "a " + word
}
