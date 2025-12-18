## shelly mock

Mock device mode for testing

### Synopsis

Mock device mode for testing without real hardware.

Create and manage mock devices for testing CLI commands
and automation scripts without physical Shelly devices.

Subcommands:
  create    - Create a new mock device
  list      - List mock devices
  delete    - Delete a mock device
  scenario  - Load a test scenario

### Examples

```
  # Create a mock device
  shelly mock create kitchen-light --model "Plus 1PM"

  # List mock devices
  shelly mock list

  # Load test scenario
  shelly mock scenario home-setup
```

### Options

```
  -h, --help   help for mock
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly mock create](shelly_mock_create.md)	 - Create a mock device
* [shelly mock delete](shelly_mock_delete.md)	 - Delete a mock device
* [shelly mock list](shelly_mock_list.md)	 - List mock devices
* [shelly mock scenario](shelly_mock_scenario.md)	 - Load a test scenario

