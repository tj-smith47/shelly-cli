## shelly power list

List power meter components

### Synopsis

List every component on a device that meters power, with its live reading.

That is any PM or PM1 power meter, any switch, cover or light that
meters its load (Plus 1PM, Plus 2PM, Plus Plug, Pro 4PM, dimmers, RGBW
PM), any EM or EM1 energy monitor, and the meters of a Gen1 device
(Shelly 1PM, Plug S, Duo bulbs, EM).

Use 'shelly power status' with a component ID for one component in
full.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output; each item carries name, type, id, power (watts) and
the full reading under em, em1 or meter.

Columns: Device, Component, Voltage, Current, Power, Energy

```
shelly power list <device> [flags]
```

### Examples

```
  # List the components that meter power on a device
  shelly power list living-room

  # Output as JSON for scripting
  shelly power list living-room -o json

  # Count the components that meter power
  shelly power list living-room -o json | jq length

  # IDs of the switch channels that meter power
  shelly power list living-room -o json | jq -r '.[] | select(.type == "switch") | .id'

  # Total power of a device in watts
  shelly power list living-room -o json | jq '[.[].power] | add'

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
      --raw                     Print the exact device response(s) as a JSON array and suppress normal output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly power](shelly_power.md)	 - Power meter operations (PM/PM1 components)

