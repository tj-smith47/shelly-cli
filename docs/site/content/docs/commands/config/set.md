---
title: "shelly config set"
description: "shelly config set"
---

## shelly config set

Set CLI configuration values

### Synopsis

Set configuration values in the Shelly CLI config file.

Use dot notation for nested values (e.g. "discovery.timeout=30s").

A key and its value may be separated with "=", ":", or a space — these are
equivalent:

  shelly config set telemetry=true
  shelly config set telemetry:true
  shelly config set telemetry true

Values are stored using each setting's real type (e.g. telemetry as a boolean,
ratelimit.global.max_concurrent as an integer), so booleans and numbers behave
correctly.

```
shelly config set <key>=<value>... [flags]
```

### Examples

```
  # Enable anonymous usage telemetry (these are equivalent)
  shelly config set telemetry=true
  shelly config set telemetry true

  # Set discovery timeout (duration)
  shelly config set discovery.timeout=30s

  # Set output format
  shelly config set output=json

  # Set multiple values
  shelly config set discovery.timeout=30s output=json
```

### Options

```
  -h, --help   help for set
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
      --raw                     Print the exact device response(s) as a JSON array and suppress normal output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly config](shelly_config.md)	 - Manage CLI configuration

