// Package yamlfmt provides YAML formatting with optional syntax highlighting.
package yamlfmt

import (
	"bytes"
	"encoding/json"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/output/synfmt"
)

// Formatter formats output as YAML with optional syntax highlighting.
type Formatter struct {
	Highlight bool // Enable syntax highlighting (disabled in --no-color/--plain)
}

// New creates a new YAML formatter with syntax highlighting.
func New() *Formatter {
	return &Formatter{
		Highlight: synfmt.ShouldHighlight(),
	}
}

// Format outputs data as YAML with optional syntax highlighting. The YAML
// carries the same keys, values and key order as the JSON encoding of data,
// so `-o yaml` and `-o json` always agree.
func (f *Formatter) Format(w io.Writer, data any) error {
	node, err := jsonShapedNode(data)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(node); err != nil {
		return err
	}
	if err := encoder.Close(); err != nil {
		return err
	}

	output := buf.String()
	if f.Highlight {
		output = synfmt.HighlightCode(output, "yaml")
	}
	_, err = io.WriteString(w, output)
	return err
}

// jsonShapedNode encodes data as JSON and reads it back as a YAML node. The
// json tags, omitempty, json:"-" and MarshalJSON methods then decide the keys,
// as they do for -o json; encoding the struct with yaml.v3 directly would use
// lowercased Go field names for every field without a yaml tag.
func jsonShapedNode(data any) (*yaml.Node, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	blockStyle(&doc)
	return &doc, nil
}

// blockStyle replaces the flow and quoting styles the JSON text gave the
// nodes with the styles yaml.v3 picks when it encodes the same Go values, so
// the output is block YAML and a string such as "on", "1.0" or "" stays quoted.
func blockStyle(n *yaml.Node) {
	n.Style = 0
	if n.Kind == yaml.ScalarNode && n.Tag == "!!str" {
		n.Style = stringStyle(n.Value)
	}
	for _, child := range n.Content {
		blockStyle(child)
	}
}

// stringStyle returns the style yaml.v3 encodes the string s with.
func stringStyle(s string) yaml.Style {
	out, err := yaml.Marshal(s)
	if err != nil || len(out) == 0 {
		return yaml.DoubleQuotedStyle
	}
	switch out[0] {
	case '"':
		return yaml.DoubleQuotedStyle
	case '\'':
		return yaml.SingleQuotedStyle
	case '|':
		return yaml.LiteralStyle
	case '>':
		return yaml.FoldedStyle
	default:
		return 0
	}
}
