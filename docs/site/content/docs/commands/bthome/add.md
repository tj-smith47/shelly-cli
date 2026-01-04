---
title: "shelly bthome add"
description: "shelly bthome add"
---

## shelly bthome add

Add a BTHome device

### Synopsis

Add a BTHome device to a Shelly gateway.

This command can either:
1. Start a discovery scan to find nearby BTHome devices
2. Add a specific device by MAC address

When scanning, the gateway will broadcast discovery requests and listen
for BTHome device advertisements. Discovered devices emit events that
can be monitored with 'shelly monitor events'.

```
shelly bthome add <device> [flags]
```

### Examples

```
  # Start 30-second discovery scan
  shelly bthome add living-room

  # Custom scan duration (60 seconds)
  shelly bthome add living-room --duration 60

  # Add specific device by MAC address
  shelly bthome add living-room --addr 3c:2e:f5:71:d5:2a

  # Add device with a name
  shelly bthome add living-room --addr 3c:2e:f5:71:d5:2a --name "Door Sensor"
```

### Options

```
      --addr string    MAC address of device to add directly
      --duration int   Discovery scan duration in seconds (default 30)
  -h, --help           help for add
  -n, --name string    Name for the device (with --addr)
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

* [shelly bthome](shelly_bthome.md)	 - Manage BTHome Bluetooth devices

