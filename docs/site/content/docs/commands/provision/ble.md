---
title: "shelly provision ble"
description: "shelly provision ble"
---

## shelly provision ble

Provision a device via Bluetooth Low Energy

### Synopsis

Provision a Shelly device using Bluetooth Low Energy (BLE).

This command allows you to configure WiFi credentials and other settings
on a new device without connecting to its AP network first.

BLE provisioning requires:
- Bluetooth hardware support on your computer
- The device to be in BLE advertising mode (typically when unconfigured)
- The device's BLE address (usually shown as ShellyXXX-YYYYYYYY)

Gen2+ devices support BLE provisioning. Gen1 devices do not have BLE capability.

```
shelly provision ble <device-address> [flags]
```

### Examples

```
  # Provision WiFi via BLE
  shelly provision ble ShellyPlus1-ABCD1234 --ssid "MyNetwork" --password "secret"

  # Set device name during provisioning
  shelly provision ble ShellyPlus1-ABCD1234 --ssid "MyNetwork" --password "secret" --name "Living Room Switch"

  # Configure with timezone
  shelly provision ble ShellyPlus1-ABCD1234 --ssid "MyNetwork" --password "secret" --timezone "America/New_York"

  # Disable cloud during provisioning
  shelly provision ble ShellyPlus1-ABCD1234 --ssid "MyNetwork" --password "secret" --no-cloud
```

### Options

```
      --cloud             Enable Shelly Cloud
  -h, --help              help for ble
      --name string       Device name to set
      --no-cloud          Disable Shelly Cloud
      --password string   WiFi password
      --ssid string       WiFi network name (required)
      --timezone string   Timezone (e.g., America/New_York)
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
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly provision](shelly_provision.md)	 - Provision device settings

