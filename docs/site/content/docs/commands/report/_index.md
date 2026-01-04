---
title: "shelly report"
description: "shelly report"
weight: 500
sidebar:
  collapsed: true
---

## shelly report

Generate reports

### Synopsis

Generate reports about devices, energy usage, or security audits.

Report types:
  devices  - Device inventory and status
  energy   - Energy consumption summary
  audit    - Security audit report

Output formats:
  json   - JSON format (default)
  text   - Human-readable text

```
shelly report [flags]
```

### Examples

```
  # Generate device report
  shelly report --type devices

  # Save report to file
  shelly report --type devices -o report.json

  # Generate energy report
  shelly report --type energy

  # Text format report
  shelly report --type devices --format text
```

### Options

```
  -f, --format string        Output format: json, text (default "json")
  -h, --help                 help for report
      --output-file string   Output file path
  -t, --type string          Report type: devices, energy, audit (default "devices")
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

