---
title: "shelly cloud events"
description: "shelly cloud events"
---

## shelly cloud events

Subscribe to real-time cloud events

### Synopsis

Subscribe to real-time events from the Shelly Cloud via WebSocket.

Displays events as they arrive from your cloud-connected devices.
Press Ctrl+C to stop listening.

Event types:
  Shelly:StatusOnChange  - Device status changed
  Shelly:Settings        - Device settings changed
  Shelly:Online          - Device came online/offline

```
shelly cloud events [flags]
```

### Examples

```
  # Watch all events
  shelly cloud events

  # Filter by device ID
  shelly cloud events --device abc123

  # Filter by event type
  shelly cloud events --event Shelly:Online

  # Output raw JSON
  shelly cloud events --raw

  # Output in JSON format
  shelly cloud events --format json
```

### Options

```
      --device string   Filter by device ID
      --event string    Filter by event type
  -f, --format string   Output format: text, json (default "text")
  -h, --help            help for events
      --raw             Output raw JSON messages
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

* [shelly cloud](shelly_cloud.md)	 - Manage cloud connection and Shelly Cloud API

