## shelly sleep

Turn device off after a delay

### Synopsis

Turn a device off after a specified delay.

Useful for:
  - Setting a sleep timer for lights
  - Scheduling devices to turn off
  - "Goodnight" automation

Press Ctrl+C to cancel before the delay expires.

```
shelly sleep <device> [flags]
```

### Examples

```
  # Turn off in 5 minutes (default)
  shelly sleep bedroom-light

  # Turn off in 30 minutes
  shelly sleep living-room -d 30m

  # Turn off in 1 hour
  shelly sleep hallway --delay 1h
```

### Options

```
  -d, --delay duration   Delay before turning off (default 5m0s)
  -h, --help             help for sleep
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

