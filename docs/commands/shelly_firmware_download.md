## shelly firmware download

Download firmware file

### Synopsis

Download firmware file for a device.

Downloads the latest available firmware for the device.
The firmware URL is determined by querying the device.

```
shelly firmware download <device> [flags]
```

### Examples

```
  # Download latest firmware for a device
  shelly firmware download living-room

  # Download to specific file
  shelly firmware download living-room --output firmware.zip

  # Download beta firmware
  shelly firmware download living-room --beta
```

### Options

```
      --beta            Download beta version
  -h, --help            help for download
      --latest          Download latest version (default true)
  -o, --output string   Output file path
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
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly firmware](shelly_firmware.md)	 - Manage device firmware

