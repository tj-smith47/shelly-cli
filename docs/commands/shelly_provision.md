## shelly provision

Discover and provision new Shelly devices

### Synopsis

Discover and provision new Shelly devices on your network.

When run without a subcommand, provision scans for unprovisioned Shelly devices
using BLE (Gen2+) and WiFi AP (Gen1). Found devices are presented for
interactive selection and provisioned with WiFi credentials automatically.

WiFi credentials are resolved in order: --from-device backup, --ssid/--password
flags, auto-detected from an existing Gen1 device, or prompted interactively.

Use --from-device to clone an existing device's full configuration (WiFi, MQTT,
cloud, light settings, schedules, etc.) onto newly provisioned devices. Use
--from-template to apply a saved device template instead.

Gen2+ devices are provisioned via BLE (parallel, no network disruption).
Gen1 devices are provisioned via their WiFi AP (sequential, requires temporary
network switch to the device's AP).

Use the subcommands for targeted provisioning of specific devices:
  wifi   - Interactive WiFi provisioning for a single device
  ble    - BLE-based provisioning for a specific device
  bulk   - Bulk provisioning from a config file

To register already-networked devices, use: shelly discover --register

```
shelly provision [flags]
```

### Examples

```
  # Auto-discover and provision all new devices
  shelly provision

  # Clone config from an existing device onto new devices
  shelly provision --from-device living-room --ap-only

  # Apply a saved template to new devices
  shelly provision --from-template bulb-config --ap-only -y

  # Provide WiFi credentials via flags (non-interactive)
  shelly provision --ssid MyNetwork --password secret --yes

  # Only discover via BLE (Gen2+ devices)
  shelly provision --ble-only

  # Only discover via WiFi AP (Gen1 devices)
  shelly provision --ap-only

  # Interactive WiFi provisioning for a single device
  shelly provision wifi living-room

  # BLE-based provisioning for new device
  shelly provision ble 192.168.33.1
```

### Options

```
      --ap-only                Only discover via WiFi AP (Gen1 devices)
      --ble-only               Only discover via BLE (Gen2+ devices)
      --from-device string     Clone config from existing device
      --from-template string   Apply saved template after provisioning
  -h, --help                   help for provision
      --name string            Device name to assign after provisioning
      --no-cloud               Disable cloud on provisioned devices
      --password string        WiFi password for provisioning
      --ssid string            WiFi SSID for provisioning
      --timeout duration       Discovery timeout (default 30s)
      --timezone string        Timezone to set on device
  -y, --yes                    Skip confirmation prompts
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly provision ble](shelly_provision_ble.md)	 - Provision a device via Bluetooth Low Energy
* [shelly provision bulk](shelly_provision_bulk.md)	 - Bulk provision from config file
* [shelly provision wifi](shelly_provision_wifi.md)	 - Interactive WiFi provisioning

