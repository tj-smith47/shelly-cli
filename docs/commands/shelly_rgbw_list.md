## shelly rgbw list

List rgbw components

### Synopsis

List all RGBW light components on the specified device with their current status.

RGBW components control color-capable lights with an additional white channel.
Each component has an ID, optional name, state (ON/OFF), RGB color values,
white channel value, brightness level, and power consumption if supported.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Name, State (ON/OFF), Color (R:G:B), White, Brightness (%), Power

```
shelly rgbw list <device> [flags]
```

### Examples

```
  # List all RGBW components on a device
  shelly rgbw list living-room

  # List RGBW components with JSON output
  shelly rgbw list living-room -o json

  # Get RGBW lights that are currently ON
  shelly rgbw list living-room -o json | jq '.[] | select(.output == true)'

  # Get current color and white values
  shelly rgbw list living-room -o json | jq '.[] | {id, r: .rgb.r, g: .rgb.g, b: .rgb.b, white}'

  # Find lights with white channel active
  shelly rgbw list living-room -o json | jq '.[] | select(.white > 0)'

  # Short forms
  shelly rgbw ls living-room
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

* [shelly rgbw](shelly_rgbw.md)	 - Control RGBW LED outputs

