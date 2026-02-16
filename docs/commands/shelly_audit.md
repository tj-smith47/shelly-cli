## shelly audit

Security audit for devices

### Synopsis

Perform a security audit on Shelly devices.

Checks performed:
  - Authentication status (password protection)
  - Cloud connection exposure
  - Firmware version (security patches)

The audit flags potential security concerns such as:
  - Devices without authentication enabled
  - Devices connected to cloud with auth disabled
  - Outdated firmware that may have vulnerabilities

Use --all to audit all registered devices.

```
shelly audit [device...] [flags]
```

### Examples

```
  # Audit a single device
  shelly audit kitchen-light

  # Audit multiple devices
  shelly audit light-1 switch-2

  # Audit all registered devices
  shelly audit --all
```

### Options

```
      --all    Audit all registered devices
  -h, --help   help for audit
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

