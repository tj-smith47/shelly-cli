---
title: "shelly debug coiot"
description: "shelly debug coiot"
---

## shelly debug coiot

Show CoIoT/CoAP status or listen for multicast updates

### Synopsis

Show CoIoT (CoAP over Internet of Things) status for a device, or listen
for multicast updates from all Gen1 devices on the network.

CoIoT is used by Gen1 devices for local discovery and real-time status
updates via multicast UDP on 224.0.1.187:5683.

Without --listen, this command shows the CoIoT configuration for a specific device:
- CoIoT enabled/disabled status
- Multicast settings
- Peer configuration (for unicast mode)
- Update period settings

With --listen, it starts a multicast listener that receives and displays
CoIoT status broadcasts from all Gen1 devices on the network.

```
shelly debug coiot [device] [flags]
```

### Examples

```
  # Show CoIoT status for a device
  shelly debug coiot living-room

  # Output as JSON
  shelly debug coiot living-room -f json

  # Listen for CoIoT multicast updates for 30 seconds (default)
  shelly debug coiot --listen

  # Listen for 2 minutes
  shelly debug coiot --listen --duration 2m

  # Stream indefinitely (until Ctrl+C)
  shelly debug coiot --listen --stream

  # Stream with raw JSON output
  shelly debug coiot --listen --stream --raw
```

### Options

```
      --duration duration   Listen duration (ignored if --stream) (default 30s)
  -f, --format string       Output format: text, json (default "text")
  -h, --help                help for coiot
  -l, --listen              Listen for CoIoT multicast updates from all Gen1 devices
      --raw                 Output raw JSON events
  -s, --stream              Stream indefinitely (until Ctrl+C)
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

