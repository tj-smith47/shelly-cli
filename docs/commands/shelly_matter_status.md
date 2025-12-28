## shelly matter status

Show Matter status

### Synopsis

Show Matter connectivity status for a Shelly device.

Displays:
- Whether Matter is enabled
- Commissionable status (can be added to a fabric)
- Number of paired fabrics
- Network information when connected

```
shelly matter status <device> [flags]
```

### Examples

```
  # Show Matter status
  shelly matter status living-room

  # Output as JSON
  shelly matter status living-room --json
```

### Options

```
  -f, --format string   Output format: text, json (default "text")
  -h, --help            help for status
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

* [shelly matter](shelly_matter.md)	 - Manage Matter connectivity

