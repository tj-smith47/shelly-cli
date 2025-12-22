## shelly power list

List power meter components

### Synopsis

List all power meter components (PM/PM1) on a device.

PM components are power meters typically found on multi-channel devices
(Shelly Pro 4PM, etc.). PM1 components are single-channel power meters
found on devices like Shelly Plus 1PM.

Use 'shelly power status' with a component ID to get real-time readings.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Type (PM or PM1)

```
shelly power list <device> [flags]
```

### Examples

```
  # List power meter components on a device
  shelly power list living-room

  # Output as JSON for scripting
  shelly power list living-room -o json

  # Get count of power meter components
  shelly power list living-room -o json | jq length

  # Get IDs of PM1 components only
  shelly power list living-room -o json | jq -r '.[] | select(.type == "PM1") | .id'

  # Check all devices for power meters
  shelly device list -o json | jq -r '.[].name' | while read dev; do
    count=$(shelly power list "$dev" -o json 2>/dev/null | jq length)
    [ "$count" -gt 0 ] && echo "$dev: $count power meters"
  done

  # Short form
  shelly power ls living-room
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly power](shelly_power.md)	 - Power meter operations (PM/PM1 components)

