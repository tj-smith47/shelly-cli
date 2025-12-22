## shelly mock scenario

Load a test scenario

### Synopsis

Load a pre-defined test scenario with multiple mock devices.

Built-in scenarios:
  home     - Basic home setup (3 devices)
  office   - Office setup (5 devices)
  minimal  - Single device for quick testing

```
shelly mock scenario <name> [flags]
```

### Examples

```
  # Load home scenario
  shelly mock scenario home

  # Load office scenario
  shelly mock scenario office
```

### Options

```
  -h, --help   help for scenario
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

* [shelly mock](shelly_mock.md)	 - Mock device mode for testing

