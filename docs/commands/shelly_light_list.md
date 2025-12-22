## shelly light list

List light components

### Synopsis

List all light components on the specified device with their current status.

Light components control dimmable lights. Each light has an ID, optional
name, state (ON/OFF), brightness level (percentage), and power consumption
if supported. Some lights also support color temperature or RGB values.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Name, State (ON/OFF), Brightness (%), Power (watts)

```
shelly light list <device> [flags]
```

### Examples

```
  # List all lights on a device
  shelly light list kitchen

  # List lights with JSON output
  shelly light list kitchen -o json

  # Get lights that are currently ON
  shelly light list kitchen -o json | jq '.[] | select(.output == true)'

  # Find lights below 50% brightness
  shelly light list kitchen -o json | jq '.[] | select(.brightness < 50)'

  # Calculate total light power consumption
  shelly light list kitchen -o json | jq '[.[].apower // 0] | add'

  # Get all light IDs
  shelly light list kitchen -o json | jq -r '.[].id'

  # Check lights across multiple devices
  for dev in kitchen bedroom living-room; do
    echo "=== $dev ==="
    shelly light list "$dev" --no-color
  done

  # Short forms
  shelly light ls kitchen
  shelly lt ls kitchen
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
      --no-headers              Hide table headers in output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly light](shelly_light.md)	 - Control light components

