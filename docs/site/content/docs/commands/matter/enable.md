---
title: "shelly matter enable"
description: "shelly matter enable"
---

## shelly matter enable

Enable Matter on a device

### Synopsis

Enable Matter connectivity on a Shelly device.

When Matter is enabled, the device can be commissioned (added) to
Matter-compatible smart home ecosystems like Apple Home, Google Home,
or Amazon Alexa.

After enabling, use 'shelly matter code' to get the pairing code
for commissioning the device.

```
shelly matter enable <device> [flags]
```

### Examples

```
  # Enable Matter
  shelly matter enable living-room

  # Then get pairing code
  shelly matter code living-room
```

### Options

```
  -h, --help   help for enable
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

* [shelly matter](shelly_matter.md)	 - Manage Matter connectivity

