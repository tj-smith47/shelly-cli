## shelly alert create

Create a monitoring alert

### Synopsis

Create a new monitoring alert for device conditions.

Conditions can be:
  - offline: Device becomes unreachable
  - online: Device becomes reachable
  - power>N: Power consumption exceeds N watts
  - temperature>N: Temperature exceeds N degrees

Actions can be:
  - notify: Desktop notification (default)
  - webhook:URL: Send HTTP POST to URL
  - command:CMD: Execute shell command

```
shelly alert create <name> [flags]
```

### Examples

```
  # Alert when device goes offline
  shelly alert create kitchen-offline --device kitchen --condition offline

  # Alert on high power consumption
  shelly alert create high-power --device heater --condition "power>2000"

  # Alert with webhook action
  shelly alert create temp-alert --device sensor --condition "temperature>30" \
    --action "webhook:http://example.com/alert"
```

### Options

```
  -a, --action string        Action when alert triggers (default "notify")
  -c, --condition string     Alert condition (required)
      --description string   Alert description
  -d, --device string        Device to monitor (required)
  -h, --help                 help for create
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

* [shelly alert](shelly_alert.md)	 - Manage monitoring alerts

