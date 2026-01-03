## shelly lora config

Configure LoRa settings

### Synopsis

Configure LoRa add-on settings on a Shelly device.

Allows setting radio parameters:
- Frequency: RF frequency in Hz (e.g., 868000000 for 868 MHz)
- Bandwidth: Channel bandwidth setting
- Data Rate: Spreading factor (higher = longer range, lower throughput)
- TX Power: Transmit power in dBm

Common frequencies:
- EU: 868 MHz (868000000)
- US: 915 MHz (915000000)
- Asia: 433 MHz (433000000)

```
shelly lora config <device> [flags]
```

### Examples

```
  # Set frequency to 868 MHz (EU)
  shelly lora config living-room --freq 868000000

  # Set transmit power to 14 dBm
  shelly lora config living-room --power 14

  # Configure multiple settings
  shelly lora config living-room --freq 915000000 --power 20 --dr 7
```

### Options

```
      --bw int      Bandwidth setting
      --dr int      Data rate / spreading factor
      --freq int    RF frequency in Hz
  -h, --help        help for config
  -i, --id int      LoRa component ID (default 0)
      --power int   Transmit power in dBm
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

* [shelly lora](shelly_lora.md)	 - Manage LoRa add-on

