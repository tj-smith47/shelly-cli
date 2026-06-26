package cmdutil

import (
	"fmt"
	"strings"
)

// KeyValue is a parsed key/value assignment from command arguments.
type KeyValue struct {
	Key   string
	Value string
}

// ParseKeyValues parses command arguments into ordered key/value assignments,
// accepting three interchangeable separator styles so callers never have to
// remember a single required form:
//
//	key=value   (inline equals)
//	key:value   (inline colon)
//	key value   (space-separated: two arguments)
//
// Inline and space-separated assignments may be mixed in one invocation. A value
// may itself contain '=' or ':' — only the first separator in an inline token
// splits it. Returns an error if a key is missing its value or is empty.
func ParseKeyValues(args []string) ([]KeyValue, error) {
	pairs := make([]KeyValue, 0, len(args))
	for i := 0; i < len(args); {
		tok := args[i]
		if k, v, ok := splitInline(tok); ok {
			pairs = append(pairs, KeyValue{Key: k, Value: v})
			i++
			continue
		}
		if tok == "" {
			return nil, fmt.Errorf("empty key in arguments")
		}
		// Bare token → key; the next argument is its value.
		if i+1 >= len(args) {
			return nil, fmt.Errorf("missing value for %q (use key=value, key:value, or key value)", tok)
		}
		next := args[i+1]
		// In a multi-assignment invocation, a bare key followed by an inline pair
		// almost always means this key's value was omitted; flag it rather than
		// silently swallowing the next assignment as this key's value. With only
		// two arguments there is no ambiguity, so the value may contain a separator.
		if len(args) > 2 && strings.ContainsAny(next, "=:") {
			return nil, fmt.Errorf("missing value for %q (use key=value, key:value, or key value)", tok)
		}
		pairs = append(pairs, KeyValue{Key: tok, Value: next})
		i += 2
	}
	if len(pairs) == 0 {
		return nil, fmt.Errorf("no key/value pairs provided")
	}
	return pairs, nil
}

// splitInline splits a token on its first '=' or ':' separator. It returns
// ok=false when the token has no separator or the key portion would be empty.
func splitInline(tok string) (key, value string, ok bool) {
	eq := strings.IndexByte(tok, '=')
	colon := strings.IndexByte(tok, ':')
	idx := -1
	switch {
	case eq >= 0 && colon >= 0:
		idx = min(eq, colon)
	case eq >= 0:
		idx = eq
	case colon >= 0:
		idx = colon
	}
	if idx <= 0 {
		return "", "", false
	}
	return tok[:idx], tok[idx+1:], true
}
