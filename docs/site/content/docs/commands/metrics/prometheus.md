---
title: "shelly metrics prometheus"
description: "shelly metrics prometheus"
---

## shelly metrics prometheus

Start Prometheus metrics exporter

### Synopsis

Start an HTTP server that exports metrics in Prometheus format.

The exporter collects metrics from all registered devices (or a specified
subset) and exposes them at /metrics for Prometheus scraping.

Metrics exported:
  Power metering (PM/PM1/EM/EM1 components):
  - shelly_power_watts: Current power consumption
  - shelly_voltage_volts: Voltage reading
  - shelly_current_amps: Current reading
  - shelly_energy_wh_total: Total energy consumption

  System metrics:
  - shelly_device_online: Device reachability (1=online, 0=offline)
  - shelly_wifi_rssi: WiFi signal strength in dBm
  - shelly_uptime_seconds: Device uptime
  - shelly_temperature_celsius: Device temperature
  - shelly_ram_free_bytes: Free RAM
  - shelly_ram_total_bytes: Total RAM

  Component state:
  - shelly_switch_on: Switch state (1=on, 0=off)

Labels include: device, component, component_id, phase

```
shelly metrics prometheus [flags]
```

### Examples

```
  # Start exporter on default port 9090
  shelly metrics prometheus

  # Start on custom port with specific devices
  shelly metrics prometheus --port 8080 --devices kitchen,living-room

  # Collect metrics every 30 seconds
  shelly metrics prometheus --interval 30s
```

### Options

```
      --devices strings     Devices to include (default: all registered)
  -h, --help                help for prometheus
      --interval duration   Metrics collection interval (default 15s)
      --port int            HTTP port for the exporter (default 9090)
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

* [shelly metrics](shelly_metrics.md)	 - Export device metrics

