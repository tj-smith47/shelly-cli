---
title: "shelly wake"
description: "shelly wake"
weight: 750
sidebar:
  collapsed: true
---

## shelly wake

Turn device on after a delay

### Synopsis

Turn a device on after a specified delay.

With --simulate sunrise, the device's lights turn on at 1% brightness when the
delay ends and rise to 100% over --duration (default 15m). Sunrise needs a
device with a dimmable light component (a dimmer, bulb, or light in white
mode); a device with only switches or relays is rejected with an error.

Useful for:
  - Waking up to lights
  - Scheduling devices to turn on
  - "Good morning" automation

Press Ctrl+C to cancel before the delay expires, or to stop a sunrise at its
current brightness.

```
shelly wake <device> [flags]
```

### Examples

```
  # Turn on in 5 minutes (default)
  shelly wake bedroom-light

  # Turn on in 7 hours (alarm)
  shelly wake living-room -d 7h

  # Turn on in 30 seconds
  shelly wake kitchen --delay 30s

  # In 7 hours, fade the bedroom light up from 1% to 100% over 15 minutes
  shelly wake bedroom-light -d 7h --simulate sunrise --duration 15m
```

### Options

```
  -d, --delay duration      Delay before turning on (default 5m0s)
      --duration duration   How long a --simulate ramp takes to reach full brightness (default 15m0s)
  -h, --help                help for wake
      --simulate string     Fade the light in instead of switching it on (supported: sunrise)
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices

