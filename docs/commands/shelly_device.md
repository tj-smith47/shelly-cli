## shelly device

Manage Shelly devices

### Synopsis

Manage Shelly devices in your registry.

Device commands allow you to add, remove, list, and manage registered devices.
Registered devices can be referenced by name in other commands.

### Examples

```
  # List all registered devices
  shelly device list

  # Get device info
  shelly dev info kitchen

  # Reboot a device
  shelly device reboot living-room
```

### Options

```
  -h, --help   help for device
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly device add](shelly_device_add.md)	 - Add a device to the registry
* [shelly device alias](shelly_device_alias.md)	 - Manage device aliases
* [shelly device config](shelly_device_config.md)	 - Manage device configuration
* [shelly device factory-reset](shelly_device_factory-reset.md)	 - Factory reset a device
* [shelly device info](shelly_device_info.md)	 - Show device information
* [shelly device list](shelly_device_list.md)	 - List registered devices
* [shelly device ping](shelly_device_ping.md)	 - Check device connectivity
* [shelly device reboot](shelly_device_reboot.md)	 - Reboot device
* [shelly device remove](shelly_device_remove.md)	 - Remove a device from the registry
* [shelly device rename](shelly_device_rename.md)	 - Rename a device in the registry
* [shelly device status](shelly_device_status.md)	 - Show device status
* [shelly device ui](shelly_device_ui.md)	 - Open device web interface in browser

