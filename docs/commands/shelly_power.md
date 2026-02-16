## shelly power

Power meter operations (PM/PM1 components)

### Synopsis

Manage and monitor power meter components.

This command works with PM and PM1 power meter components found on
Shelly devices that include energy metering functionality:
  - Shelly Plus PM (single-phase power meter)
  - Shelly Pro series with PM components
  - Various Plus/Pro devices with built-in power metering

PM/PM1 components provide real-time measurements including:
  - Voltage, current, and power
  - Frequency
  - Accumulated energy (total and by-minute)
  - Return energy (for bidirectional meters)

For professional energy monitors (EM/EM1 components), use 'shelly energy'.

### Examples

```
  # List power meter components
  shelly power list kitchen

  # Get current power status
  shelly pm status living-room

  # Check power consumption with JSON output
  shelly power status kitchen -o json
```

### Options

```
  -h, --help   help for power
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly power list](shelly_power_list.md)	 - List power meter components
* [shelly power status](shelly_power_status.md)	 - Show power meter status

