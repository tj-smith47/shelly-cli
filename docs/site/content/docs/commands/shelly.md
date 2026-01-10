---
title: "shelly shelly"
description: "shelly"
weight: 1
---

## shelly

CLI for controlling Shelly smart home devices

### Synopsis

Shelly CLI - Control your Shelly smart home devices from the command line.

This tool provides a comprehensive interface for discovering, monitoring,
and controlling Shelly devices on your local network.

### Examples

```
  # Initialize configuration
  shelly init

  # Discover and control devices
  shelly discover scan
  shelly switch on kitchen

  # Pipe output to jq for processing
  shelly device list -o json | jq '.[].name'

  # Pipe device names to batch commands
  echo -e "kitchen\nbedroom" | shelly batch on

  # Launch interactive dashboard
  shelly dash
```

### Options

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
  -h, --help                    help for shelly
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

* [shelly action](shelly_action.md)	 - Manage Gen1 device action URLs
* [shelly alert](shelly_alert.md)	 - Manage monitoring alerts
* [shelly alias](shelly_alias.md)	 - Manage command aliases
* [shelly api](shelly_api.md)	 - Execute API calls on Shelly devices
* [shelly audit](shelly_audit.md)	 - Security audit for devices
* [shelly auth](shelly_auth.md)	 - Manage device authentication
* [shelly backup](shelly_backup.md)	 - Backup and restore device configurations
* [shelly batch](shelly_batch.md)	 - Execute commands on multiple devices
* [shelly benchmark](shelly_benchmark.md)	 - Test device performance
* [shelly bthome](shelly_bthome.md)	 - Manage BTHome Bluetooth devices
* [shelly cache](shelly_cache.md)	 - Manage CLI cache
* [shelly cert](shelly_cert.md)	 - Manage device TLS certificates
* [shelly cloud](shelly_cloud.md)	 - Manage cloud connection and Shelly Cloud API
* [shelly completion](shelly_completion.md)	 - Generate shell completion scripts
* [shelly config](shelly_config.md)	 - Manage CLI configuration
* [shelly cover](shelly_cover.md)	 - Control cover/roller components
* [shelly dash](shelly_dash.md)	 - Launch interactive TUI dashboard
* [shelly debug](shelly_debug.md)	 - Debug and diagnostic commands
* [shelly device](shelly_device.md)	 - Manage Shelly devices
* [shelly discover](shelly_discover.md)	 - Discover Shelly devices on the network
* [shelly doctor](shelly_doctor.md)	 - Check system health and diagnose issues
* [shelly energy](shelly_energy.md)	 - Energy monitoring operations (EM/EM1 components)
* [shelly ethernet](shelly_ethernet.md)	 - Manage device Ethernet configuration
* [shelly export](shelly_export.md)	 - Export fleet data for infrastructure tools
* [shelly feedback](shelly_feedback.md)	 - Report issues or request features
* [shelly firmware](shelly_firmware.md)	 - Manage device firmware
* [shelly fleet](shelly_fleet.md)	 - Cloud-based fleet management
* [shelly group](shelly_group.md)	 - Manage device groups
* [shelly init](shelly_init.md)	 - Initialize shelly CLI for first-time use
* [shelly input](shelly_input.md)	 - Manage input components
* [shelly kvs](shelly_kvs.md)	 - Manage device key-value storage
* [shelly light](shelly_light.md)	 - Control light components
* [shelly log](shelly_log.md)	 - Manage CLI logs
* [shelly lora](shelly_lora.md)	 - Manage LoRa add-on
* [shelly matter](shelly_matter.md)	 - Manage Matter connectivity
* [shelly metrics](shelly_metrics.md)	 - Export device metrics
* [shelly migrate](shelly_migrate.md)	 - Migrate configuration between devices
* [shelly mock](shelly_mock.md)	 - Mock device mode for testing
* [shelly modbus](shelly_modbus.md)	 - Manage Modbus-TCP configuration
* [shelly monitor](shelly_monitor.md)	 - Real-time device monitoring
* [shelly mqtt](shelly_mqtt.md)	 - Manage device MQTT configuration
* [shelly off](shelly_off.md)	 - Turn off a device (auto-detects type)
* [shelly on](shelly_on.md)	 - Turn on a device (auto-detects type)
* [shelly party](shelly_party.md)	 - Party mode - flash lights!
* [shelly plugin](shelly_plugin.md)	 - Manage CLI plugins
* [shelly power](shelly_power.md)	 - Power meter operations (PM/PM1 components)
* [shelly profile](shelly_profile.md)	 - Device profile information
* [shelly provision](shelly_provision.md)	 - Provision device settings
* [shelly qr](shelly_qr.md)	 - Generate device QR code
* [shelly repl](shelly_repl.md)	 - Launch interactive REPL
* [shelly report](shelly_report.md)	 - Generate reports
* [shelly rgb](shelly_rgb.md)	 - Control RGB light components
* [shelly rgbw](shelly_rgbw.md)	 - Control RGBW LED outputs
* [shelly scene](shelly_scene.md)	 - Manage device scenes
* [shelly schedule](shelly_schedule.md)	 - Manage device schedules
* [shelly script](shelly_script.md)	 - Manage device scripts
* [shelly sensor](shelly_sensor.md)	 - Manage device sensors
* [shelly sensoraddon](shelly_sensoraddon.md)	 - Manage Sensor Add-on peripherals
* [shelly shell](shelly_shell.md)	 - Interactive shell for a specific device
* [shelly sleep](shelly_sleep.md)	 - Turn device off after a delay
* [shelly status](shelly_status.md)	 - Show device status (quick overview)
* [shelly switch](shelly_switch.md)	 - Control switch components
* [shelly sync](shelly_sync.md)	 - Synchronize device configurations
* [shelly template](shelly_template.md)	 - Manage device configuration templates
* [shelly theme](shelly_theme.md)	 - Manage CLI color themes
* [shelly thermostat](shelly_thermostat.md)	 - Manage thermostats
* [shelly toggle](shelly_toggle.md)	 - Toggle a device (auto-detects type)
* [shelly update](shelly_update.md)	 - Update shelly to the latest version
* [shelly version](shelly_version.md)	 - Print version information
* [shelly virtual](shelly_virtual.md)	 - Manage virtual components
* [shelly wait](shelly_wait.md)	 - Wait for a duration
* [shelly wake](shelly_wake.md)	 - Turn device on after a delay
* [shelly webhook](shelly_webhook.md)	 - Manage device webhooks
* [shelly wifi](shelly_wifi.md)	 - Manage device WiFi configuration
* [shelly zigbee](shelly_zigbee.md)	 - Manage Zigbee connectivity
* [shelly zwave](shelly_zwave.md)	 - Z-Wave device utilities

