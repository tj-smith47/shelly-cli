---
title: "shelly benchmark"
description: "shelly benchmark"
weight: 90
sidebar:
  collapsed: true
---

## shelly benchmark

Test device performance

### Synopsis

Measure device performance including API latency and response times.

The benchmark runs multiple iterations to collect statistics on:
  - Ping latency (basic connectivity)
  - RPC latency (API call response time)

Results include min, max, average, and percentile statistics (P50, P95, P99).

```
shelly benchmark <device> [flags]
```

### Examples

```
  # Basic benchmark (10 iterations)
  shelly benchmark kitchen-light

  # Extended benchmark
  shelly benchmark kitchen-light --iterations 50

  # JSON output for logging
  shelly benchmark kitchen-light --json
```

### Options

```
  -h, --help             help for benchmark
  -n, --iterations int   Number of iterations (default 10)
      --warmup int       Number of warmup iterations (not counted) (default 2)
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

