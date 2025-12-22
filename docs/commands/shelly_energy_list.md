## shelly energy list

List energy monitoring components

### Synopsis

List all energy monitoring components (EM/EM1) on a device.

Shows component IDs and types for all energy monitors found on the device.
EM components are 3-phase monitors (Shelly Pro 3EM, etc.), EM1 components
are single-phase monitors (Shelly EM, Shelly Plus 1PM, etc.).

Use 'shelly energy status' with a component ID to get real-time readings.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Type

```
shelly energy list <device> [flags]
```

### Examples

```
  # List energy monitoring components on a device
  shelly energy list shelly-3em-pro

  # Output as JSON for scripting
  shelly energy list shelly-3em-pro -o json

  # Get IDs of 3-phase monitors
  shelly energy list shelly-3em-pro -o json | jq -r '.[] | select(.type | contains("3-phase")) | .id'

  # Count total energy components
  shelly energy list shelly-3em-pro -o json | jq length

  # Short form
  shelly energy ls shelly-3em-pro
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

* [shelly energy](shelly_energy.md)	 - Energy monitoring operations (EM/EM1 components)

