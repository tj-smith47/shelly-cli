## shelly action list

List action URLs for a Gen1 device

### Synopsis

List all configured action URLs for a Gen1 Shelly device.

Gen1 devices use HTTP-based settings for action URLs. This command shows
all configured actions and their target URLs.

Note: This feature is currently in development. Gen1 device support requires
direct HTTP communication rather than the RPC protocol used by Gen2 devices.

Workaround: Access the device's web interface at http://<device-ip>/settings
to view and configure action URLs.

```
shelly action list <device> [flags]
```

### Examples

```
  # List actions for a device
  shelly action list living-room

  # Workaround: use curl to get settings
  curl http://192.168.1.100/settings | jq '.actions'
```

### Options

```
  -h, --help   help for list
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

* [shelly action](shelly_action.md)	 - Manage Gen1 device action URLs

