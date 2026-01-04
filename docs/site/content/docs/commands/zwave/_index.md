---
title: "shelly zwave"
description: "shelly zwave"
weight: 750
sidebar:
  collapsed: true
---

## shelly zwave

Z-Wave device utilities

### Synopsis

Utilities for working with Shelly Wave Z-Wave devices.

Shelly Wave devices are Z-Wave end devices that require a third-party
gateway/hub for full operation. They support both standard Z-Wave mesh
networks and Z-Wave Long Range (ZWLR) star topology.

Supported gateways: Home Assistant (Z-Wave JS), Hubitat, HomeSeer,
SmartThings, Vera/ezlo, OpenHAB, and other Z-Wave certified controllers.

Note: Many Wave devices also include WiFi or Ethernet connectivity,
allowing direct control via the standard Gen2 RPC API.

### Examples

```
  # Show Z-Wave device info
  shelly zwave info SNSW-001P16ZW

  # Show inclusion instructions
  shelly zwave inclusion SNSW-001P16ZW --mode button

  # Show exclusion instructions
  shelly zwave exclusion SNSW-001P16ZW

  # Show factory reset instructions
  shelly zwave reset SNSW-001P16ZW

  # Show common configuration parameters
  shelly zwave config
```

### Options

```
  -h, --help   help for zwave
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
* [shelly zwave config](shelly_zwave_config.md)	 - Show common configuration parameters
* [shelly zwave exclusion](shelly_zwave_exclusion.md)	 - Show exclusion instructions
* [shelly zwave inclusion](shelly_zwave_inclusion.md)	 - Show inclusion instructions
* [shelly zwave info](shelly_zwave_info.md)	 - Show Z-Wave device information
* [shelly zwave reset](shelly_zwave_reset.md)	 - Show factory reset instructions

