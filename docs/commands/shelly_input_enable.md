## shelly input enable

Enable input component

### Synopsis

Enable an input component on a Shelly device.

When enabled, the input will respond to physical button presses or switch
state changes and trigger associated actions.

```
shelly input enable <device> [flags]
```

### Examples

```
  # Enable input on a device
  shelly input enable kitchen

  # Enable specific input by ID
  shelly input enable living-room --id 1

  # Using alias
  shelly input on bedroom
```

### Options

```
  -h, --help     help for enable
  -i, --id int   Input component ID (default 0)
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

* [shelly input](shelly_input.md)	 - Manage input components

