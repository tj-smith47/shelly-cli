## shelly zwave config

Show common configuration parameters

### Synopsis

Show common Z-Wave configuration parameters for Wave devices.

These parameters can be configured via your Z-Wave gateway's
configuration interface. Actual parameters vary by device model.

```
shelly zwave config [flags]
```

### Examples

```
  # Show common parameters
  shelly zwave config

  # JSON output
  shelly zwave config -o json
```

### Options

```
  -h, --help            help for config
  -o, --output string   Output format: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --offline                 Only read from cache, error on cache miss
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly zwave](shelly_zwave.md)	 - Z-Wave device utilities

