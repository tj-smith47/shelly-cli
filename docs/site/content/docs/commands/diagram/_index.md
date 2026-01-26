---
title: "shelly diagram"
description: "shelly diagram"
weight: 220
sidebar:
  collapsed: true
---

## shelly diagram

Display ASCII wiring diagrams for Shelly devices

### Synopsis

Display ASCII wiring diagrams showing terminal connections and wiring
layouts for Shelly device models. Useful for installation reference.

Use --style to choose between schematic (circuit-style), compact (minimal box),
or detailed (installer-friendly with annotations) diagram styles.

```
shelly diagram [flags]
```

### Examples

```
  # Show wiring diagram for Shelly Plus 1
  shelly diagram -m plus-1

  # Compact layout for Pro 4PM
  shelly diagram -m pro-4pm -s compact

  # Detailed installer view for Dimmer 2
  shelly diagram -m dimmer-2 -s detailed

  # Disambiguate model with generation
  shelly diagram -m 1 -g 1

  # Show Gen3 variant
  shelly diagram -m 1pm-gen3
```

### Options

```
  -g, --generation string   Device generation (1, 2, 3, 4, gen1, gen2, gen3, gen4)
  -h, --help                help for diagram
  -m, --model string        Device model slug (e.g., plus-1, pro-4pm, dimmer-2)
  -s, --style string        Diagram style (schematic, compact, detailed) (default "schematic")
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

