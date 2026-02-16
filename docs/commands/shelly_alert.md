## shelly alert

Manage monitoring alerts

### Synopsis

Manage monitoring alerts for device conditions.

Alerts notify you when devices meet specified conditions like going offline,
exceeding power thresholds, or temperature changes.

### Examples

```
  # Create an offline alert
  shelly alert create kitchen-offline --device kitchen --condition offline

  # List all alerts
  shelly alert list

  # Test an alert
  shelly alert test kitchen-offline

  # Snooze an alert for 1 hour
  shelly alert snooze kitchen-offline --duration 1h

  # Start monitoring alerts
  shelly alert watch
```

### Options

```
  -h, --help   help for alert
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
* [shelly alert create](shelly_alert_create.md)	 - Create a monitoring alert
* [shelly alert list](shelly_alert_list.md)	 - List configured alerts
* [shelly alert snooze](shelly_alert_snooze.md)	 - Snooze an alert temporarily
* [shelly alert test](shelly_alert_test.md)	 - Test an alert by triggering it
* [shelly alert watch](shelly_alert_watch.md)	 - Monitor alerts in real-time

