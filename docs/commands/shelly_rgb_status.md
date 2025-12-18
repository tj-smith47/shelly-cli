## shelly rgb status

Show rgb status

### Synopsis

Show the current status of a rgb component on the specified device.

```
shelly rgb status <device> [flags]
```

### Examples

```
  # Show rgb status
  shelly rgb status <device>

  # Show status with JSON output
  shelly rgb st <device> -o json
```

### Options

```
  -h, --help     help for status
  -i, --id int   RGB component ID (default 0)
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

* [shelly rgb](shelly_rgb.md)	 - Control RGB light components

