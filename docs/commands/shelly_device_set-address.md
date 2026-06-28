## shelly device set-address

Update a registered device's IP address

### Synopsis

Change the address of a device already in the registry.

Use this when a device moves to a new IP (DHCP lease change, subnet move, or a
manual re-IP). Only the address changes — the device's name, generation, model,
auth, and every group membership are preserved. This is the safe alternative to
'device remove' + 'device add', which would drop the device from its groups.

The new address is verified by default; pass --no-verify to pre-stage an address
for a device that is not yet reachable there.

```
shelly device set-address <name> <address> [flags]
```

### Examples

```
  # Re-point a device at its new IP
  shelly device set-address guest-bath 10.23.47.219

  # Pre-stage an address without a reachability check
  shelly device set-address guest-bath 10.23.47.219 --no-verify

  # Short form
  shelly dev set-addr bedroom 192.168.1.42
```

### Options

```
  -h, --help        help for set-address
      --no-verify   Skip the reachability check at the new address
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

* [shelly device](shelly_device.md)	 - Manage Shelly devices

