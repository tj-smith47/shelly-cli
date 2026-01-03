## shelly monitor power

Monitor power consumption in real-time

### Synopsis

Monitor a device's power consumption in real-time.

Shows power (W), voltage (V), current (A), and energy (Wh) for all
energy meters and power meters on the device.
Press Ctrl+C to stop monitoring.

```
shelly monitor power <device> [flags]
```

### Examples

```
  # Monitor power consumption
  shelly monitor power living-room

  # Monitor with 1-second interval
  shelly monitor power living-room --interval 1s

  # Monitor for a specific number of updates
  shelly monitor power living-room --count 10
```

### Options

```
  -n, --count int           Number of updates (0 = unlimited)
  -h, --help                help for power
  -i, --interval duration   Refresh interval (default 2s)
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly monitor](shelly_monitor.md)	 - Real-time device monitoring

