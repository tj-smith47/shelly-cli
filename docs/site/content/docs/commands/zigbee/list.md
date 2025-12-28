---
title: "shelly zigbee list"
description: "shelly zigbee list"
---

## shelly zigbee list

List Zigbee-capable devices

### Synopsis

List all Zigbee-capable Shelly devices on the network.

Scans configured devices to find those with Zigbee support
and shows their current Zigbee status.

Note: This only shows devices in your Shelly CLI config, not
devices paired to Zigbee coordinators.

```
shelly zigbee list [flags]
```

### Examples

```
  # List Zigbee-capable devices
  shelly zigbee list

  # Output as JSON
  shelly zigbee list --json
```

### Options

```
  -f, --format string   Output format: text, json (default "text")
  -h, --help            help for list
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

* [shelly zigbee](shelly_zigbee.md)	 - Manage Zigbee connectivity

