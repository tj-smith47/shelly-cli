## shelly rgbw

Control RGBW LED outputs

### Synopsis

Control RGBW LED outputs on Shelly devices.

RGBW components support RGB color channels plus a separate white channel,
providing more control than RGB-only outputs.

### Examples

```
  # Turn on RGBW with warm white
  shelly rgbw on kitchen --rgb 255,200,150 --white 128

  # Set color and brightness
  shelly rgbw set kitchen --rgb 255,0,0 --brightness 75

  # Toggle RGBW state
  shelly rgbw toggle kitchen
```

### Options

```
  -h, --help   help for rgbw
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
* [shelly rgbw list](shelly_rgbw_list.md)	 - List rgbw components
* [shelly rgbw off](shelly_rgbw_off.md)	 - Turn rgbw off
* [shelly rgbw on](shelly_rgbw_on.md)	 - Turn rgbw on
* [shelly rgbw set](shelly_rgbw_set.md)	 - Set RGBW parameters
* [shelly rgbw status](shelly_rgbw_status.md)	 - Show rgbw status
* [shelly rgbw toggle](shelly_rgbw_toggle.md)	 - Toggle rgbw on/off

