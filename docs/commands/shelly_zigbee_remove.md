## shelly zigbee remove

Leave Zigbee network

### Synopsis

Leave the current Zigbee network and disable Zigbee.

This causes the device to leave its Zigbee network and disables
Zigbee functionality. The device will no longer be controllable
through Zigbee coordinators.

Note: The device will still be accessible via WiFi/HTTP.

```
shelly zigbee remove <device> [flags]
```

### Examples

```
  # Leave Zigbee network
  shelly zigbee remove living-room

  # Leave without confirmation
  shelly zigbee remove living-room --yes
```

### Options

```
  -h, --help   help for remove
  -y, --yes    Skip confirmation
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

* [shelly zigbee](shelly_zigbee.md)	 - Manage Zigbee connectivity

