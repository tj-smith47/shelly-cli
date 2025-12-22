## shelly alert watch

Monitor alerts in real-time

### Synopsis

Monitor configured alerts and trigger actions when conditions are met.

This command runs continuously, polling device status at the specified interval
and executing alert actions when conditions are triggered.

Conditions supported:
  - offline: Device becomes unreachable
  - online: Device becomes reachable
  - power>N: Power consumption exceeds N watts
  - power<N: Power consumption below N watts
  - temperature>N: Temperature exceeds N degrees
  - temperature<N: Temperature below N degrees

Actions supported:
  - notify: Print to console (default)
  - webhook:URL: Send HTTP POST to URL with alert JSON
  - command:CMD: Execute shell command

```
shelly alert watch [flags]
```

### Examples

```
  # Monitor alerts every 30 seconds
  shelly alert watch

  # Monitor with custom interval
  shelly alert watch --interval 1m

  # Run once and exit (for cron)
  shelly alert watch --once
```

### Options

```
  -h, --help                help for watch
  -i, --interval duration   Check interval (default 30s)
      --once                Run once and exit (for cron/scheduled tasks)
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
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly alert](shelly_alert.md)	 - Manage monitoring alerts

