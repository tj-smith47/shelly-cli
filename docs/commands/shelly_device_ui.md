## shelly device ui

Open device web interface in browser

### Synopsis

Open a Shelly device's web interface in your default browser.

The device can be specified by name (from config) or by IP address/hostname.

```
shelly device ui <device> [flags]
```

### Examples

```
  # Open web interface by device name
  shelly device ui living-room

  # Open web interface by IP address
  shelly device ui 192.168.1.100

  # Using the 'web' alias
  shelly device web kitchen

  # Copy URL to clipboard instead of opening
  shelly device ui living-room --copy-url
```

### Options

```
      --copy-url   Copy URL to clipboard instead of opening browser
  -h, --help       help for ui
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

* [shelly device](shelly_device.md)	 - Manage Shelly devices

