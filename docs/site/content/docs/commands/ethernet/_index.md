---
title: "shelly ethernet"
description: "shelly ethernet"
weight: 260
sidebar:
  collapsed: true
---

## shelly ethernet

Manage device Ethernet configuration

### Synopsis

Manage device Ethernet configuration settings.

Ethernet is available on Shelly Pro devices that have an Ethernet port.
It provides wired network connectivity as an alternative to WiFi.

### Examples

```
  # Show Ethernet status
  shelly ethernet status living-room-pro

  # Configure Ethernet with DHCP
  shelly ethernet set living-room-pro --enable

  # Configure Ethernet with static IP
  shelly ethernet set living-room-pro --enable --static-ip "192.168.1.50" \
    --gateway "192.168.1.1" --netmask "255.255.255.0"
```

### Options

```
  -h, --help   help for ethernet
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly ethernet set](shelly_ethernet_set.md)	 - Configure Ethernet connection
* [shelly ethernet status](shelly_ethernet_status.md)	 - Show Ethernet status

