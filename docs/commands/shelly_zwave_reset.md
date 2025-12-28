## shelly zwave reset

Show factory reset instructions

### Synopsis

Show factory reset instructions for a Z-Wave device.

WARNING: Factory reset should only be used when the gateway is missing
or inoperable. All custom parameters, associations, and routing
information will be lost.

```
shelly zwave reset <model> [flags]
```

### Examples

```
  # Show factory reset instructions
  shelly zwave reset SNSW-001P16ZW
```

### Options

```
  -h, --help   help for reset
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

* [shelly zwave](shelly_zwave.md)	 - Z-Wave device utilities

