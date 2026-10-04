---
title: "shelly energy status"
description: "shelly energy status"
---

## shelly energy status

Show energy monitor status

### Synopsis

Show the live power and energy reading of one component on a device.

Reads whichever component meters power: an EM (3-phase) or EM1
(single-phase) energy monitor, a PM or PM1 power meter, a switch, cover
or light that meters its load (Plus 1PM, Plus 2PM, Plus Plug, Pro 4PM,
dimmers, RGBW PM), or the meters of a Gen1 device (Shelly 1PM, Plug S,
Duo bulbs, EM). An EM reading shows per-phase data and totals; the
others show voltage, current, power, frequency and accumulated energy,
as far as the component reports them.

Without an ID the first component that meters power is shown, energy
monitors first; 'shelly energy list' lists them all. Use --type when two
component types share an ID.

With --all, shows every power reading on every registered device as one
list. Devices that are offline or meter nothing are skipped with a note
on stderr. With -o json or -o yaml each reading carries name, type, id,
power (watts) and the full reading under em, em1 or meter.

```
shelly energy status [device] [id] [flags]
```

### Examples

```
  # Show energy monitor status
  shelly energy status shelly-3em-pro

  # Show specific component by ID
  shelly energy status shelly-em 0

  # Specify component type explicitly
  shelly energy status shelly-em --type em1

  # A switch that meters its load (Plus 1PM, Plus 2PM channel 1)
  shelly energy status kitchen 1

  # Output as JSON for scripting
  shelly energy status shelly-3em-pro -o json

  # Show every power reading on every registered device
  shelly energy status --all

  # Every power reading as one JSON list
  shelly energy status --all -o json
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

* [shelly energy](shelly_energy.md)	 - Energy monitoring operations (EM/EM1 components)

