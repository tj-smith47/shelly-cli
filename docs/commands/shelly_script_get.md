## shelly script get

Get script code or status

### Synopsis

Get the source code or status of a script on a Gen2+ device.

By default, displays the script source code. Use --status to show
detailed status including memory usage and errors.

```
shelly script get <device> <id> [flags]
```

### Examples

```
  # Get script code
  shelly script get living-room 1

  # Get script status
  shelly script get living-room 1 --status
```

### Options

```
  -h, --help     help for get
      --status   Show script status instead of code
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

* [shelly script](shelly_script.md)	 - Manage device scripts

