## shelly link set

Set a device power link

### Synopsis

Set a parent-child power link between devices.

The child device is powered by a switch on the parent device. When the
child is offline, its state can be derived from the parent switch state.

```
shelly link set <child-device> <parent-device> [flags]
```

### Examples

```
  # Link bulb to switch:0 on bedroom-2pm
  shelly link set bulb-duo bedroom-2pm

  # Link to a specific switch ID
  shelly link set garage-light garage-switch --switch-id 1

  # Update an existing link
  shelly link set bulb-duo new-switch
```

### Options

```
  -h, --help            help for set
      --switch-id int   Switch component ID on the parent device
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

* [shelly link](shelly_link.md)	 - Manage device power links

