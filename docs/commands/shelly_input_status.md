## shelly input status

Show input status

### Synopsis

Show the current status of a input component on the specified device.

```
shelly input status <device> [flags]
```

### Examples

```
  # Show input status
  shelly input status <device>

  # Show status with JSON output
  shelly input st <device> -o json
```

### Options

```
  -h, --help     help for status
  -i, --id int   Input component ID (default 0)
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

* [shelly input](shelly_input.md)	 - Manage input components

