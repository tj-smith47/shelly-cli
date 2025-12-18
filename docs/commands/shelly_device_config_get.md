## shelly device config get

Get device configuration

### Synopsis

Get configuration for a device or specific component.

Without a component argument, returns the full device configuration.
With a component argument (e.g., "switch:0", "sys", "wifi"), returns
only that component's configuration.

```
shelly device config get <device> [component] [flags]
```

### Examples

```
  # Get full device configuration
  shelly config get living-room

  # Get switch:0 configuration
  shelly config get living-room switch:0

  # Get system configuration
  shelly config get living-room sys

  # Get WiFi configuration
  shelly config get living-room wifi

  # Output as JSON
  shelly config get living-room -o json

  # Output as YAML
  shelly config get living-room -o yaml
```

### Options

```
  -h, --help   help for get
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

* [shelly device config](shelly_device_config.md)	 - Manage device configuration

