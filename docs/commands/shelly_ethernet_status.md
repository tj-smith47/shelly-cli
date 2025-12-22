## shelly ethernet status

Show Ethernet status

### Synopsis

Show the current Ethernet status for a device.

Displays connection status and IP address. Only available on Pro devices
with an Ethernet port.

```
shelly ethernet status <device> [flags]
```

### Examples

```
  # Show Ethernet status
  shelly ethernet status living-room-pro

  # Output as JSON
  shelly ethernet status living-room-pro -o json
```

### Options

```
  -h, --help   help for status
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

* [shelly ethernet](shelly_ethernet.md)	 - Manage device Ethernet configuration

