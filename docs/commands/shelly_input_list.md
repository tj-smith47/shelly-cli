## shelly input list

List input components

### Synopsis

List all input components on a Shelly device with their current state.

Input components represent physical inputs (buttons, switches, sensors)
connected to the device. Each input has an ID, optional name, type
(button, switch, etc.), and current state (active/inactive).

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Name, Type, State (active/inactive)

```
shelly input list <device> [flags]
```

### Examples

```
  # List all inputs on a device
  shelly input list living-room

  # Output as JSON for scripting
  shelly input list living-room -o json

  # Get active inputs only
  shelly input list living-room -o json | jq '.[] | select(.state == true)'

  # List inputs by type
  shelly input list living-room -o json | jq '.[] | select(.type == "button")'

  # Get input IDs only
  shelly input list living-room -o json | jq -r '.[].id'

  # Monitor input state across multiple devices
  for dev in switch-1 switch-2; do
    echo "=== $dev ==="
    shelly input list "$dev" --no-color
  done

  # Short form
  shelly input ls living-room
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
      --offline                 Only read from cache, error on cache miss
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly input](shelly_input.md)	 - Manage input components

