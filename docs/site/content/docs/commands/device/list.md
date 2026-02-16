---
title: "shelly device list"
description: "shelly device list"
---

## shelly device list

List registered devices

### Synopsis

List all devices registered in the local registry.

The registry stores device information including name, address, model,
generation, platform, and authentication credentials. Use filters to
narrow results by device generation, device type, or platform.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting and piping to tools like jq.

Columns: Name, Address, Platform, Type, Model, Generation, Auth

```
shelly device list [flags]
```

### Examples

```
  # List all registered devices
  shelly device list

  # List only Gen2 devices
  shelly device list --generation 2

  # List devices by type
  shelly device list --type SHSW-1

  # List only Shelly devices (exclude plugin-managed)
  shelly device list --platform shelly

  # List only Tasmota devices (from shelly-tasmota plugin)
  shelly device list --platform tasmota

  # Show firmware versions and sort updates first
  shelly device list --version --updates-first

  # Output as JSON for scripting
  shelly device list -o json

  # Pipe to jq to extract device names
  shelly device list -o json | jq -r '.[].name'

  # Parse table output in scripts (disable colors)
  shelly device list --no-color | tail -n +2 | while read name addr _; do
    echo "Device: $name at $addr"
  done

  # Export to CSV via jq
  shelly device list -o json | jq -r '.[] | [.name,.address,.model] | @csv'

  # Short form
  shelly dev ls
```

### Options

```
  -g, --generation int    Filter by generation (1, 2, or 3)
  -h, --help              help for list
  -p, --platform string   Filter by platform (e.g., shelly, tasmota)
      --refresh           Force refresh device metadata from hardware
  -t, --type string       Filter by device type
  -u, --updates-first     Sort devices with available updates first
  -V, --version           Show firmware version information
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
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly device](shelly_device.md)	 - Manage Shelly devices

