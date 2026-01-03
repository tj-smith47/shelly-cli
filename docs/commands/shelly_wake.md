## shelly wake

Turn device on after a delay

### Synopsis

Turn a device on after a specified delay.

Useful for:
  - Waking up to lights
  - Scheduling devices to turn on
  - "Good morning" automation

Press Ctrl+C to cancel before the delay expires.

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
```

### Options

```
  -d, --delay duration   Delay before turning on (default 5m0s)
  -h, --help             help for wake
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

