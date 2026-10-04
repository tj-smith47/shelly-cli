---
title: "shelly energy status"
description: "shelly energy status"
---

## shelly energy status

Show energy monitor status

### Synopsis

Show current status of an energy monitoring component.

Displays real-time measurements including voltage, current, power,
power factor, and frequency. For 3-phase EM components, shows
per-phase data and totals.

With --all, shows every EM and EM1 component on every registered
device as one list. Devices that are offline or have no EM or EM1
component are skipped with a note on stderr. With -o json or -o yaml
the list is printed as one array whose items carry name, type, id,
power (watts) and the full em or em1 reading.

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

  # Output as JSON for scripting
  shelly energy status shelly-3em-pro -o json

  # Show every energy monitor on every registered device
  shelly energy status --all

  # Every energy monitor as one JSON list
  shelly energy status --all -o json
```

### Options

```
  -a, --all           Target all registered devices
  -h, --help          help for status
      --type string   Component type (auto, em, em1) (default "auto")
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

