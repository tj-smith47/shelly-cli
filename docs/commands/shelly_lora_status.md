## shelly lora status

Show LoRa add-on status

### Synopsis

Show LoRa add-on status for a Shelly device.

Displays the current LoRa configuration and signal quality
information from the last received packet.

```
shelly lora status <device> [flags]
```

### Examples

```
  # Show LoRa status
  shelly lora status living-room

  # Specify component ID (default: 100)
  shelly lora status living-room --id 100

  # Output as JSON
  shelly lora status living-room --json
```

### Options

```
  -h, --help     help for status
      --id int   LoRa component ID (default 100)
      --json     Output as JSON
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly lora](shelly_lora.md)	 - Manage LoRa add-on

