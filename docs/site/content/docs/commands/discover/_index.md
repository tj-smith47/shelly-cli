---
title: "shelly discover"
description: "shelly discover"
weight: 240
sidebar:
  collapsed: true
---

## shelly discover

Discover Shelly devices on the network

### Synopsis

Discover Shelly devices using various protocols.

By default, uses HTTP subnet scanning which works reliably even when
multicast is blocked. Automatically detects the local subnet.

Available discovery methods (--method):
  http   - HTTP subnet scanning (default, works everywhere)
  mdns   - mDNS/Zeroconf discovery (Gen2+ devices)
  ble    - Bluetooth Low Energy discovery (provisioning mode)
  coiot  - CoIoT/CoAP discovery (Gen1 devices)

Plugin-managed devices (e.g., Tasmota, ESPHome) can also be discovered
if the corresponding plugin is installed. Use --skip-plugins to disable
plugin detection, or --platform to filter by specific platform.

```
shelly discover [flags]
```

### Examples

```
  # Discover devices via HTTP scan (default, auto-detects subnet)
  shelly discover

  # Specify subnet for HTTP scan
  shelly discover --subnet 192.168.1.0/24

  # Use mDNS instead of HTTP scan
  shelly discover --method mdns

  # Use BLE discovery
  shelly discover --method ble

  # Auto-register discovered devices
  shelly discover --register

  # Skip plugin detection (Shelly-only)
  shelly discover --skip-plugins

  # Discover only Tasmota devices
  shelly discover --platform tasmota
```

### Options

```
  -h, --help               help for discover
  -m, --method string      Discovery method: http, mdns, ble, coiot (default "http")
  -p, --platform string    Only discover devices of this platform (e.g., tasmota)
      --register           Auto-register discovered devices
      --skip-existing      Skip devices already registered (default true)
      --skip-plugins       Skip plugin detection (Shelly-only discovery)
      --subnet string      Subnet to scan (auto-detected if not specified)
  -t, --timeout duration   Discovery timeout (default 2m0s)
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
* [shelly discover ble](shelly_discover_ble.md)	 - Discover devices using Bluetooth Low Energy
* [shelly discover coiot](shelly_discover_coiot.md)	 - Discover devices via CoIoT
* [shelly discover http](shelly_discover_http.md)	 - Discover devices via HTTP subnet scanning
* [shelly discover mdns](shelly_discover_mdns.md)	 - Discover devices using mDNS/Zeroconf

