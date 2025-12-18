## shelly light on

Turn light on

### Synopsis

Turn on a light component on the specified device.

```
shelly light on <device> [flags]
```

### Examples

```
  # Turn on light
  shelly light on <device>

  # Turn on specific light ID
  shelly light on <device> --id 1
```

### Options

```
  -h, --help     help for on
  -i, --id int   Light component ID (default 0)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly light](shelly_light.md)	 - Control light components

