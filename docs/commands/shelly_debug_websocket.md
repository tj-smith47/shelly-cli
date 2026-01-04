## shelly debug websocket

Debug WebSocket connection and stream events

### Synopsis

Debug WebSocket connection and stream real-time events from a Shelly device.

This command connects to a Gen2+ device via WebSocket and streams all
notifications (state changes, sensor updates, button presses, etc.) in real-time.

Gen2+ devices support WebSocket at ws://<device>/rpc for bidirectional
communication and event notifications.

```
shelly debug websocket <device> [flags]
```

### Examples

```
  # Stream events for 30 seconds (default)
  shelly debug websocket living-room

  # Stream events for 5 minutes
  shelly debug websocket living-room --duration 5m

  # Stream events indefinitely (until Ctrl+C)
  shelly debug websocket living-room --duration 0

  # Raw JSON output
  shelly debug websocket living-room --raw
```

### Options

```
      --duration duration   Monitoring duration (0 for indefinite) (default 30s)
  -h, --help                help for websocket
      --raw                 Output raw JSON events
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --offline                 Only read from cache, error on cache miss
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly debug](shelly_debug.md)	 - Debug and diagnostic commands

