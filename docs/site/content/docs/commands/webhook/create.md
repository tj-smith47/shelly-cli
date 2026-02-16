---
title: "shelly webhook create"
description: "shelly webhook create"
---

## shelly webhook create

Create a webhook

### Synopsis

Create a new webhook for a device.

Webhooks are triggered by events and send HTTP requests to the specified URLs.
Common events include "switch.on", "switch.off", "input.toggle", etc.

```
shelly webhook create <device> [flags]
```

### Examples

```
  # Create a webhook for switch on event
  shelly webhook create living-room --event "switch.on" --url "http://example.com/hook"

  # Create with multiple URLs
  shelly webhook create living-room --event "switch.off" --url "http://a.com" --url "http://b.com"

  # Create disabled webhook
  shelly webhook create living-room --event "input.toggle" --url "http://example.com" --disable
```

### Options

```
      --cid int           Component ID (default: 0)
      --disable           Create webhook in disabled state
      --event string      Event type (e.g., switch.on, input.toggle)
  -h, --help              help for create
      --name string       Webhook name (optional)
      --url stringArray   Webhook URL (can be specified multiple times)
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

* [shelly webhook](shelly_webhook.md)	 - Manage device webhooks

