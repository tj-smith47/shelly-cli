---
title: "shelly wifi"
description: "shelly wifi"
weight: 730
sidebar:
  collapsed: true
---

## shelly wifi

Manage device WiFi configuration

### Synopsis

Manage device WiFi configuration settings.

Get WiFi status, scan for networks, and configure WiFi settings including
station mode (connecting to a network) and access point mode.

### Examples

```
  # Show WiFi status
  shelly wifi status living-room

  # Scan for available networks
  shelly wifi scan living-room

  # Configure WiFi connection
  shelly wifi set living-room --ssid "MyNetwork" --password "secret"

  # Configure access point
  shelly wifi ap living-room --enable --ssid "ShellyAP"
```

### Options

```
  -h, --help   help for wifi
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly wifi ap](shelly_wifi_ap.md)	 - Configure WiFi access point
* [shelly wifi scan](shelly_wifi_scan.md)	 - Scan for available WiFi networks
* [shelly wifi set](shelly_wifi_set.md)	 - Configure WiFi connection
* [shelly wifi status](shelly_wifi_status.md)	 - Show WiFi status

