## shelly firmware update

Update device firmware

### Synopsis

Update device firmware to the latest version.

By default, updates to the latest stable version. Use --beta for beta firmware
or --url for a custom firmware file.

Use --list to show available updates before prompting for confirmation.
This is useful for reviewing what version will be installed.

Supports both native Shelly devices and plugin-managed devices (Tasmota, etc.).
Plugin devices are automatically detected and updated using the appropriate plugin.

Use --all to update all registered devices. The --staged flag allows percentage-based
rollouts (e.g., --staged 25 updates 25% of devices).

```
shelly firmware update [device] [flags]
```

### Examples

```
  # Update to latest stable
  shelly firmware update living-room

  # Show update info before prompting
  shelly firmware update living-room --list

  # Update to beta
  shelly firmware update living-room --beta

  # Update from custom URL
  shelly firmware update living-room --url http://example.com/firmware.zip

  # Update plugin-managed device (Tasmota)
  shelly firmware update tasmota-plug --url http://ota.tasmota.com/tasmota.bin.gz

  # Update all devices
  shelly firmware update --all

  # Staged rollout (25% of devices)
  shelly firmware update --all --staged 25
```

### Options

```
      --all            Update all registered devices
      --beta           Update to beta firmware
  -h, --help           help for update
  -l, --list           Show available updates before prompting
      --parallel int   Number of devices to update in parallel (default 3)
      --staged int     Percentage of devices to update (for staged rollouts) (default 100)
      --url string     Custom firmware URL
  -y, --yes            Skip confirmation prompt
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
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

* [shelly firmware](shelly_firmware.md)	 - Manage device firmware

