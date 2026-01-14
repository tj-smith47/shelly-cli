## shelly rgb off

Turn rgb off

### Synopsis

Turn off a rgb component on the specified device.

```
shelly rgb off <device> [flags]
```

### Examples

```
  # Turn off rgb
  shelly rgb off <device>

  # Turn off specific rgb by ID
  shelly rgb off <device> --id 1

  # Turn off rgb by name
  shelly rgb off <device> --name "Kitchen Light"
```

### Options

```
  -h, --help          help for off
  -i, --id int        RGB component ID (default 0)
  -n, --name string   RGB name (alternative to --id)
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

* [shelly rgb](shelly_rgb.md)	 - Control RGB light components

