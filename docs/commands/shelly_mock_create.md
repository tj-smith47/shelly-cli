## shelly mock create

Create a mock device

### Synopsis

Create a new mock device for testing.

```
shelly mock create <name> [flags]
```

### Examples

```
  # Create a mock Plus 1PM
  shelly mock create kitchen --model "Plus 1PM"

  # Create with specific firmware
  shelly mock create bedroom --model "Plus 2PM" --firmware "1.0.8"
```

### Options

```
      --firmware string   Firmware version (default "1.0.0")
  -h, --help              help for create
      --model string      Device model (default "Plus 1PM")
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly mock](shelly_mock.md)	 - Mock device mode for testing

