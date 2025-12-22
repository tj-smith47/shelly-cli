## shelly theme list

List available themes

### Synopsis

List all available color themes.

The CLI includes 280+ themes from bubbletint. Use --filter to search
for themes by name pattern (case-insensitive).

Use 'shelly theme set <name>' to apply a theme.
Use 'shelly theme preview <name>' to see a theme before applying.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: Theme (name), Current (checkmark if active)

```
shelly theme list [flags]
```

### Examples

```
  # List all themes
  shelly theme list

  # Filter themes by name pattern
  shelly theme list --filter dark
  shelly theme list --filter nord

  # Output as JSON
  shelly theme list -o json

  # Get theme names only
  shelly theme list -o json | jq -r '.[].id'

  # Find current theme
  shelly theme list -o json | jq -r '.[] | select(.current) | .id'

  # Count themes matching pattern
  shelly theme list --filter monokai -o json | jq length

  # Random theme selection
  shelly theme set "$(shelly theme list -o json | jq -r '.[].id' | shuf -n1)"

  # Short form
  shelly theme ls
```

### Options

```
      --filter string   Filter themes by name pattern
  -h, --help            help for list
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly theme](shelly_theme.md)	 - Manage CLI color themes

