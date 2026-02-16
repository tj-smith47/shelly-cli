## shelly thermostat disable

Disable thermostat

### Synopsis

Disable a thermostat component.

When disabled, the thermostat will not actively control the valve
position based on temperature. The valve will typically remain
in its current position or close.

```
shelly thermostat disable <device> [flags]
```

### Examples

```
  # Disable thermostat
  shelly thermostat disable gateway

  # Disable specific thermostat
  shelly thermostat disable gateway --id 1
```

### Options

```
  -h, --help     help for disable
  -i, --id int   Thermostat component ID (default 0)
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

* [shelly thermostat](shelly_thermostat.md)	 - Manage thermostats

