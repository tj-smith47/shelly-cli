---
title: "shelly provision"
description: "shelly provision"
weight: 510
sidebar:
  collapsed: true
---

## shelly provision

Discover and provision new Shelly devices

### Synopsis

Discover and provision new Shelly devices on your network.

When run without a subcommand, provision scans for unprovisioned Shelly devices
using BLE (Gen2+), WiFi AP (Gen1), and network discovery (mDNS/CoIoT). Found
devices are presented for interactive selection and provisioned with WiFi
credentials automatically.

Gen2+ devices are provisioned via BLE (parallel, no network disruption).
Gen1 devices are provisioned via their WiFi AP (sequential, requires temporary
network switch to the device's AP).

Already-networked but unregistered devices are simply registered in the config.

Use the subcommands for targeted provisioning of specific devices:
  wifi   - Interactive WiFi provisioning for a single device
  ble    - BLE-based provisioning for a specific device
  bulk   - Bulk provisioning from a config file

```
shelly provision [flags]
```

### Examples

```
  # Auto-discover and provision all new devices
  shelly provision

  # Provide WiFi credentials via flags (non-interactive)
  shelly provision --ssid MyNetwork --password secret --yes

  # Only discover via BLE (Gen2+ devices)
  shelly provision --ble-only

  # Only discover via WiFi AP (Gen1 devices)
  shelly provision --ap-only

  # Only register already-networked devices (no provisioning)
  shelly provision --register-only

  # Scan a specific subnet for devices
  shelly provision --subnet 192.168.1.0/24

  # Interactive WiFi provisioning for a single device
  shelly provision wifi living-room

  # Bulk provision from config file
  shelly provision bulk devices.yaml

  # BLE-based provisioning for new device
  shelly provision ble 192.168.33.1
```

### Options

```
      --ap-only            Only discover via WiFi AP (Gen1 devices)
      --ble-only           Only discover via BLE (Gen2+ devices)
  -h, --help               help for provision
      --name string        Device name to assign after provisioning
      --network-only       Only discover already-networked devices
      --no-cloud           Disable cloud on provisioned devices
      --password string    WiFi password for provisioning
      --register-only      Only register devices (skip provisioning)
      --ssid string        WiFi SSID for provisioning
      --subnet string      Subnet to scan (e.g., 192.168.1.0/24)
      --timeout duration   Discovery timeout (default 30s)
      --timezone string    Timezone to set on device
  -y, --yes                Skip confirmation prompts
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
* [shelly provision ble](shelly_provision_ble.md)	 - Provision a device via Bluetooth Low Energy
* [shelly provision bulk](shelly_provision_bulk.md)	 - Bulk provision from config file
* [shelly provision wifi](shelly_provision_wifi.md)	 - Interactive WiFi provisioning

