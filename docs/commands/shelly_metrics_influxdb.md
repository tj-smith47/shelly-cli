## shelly metrics influxdb

Output metrics in InfluxDB line protocol

### Synopsis

Output device metrics in InfluxDB line protocol format.

Outputs power, voltage, current, and energy metrics from all registered
devices (or a specified subset) in InfluxDB line protocol format suitable
for piping to InfluxDB or Telegraf.

Format: measurement,tags field=value,field=value timestamp

Use --continuous to stream metrics at regular intervals.

```
shelly metrics influxdb [flags]
```

### Examples

```
  # Output metrics once to stdout
  shelly metrics influxdb

  # Output for specific devices with custom measurement name
  shelly metrics influxdb --devices kitchen --measurement home_power

  # Stream metrics every 10 seconds
  shelly metrics influxdb --continuous --interval 10s

  # Add custom tags
  shelly metrics influxdb --tags location=home,floor=1

  # Pipe directly to InfluxDB (requires influx CLI)
  shelly metrics influxdb | influx write -b mybucket
```

### Options

```
  -c, --continuous           Stream metrics continuously
      --devices strings      Devices to include (default: all registered)
  -h, --help                 help for influxdb
  -i, --interval duration    Collection interval for continuous mode (default 10s)
  -m, --measurement string   Measurement name (default "shelly")
  -o, --output string        Output file (default: stdout)
  -t, --tags strings         Additional tags (key=value)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --offline                 Only read from cache, error on cache miss
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly metrics](shelly_metrics.md)	 - Export device metrics

