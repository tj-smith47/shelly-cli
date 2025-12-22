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

  # Turn on specific switch ID
  shelly switch on <device> --id 1
```

### Options

```
  -h, --help     help for on
  -i, --id int   Switch component ID (default 0)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly switch](shelly_switch.md)	 - Control switch components

