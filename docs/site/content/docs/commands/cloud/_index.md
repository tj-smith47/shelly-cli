---
title: "shelly cloud"
description: "shelly cloud"
weight: 130
sidebar:
  collapsed: true
---

## shelly cloud

Manage cloud connection and Shelly Cloud API

### Synopsis

Manage device cloud connection and interact with the Shelly Cloud API.

Device cloud commands:
  status    Show device cloud connection status
  enable    Enable cloud connection on a device
  disable   Disable cloud connection on a device

Cloud API commands (requires login):
  login       Authenticate with Shelly Cloud
  logout      Clear cloud credentials
  auth-status Show authentication status
  token       Show/manage access token
  devices     List cloud-registered devices
  device      Show cloud device details
  control     Control devices via cloud
  events      Subscribe to real-time cloud events

### Examples

```
  # Device cloud configuration
  shelly cloud status living-room
  shelly cloud enable living-room
  shelly cloud disable living-room

  # Cloud API authentication
  shelly cloud login
  shelly cloud auth-status
  shelly cloud logout

  # Cloud API device management
  shelly cloud devices
  shelly cloud device abc123
  shelly cloud control abc123 on
```

### Options

```
  -h, --help   help for cloud
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
* [shelly cloud auth-status](shelly_cloud_auth-status.md)	 - Show cloud authentication status
* [shelly cloud control](shelly_cloud_control.md)	 - Control a device via cloud
* [shelly cloud device](shelly_cloud_device.md)	 - Show cloud device details
* [shelly cloud devices](shelly_cloud_devices.md)	 - List cloud-registered devices
* [shelly cloud disable](shelly_cloud_disable.md)	 - Disable cloud connection
* [shelly cloud enable](shelly_cloud_enable.md)	 - Enable cloud connection
* [shelly cloud events](shelly_cloud_events.md)	 - Subscribe to real-time cloud events
* [shelly cloud login](shelly_cloud_login.md)	 - Authenticate with Shelly Cloud
* [shelly cloud logout](shelly_cloud_logout.md)	 - Clear cloud credentials
* [shelly cloud status](shelly_cloud_status.md)	 - Show cloud connection status
* [shelly cloud token](shelly_cloud_token.md)	 - Show or manage cloud token

