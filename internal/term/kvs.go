package term

import (
	"encoding/json"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayKVSRaw prints just the raw value from a KVS result.
func DisplayKVSRaw(ios *iostreams.IOStreams, result *shelly.KVSGetResult) {
	switch v := result.Value.(type) {
	case string:
		ios.Println(v)
	case nil:
		ios.Println("null")
	default:
		// For other types (numbers, bools), use JSON encoding
		data, err := json.Marshal(v)
		if err != nil {
			ios.Printf("%v\n", v)
			return
		}
		ios.Println(string(data))
	}
}

// DisplayKVSResult prints a formatted KVS get result.
func DisplayKVSResult(ios *iostreams.IOStreams, key string, result *shelly.KVSGetResult) {
	label := theme.Highlight()
	ios.Printf("%s: %s\n", label.Render("Key"), key)
	ios.Printf("%s: %s\n", label.Render("Value"), output.FormatJSONValue(result.Value))
	ios.Printf("%s: %s\n", label.Render("Type"), output.ValueType(result.Value))
	ios.Printf("%s: %s\n", label.Render("Etag"), result.Etag)
}

// DisplayKVSKeys prints a table of KVS keys.
func DisplayKVSKeys(ios *iostreams.IOStreams, result *shelly.KVSListResult) {
	if len(result.Keys) == 0 {
		ios.NoResults("No keys stored")
		return
	}

	ios.Title("KVS Keys")
	ios.Println()

	table := output.NewTable("Key")
	for _, key := range result.Keys {
		table.AddRow(key)
	}
	printTable(ios, table)

	ios.Printf("\n%d key(s), revision %d\n", len(result.Keys), result.Rev)
}

// DisplayKVSItems prints a table of KVS key-value pairs.
func DisplayKVSItems(ios *iostreams.IOStreams, items []shelly.KVSItem) {
	ios.Title("KVS Data")
	ios.Println()

	table := output.NewTable("Key", "Value", "Type")
	for _, item := range items {
		table.AddRow(item.Key, output.FormatDisplayValue(item.Value), output.ValueType(item.Value))
	}
	printTable(ios, table)

	ios.Printf("\n%d key(s)\n", len(items))
}
