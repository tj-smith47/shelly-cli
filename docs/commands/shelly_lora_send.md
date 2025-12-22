## shelly lora send

Send data over LoRa

### Synopsis

Send data over LoRa RF on a Shelly device.

Transmits data through the LoRa add-on. Data can be provided as:
- Plain text (default)
- Hexadecimal bytes with --hex flag

The data is base64-encoded before transmission as required by
the LoRa.SendBytes API.

```
shelly lora send <device> <data> [flags]
```

### Examples

```
  # Send a text message
  shelly lora send living-room "Hello World"

  # Send hex data
  shelly lora send living-room "48656c6c6f" --hex

  # Specify component ID
  shelly lora send living-room "test" --id 100
```

### Options

```
  -h, --help     help for send
      --hex      Data is hexadecimal
      --id int   LoRa component ID (default 100)
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

