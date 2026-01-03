---
title: "shelly export"
description: "shelly export"
weight: 230
sidebar:
  collapsed: true
---

## shelly export

Export fleet data for infrastructure tools

### Synopsis

Export device fleet data for infrastructure-as-code tools.

Supports exporting to CSV, Ansible inventory, and Terraform configuration
formats. Useful for documentation and fleet management workflows.

For single-device configuration export (JSON/YAML), use:
  shelly device config export <device> <file> [--format json|yaml]

### Examples

```
  # Export device list as CSV
  shelly export csv living-room bedroom kitchen devices.csv

  # Export as Ansible inventory
  shelly export ansible @all inventory.yaml

  # Export as Terraform config
  shelly export terraform @all shelly.tf
```

### Options

```
  -h, --help   help for export
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly export ansible](shelly_export_ansible.md)	 - Export devices as Ansible inventory
* [shelly export csv](shelly_export_csv.md)	 - Export device list as CSV
* [shelly export terraform](shelly_export_terraform.md)	 - Export devices as Terraform configuration

