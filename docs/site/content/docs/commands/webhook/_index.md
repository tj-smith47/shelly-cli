---
title: "shelly webhook"
description: "shelly webhook"
weight: 760
sidebar:
  collapsed: true
---

## shelly webhook

Manage device webhooks

### Synopsis

Manage webhooks for devices.

Webhooks allow the device to send HTTP requests when events occur,
enabling integration with external services and automation systems.

### Examples

```
  # List all webhooks
  shelly webhook list living-room

  # Create a webhook
  shelly webhook create living-room --event "switch.on" --url "http://example.com/hook"

  # Delete a webhook
  shelly webhook delete living-room 1

  # Update a webhook
  shelly webhook update living-room 1 --disable
```

### Options

```
  -h, --help   help for webhook
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
* [shelly webhook create](shelly_webhook_create.md)	 - Create a webhook
* [shelly webhook delete](shelly_webhook_delete.md)	 - Delete a webhook
* [shelly webhook list](shelly_webhook_list.md)	 - List webhooks
* [shelly webhook server](shelly_webhook_server.md)	 - Start a local webhook receiver server
* [shelly webhook update](shelly_webhook_update.md)	 - Update a webhook

