## shelly rgbw on

Turn rgbw on

### Synopsis

Turn on a rgbw component on the specified device.

```
shelly rgbw on <device> [flags]
```

### Examples

```
  # Turn on rgbw
  shelly rgbw on <device>

  # Turn on specific rgbw by ID
  shelly rgbw on <device> --id 1

  # Turn on rgbw by name
  shelly rgbw on <device> --name "Kitchen Light"
```

### Options

```
  -h, --help          help for on
  -i, --id int        RGBW component ID (default 0)
  -n, --name string   RGBW name (alternative to --id)
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

* [shelly rgbw](shelly_rgbw.md)	 - Control RGBW LED outputs

