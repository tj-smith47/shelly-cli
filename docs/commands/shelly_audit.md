## shelly audit

Security audit for devices

### Synopsis

Perform a security audit on Shelly devices.

Checks performed:
  - auth:     Authentication status (password protection)
  - cloud:    Cloud connection exposure (cloud connected without a password)
  - firmware: Firmware version (security patches)

Every check runs by default. Pass one or more --check-<name> flags to run
only those checks.

With no device named, every registered device is audited (the same as --all).

Use -o json or -o yaml for one structured result per device, suitable for
scripting.

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
  shelly audit

  # Only check firmware and authentication, as JSON
  shelly audit --check-firmware --check-auth -o json

  # Only check cloud exposure on one device
  shelly audit kitchen-light --check-cloud
```

### Options

```
      --all              Audit all registered devices (the default when no device is named)
      --check-auth       Check authentication (password protection)
      --check-cloud      Check cloud connection exposure
      --check-firmware   Check for firmware updates
  -h, --help             help for audit
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

