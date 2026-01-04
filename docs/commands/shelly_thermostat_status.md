## shelly thermostat status

Show thermostat status

### Synopsis

Show detailed status of a thermostat component.

Displays:
- Current temperature
- Target temperature
- Valve position (0-100%)
- Operating mode
- Boost/Override status
- Errors and flags

```
shelly thermostat status <device> [flags]
```

### Examples

```
  # Show thermostat status (default ID 0)
  shelly thermostat status gateway

  # Show specific thermostat
  shelly thermostat status gateway --id 1

  # Output as JSON
  shelly thermostat status gateway --json
```

### Options

```
  -h, --help     help for status
  -i, --id int   Thermostat component ID (default 0)
      --json     Output as JSON
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
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

