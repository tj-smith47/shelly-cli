## shelly lora

Manage LoRa add-on

### Synopsis

Manage LoRa add-on connectivity on Shelly devices.

LoRa (Long Range) is a wireless modulation technique that enables
long-distance communication with low power consumption. The Shelly
LoRa add-on extends device connectivity for scenarios where WiFi
is not available or practical.

LoRa features:
- Long range (up to 15km in ideal conditions)
- Low power consumption
- Point-to-point or star network topology
- Configurable frequency, bandwidth, and spreading factor

### Examples

```
  # Show LoRa add-on status
  shelly lora status living-room

  # Configure LoRa settings
  shelly lora config living-room --freq 868000000 --power 14

  # Send a message
  shelly lora send living-room "Hello World"
```

### Options

```
  -h, --help   help for lora
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
* [shelly lora config](shelly_lora_config.md)	 - Configure LoRa settings
* [shelly lora send](shelly_lora_send.md)	 - Send data over LoRa
* [shelly lora status](shelly_lora_status.md)	 - Show LoRa add-on status

