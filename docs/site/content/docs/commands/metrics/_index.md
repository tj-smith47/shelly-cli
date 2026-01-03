---
title: "shelly metrics"
description: "shelly metrics"
weight: 350
sidebar:
  collapsed: true
---

## shelly metrics

Export device metrics

### Synopsis

Export metrics from Shelly devices in various formats.

Supports multiple output formats for integration with monitoring systems:
  - Prometheus: Start an HTTP exporter for Prometheus scraping
  - JSON: Output metrics in JSON format for custom integrations
  - InfluxDB: Output in InfluxDB line protocol for time-series databases

All formats export: power, voltage, current, energy, temperature, and device status.

### Examples

```
  # Start Prometheus exporter
  shelly metrics prometheus --devices kitchen,bedroom

  # Export metrics as JSON
  shelly metric json kitchen

  # Export in InfluxDB line protocol
  shelly metrics influxdb kitchen
```

### Options

```
  -h, --help   help for metrics
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
* [shelly metrics influxdb](shelly_metrics_influxdb.md)	 - Output metrics in InfluxDB line protocol
* [shelly metrics json](shelly_metrics_json.md)	 - Output metrics as JSON
* [shelly metrics prometheus](shelly_metrics_prometheus.md)	 - Start Prometheus metrics exporter

