## shelly virtual create

Create a virtual component

### Synopsis

Create a new virtual component on a Shelly Gen2+ device.

Available types:
  boolean  - True/false value
  number   - Numeric value with optional min/max
  text     - Text string value
  enum     - Selection from predefined options
  button   - Triggerable button
  group    - Component group

Virtual components are automatically assigned IDs in the range 200-299.

```
shelly virtual create <device> <type> [flags]
```

### Examples

```
  # Create a boolean component
  shelly virtual create kitchen boolean --name "Away Mode"

  # Create a number component
  shelly virtual create kitchen number --name "Temperature Offset"

  # Create with specific ID
  shelly virtual create kitchen boolean --name "Override" --id 205
```

### Options

```
  -h, --help          help for create
      --id int        Specific component ID (200-299, auto-assigned if not specified)
      --name string   Component display name
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

* [shelly virtual](shelly_virtual.md)	 - Manage virtual components

