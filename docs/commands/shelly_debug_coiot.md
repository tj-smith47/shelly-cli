## shelly debug coiot

Show CoIoT/CoAP status

### Synopsis

Show CoIoT (CoAP over Internet of Things) status for a device.

CoIoT is used by Gen1 and some Gen2 devices for local discovery and
real-time status updates via multicast UDP.

This command shows:
- CoIoT enabled/disabled status
- Multicast settings
- Peer configuration (for unicast mode)
- Update period settings

```
shelly debug coiot <device> [flags]
```

### Examples

```
  # Show CoIoT status
  shelly debug coiot living-room

  # Output as JSON
  shelly debug coiot living-room --json
```

### Options

```
  -h, --help   help for coiot
      --json   Output as JSON
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

* [shelly debug](shelly_debug.md)	 - Debug and diagnostic commands

