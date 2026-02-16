## shelly monitor status

Monitor device status in real-time

### Synopsis

Monitor a device's status in real-time with automatic refresh.

Status includes switches, covers, lights, energy meters, and other components.
Press Ctrl+C to stop monitoring.

```
shelly monitor status <device> [flags]
```

### Examples

```
  # Monitor device status every 2 seconds
  shelly monitor status living-room

  # Monitor with custom interval
  shelly monitor status living-room --interval 5s

  # Monitor for a specific number of updates
  shelly monitor status living-room --count 10
```

### Options

```
  -n, --count int           Number of updates (0 = unlimited)
  -h, --help                help for status
  -i, --interval duration   Refresh interval (default 2s)
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

