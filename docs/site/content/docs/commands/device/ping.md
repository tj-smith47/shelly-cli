---
title: "shelly device ping"
description: "shelly device ping"
---

## shelly device ping

Check device connectivity

### Synopsis

Check if a device is reachable and responding.

The ping command attempts to connect to the device and retrieve its info.
This is useful for verifying network connectivity and device availability.

Use -c to send multiple pings and show statistics.

```
shelly device ping <device> [flags]
```

### Examples

```
  # Ping a registered device
  shelly device ping living-room

  # Ping by IP address
  shelly device ping 192.168.1.100

  # Ping multiple times with statistics
  shelly device ping kitchen -c 5

  # Ping with custom timeout
  shelly device ping slow-device --timeout 10s
```

### Options

```
  -c, --count int          Number of pings to send (default 1)
  -h, --help               help for ping
      --timeout duration   Timeout for each ping (default 5s)
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

* [shelly device](shelly_device.md)	 - Manage Shelly devices

