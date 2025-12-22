## shelly matter code

Show Matter pairing code

### Synopsis

Show the Matter pairing code for commissioning a device.

Displays the commissioning information needed to add the device
to a Matter fabric (Apple Home, Google Home, etc.):
- Manual pairing code (11-digit number)
- QR code data (for compatible apps)
- Discriminator and setup PIN

If the pairing code is not available via the API, check the device
label or web UI at http://<device-ip>/matter for the QR code.

```
shelly matter code <device> [flags]
```

### Examples

```
  # Show pairing code
  shelly matter code living-room

  # Output as JSON
  shelly matter code living-room --json
```

### Options

```
  -h, --help   help for code
      --json   Output as JSON
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

* [shelly matter](shelly_matter.md)	 - Manage Matter connectivity

