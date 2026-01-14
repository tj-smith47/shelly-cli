## shelly switch on

Turn switch on

### Synopsis

Turn on a switch component on the specified device.

```
shelly switch on <device> [flags]
```

### Examples

```
  # Turn on switch
  shelly switch on <device>

  # Turn on specific switch by ID
  shelly switch on <device> --id 1

  # Turn on switch by name
  shelly switch on <device> --name "Kitchen Light"
```

### Options

```
  -h, --help          help for on
  -i, --id int        Switch component ID (default 0)
  -n, --name string   Switch name (alternative to --id)
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

* [shelly switch](shelly_switch.md)	 - Control switch components

