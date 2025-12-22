## shelly metrics json

Output metrics as JSON

### Synopsis

Output device metrics in JSON format.

Outputs power, voltage, current, and energy metrics from all registered
devices (or a specified subset) in a machine-readable JSON format.

Use --continuous to stream metrics at regular intervals, or run once
for a single snapshot.

```
shelly metrics json [flags]
```

### Examples

```
  # Output metrics once to stdout
  shelly metrics json

  # Output for specific devices
  shelly metrics json --devices kitchen,living-room

  # Stream metrics every 10 seconds
  shelly metrics json --continuous --interval 10s

  # Save to file
  shelly metrics json --output metrics.json
```

### Options

```
  -c, --continuous          Stream metrics continuously
      --devices strings     Devices to include (default: all registered)
  -h, --help                help for json
  -i, --interval duration   Collection interval for continuous mode (default 10s)
  -o, --output string       Output file (default: stdout)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly metrics](shelly_metrics.md)	 - Export device metrics

