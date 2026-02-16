## shelly monitor events

Monitor device events in real-time

### Synopsis

Monitor device events via WebSocket subscription.

Events include state changes, notifications, and status updates.
Press Ctrl+C to stop monitoring.

```
shelly monitor events <device> [flags]
```

### Examples

```
  # Monitor all events
  shelly monitor events living-room

  # Filter events by component
  shelly monitor events living-room --filter switch

  # Output events as JSON
  shelly monitor events living-room -o json
```

### Options

```
  -f, --filter string   Filter events by component type
  -h, --help            help for events
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
  -F, --fields                  Print available field names for use with --jq and --template
  -Q, --jq stringArray          Apply jq expression to filter output (repeatable, joined with |)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --offline                 Only read from cache, error on cache miss
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly monitor](shelly_monitor.md)	 - Real-time device monitoring

