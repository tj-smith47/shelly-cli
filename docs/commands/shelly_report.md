## shelly report

Generate reports

### Synopsis

Generate a report about every registered device: an inventory, the current
power draw, or a security audit. All devices are queried at the same time, Gen1
and Gen2+ alike, and each device has one row, sorted by name.

Report types:
  devices  - name, ip, model, generation, firmware, mac and online for each
             device; summary: total, online, offline
  energy   - online, reporting (the device has a power meter) and power_w for
             each device; summary: total, online, offline, devices_reporting,
             total_power_w
  audit    - the checks of 'shelly audit': reachable, auth_enabled,
             cloud_connected, firmware_current, firmware_available,
             firmware_outdated, issues and warnings for each device; summary:
             devices_scanned, reachable, unreachable, auth_enabled,
             auth_disabled, cloud_connected, outdated_firmware, issues, warnings

A value the device did not report (it is offline, or the check failed) is
null in the audit rows. The document has timestamp, report_type, devices and
summary.

Output formats (--format, or the global -o):
  json   - JSON (default)
  yaml   - YAML
  text   - human-readable table (-o table is the same)

Progress messages go to stderr, so stdout carries only the report.

```
shelly report [flags]
```

### Examples

```
  # Device inventory as JSON
  shelly report --type devices -o json

  # Names of the devices that are offline
  shelly report --type devices -o json | jq -r '.devices[] | select(.online == false) | .name'

  # Current power draw of every device, as a table
  shelly report --type energy -o text

  # Total power in watts
  shelly report --type energy -o json | jq '.summary.total_power_w'

  # Security audit as YAML
  shelly report --type audit -o yaml

  # Devices with a firmware update available
  shelly report --type audit -o json | jq -r '.devices[] | select(.firmware_outdated == true) | .name'

  # Save a report to a file
  shelly report --type devices --output-file report.json
```

### Options

```
  -f, --format string        Output format: json, yaml, text, table (default "json")
  -h, --help                 help for report
      --output-file string   Output file path
  -t, --type string          Report type: devices, energy, audit (default "devices")
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

