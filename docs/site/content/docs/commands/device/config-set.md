---
title: "shelly device config set"
description: "shelly device config set"
---

## shelly device config set

Set device configuration

### Synopsis

Set configuration values for a device component.

Specify key=value pairs to update. Only the specified keys will be modified.

A key and its value may be separated with "=", ":", or a space — these are
equivalent ("name=Light", "name:Light", and "name Light" all set name to Light).

WiFi station keys (sta, sta1 and their fields, such as sta.ssid) are refused:
a station write needs the network's password planned, so set stations with
'shelly wifi set'. The other wifi keys (ap, roam) are accepted.

```
shelly device config set <device> <component> <key>=<value>... [flags]
```

### Examples

```
  # Set switch name (these are equivalent)
  shelly device config set living-room switch:0 name="Main Light"
  shelly device config set living-room switch:0 name "Main Light"

  # Set multiple values
  shelly device config set living-room switch:0 name="Light" initial_state=on

  # Set light brightness default
  shelly device config set living-room light:0 default.brightness=50
```

### Options

```
  -h, --help   help for set
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
  -F, --fields                  Print available field names for use with --jq and --template
  -Q, --jq stringArray          Apply jq expression to filter output (repeatable, joined with |)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --offline                 Only read from cache, error on cache miss
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --raw                     Print the exact device response(s) as a JSON array and suppress normal output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly device config](shelly_device_config.md)	 - Manage device configuration

