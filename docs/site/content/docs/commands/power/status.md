---
title: "shelly power status"
description: "shelly power status"
---

## shelly power status

Show power meter status

### Synopsis

Show the live power reading of one component on a device.

Reads whichever component meters power: a PM or PM1 power meter, a
switch, cover or light that meters its load (Plus 1PM, Plus 2PM, Plus
Plug, Pro 4PM, dimmers, RGBW PM), an EM or EM1 energy monitor, or the
meters of a Gen1 device (Shelly 1PM, Plug S, Duo bulbs, EM). Shows
voltage, current, power, frequency and accumulated energy, as far as
the component reports them.

Without an ID the first component that meters power is shown; 'shelly
power list' lists them all. Use --type when two component types share
an ID.

With --all, shows every power reading on every registered device as one
list. Devices that are offline or meter nothing are skipped with a note
on stderr. With -o json or -o yaml each reading carries name, type, id,
power (watts) and the full reading under em, em1 or meter.

```
shelly power status [device] [id] [flags]
```

### Examples

```
  # Show the first power reading of a device
  shelly power status living-room

  # Show channel 1 of a two-channel switch
  shelly power status living-room 1

  # Pick the component type explicitly
  shelly power status living-room 0 --type pm1

  # Output as JSON for scripting
  shelly power status living-room -o json

  # Read just the power in watts
  shelly power status living-room -o json | jq '.power'

  # Every power reading on every registered device
  shelly power status --all
```

### Options

```
  -a, --all           Target all registered devices
  -h, --help          help for status
      --type string   Component type: auto, or one of em, em1, pm, pm1, switch, cover, light, rgb, rgbw, cct, meter (Gen1), emeter (Gen1) (default "auto")
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

* [shelly power](shelly_power.md)	 - Power meter operations (PM/PM1 components)

